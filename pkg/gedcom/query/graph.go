package query

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// Graph represents the graph structure of a GEDCOM tree.
type Graph struct {
	// Core data
	tree *gedcom.GedcomTree

	// ID mapping: XREF string <-> uint32 ID (for memory efficiency)
	xrefToID map[string]uint32 // XREF string -> uint32 ID
	idToXref map[uint32]string // uint32 ID -> XREF string
	nextID   uint32            // Next available ID

	// Node storage (using uint32 IDs internally for memory efficiency)
	nodes        map[uint32]GraphNode // All nodes by ID
	individuals  map[uint32]*IndividualNode
	families     map[uint32]*FamilyNode
	notes        map[uint32]*NoteNode
	sources      map[uint32]*SourceNode
	repositories map[uint32]*RepositoryNode
	events       map[uint32]*EventNode

	// Edge storage
	edges     map[string]*Edge   // All edges by ID (edge IDs remain strings)
	edgeIndex map[uint32][]*Edge // Edges by node ID (using uint32 for consistency)

	// Lazy loading support
	nodeMetadata map[uint32]*NodeMetadata // Node metadata (skeleton) - always loaded
	edgesLoaded  map[uint32]bool          // Track which nodes have edges loaded
	edgesLoading map[uint32]bool          // Track which nodes are currently loading edges (prevents recursion)
	lazyMode     bool                     // If true, use lazy loading; if false, load everything upfront

	// Graph partitioning support
	components     map[uint32][]uint32 // ComponentID -> []nodeIDs
	componentCount uint32              // Number of components found

	// Thread safety
	mu sync.RWMutex

	// Metadata
	properties map[string]interface{}

	// Performance optimizations
	cache   *queryCache
	indexes *FilterIndexes

	// Hybrid storage support
	hybridStorage *HybridStorage
	hybridMode    bool
	queryHelpers  *HybridQueryHelpers // SQLite query helpers
	hybridCache   *HybridCache        // LRU cache for hybrid storage
}

// NewGraph creates a new empty graph.
func NewGraph(tree *gedcom.GedcomTree) *Graph {
	return &Graph{
		tree:           tree,
		xrefToID:       make(map[string]uint32),
		idToXref:       make(map[uint32]string),
		nextID:         1, // Start IDs at 1 (0 reserved for invalid)
		nodes:          make(map[uint32]GraphNode),
		individuals:    make(map[uint32]*IndividualNode),
		families:       make(map[uint32]*FamilyNode),
		notes:          make(map[uint32]*NoteNode),
		sources:        make(map[uint32]*SourceNode),
		repositories:   make(map[uint32]*RepositoryNode),
		events:         make(map[uint32]*EventNode),
		edges:          make(map[string]*Edge),
		edgeIndex:      make(map[uint32][]*Edge),
		nodeMetadata:   make(map[uint32]*NodeMetadata),
		edgesLoaded:    make(map[uint32]bool),
		edgesLoading:   make(map[uint32]bool),
		lazyMode:       false, // Default: eager loading (backward compatible)
		components:     make(map[uint32][]uint32),
		componentCount: 0,
		properties:     make(map[string]interface{}),
		cache:          newQueryCache(1000), // Default cache size
		indexes:        newFilterIndexes(),
	}
}

// GetNode returns a node by XREF ID (external API - still uses strings).
func (g *Graph) GetNode(xrefID string) GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.nodes[id]
}

// GetIndividual returns an IndividualNode by xref ID (external API - still uses strings).
// If lazy mode is enabled, loads the node on-demand.
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetIndividual(xrefID string) *IndividualNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getIndividualFromHybrid(xrefID)
	}

	g.mu.RLock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		g.mu.RUnlock()
		return nil
	}
	node := g.individuals[id]
	lazyMode := g.lazyMode
	g.mu.RUnlock()

	// If lazy mode and node not loaded, load it
	if lazyMode && node == nil {
		loadedNode, err := g.ensureNodeLoaded(xrefID)
		if err != nil || loadedNode == nil {
			return nil
		}
		if indiNode, ok := loadedNode.(*IndividualNode); ok {
			// Ensure edges are loaded when node is accessed (without holding lock)
			g.ensureEdgesLoadedUnlocked(indiNode)
			return indiNode
		}
		return nil
	}

	// Ensure edges are loaded if lazy mode (without holding lock)
	if lazyMode && node != nil {
		g.ensureEdgesLoadedUnlocked(node)
	}

	return node
}

