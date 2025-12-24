package query

import (
	"fmt"
)

// RelationshipResult represents the relationship between two individuals.
type RelationshipResult struct {
	RelationshipType string
	Degree           int
	Removal          int
	Path             *Path
	AllPaths         []*Path
	IsDirect         bool
	IsAncestral      bool
	IsDescendant     bool
	IsCollateral     bool
}

// CalculateRelationship calculates the relationship between two individuals.
func (g *Graph) CalculateRelationship(fromXref, toXref string) (*RelationshipResult, error) {
	fromID := g.GetNodeID(fromXref)
	toID := g.GetNodeID(toXref)
	fromNode := g.individuals[fromID]
	toNode := g.individuals[toID]

	if fromNode == nil {
		return nil, fmt.Errorf("individual %s not found", fromXref)
	}
	if toNode == nil {
		return nil, fmt.Errorf("individual %s not found", toXref)
	}

	result := &RelationshipResult{}

	// Find shortest path (using XREF strings)
	path, err := g.ShortestPath(fromXref, toXref)
	if err != nil {
		return nil, err
	}
	result.Path = path

	// Find all paths (limited to reasonable number)
	allPaths, _ := g.AllPaths(fromXref, toXref, 10)
	result.AllPaths = allPaths

	// Check if direct relationship (parent, child, sibling, spouse)
	result.IsDirect = g.isDirectRelationship(fromNode, toNode)

	// Check if ancestral relationship (using XREF strings)
	result.IsAncestral = g.isAncestralRelationship(fromXref, toXref)

	// Check if descendant relationship
	result.IsDescendant = g.isAncestralRelationship(toXref, fromXref)

	// Check if collateral (cousins, uncles, etc.)
	result.IsCollateral = !result.IsDirect && !result.IsAncestral && !result.IsDescendant

	// Calculate relationship degree and type
	if result.IsDirect {
		result.RelationshipType = g.getDirectRelationshipType(fromNode, toNode)
		result.Degree = 0
		result.Removal = 0
	} else if result.IsAncestral || result.IsDescendant {
		result.RelationshipType = g.getAncestralRelationshipType(result.IsAncestral)
		result.Degree = g.calculateGenerations(path)
		result.Removal = 0
	} else if result.IsCollateral {
		// Find common ancestor (using XREF strings)
		commonAncestors, _ := g.CommonAncestors(fromXref, toXref)
		if len(commonAncestors) > 0 {
			lca, _ := g.LowestCommonAncestor(fromXref, toXref)
			if lca != nil {
				// Calculate degree: generations from LCA to both individuals
				fromDepth := g.getAncestorDepth(fromXref, lca.ID())
				toDepth := g.getAncestorDepth(toXref, lca.ID())
				result.Degree = min(fromDepth, toDepth) - 1
				result.Removal = abs(fromDepth - toDepth)
				result.RelationshipType = g.getCollateralRelationshipType(result.Degree, result.Removal)
			}
		}
	}

	return result, nil
}

// isDirectRelationship checks if two individuals have a direct relationship.
func (g *Graph) isDirectRelationship(from, to *IndividualNode) bool {
	// Check if parent (from is parent of to)
	parents := to.getParentsFromEdges()
	for _, parent := range parents {
		if parent.ID() == from.ID() {
			return true
		}
	}

	// Check if child (to is child of from)
	children := from.getChildrenFromEdges()
	for _, child := range children {
		if child.ID() == to.ID() {
			return true
		}
	}

	// Check if sibling
	siblings := from.getSiblingsFromEdges()
	for _, sibling := range siblings {
		if sibling.ID() == to.ID() {
			return true
		}
	}

	// Check if spouse
	spouses := from.getSpousesFromEdges()
	for _, spouse := range spouses {
		if spouse.ID() == to.ID() {
			return true
		}
	}

	return false
}

// isAncestralRelationship checks if toXref is an ancestor of fromXref.
func (g *Graph) isAncestralRelationship(fromXref, toXref string) bool {
	fromID := g.GetNodeID(fromXref)
	fromNode := g.individuals[fromID]
	if fromNode == nil {
		return false
	}

	// Find all ancestors of fromNode (using uint32 IDs)
	ancestors := g.findAllAncestors(fromNode, make(map[uint32]bool))

	// Check if toXref is in ancestors
	toID := g.GetNodeID(toXref)
	return ancestors[toID]
}

// getDirectRelationshipType returns the type of direct relationship.
func (g *Graph) getDirectRelationshipType(from, to *IndividualNode) string {
	// Check if parent (from is parent of to)
	parents := to.getParentsFromEdges()
	for _, parent := range parents {
		if parent.ID() == from.ID() {
			return "parent"
		}
	}

	// Check if child (to is child of from)
	children := from.getChildrenFromEdges()
	for _, child := range children {
		if child.ID() == to.ID() {
			return "child"
		}
	}

	// Check if sibling
	siblings := from.getSiblingsFromEdges()
	for _, sibling := range siblings {
		if sibling.ID() == to.ID() {
			return "sibling"
		}
	}

	// Check if spouse
	spouses := from.getSpousesFromEdges()
	for _, spouse := range spouses {
		if spouse.ID() == to.ID() {
			return "spouse"
		}
	}

	return "unknown"
}

// getAncestralRelationshipType returns the type of ancestral relationship.
func (g *Graph) getAncestralRelationshipType(isAncestral bool) string {
	if isAncestral {
		return "ancestor"
	}
	return "descendant"
}

// getCollateralRelationshipType returns the type of collateral relationship.
func (g *Graph) getCollateralRelationshipType(degree, removal int) string {
	if degree == 0 && removal == 0 {
		return "sibling"
	}
	if degree == 1 && removal == 0 {
		return "cousin"
	}
	if degree == 1 && removal == 1 {
		return "cousin once removed"
	}
	if degree == 1 && removal > 1 {
		return fmt.Sprintf("cousin %d times removed", removal)
	}
	if degree > 1 && removal == 0 {
		return fmt.Sprintf("%dth cousin", degree)
	}
	if degree > 1 && removal > 0 {
		return fmt.Sprintf("%dth cousin %d times removed", degree, removal)
	}
	return "distant relative"
}

// calculateGenerations counts the number of generations in a path.
func (g *Graph) calculateGenerations(path *Path) int {
	count := 0
	for _, edge := range path.Edges {
		if edge.EdgeType == EdgeTypeFAMC || edge.EdgeType == EdgeTypeCHIL {
			count++
		}
	}
	return count
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
