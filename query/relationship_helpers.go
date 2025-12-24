package query

// Parents returns all parents of this individual.
// Computes parents from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Parents() []*IndividualNode {
	return node.getParentsFromEdges()
}

// getParentsFromEdges computes parents from edges (replaces cached Parents field).
func (node *IndividualNode) getParentsFromEdges() []*IndividualNode {
	parents := make([]*IndividualNode, 0)
	seen := make(map[string]bool)

	// Find parents via FAMC edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
			famNode := edge.Family
			// Get parents from family's HUSB/WIFE edges
			for _, famEdge := range famNode.OutEdges() {
				if (famEdge.EdgeType == EdgeTypeHUSB || famEdge.EdgeType == EdgeTypeWIFE) {
					if indiNode, ok := famEdge.To.(*IndividualNode); ok {
						if !seen[indiNode.ID()] {
							seen[indiNode.ID()] = true
							parents = append(parents, indiNode)
						}
					}
				}
			}
		}
	}

	return parents
}

// Children returns all children of this individual.
// Computes children from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Children() []*IndividualNode {
	return node.getChildrenFromEdges()
}

// getChildrenFromEdges computes children from edges (replaces cached Children field).
func (node *IndividualNode) getChildrenFromEdges() []*IndividualNode {
	children := make([]*IndividualNode, 0)
	seen := make(map[string]bool)

	// Find children via FAMS -> Family -> CHIL edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
			famNode := edge.Family
			// Get children from family's CHIL edges
			for _, famEdge := range famNode.OutEdges() {
				if famEdge.EdgeType == EdgeTypeCHIL {
					if indiNode, ok := famEdge.To.(*IndividualNode); ok {
						if !seen[indiNode.ID()] {
							seen[indiNode.ID()] = true
							children = append(children, indiNode)
						}
					}
				}
			}
		}
	}

	return children
}

// Spouses returns all spouses of this individual.
// Computes spouses from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Spouses() []*IndividualNode {
	return node.getSpousesFromEdges()
}

// getSpousesFromEdges computes spouses from edges (replaces cached Spouses field).
func (node *IndividualNode) getSpousesFromEdges() []*IndividualNode {
	spouses := make([]*IndividualNode, 0)
	seen := make(map[string]bool)

	// Find spouses via FAMS edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
			famNode := edge.Family
			// Get spouse from family's HUSB/WIFE edges (the other spouse)
			for _, famEdge := range famNode.OutEdges() {
				if (famEdge.EdgeType == EdgeTypeHUSB || famEdge.EdgeType == EdgeTypeWIFE) {
					if indiNode, ok := famEdge.To.(*IndividualNode); ok {
						if indiNode.ID() != node.ID() && !seen[indiNode.ID()] {
							seen[indiNode.ID()] = true
							spouses = append(spouses, indiNode)
						}
					}
				}
			}
		}
	}

	return spouses
}

// Siblings returns all siblings of this individual.
// Computes siblings from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *IndividualNode) Siblings() []*IndividualNode {
	return node.getSiblingsFromEdges()
}

// getSiblingsFromEdges computes siblings from edges (replaces cached Siblings field).
func (node *IndividualNode) getSiblingsFromEdges() []*IndividualNode {
	siblings := make([]*IndividualNode, 0)
	seen := make(map[string]bool)

	// Find siblings via shared FAMC (parent families)
	parentFamilies := make(map[string]*FamilyNode)
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
			parentFamilies[edge.Family.ID()] = edge.Family
		}
	}

	// Get siblings from parent families' children
	for _, famNode := range parentFamilies {
		for _, famEdge := range famNode.OutEdges() {
			if famEdge.EdgeType == EdgeTypeCHIL {
				if indiNode, ok := famEdge.To.(*IndividualNode); ok {
					if indiNode.ID() != node.ID() && !seen[indiNode.ID()] {
						seen[indiNode.ID()] = true
						siblings = append(siblings, indiNode)
					}
				}
			}
		}
	}

	return siblings
}

// Husband returns the husband of this family.
// Computes husband from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *FamilyNode) Husband() *IndividualNode {
	return node.getHusbandFromEdges()
}

// getHusbandFromEdges computes husband from edges (replaces cached Husband field).
func (node *FamilyNode) getHusbandFromEdges() *IndividualNode {
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeHUSB {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				return indiNode
			}
		}
	}
	return nil
}

// Wife returns the wife of this family.
// Computes wife from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *FamilyNode) Wife() *IndividualNode {
	return node.getWifeFromEdges()
}

// getWifeFromEdges computes wife from edges (replaces cached Wife field).
func (node *FamilyNode) getWifeFromEdges() *IndividualNode {
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeWIFE {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				return indiNode
			}
		}
	}
	return nil
}

// Children returns all children of this family.
// Computes children from edges (edge-based traversal).
// This is the recommended way to get relationships - use graph nodes, not records.
func (node *FamilyNode) Children() []*IndividualNode {
	return node.getChildrenFromEdges()
}

// getChildrenFromEdges computes children from edges (replaces cached Children field).
func (node *FamilyNode) getChildrenFromEdges() []*IndividualNode {
	children := make([]*IndividualNode, 0)
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeCHIL {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				children = append(children, indiNode)
			}
		}
	}
	return children
}