// getIndividualFromHybrid loads an individual from hybrid storage
func (g *Graph) getIndividualFromHybrid(xrefID string) *IndividualNode {
	// Check cache first
	if g.hybridCache != nil {
		// Check XREF -> ID cache
		if nodeID, found := g.hybridCache.GetXrefToID(xrefID); found {
			// Check node cache
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if indiNode, ok := node.(*IndividualNode); ok {
					return indiNode
				}
			}
		}
	}

	// Get node ID from SQLite (with caching)
	var nodeID uint32
	var err error
	if g.hybridCache != nil {
		if cachedID, found := g.hybridCache.GetXrefToID(xrefID); found {
			nodeID = cachedID
		} else {
			nodeID, err = g.queryHelpers.FindByXref(xrefID)
			if err != nil || nodeID == 0 {
				return nil
			}
			// Cache the mapping
			g.hybridCache.SetXrefToID(xrefID, nodeID)
		}
	} else {
		nodeID, err = g.queryHelpers.FindByXref(xrefID)
		if err != nil || nodeID == 0 {
			return nil
		}
	}

	// Check if already in memory
	g.mu.RLock()
	if node, exists := g.individuals[nodeID]; exists {
		g.mu.RUnlock()
		// Update cache
		if g.hybridCache != nil {
			g.hybridCache.SetNode(nodeID, node)
		}
		return node
	}
	g.mu.RUnlock()

	// Load from BadgerDB
	badgerDB := g.hybridStorage.BadgerDB()
	key := fmt.Sprintf("node:%d", nodeID)

	// Deserialize node (reconstructs record from tree)
	var nodeDataBytes []byte
	err = badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			nodeDataBytes = make([]byte, len(val))
			copy(nodeDataBytes, val)
			return nil
		})
	})
	if err != nil {
		return nil
	}

	node, err := DeserializeNode(nodeDataBytes, g)
	if err != nil || node == nil {
		return nil
	}

	indiNode, ok := node.(*IndividualNode)
	if !ok {
		return nil
	}

	// Add to in-memory cache
	g.mu.Lock()
	g.xrefToID[xrefID] = nodeID
	g.idToXref[nodeID] = xrefID
	g.nodes[nodeID] = indiNode
	g.individuals[nodeID] = indiNode
	g.mu.Unlock()

	// Update hybrid cache
	if g.hybridCache != nil {
		g.hybridCache.SetNode(nodeID, indiNode)
		g.hybridCache.SetXrefToID(xrefID, nodeID)
	}

	// Load edges
	g.loadEdgesFromHybrid(nodeID, indiNode)

	return indiNode
}

// GetFamily returns a FamilyNode by xref ID (external API - still uses strings).
// If lazy mode is enabled, loads the node on-demand.
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetFamily(xrefID string) *FamilyNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getFamilyFromHybrid(xrefID)
	}

	g.mu.RLock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		g.mu.RUnlock()
		return nil
	}
	node := g.families[id]
	lazyMode := g.lazyMode
	g.mu.RUnlock()

	// If lazy mode and node not loaded, load it
	if lazyMode && node == nil {
		loadedNode, err := g.ensureNodeLoaded(xrefID)
		if err != nil || loadedNode == nil {
			return nil
		}
		if famNode, ok := loadedNode.(*FamilyNode); ok {
			// Ensure edges are loaded when node is accessed (without holding lock)
			g.ensureEdgesLoadedUnlocked(famNode)
			return famNode
		}
		return nil
	}

	// Ensure edges are loaded if lazy mode (without holding lock)
	if lazyMode && node != nil {
		g.ensureEdgesLoadedUnlocked(node)
	}

	return node
}

