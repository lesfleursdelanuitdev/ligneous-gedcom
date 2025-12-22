package query

import (
	"fmt"
)

// CommonAncestors finds all common ancestors of two individuals.
func (g *Graph) CommonAncestors(indi1Xref, indi2Xref string) ([]*IndividualNode, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	indi1ID := g.xrefToID[indi1Xref]
	indi2ID := g.xrefToID[indi2Xref]

	indi1 := g.individuals[indi1ID]
	indi2 := g.individuals[indi2ID]

	if indi1 == nil {
		return nil, fmt.Errorf("individual %s not found", indi1Xref)
	}
	if indi2 == nil {
		return nil, fmt.Errorf("individual %s not found", indi2Xref)
	}

	// Find all ancestors of indi1 (using uint32 IDs)
	ancestors1 := g.findAllAncestors(indi1, make(map[uint32]bool))

	// Find all ancestors of indi2 (using uint32 IDs)
	ancestors2 := g.findAllAncestors(indi2, make(map[uint32]bool))

	// Find intersection
	common := make([]*IndividualNode, 0)
	for id := range ancestors1 {
		if ancestors2[id] {
			if node := g.individuals[id]; node != nil {
				common = append(common, node)
			}
		}
	}

	return common, nil
}

// findAllAncestors finds all ancestors of an individual recursively (using uint32 IDs).
func (g *Graph) findAllAncestors(indi *IndividualNode, visited map[uint32]bool) map[uint32]bool {
	// Get uint32 ID for this individual
	indiXref := indi.ID()
	indiID := g.getID(indiXref)
	if indiID == 0 {
		return visited
	}

	if visited[indiID] {
		return visited
	}

	visited[indiID] = true

	// Find parents via FAMC edges
	for _, edge := range indi.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
			famNode := edge.Family
			husband := famNode.getHusbandFromEdges()
			if husband != nil {
				g.findAllAncestors(husband, visited)
			}
			wife := famNode.getWifeFromEdges()
			if wife != nil {
				g.findAllAncestors(wife, visited)
			}
		}
	}

	return visited
}

// LowestCommonAncestor finds the lowest (most recent) common ancestor of two individuals.
// The LCA is the common ancestor that is closest to both individuals (most recent).
// This is the one that minimizes the maximum distance from both individuals.
func (g *Graph) LowestCommonAncestor(indi1ID, indi2ID string) (*IndividualNode, error) {
	commonAncestors, err := g.CommonAncestors(indi1ID, indi2ID)
	if err != nil {
		return nil, err
	}

	if len(commonAncestors) == 0 {
		return nil, fmt.Errorf("no common ancestors found")
	}

	// Find the lowest common ancestor (most recent)
	// This is the one that minimizes the maximum distance from both individuals
	lowest := commonAncestors[0]
	minMaxDepth := max(g.getAncestorDepth(indi1ID, lowest.ID()), g.getAncestorDepth(indi2ID, lowest.ID()))

	for _, ancestor := range commonAncestors[1:] {
		depth1 := g.getAncestorDepth(indi1ID, ancestor.ID())
		depth2 := g.getAncestorDepth(indi2ID, ancestor.ID())
		maxDepth := max(depth1, depth2)

		// The LCA is the one with the minimum maximum depth (closest to both)
		if maxDepth < minMaxDepth {
			minMaxDepth = maxDepth
			lowest = ancestor
		}
	}

	return lowest, nil
}

// getAncestorDepth calculates the depth (generations) from descendant to ancestor.
func (g *Graph) getAncestorDepth(descendantID, ancestorID string) int {
	if descendantID == ancestorID {
		return 0
	}

	path, err := g.ShortestPath(descendantID, ancestorID)
	if err != nil {
		return -1
	}

	// Count only blood relationship edges
	depth := 0
	for _, edge := range path.Edges {
		if edge.EdgeType == EdgeTypeFAMC || edge.EdgeType == EdgeTypeCHIL {
			depth++
		}
	}

	return depth
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
