package query

import (
	"fmt"
	"sync"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// NodeMetadata stores minimal information about a node for lazy loading.
type NodeMetadata struct {
	XrefID     string
	NodeType   NodeType
	ComponentID uint32 // For graph partitioning (0 = unassigned)
	Loaded     bool    // Whether the node is currently loaded in memory
}

// LazyGraph extends Graph with lazy loading capabilities.
// This is embedded in Graph to add lazy loading without breaking existing code.
type LazyGraph struct {
	// Node metadata (skeleton) - always loaded
	nodeMetadata map[uint32]*NodeMetadata // uint32 ID -> metadata

	// Component information (for partitioning)
	components map[uint32][]uint32 // ComponentID -> []nodeIDs
	componentCount uint32

	// Lazy loading state
	lazyMode bool // If true, use lazy loading; if false, load everything upfront

	// Edge loading state (track which nodes have edges loaded)
	edgesLoaded map[uint32]bool // nodeID -> true if edges loaded

	mu sync.RWMutex
}

// ensureNodeLoaded ensures a node is loaded in memory, loading it if necessary.
func (g *Graph) ensureNodeLoaded(xrefID string) (GraphNode, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Get internal ID
	internalID := g.xrefToID[xrefID]
	if internalID == 0 {
		return nil, fmt.Errorf("node %s not found", xrefID)
	}

	// Check if already loaded
	if node := g.nodes[internalID]; node != nil {
		return node, nil
	}

	// Load from GEDCOM tree
	return g.loadNodeFromTree(xrefID, internalID)
}

// loadNodeFromTree loads a node from the GEDCOM tree and adds it to the graph.
// Must be called with lock held.
func (g *Graph) loadNodeFromTree(xrefID string, internalID uint32) (GraphNode, error) {
	// Get metadata to determine node type
	meta := g.nodeMetadata[internalID]
	if meta == nil {
		return nil, fmt.Errorf("node metadata not found for %s", xrefID)
	}

	var node GraphNode

	// Load node based on type
	switch meta.NodeType {
	case NodeTypeIndividual:
		record := g.tree.GetIndividual(xrefID)
		if record == nil {
			return nil, fmt.Errorf("individual %s not found in GEDCOM tree", xrefID)
		}
		if indi, ok := record.(*gedcom.IndividualRecord); ok {
			node = NewIndividualNode(xrefID, indi)
		} else {
			return nil, fmt.Errorf("invalid individual record type for %s", xrefID)
		}

	case NodeTypeFamily:
		record := g.tree.GetFamily(xrefID)
		if record == nil {
			return nil, fmt.Errorf("family %s not found in GEDCOM tree", xrefID)
		}
		if fam, ok := record.(*gedcom.FamilyRecord); ok {
			node = NewFamilyNode(xrefID, fam)
		} else {
			return nil, fmt.Errorf("invalid family record type for %s", xrefID)
		}

	case NodeTypeNote:
		allNotes := g.tree.GetAllNotes()
		if record, exists := allNotes[xrefID]; exists {
			if note, ok := record.(*gedcom.NoteRecord); ok {
				node = NewNoteNode(xrefID, note)
			}
		}

	case NodeTypeSource:
		allSources := g.tree.GetAllSources()
		if record, exists := allSources[xrefID]; exists {
			if source, ok := record.(*gedcom.SourceRecord); ok {
				node = NewSourceNode(xrefID, source)
			}
		}

	case NodeTypeRepository:
		allRepos := g.tree.GetAllRepositories()
		if record, exists := allRepos[xrefID]; exists {
			if repo, ok := record.(*gedcom.RepositoryRecord); ok {
				node = NewRepositoryNode(xrefID, repo)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported node type %s for %s", meta.NodeType, xrefID)
	}

	if node == nil {
		return nil, fmt.Errorf("failed to load node %s from GEDCOM tree", xrefID)
	}

	// Add to graph (using existing AddNode logic but skip ID creation since it exists)
	if err := g.addNodeInternal(node); err != nil {
		return nil, err
	}

	// Mark as loaded
	meta.Loaded = true

	return node, nil
}

// ensureEdgesLoadedUnlocked ensures edges for a node are loaded, loading them if necessary.
// Must NOT be called while holding g.mu lock (to avoid deadlock).
func (g *Graph) ensureEdgesLoadedUnlocked(node GraphNode) error {
	xrefID := node.ID()
	
	// Quick check without lock
	g.mu.RLock()
	internalID := g.xrefToID[xrefID]
	if internalID == 0 {
		g.mu.RUnlock()
		return fmt.Errorf("node %s not found", xrefID)
	}

	// Check if edges already loaded
	if g.edgesLoaded != nil && g.edgesLoaded[internalID] {
		g.mu.RUnlock()
		return nil
	}
	
	// Check if edges are currently being loaded (prevents recursion)
	if g.edgesLoading != nil && g.edgesLoading[internalID] {
		g.mu.RUnlock()
		return nil // Return early to prevent infinite recursion
	}
	g.mu.RUnlock()

	// Need to load edges - acquire write lock
	g.mu.Lock()
	// Double-check after acquiring lock
	if g.edgesLoaded != nil && g.edgesLoaded[internalID] {
		g.mu.Unlock()
		return nil
	}
	
	// Mark as loading to prevent recursion
	if g.edgesLoading == nil {
		g.edgesLoading = make(map[uint32]bool)
	}
	g.edgesLoading[internalID] = true
	g.mu.Unlock()
	
	// Load edges (without holding lock to allow GetIndividual/GetFamily calls)
	defer func() {
		g.mu.Lock()
		delete(g.edgesLoading, internalID)
		if g.edgesLoaded == nil {
			g.edgesLoaded = make(map[uint32]bool)
		}
		g.edgesLoaded[internalID] = true
		g.mu.Unlock()
	}()

	// Load edges on-demand based on node type
	// Note: We don't hold the lock while loading because loadIndividualEdges/loadFamilyEdges
	// may call GetIndividual/GetFamily which need locks
	var err error
	switch node.NodeType() {
	case NodeTypeIndividual:
		err = g.loadIndividualEdgesUnlocked(node.(*IndividualNode))
	case NodeTypeFamily:
		err = g.loadFamilyEdgesUnlocked(node.(*FamilyNode))
	case NodeTypeNote, NodeTypeSource, NodeTypeRepository:
		err = g.loadReferenceEdgesUnlocked(node)
	case NodeTypeEvent:
		// Event edges are loaded when the owner's edges are loaded
		err = nil
	}
	
	return err
}

// loadIndividualEdgesUnlocked loads edges for an individual node on-demand.
// Must NOT be called while holding g.mu lock.
func (g *Graph) loadIndividualEdgesUnlocked(node *IndividualNode) error {
	xrefID := node.ID()

	// Load FAMC edges (child -> family)
	indi := node.Individual
	if indi != nil {
		famcXrefs := indi.GetValues("FAMC")
		for i, famcXref := range famcXrefs {
			famNode := g.GetFamily(famcXref)
			if famNode != nil {
				edgeID := fmt.Sprintf("%s_FAMC_%s_%d", xrefID, famcXref, i)
				edge := NewEdgeWithFamily(edgeID, node, famNode, EdgeTypeFAMC, famNode)
				if err := g.addEdgeInternal(edge); err != nil {
					// Continue on error (edge might already exist)
				}
			}
		}

		// Load FAMS edges (spouse -> family)
		famsXrefs := indi.GetValues("FAMS")
		for i, famsXref := range famsXrefs {
			famNode := g.GetFamily(famsXref)
			if famNode != nil {
				edgeID := fmt.Sprintf("%s_FAMS_%s_%d", xrefID, famsXref, i)
				edge := NewEdgeWithFamily(edgeID, node, famNode, EdgeTypeFAMS, famNode)
				if err := g.addEdgeInternal(edge); err != nil {
					// Continue on error
				}
			}
		}

		// Load NOTE edges
		notes := indi.GetNotes()
		for i, noteXref := range notes {
			noteNode := g.GetNote(noteXref)
			if noteNode != nil {
				edgeID := fmt.Sprintf("%s_NOTE_%s_%d", xrefID, noteXref, i)
				edge := NewEdge(edgeID, node, noteNode, EdgeTypeNOTE)
				if err := g.addEdgeInternal(edge); err != nil {
					// Continue on error
				}
			}
		}

		// Load SOUR edges
		sources := indi.GetSources()
		for i, sourceXref := range sources {
			sourceNode := g.GetSource(sourceXref)
			if sourceNode != nil {
				edgeID := fmt.Sprintf("%s_SOUR_%s_%d", xrefID, sourceXref, i)
				edge := NewEdge(edgeID, node, sourceNode, EdgeTypeSOUR)
				if err := g.addEdgeInternal(edge); err != nil {
					// Continue on error
				}
			}
		}
	}

	return nil
}

// loadFamilyEdgesUnlocked loads edges for a family node on-demand.
// Must NOT be called while holding g.mu lock.
func (g *Graph) loadFamilyEdgesUnlocked(node *FamilyNode) error {
	xrefID := node.ID()

	fam := node.Family
	if fam == nil {
		return nil
	}

	// Load HUSB edge
	husbandXref := fam.GetHusband()
	if husbandXref != "" {
		husbandNode := g.GetIndividual(husbandXref)
		if husbandNode != nil {
			edgeID := fmt.Sprintf("%s_HUSB_%s", xrefID, husbandXref)
			edge := NewEdgeWithFamily(edgeID, node, husbandNode, EdgeTypeHUSB, node)
			if err := g.AddEdge(edge); err != nil {
				// Continue on error
			}
			// Also create reverse FAMS edge
			edgeID2 := fmt.Sprintf("%s_FAMS_%s", husbandXref, xrefID)
			edge2 := NewEdgeWithFamily(edgeID2, husbandNode, node, EdgeTypeFAMS, node)
			if err := g.AddEdge(edge2); err != nil {
				// Continue on error
			}
		}
	}

	// Load WIFE edge
	wifeXref := fam.GetWife()
	if wifeXref != "" {
		wifeNode := g.GetIndividual(wifeXref)
		if wifeNode != nil {
			edgeID := fmt.Sprintf("%s_WIFE_%s", xrefID, wifeXref)
			edge := NewEdgeWithFamily(edgeID, node, wifeNode, EdgeTypeWIFE, node)
			if err := g.AddEdge(edge); err != nil {
				// Continue on error
			}
			// Also create reverse FAMS edge
			edgeID2 := fmt.Sprintf("%s_FAMS_%s", wifeXref, xrefID)
			edge2 := NewEdgeWithFamily(edgeID2, wifeNode, node, EdgeTypeFAMS, node)
			if err := g.AddEdge(edge2); err != nil {
				// Continue on error
			}
		}
	}

	// Load CHIL edges
	children := fam.GetChildren()
	for i, childXref := range children {
		childNode := g.GetIndividual(childXref)
		if childNode != nil {
			edgeID := fmt.Sprintf("%s_CHIL_%s_%d", xrefID, childXref, i)
			edge := NewEdgeWithFamily(edgeID, node, childNode, EdgeTypeCHIL, node)
			if err := g.AddEdge(edge); err != nil {
				// Continue on error
			}
			// Also create reverse FAMC edge
			edgeID2 := fmt.Sprintf("%s_FAMC_%s_%d", childXref, xrefID, i)
			edge2 := NewEdgeWithFamily(edgeID2, childNode, node, EdgeTypeFAMC, node)
			if err := g.AddEdge(edge2); err != nil {
				// Continue on error
			}
		}
	}

	// Load NOTE and SOUR edges
	notes := fam.GetNotes()
	for i, noteXref := range notes {
		noteNode := g.GetNote(noteXref)
		if noteNode != nil {
			edgeID := fmt.Sprintf("%s_NOTE_%s_%d", xrefID, noteXref, i)
			edge := NewEdge(edgeID, node, noteNode, EdgeTypeNOTE)
			if err := g.AddEdge(edge); err != nil {
				// Continue on error
			}
		}
	}

	sources := fam.GetSources()
	for i, sourceXref := range sources {
		sourceNode := g.GetSource(sourceXref)
		if sourceNode != nil {
			edgeID := fmt.Sprintf("%s_SOUR_%s_%d", xrefID, sourceXref, i)
			edge := NewEdge(edgeID, node, sourceNode, EdgeTypeSOUR)
			if err := g.AddEdge(edge); err != nil {
				// Continue on error
			}
		}
	}

	return nil
}

// loadReferenceEdgesUnlocked loads edges for note/source/repository nodes.
// Must NOT be called while holding g.mu lock.
func (g *Graph) loadReferenceEdgesUnlocked(node GraphNode) error {
	// Reference nodes typically don't have outgoing edges
	// They're referenced by other nodes
	return nil
}