// getFamilyFromHybrid loads a family from hybrid storage
func (g *Graph) getFamilyFromHybrid(xrefID string) *FamilyNode {
	// Check cache first
	if g.hybridCache != nil {
		if nodeID, found := g.hybridCache.GetXrefToID(xrefID); found {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if famNode, ok := node.(*FamilyNode); ok {
					return famNode
				}
			}
		}
	}

	// Get node ID from SQLite
	var nodeID uint32
	var err error
	if g.hybridCache != nil {
		if cachedID, found := g.hybridCache.GetXrefToID(xrefID); found {
			nodeID = cachedID
		} else {
			nodeID, err = g.queryHelpers.FindByXref(xrefID)
			if err != nil || nodeID == 0 {
				return nil
			}
			g.hybridCache.SetXrefToID(xrefID, nodeID)
		}
	} else {
		nodeID, err = g.queryHelpers.FindByXref(xrefID)
		if err != nil || nodeID == 0 {
			return nil
		}
	}

	// Check if already in memory
	g.mu.RLock()
	if node, exists := g.families[nodeID]; exists {
		g.mu.RUnlock()
		if g.hybridCache != nil {
			g.hybridCache.SetNode(nodeID, node)
		}
		return node
	}
	g.mu.RUnlock()

	// Load from BadgerDB
	badgerDB := g.hybridStorage.BadgerDB()
	key := fmt.Sprintf("node:%d", nodeID)

	var nodeDataBytes []byte
	err = badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			nodeDataBytes = make([]byte, len(val))
			copy(nodeDataBytes, val)
			return nil
		})
	})
	if err != nil {
		return nil
	}

	node, err := DeserializeNode(nodeDataBytes, g)
	if err != nil || node == nil {
		return nil
	}

	famNode, ok := node.(*FamilyNode)
	if !ok {
		return nil
	}

	// Add to in-memory cache
	g.mu.Lock()
	g.xrefToID[xrefID] = nodeID
	g.idToXref[nodeID] = xrefID
	g.nodes[nodeID] = famNode
	g.families[nodeID] = famNode
	g.mu.Unlock()

	// Update hybrid cache
	if g.hybridCache != nil {
		g.hybridCache.SetNode(nodeID, famNode)
		g.hybridCache.SetXrefToID(xrefID, nodeID)
	}

	// Load edges
	g.loadEdgesFromHybrid(nodeID, famNode)

	return famNode
}

// GetNote returns a NoteNode by xref ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetNote(xrefID string) *NoteNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getNoteFromHybrid(xrefID)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.notes[id]
}

// getNoteFromHybrid loads a note from hybrid storage
func (g *Graph) getNoteFromHybrid(xrefID string) *NoteNode {
	// Check cache first
	if g.hybridCache != nil {
		if nodeID, found := g.hybridCache.GetXrefToID(xrefID); found {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if noteNode, ok := node.(*NoteNode); ok {
					return noteNode
				}
			}
		}
	}

	// Get node ID from SQLite
	var nodeID uint32
	var err error
	if g.hybridCache != nil {
		if cachedID, found := g.hybridCache.GetXrefToID(xrefID); found {
			nodeID = cachedID
		} else {
			nodeID, err = g.queryHelpers.FindByXref(xrefID)
			if err != nil || nodeID == 0 {
				return nil
			}
			g.hybridCache.SetXrefToID(xrefID, nodeID)
		}
	} else {
		nodeID, err = g.queryHelpers.FindByXref(xrefID)
		if err != nil || nodeID == 0 {
			return nil
		}
	}

	// Check if already in memory
	g.mu.RLock()
	if node, exists := g.notes[nodeID]; exists {
		g.mu.RUnlock()
		if g.hybridCache != nil {
			g.hybridCache.SetNode(nodeID, node)
		}
		return node
	}
	g.mu.RUnlock()

	// Load from BadgerDB
	badgerDB := g.hybridStorage.BadgerDB()
	key := fmt.Sprintf("node:%d", nodeID)

	var nodeDataBytes []byte
	err = badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			nodeDataBytes = make([]byte, len(val))
			copy(nodeDataBytes, val)
			return nil
		})
	})
	if err != nil {
		return nil
	}

	node, err := DeserializeNode(nodeDataBytes, g)
	if err != nil || node == nil {
		return nil
	}

	noteNode, ok := node.(*NoteNode)
	if !ok {
		return nil
	}

	// Add to in-memory cache
	g.mu.Lock()
	g.xrefToID[xrefID] = nodeID
	g.idToXref[nodeID] = xrefID
	g.nodes[nodeID] = noteNode
	g.notes[nodeID] = noteNode
	g.mu.Unlock()

	// Update hybrid cache
	if g.hybridCache != nil {
		g.hybridCache.SetNode(nodeID, noteNode)
		g.hybridCache.SetXrefToID(xrefID, nodeID)
	}

	// Load edges
	g.loadEdgesFromHybrid(nodeID, noteNode)

	return noteNode
}

