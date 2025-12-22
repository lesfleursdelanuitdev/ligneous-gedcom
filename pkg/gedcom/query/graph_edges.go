package query

import (
	"fmt"
)

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

