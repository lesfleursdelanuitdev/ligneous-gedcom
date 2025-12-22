package query

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