// GetSource returns a SourceNode by xref ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetSource(xrefID string) *SourceNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getSourceFromHybrid(xrefID)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.sources[id]
}

// getSourceFromHybrid loads a source from hybrid storage
func (g *Graph) getSourceFromHybrid(xrefID string) *SourceNode {
	// Check cache first
	if g.hybridCache != nil {
		if nodeID, found := g.hybridCache.GetXrefToID(xrefID); found {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if sourceNode, ok := node.(*SourceNode); ok {
					return sourceNode
				}
			}
		}
	}

	// Get node ID from SQLite
	var nodeID uint32
	var err error
	if g.hybridCache != nil {
		if cachedID, found := g.hybridCache.GetXrefToID(xrefID); found {
			nodeID = cachedID
		} else {
			nodeID, err = g.queryHelpers.FindByXref(xrefID)
			if err != nil || nodeID == 0 {
				return nil
			}
			g.hybridCache.SetXrefToID(xrefID, nodeID)
		}
	} else {
		nodeID, err = g.queryHelpers.FindByXref(xrefID)
		if err != nil || nodeID == 0 {
			return nil
		}
	}

	// Check if already in memory
	g.mu.RLock()
	if node, exists := g.sources[nodeID]; exists {
		g.mu.RUnlock()
		if g.hybridCache != nil {
			g.hybridCache.SetNode(nodeID, node)
		}
		return node
	}
	g.mu.RUnlock()

	// Load from BadgerDB
	badgerDB := g.hybridStorage.BadgerDB()
	key := fmt.Sprintf("node:%d", nodeID)

	var nodeDataBytes []byte
	err = badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			nodeDataBytes = make([]byte, len(val))
			copy(nodeDataBytes, val)
			return nil
		})
	})
	if err != nil {
		return nil
	}

	node, err := DeserializeNode(nodeDataBytes, g)
	if err != nil || node == nil {
		return nil
	}

	sourceNode, ok := node.(*SourceNode)
	if !ok {
		return nil
	}

	// Add to in-memory cache
	g.mu.Lock()
	g.xrefToID[xrefID] = nodeID
	g.idToXref[nodeID] = xrefID
	g.nodes[nodeID] = sourceNode
	g.sources[nodeID] = sourceNode
	g.mu.Unlock()

	// Update hybrid cache
	if g.hybridCache != nil {
		g.hybridCache.SetNode(nodeID, sourceNode)
		g.hybridCache.SetXrefToID(xrefID, nodeID)
	}

	// Load edges
	g.loadEdgesFromHybrid(nodeID, sourceNode)

	return sourceNode
}

// GetRepository returns a RepositoryNode by xref ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetRepository(xrefID string) *RepositoryNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getRepositoryFromHybrid(xrefID)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.repositories[id]
}

// getRepositoryFromHybrid loads a repository from hybrid storage
func (g *Graph) getRepositoryFromHybrid(xrefID string) *RepositoryNode {
	// Check cache first
	if g.hybridCache != nil {
		if nodeID, found := g.hybridCache.GetXrefToID(xrefID); found {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if repoNode, ok := node.(*RepositoryNode); ok {
					return repoNode
				}
			}
		}
	}

	// Get node ID from SQLite
	var nodeID uint32
	var err error
	if g.hybridCache != nil {
		if cachedID, found := g.hybridCache.GetXrefToID(xrefID); found {
			nodeID = cachedID
		} else {
			nodeID, err = g.queryHelpers.FindByXref(xrefID)
			if err != nil || nodeID == 0 {
				return nil
			}
			g.hybridCache.SetXrefToID(xrefID, nodeID)
		}
	} else {
		nodeID, err = g.queryHelpers.FindByXref(xrefID)
		if err != nil || nodeID == 0 {
			return nil
		}
	}

	// Check if already in memory
	g.mu.RLock()
	if node, exists := g.repositories[nodeID]; exists {
		g.mu.RUnlock()
		if g.hybridCache != nil {
			g.hybridCache.SetNode(nodeID, node)
		}
		return node
	}
	g.mu.RUnlock()

	// Load from BadgerDB
	badgerDB := g.hybridStorage.BadgerDB()
	key := fmt.Sprintf("node:%d", nodeID)

	var nodeDataBytes []byte
	err = badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			nodeDataBytes = make([]byte, len(val))
			copy(nodeDataBytes, val)
			return nil
		})
	})
	if err != nil {
		return nil
	}

	node, err := DeserializeNode(nodeDataBytes, g)
	if err != nil || node == nil {
		return nil
	}

	repoNode, ok := node.(*RepositoryNode)
	if !ok {
		return nil
	}

	// Add to in-memory cache
	g.mu.Lock()
	g.xrefToID[xrefID] = nodeID
	g.idToXref[nodeID] = xrefID
	g.nodes[nodeID] = repoNode
	g.repositories[nodeID] = repoNode
	g.mu.Unlock()

	// Update hybrid cache
	if g.hybridCache != nil {
		g.hybridCache.SetNode(nodeID, repoNode)
		g.hybridCache.SetXrefToID(xrefID, nodeID)
	}

	// Load edges
	g.loadEdgesFromHybrid(nodeID, repoNode)

	return repoNode
}

