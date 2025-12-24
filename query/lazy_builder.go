package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// BuildGraphLazy constructs a graph skeleton with lazy loading enabled.
// Nodes and edges are loaded on-demand when accessed.
func BuildGraphLazy(tree *types.GedcomTree) (*Graph, error) {
	graph := NewGraph(tree)
	graph.lazyMode = true

	// Phase 1: Build skeleton (node metadata only)
	if err := createNodeSkeleton(graph, tree); err != nil {
		return nil, fmt.Errorf("failed to create node skeleton: %w", err)
	}

	// Phase 2: Detect connected components (for partitioning)
	if err := detectComponents(graph, tree); err != nil {
		return nil, fmt.Errorf("failed to detect components: %w", err)
	}

	// Phase 3: Build indexes (can work with skeleton)
	// Note: Indexes may need nodes loaded, so we'll build them lazily too
	// graph.indexes.buildIndexes(graph) // Skip for now, build on-demand

	return graph, nil
}

// createNodeSkeleton creates metadata for all nodes without loading them.
func createNodeSkeleton(graph *Graph, tree *types.GedcomTree) error {
	// Create metadata for individuals
	individuals := tree.GetAllIndividuals()
	for xrefID := range individuals {
		internalID := graph.getOrCreateID(xrefID)
		graph.nodeMetadata[internalID] = &NodeMetadata{
			XrefID:     xrefID,
			NodeType:   NodeTypeIndividual,
			ComponentID: 0, // Will be assigned during component detection
			Loaded:     false,
		}
	}

	// Create metadata for families
	families := tree.GetAllFamilies()
	for xrefID := range families {
		internalID := graph.getOrCreateID(xrefID)
		graph.nodeMetadata[internalID] = &NodeMetadata{
			XrefID:     xrefID,
			NodeType:   NodeTypeFamily,
			ComponentID: 0,
			Loaded:     false,
		}
	}

	// Create metadata for notes
	notes := tree.GetAllNotes()
	for xrefID := range notes {
		internalID := graph.getOrCreateID(xrefID)
		graph.nodeMetadata[internalID] = &NodeMetadata{
			XrefID:     xrefID,
			NodeType:   NodeTypeNote,
			ComponentID: 0,
			Loaded:     false,
		}
	}

	// Create metadata for sources
	sources := tree.GetAllSources()
	for xrefID := range sources {
		internalID := graph.getOrCreateID(xrefID)
		graph.nodeMetadata[internalID] = &NodeMetadata{
			XrefID:     xrefID,
			NodeType:   NodeTypeSource,
			ComponentID: 0,
			Loaded:     false,
		}
	}

	// Create metadata for repositories
	repositories := tree.GetAllRepositories()
	for xrefID := range repositories {
		internalID := graph.getOrCreateID(xrefID)
		graph.nodeMetadata[internalID] = &NodeMetadata{
			XrefID:     xrefID,
			NodeType:   NodeTypeRepository,
			ComponentID: 0,
			Loaded:     false,
		}
	}

	return nil
}

// detectComponents identifies connected components in the graph skeleton.
// Uses BFS to find all connected nodes via family relationships.
func detectComponents(graph *Graph, tree *types.GedcomTree) error {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	visited := make(map[uint32]bool)
	componentID := uint32(1)

	// Start from each unvisited individual
	individuals := tree.GetAllIndividuals()
	for xrefID := range individuals {
		internalID := graph.xrefToID[xrefID]
		if internalID == 0 || visited[internalID] {
			continue
		}

		// BFS to find all connected nodes
		component := make([]uint32, 0)
		queue := []uint32{internalID}
		visited[internalID] = true

		for len(queue) > 0 {
			currentID := queue[0]
			queue = queue[1:]
			component = append(component, currentID)

			// Get metadata
			meta := graph.nodeMetadata[currentID]
			if meta == nil {
				continue
			}

			// Find connected nodes via family relationships
			// For individuals: find families via FAMC/FAMS
			if meta.NodeType == NodeTypeIndividual {
				record := tree.GetIndividual(meta.XrefID)
				if record != nil {
					if indi, ok := record.(*types.IndividualRecord); ok {
						// FAMC edges (child -> family)
						famcXrefs := indi.GetValues("FAMC")
						for _, famcXref := range famcXrefs {
							famID := graph.xrefToID[famcXref]
							if famID != 0 && !visited[famID] {
								visited[famID] = true
								queue = append(queue, famID)
							}
						}

						// FAMS edges (spouse -> family)
						famsXrefs := indi.GetValues("FAMS")
						for _, famsXref := range famsXrefs {
							famID := graph.xrefToID[famsXref]
							if famID != 0 && !visited[famID] {
								visited[famID] = true
								queue = append(queue, famID)
							}
						}
					}
				}
			}

			// For families: find individuals via HUSB/WIFE/CHIL
			if meta.NodeType == NodeTypeFamily {
				record := tree.GetFamily(meta.XrefID)
				if record != nil {
					if fam, ok := record.(*types.FamilyRecord); ok {
						// Husband
						if husbXref := fam.GetHusband(); husbXref != "" {
							husbID := graph.xrefToID[husbXref]
							if husbID != 0 && !visited[husbID] {
								visited[husbID] = true
								queue = append(queue, husbID)
							}
						}

						// Wife
						if wifeXref := fam.GetWife(); wifeXref != "" {
							wifeID := graph.xrefToID[wifeXref]
							if wifeID != 0 && !visited[wifeID] {
								visited[wifeID] = true
								queue = append(queue, wifeID)
							}
						}

						// Children
						children := fam.GetChildren()
						for _, childXref := range children {
							childID := graph.xrefToID[childXref]
							if childID != 0 && !visited[childID] {
								visited[childID] = true
								queue = append(queue, childID)
							}
						}
					}
				}
			}
		}

		// Assign component ID to all nodes in this component
		if len(component) > 0 {
			graph.components[componentID] = component
			for _, nodeID := range component {
				if meta := graph.nodeMetadata[nodeID]; meta != nil {
					meta.ComponentID = componentID
				}
			}
			componentID++
		}
	}

	graph.componentCount = componentID - 1
	return nil
}

// LoadComponent loads all nodes and edges in a component into memory.
func (g *Graph) LoadComponent(componentID uint32) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	component := g.components[componentID]
	if component == nil {
		return fmt.Errorf("component %d not found", componentID)
	}

	// Load all nodes in component
	for _, nodeID := range component {
		meta := g.nodeMetadata[nodeID]
		if meta == nil || meta.Loaded {
			continue
		}

		node, err := g.loadNodeFromTree(meta.XrefID, nodeID)
		if err != nil {
			continue // Skip on error
		}

		// Mark as loaded
		meta.Loaded = true

		// Load edges for this node
		if err := g.ensureEdgesLoadedUnlocked(node); err != nil {
			// Continue on error
		}
	}

	return nil
}

// GetComponentID returns the component ID for a node.
func (g *Graph) GetComponentID(xrefID string) uint32 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	internalID := g.xrefToID[xrefID]
	if internalID == 0 {
		return 0
	}

	meta := g.nodeMetadata[internalID]
	if meta == nil {
		return 0
	}

	return meta.ComponentID
}

// GetComponentCount returns the number of connected components.
func (g *Graph) GetComponentCount() uint32 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.componentCount
}

// GetComponentSize returns the number of nodes in a component.
func (g *Graph) GetComponentSize(componentID uint32) int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	component := g.components[componentID]
	if component == nil {
		return 0
	}

	return len(component)
}