// GetEvent returns an EventNode by event ID (external API - still uses strings).
// If hybrid mode is enabled, loads the node from BadgerDB.
func (g *Graph) GetEvent(eventID string) *EventNode {
	// If hybrid mode, load from BadgerDB
	if g.hybridMode && g.hybridStorage != nil {
		return g.getEventFromHybrid(eventID)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[eventID]
	if id == 0 {
		return nil
	}
	return g.events[id]
}

// getEventFromHybrid loads an event from hybrid storage
func (g *Graph) getEventFromHybrid(eventID string) *EventNode {
	// Check cache first
	if g.hybridCache != nil {
		if nodeID, found := g.hybridCache.GetXrefToID(eventID); found {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if eventNode, ok := node.(*EventNode); ok {
					return eventNode
				}
			}
		}
	}

	// Get node ID from SQLite
	var nodeID uint32
	var err error
	if g.hybridCache != nil {
		if cachedID, found := g.hybridCache.GetXrefToID(eventID); found {
			nodeID = cachedID
		} else {
			nodeID, err = g.queryHelpers.FindByXref(eventID)
			if err != nil || nodeID == 0 {
				return nil
			}
			g.hybridCache.SetXrefToID(eventID, nodeID)
		}
	} else {
		nodeID, err = g.queryHelpers.FindByXref(eventID)
		if err != nil || nodeID == 0 {
			return nil
		}
	}

	// Check if already in memory
	g.mu.RLock()
	if node, exists := g.events[nodeID]; exists {
		g.mu.RUnlock()
		if g.hybridCache != nil {
			g.hybridCache.SetNode(nodeID, node)
		}
		return node
	}
	g.mu.RUnlock()

	// Load from BadgerDB
	badgerDB := g.hybridStorage.BadgerDB()
	key := fmt.Sprintf("node:%d", nodeID)

	var nodeDataBytes []byte
	err = badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			nodeDataBytes = make([]byte, len(val))
			copy(nodeDataBytes, val)
			return nil
		})
	})
	if err != nil {
		return nil
	}

	node, err := DeserializeNode(nodeDataBytes, g)
	if err != nil || node == nil {
		return nil
	}

	eventNode, ok := node.(*EventNode)
	if !ok {
		return nil
	}

	// Add to in-memory cache
	g.mu.Lock()
	g.xrefToID[eventID] = nodeID
	g.idToXref[nodeID] = eventID
	g.nodes[nodeID] = eventNode
	g.events[nodeID] = eventNode
	g.mu.Unlock()

	// Update hybrid cache
	if g.hybridCache != nil {
		g.hybridCache.SetNode(nodeID, eventNode)
		g.hybridCache.SetXrefToID(eventID, nodeID)
	}

	// Load edges
	g.loadEdgesFromHybrid(nodeID, eventNode)

	return eventNode
}

// GetAllNodes returns all nodes in the graph (external API - returns XREF strings).
func (g *Graph) GetAllNodes() map[string]GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]GraphNode)
	for id, node := range g.nodes {
		xrefID := g.idToXref[id]
		if xrefID != "" {
			result[xrefID] = node
		}
	}
	return result
}

// GetAllIndividuals returns all IndividualNodes (external API - returns XREF strings).
// If hybrid mode is enabled, queries SQLite for all individual IDs.
// NOTE: In hybrid mode, this returns an empty map because nodes are loaded on-demand.
// Use FilterQuery.executeHybrid() for efficient querying in hybrid mode.
func (g *Graph) GetAllIndividuals() map[string]*IndividualNode {
	// If hybrid mode, return empty map - nodes loaded on-demand via GetIndividual
	// Note: We check queryHelpers instead of hybridStorage because queryHelpers
	// is required for SQLite queries, and this method is primarily used for
	// in-memory iteration. FilterQuery.executeHybrid() handles hybrid mode queries.
	if g.hybridMode && g.queryHelpers != nil {
		return make(map[string]*IndividualNode)
	}
	return g.getAllIndividualsInMemory()
}

// getAllIndividualsInMemory returns in-memory individuals
func (g *Graph) getAllIndividualsInMemory() map[string]*IndividualNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*IndividualNode)
	for id, node := range g.individuals {
		xrefID := g.idToXref[id]
		if xrefID != "" {
			result[xrefID] = node
		}
	}
	return result
}

// GetAllFamilies returns all FamilyNodes (external API - returns XREF strings).
func (g *Graph) GetAllFamilies() map[string]*FamilyNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*FamilyNode)
	for id, node := range g.families {
		xrefID := g.idToXref[id]
		if xrefID != "" {
			result[xrefID] = node
		}
	}
	return result
}

// AddNode adds a node to the graph (external API - still uses XREF strings).
func (g *Graph) AddNode(node GraphNode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	xrefID := node.ID()
	if xrefID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	// Get or create uint32 ID for this XREF
	internalID := g.getOrCreateID(xrefID)

	// Check if node already exists
	if _, exists := g.nodes[internalID]; exists {
		return fmt.Errorf("node with ID %s already exists", xrefID)
	}

	// Add to nodes map (using uint32 ID)
	g.nodes[internalID] = node

	// Add to type-specific map (using uint32 ID)
	switch node.NodeType() {
	case NodeTypeIndividual:
		if indiNode, ok := node.(*IndividualNode); ok {
			g.individuals[internalID] = indiNode
		}
	case NodeTypeFamily:
		if famNode, ok := node.(*FamilyNode); ok {
			g.families[internalID] = famNode
		}
	case NodeTypeNote:
		if noteNode, ok := node.(*NoteNode); ok {
			g.notes[internalID] = noteNode
		}
	case NodeTypeSource:
		if sourceNode, ok := node.(*SourceNode); ok {
			g.sources[internalID] = sourceNode
		}
	case NodeTypeRepository:
		if repoNode, ok := node.(*RepositoryNode); ok {
			g.repositories[internalID] = repoNode
		}
	case NodeTypeEvent:
		if eventNode, ok := node.(*EventNode); ok {
			g.events[internalID] = eventNode
		}
	}

	return nil
}

// AddEdge adds an edge to the graph.
func (g *Graph) AddEdge(edge *Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if edge == nil {
		return fmt.Errorf("edge cannot be nil")
	}

	if edge.ID == "" {
		return fmt.Errorf("edge ID cannot be empty")
	}

	if edge.From == nil || edge.To == nil {
		return fmt.Errorf("edge must have both From and To nodes")
	}

	// Check if edge already exists
	if _, exists := g.edges[edge.ID]; exists {
		return fmt.Errorf("edge with ID %s already exists", edge.ID)
	}

	// Add to edges map
	g.edges[edge.ID] = edge

	// Add to edge index (using uint32 IDs)
	fromXref := edge.From.ID()
	toXref := edge.To.ID()
	fromID := g.getOrCreateID(fromXref)
	toID := g.getOrCreateID(toXref)

	g.edgeIndex[fromID] = append(g.edgeIndex[fromID], edge)
	g.edgeIndex[toID] = append(g.edgeIndex[toID], edge)

	// Add to node's edge lists
	edge.From.AddOutEdge(edge)
	edge.To.AddInEdge(edge)

	// If bidirectional, also add reverse
	if edge.IsBidirectional() {
		edge.To.AddOutEdge(edge)
		edge.From.AddInEdge(edge)
	}

	return nil
}

// GetEdges returns all edges for a given node XREF ID (external API - still uses strings).
func (g *Graph) GetEdges(xrefID string) []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	id := g.xrefToID[xrefID]
	if id == 0 {
		return nil
	}
	return g.edgeIndex[id]
}

// GetEdge returns an edge by ID.
func (g *Graph) GetEdge(edgeID string) *Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edges[edgeID]
}

// GetAllEdges returns all edges in the graph.
func (g *Graph) GetAllEdges() map[string]*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]*Edge)
	for k, v := range g.edges {
		result[k] = v
	}
	return result
}

// NodeCount returns the total number of nodes.
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the total number of edges.
func (g *Graph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.edges)
}

// Tree returns the underlying GEDCOM tree.
func (g *Graph) Tree() *gedcom.GedcomTree {
	return g.tree
}

// loadEdgesFromHybrid loads edges for a node from BadgerDB
func (g *Graph) loadEdgesFromHybrid(nodeID uint32, node GraphNode) {
	badgerDB := g.hybridStorage.BadgerDB()
	key := fmt.Sprintf("edges:%d:out", nodeID)

	err := badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err // No edges found
		}
		return item.Value(func(val []byte) error {
			var edges []EdgeData
			if err := deserialize(val, &edges); err != nil {
				return err
			}

			// Convert EdgeData to Edge objects and add to node
			for _, edgeData := range edges {
				// Get target node
				toXref, err := g.queryHelpers.FindXrefByID(edgeData.ToID)
				if err != nil || toXref == "" {
					continue
				}

				// Get target node object based on edge type
				var toNode GraphNode
				switch edgeData.EdgeType {
				case EdgeTypeHUSB, EdgeTypeWIFE, EdgeTypeCHIL, EdgeTypeFAMC, EdgeTypeFAMS:
					// Family-related edges - target is individual
					toNode = g.GetIndividual(toXref)
				case EdgeTypeNOTE:
					// NOTE edge - target is note
					toNode = g.GetNote(toXref)
				case EdgeTypeSOUR:
					// SOUR edge - target is source
					toNode = g.GetSource(toXref)
				case EdgeTypeREPO:
					// REPO edge - target is repository
					toNode = g.GetRepository(toXref)
				case EdgeTypeHasEvent:
					// has_event edge - target is event
					toNode = g.GetEvent(toXref)
				default:
					// Other edge types - try GetNode
					toNode = g.GetNode(toXref)
				}

				if toNode == nil {
					continue
				}

				// Create edge
				var edge *Edge
				if edgeData.FamilyID != 0 {
					famXref, _ := g.queryHelpers.FindXrefByID(edgeData.FamilyID)
					if famXref != "" {
						famNode := g.GetFamily(famXref)
						if famNode != nil {
							edgeID := fmt.Sprintf("%s_%s_%s", node.ID(), edgeData.EdgeType, toXref)
							edge = NewEdgeWithFamily(edgeID, node, toNode, edgeData.EdgeType, famNode)
						}
					}
				}
				if edge == nil {
					edgeID := fmt.Sprintf("%s_%s_%s", node.ID(), edgeData.EdgeType, toXref)
					edge = NewEdge(edgeID, node, toNode, edgeData.EdgeType)
				}

				// Add edge to node using the interface methods
				// Note: We only add the forward edge here. Reverse edges will be added
				// when the target node's edges are loaded, to avoid circular dependencies
				// and ensure both nodes are fully initialized.
				node.AddOutEdge(edge)
			}
			return nil
		})
	})
	if err != nil {
		// No edges or error - that's okay
		return
	}
}

// reverseEdgeType returns the reverse edge type
func reverseEdgeType(edgeType EdgeType) EdgeType {
	switch edgeType {
	case EdgeTypeHUSB:
		return EdgeTypeFAMS
	case EdgeTypeWIFE:
		return EdgeTypeFAMS
	case EdgeTypeCHIL:
		return EdgeTypeFAMC
	case EdgeTypeFAMC:
		return EdgeTypeCHIL
	case EdgeTypeFAMS:
		// FAMS can reverse to HUSB or WIFE - we'll use FAMS for now
		return EdgeTypeFAMS
	default:
		return edgeType
	}
}
