package query

import (
	"fmt"
	"sort"
	"strings"
)

// AddNodeIncremental adds a node to the graph and updates relationships incrementally.
// This is more efficient than rebuilding the entire graph.
func (g *Graph) AddNodeIncremental(node GraphNode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Add the node using existing AddNode logic
	if err := g.addNodeInternal(node); err != nil {
		return err
	}

	// Update indexes if it's an individual
	if indiNode, ok := node.(*IndividualNode); ok {
		g.updateIndexesForIndividual(indiNode)
	}

	// Invalidate cache (relationships may have changed)
	g.cache.clear()

	return nil
}

// RemoveNodeIncremental removes a node from the graph and cleans up relationships.
func (g *Graph) RemoveNodeIncremental(nodeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	node := g.nodes[nodeID]
	if node == nil {
		return fmt.Errorf("node %s not found", nodeID)
	}

	// Remove all edges connected to this node
	edgesToRemove := make([]*Edge, 0)
	for _, edge := range node.OutEdges() {
		edgesToRemove = append(edgesToRemove, edge)
	}
	for _, edge := range node.InEdges() {
		edgesToRemove = append(edgesToRemove, edge)
	}

	// Remove edges (this will update relationships)
	for _, edge := range edgesToRemove {
		if err := g.removeEdgeInternal(edge.ID); err != nil {
			// Continue even if edge removal fails
			continue
		}
	}

	// Remove from type-specific maps
	switch node.NodeType() {
	case NodeTypeIndividual:
		delete(g.individuals, nodeID)
		// Update indexes (remove from all indexes)
		g.removeFromIndexes(nodeID)
	case NodeTypeFamily:
		delete(g.families, nodeID)
	case NodeTypeNote:
		delete(g.notes, nodeID)
	case NodeTypeSource:
		delete(g.sources, nodeID)
	case NodeTypeRepository:
		delete(g.repositories, nodeID)
	case NodeTypeEvent:
		delete(g.events, nodeID)
	}

	// Remove from main nodes map
	delete(g.nodes, nodeID)

	// Invalidate cache
	g.cache.clear()

	return nil
}

// AddEdgeIncremental adds an edge to the graph and updates relationships incrementally.
func (g *Graph) AddEdgeIncremental(edge *Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Add the edge using existing AddEdge logic
	if err := g.addEdgeInternal(edge); err != nil {
		return err
	}

	// Update cached relationships based on edge type
	g.updateRelationshipsForEdge(edge)

	// Invalidate cache
	g.cache.clear()

	return nil
}

// RemoveEdgeIncremental removes an edge from the graph and updates relationships.
func (g *Graph) RemoveEdgeIncremental(edgeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	edge := g.edges[edgeID]
	if edge == nil {
		return fmt.Errorf("edge %s not found", edgeID)
	}

	// Store edge info before removal (for relationship updates)
	fromNode := edge.From
	toNode := edge.To
	edgeType := edge.EdgeType

	// Remove the edge
	if err := g.removeEdgeInternal(edgeID); err != nil {
		return err
	}

	// Update cached relationships
	g.updateRelationshipsAfterEdgeRemoval(fromNode, toNode, edgeType)

	// Invalidate cache
	g.cache.clear()

	return nil
}

// Internal helper methods (must be called with lock held)

func (g *Graph) addNodeInternal(node GraphNode) error {
	id := node.ID()
	if id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	// Check if node already exists
	if _, exists := g.nodes[id]; exists {
		return fmt.Errorf("node with ID %s already exists", id)
	}

	// Add to nodes map
	g.nodes[id] = node

	// Add to type-specific map
	switch node.NodeType() {
	case NodeTypeIndividual:
		if indiNode, ok := node.(*IndividualNode); ok {
			g.individuals[id] = indiNode
		}
	case NodeTypeFamily:
		if famNode, ok := node.(*FamilyNode); ok {
			g.families[id] = famNode
		}
	case NodeTypeNote:
		if noteNode, ok := node.(*NoteNode); ok {
			g.notes[id] = noteNode
		}
	case NodeTypeSource:
		if sourceNode, ok := node.(*SourceNode); ok {
			g.sources[id] = sourceNode
		}
	case NodeTypeRepository:
		if repoNode, ok := node.(*RepositoryNode); ok {
			g.repositories[id] = repoNode
		}
	case NodeTypeEvent:
		if eventNode, ok := node.(*EventNode); ok {
			g.events[id] = eventNode
		}
	}

	return nil
}

func (g *Graph) addEdgeInternal(edge *Edge) error {
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

	// Add to edge index
	fromID := edge.From.ID()
	toID := edge.To.ID()

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

func (g *Graph) removeEdgeInternal(edgeID string) error {
	edge := g.edges[edgeID]
	if edge == nil {
		return fmt.Errorf("edge %s not found", edgeID)
	}

	fromID := edge.From.ID()
	toID := edge.To.ID()

	// Remove from edge index
	g.removeFromEdgeIndex(fromID, edge)
	g.removeFromEdgeIndex(toID, edge)

	// Remove from node's edge lists
	edge.From.RemoveOutEdge(edge)
	edge.To.RemoveInEdge(edge)

	// If bidirectional, also remove reverse
	if edge.IsBidirectional() {
		edge.To.RemoveOutEdge(edge)
		edge.From.RemoveInEdge(edge)
	}

	// Remove from edges map
	delete(g.edges, edgeID)

	return nil
}

func (g *Graph) removeFromEdgeIndex(nodeID string, edge *Edge) {
	edges := g.edgeIndex[nodeID]
	for i, e := range edges {
		if e.ID == edge.ID {
			// Remove by swapping with last element and truncating
			edges[i] = edges[len(edges)-1]
			g.edgeIndex[nodeID] = edges[:len(edges)-1]
			return
		}
	}
}

// updateRelationshipsForEdge updates cached relationships when an edge is added.
func (g *Graph) updateRelationshipsForEdge(edge *Edge) {
	switch edge.EdgeType {
	case EdgeTypeHUSB:
		// Family -> Individual (husband)
		if famNode, ok := edge.From.(*FamilyNode); ok {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				famNode.Husband = indiNode
				// Update spouse relationship
				if famNode.Wife != nil {
					g.addSpouseRelationship(indiNode, famNode.Wife)
					g.addSpouseRelationship(famNode.Wife, indiNode)
				}
			}
		}

	case EdgeTypeWIFE:
		// Family -> Individual (wife)
		if famNode, ok := edge.From.(*FamilyNode); ok {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				famNode.Wife = indiNode
				// Update spouse relationship
				if famNode.Husband != nil {
					g.addSpouseRelationship(indiNode, famNode.Husband)
					g.addSpouseRelationship(famNode.Husband, indiNode)
				}
			}
		}

	case EdgeTypeCHIL:
		// Family -> Individual (child)
		if famNode, ok := edge.From.(*FamilyNode); ok {
			if indiNode, ok := edge.To.(*IndividualNode); ok {
				// Add to family's children
				if !g.containsChild(famNode.Children, indiNode) {
					famNode.Children = append(famNode.Children, indiNode)
				}
				// Update parent relationships
				if famNode.Husband != nil {
					g.addParentChildRelationship(famNode.Husband, indiNode)
				}
				if famNode.Wife != nil {
					g.addParentChildRelationship(famNode.Wife, indiNode)
				}
				// Update sibling relationships
				for _, sibling := range famNode.Children {
					if sibling.ID() != indiNode.ID() {
						g.addSiblingRelationship(indiNode, sibling)
						g.addSiblingRelationship(sibling, indiNode)
					}
				}
			}
		}

	case EdgeTypeFAMC:
		// Individual -> Family (child of family)
		if indiNode, ok := edge.From.(*IndividualNode); ok {
			if edge.Family != nil {
				famNode := edge.Family
				// Update parent relationships
				if famNode.Husband != nil {
					g.addParentChildRelationship(famNode.Husband, indiNode)
				}
				if famNode.Wife != nil {
					g.addParentChildRelationship(famNode.Wife, indiNode)
				}
				// Update sibling relationships
				for _, sibling := range famNode.Children {
					if sibling.ID() != indiNode.ID() {
						g.addSiblingRelationship(indiNode, sibling)
						g.addSiblingRelationship(sibling, indiNode)
					}
				}
			}
		}

	case EdgeTypeFAMS:
		// Individual -> Family (spouse in family)
		if indiNode, ok := edge.From.(*IndividualNode); ok {
			if edge.Family != nil {
				famNode := edge.Family
				// Update spouse relationships
				if famNode.Husband != nil && famNode.Husband.ID() != indiNode.ID() {
					g.addSpouseRelationship(indiNode, famNode.Husband)
					g.addSpouseRelationship(famNode.Husband, indiNode)
				}
				if famNode.Wife != nil && famNode.Wife.ID() != indiNode.ID() {
					g.addSpouseRelationship(indiNode, famNode.Wife)
					g.addSpouseRelationship(famNode.Wife, indiNode)
				}
				// Update child relationships
				for _, child := range famNode.Children {
					g.addParentChildRelationship(indiNode, child)
				}
			}
		}
	}
}

// updateRelationshipsAfterEdgeRemoval updates cached relationships when an edge is removed.
func (g *Graph) updateRelationshipsAfterEdgeRemoval(fromNode, toNode GraphNode, edgeType EdgeType) {
	switch edgeType {
	case EdgeTypeHUSB:
		if famNode, ok := fromNode.(*FamilyNode); ok {
			famNode.Husband = nil
			// Remove spouse relationships
			if famNode.Wife != nil {
				g.removeSpouseRelationship(famNode.Wife, toNode)
			}
		}

	case EdgeTypeWIFE:
		if famNode, ok := fromNode.(*FamilyNode); ok {
			famNode.Wife = nil
			// Remove spouse relationships
			if famNode.Husband != nil {
				g.removeSpouseRelationship(famNode.Husband, toNode)
			}
		}

	case EdgeTypeCHIL:
		if famNode, ok := fromNode.(*FamilyNode); ok {
			// Remove from family's children
			famNode.Children = g.removeChild(famNode.Children, toNode)
			// Remove parent-child relationships
			if indiNode, ok := toNode.(*IndividualNode); ok {
				if famNode.Husband != nil {
					g.removeParentChildRelationship(famNode.Husband, indiNode)
				}
				if famNode.Wife != nil {
					g.removeParentChildRelationship(famNode.Wife, indiNode)
				}
				// Remove sibling relationships
				for _, sibling := range famNode.Children {
					g.removeSiblingRelationship(indiNode, sibling)
					g.removeSiblingRelationship(sibling, indiNode)
				}
			}
		}

	case EdgeTypeFAMC:
		if indiNode, ok := fromNode.(*IndividualNode); ok {
			// Remove parent relationships
			// Need to find the family to get parents
			for _, edge := range indiNode.OutEdges() {
				if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
					famNode := edge.Family
					if famNode.Husband != nil {
						g.removeParentChildRelationship(famNode.Husband, indiNode)
					}
					if famNode.Wife != nil {
						g.removeParentChildRelationship(famNode.Wife, indiNode)
					}
					// Remove sibling relationships
					for _, sibling := range famNode.Children {
						if sibling.ID() != indiNode.ID() {
							g.removeSiblingRelationship(indiNode, sibling)
							g.removeSiblingRelationship(sibling, indiNode)
						}
					}
				}
			}
		}

	case EdgeTypeFAMS:
		if indiNode, ok := fromNode.(*IndividualNode); ok {
			// Remove spouse and child relationships
			// Need to find the family
			for _, edge := range indiNode.OutEdges() {
				if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
					famNode := edge.Family
					if famNode.Husband != nil && famNode.Husband.ID() != indiNode.ID() {
						g.removeSpouseRelationship(indiNode, famNode.Husband)
					}
					if famNode.Wife != nil && famNode.Wife.ID() != indiNode.ID() {
						g.removeSpouseRelationship(indiNode, famNode.Wife)
					}
					for _, child := range famNode.Children {
						g.removeParentChildRelationship(indiNode, child)
					}
				}
			}
		}
	}
}

// Helper functions for relationship management

func (g *Graph) addSpouseRelationship(indi1, indi2 *IndividualNode) {
	if !g.containsSpouse(indi1.Spouses, indi2) {
		indi1.Spouses = append(indi1.Spouses, indi2)
	}
}

func (g *Graph) removeSpouseRelationship(indi1 *IndividualNode, indi2 GraphNode) {
	if indi2Node, ok := indi2.(*IndividualNode); ok {
		indi1.Spouses = g.removeSpouse(indi1.Spouses, indi2Node)
	}
}

func (g *Graph) addParentChildRelationship(parent, child *IndividualNode) {
	if !g.containsChild(parent.Children, child) {
		parent.Children = append(parent.Children, child)
	}
	if !g.containsParent(child.Parents, parent) {
		child.Parents = append(child.Parents, parent)
	}
}

func (g *Graph) removeParentChildRelationship(parent *IndividualNode, child GraphNode) {
	if childNode, ok := child.(*IndividualNode); ok {
		parent.Children = g.removeChild(parent.Children, childNode)
		childNode.Parents = g.removeParent(childNode.Parents, parent)
	}
}

func (g *Graph) addSiblingRelationship(sib1, sib2 *IndividualNode) {
	if sib1.ID() != sib2.ID() && !g.containsSibling(sib1.Siblings, sib2) {
		sib1.Siblings = append(sib1.Siblings, sib2)
	}
}

func (g *Graph) removeSiblingRelationship(sib1 *IndividualNode, sib2 GraphNode) {
	if sib2Node, ok := sib2.(*IndividualNode); ok {
		sib1.Siblings = g.removeSibling(sib1.Siblings, sib2Node)
	}
}

// Contains checks

func (g *Graph) containsSpouse(spouses []*IndividualNode, spouse *IndividualNode) bool {
	for _, s := range spouses {
		if s.ID() == spouse.ID() {
			return true
		}
	}
	return false
}

func (g *Graph) containsChild(children []*IndividualNode, child *IndividualNode) bool {
	for _, c := range children {
		if c.ID() == child.ID() {
			return true
		}
	}
	return false
}

func (g *Graph) containsParent(parents []*IndividualNode, parent *IndividualNode) bool {
	for _, p := range parents {
		if p.ID() == parent.ID() {
			return true
		}
	}
	return false
}

func (g *Graph) containsSibling(siblings []*IndividualNode, sibling *IndividualNode) bool {
	for _, s := range siblings {
		if s.ID() == sibling.ID() {
			return true
		}
	}
	return false
}

// Remove helpers

func (g *Graph) removeSpouse(spouses []*IndividualNode, spouse *IndividualNode) []*IndividualNode {
	result := make([]*IndividualNode, 0, len(spouses))
	for _, s := range spouses {
		if s.ID() != spouse.ID() {
			result = append(result, s)
		}
	}
	return result
}

func (g *Graph) removeChild(children []*IndividualNode, child GraphNode) []*IndividualNode {
	childID := child.ID()
	result := make([]*IndividualNode, 0, len(children))
	for _, c := range children {
		if c.ID() != childID {
			result = append(result, c)
		}
	}
	return result
}

func (g *Graph) removeParent(parents []*IndividualNode, parent *IndividualNode) []*IndividualNode {
	result := make([]*IndividualNode, 0, len(parents))
	for _, p := range parents {
		if p.ID() != parent.ID() {
			result = append(result, p)
		}
	}
	return result
}

func (g *Graph) removeSibling(siblings []*IndividualNode, sibling *IndividualNode) []*IndividualNode {
	result := make([]*IndividualNode, 0, len(siblings))
	for _, s := range siblings {
		if s.ID() != sibling.ID() {
			result = append(result, s)
		}
	}
	return result
}

// Index management

func (g *Graph) updateIndexesForIndividual(indiNode *IndividualNode) {
	if indiNode.Individual == nil {
		return
	}

	indi := indiNode.Individual
	xrefID := indiNode.ID()

	// Update name index
	name := strings.ToLower(indi.GetName())
	if name != "" {
		g.indexes.nameIndex[name] = append(g.indexes.nameIndex[name], xrefID)
		words := strings.Fields(name)
		for _, word := range words {
			if len(word) > 2 {
				g.indexes.nameIndex[word] = append(g.indexes.nameIndex[word], xrefID)
			}
		}
	}

	// Update birth date index
	birthDate, err := indi.GetBirthDateParsed()
	if err == nil && birthDate != nil && birthDate.IsValid() {
		g.indexes.birthDateIndex = append(g.indexes.birthDateIndex, &dateIndexEntry{
			xrefID:    xrefID,
			birthDate: birthDate,
		})
		// Re-sort (could be optimized with insertion sort)
		sort.Slice(g.indexes.birthDateIndex, func(i, j int) bool {
			dateI := g.indexes.birthDateIndex[i].birthDate.Earliest()
			dateJ := g.indexes.birthDateIndex[j].birthDate.Earliest()
			return dateI.Before(dateJ)
		})
	}

	// Update place index
	birthPlace := strings.ToLower(indi.GetBirthPlace())
	if birthPlace != "" {
		g.indexes.placeIndex[birthPlace] = append(g.indexes.placeIndex[birthPlace], xrefID)
		words := strings.Fields(birthPlace)
		for _, word := range words {
			if len(word) > 2 {
				g.indexes.placeIndex[word] = append(g.indexes.placeIndex[word], xrefID)
			}
		}
	}

	// Update sex index
	sex := strings.ToUpper(indi.GetSex())
	if sex != "" {
		g.indexes.sexIndex[sex] = append(g.indexes.sexIndex[sex], xrefID)
	}

	// Update boolean indexes
	g.indexes.hasChildrenIndex[xrefID] = len(indiNode.Children) > 0
	g.indexes.hasSpouseIndex[xrefID] = len(indiNode.Spouses) > 0
	g.indexes.livingIndex[xrefID] = indi.GetDeathDate() == ""
}

func (g *Graph) removeFromIndexes(xrefID string) {
	// Remove from name index
	for key, xrefIDs := range g.indexes.nameIndex {
		g.indexes.nameIndex[key] = g.removeFromSlice(xrefIDs, xrefID)
		if len(g.indexes.nameIndex[key]) == 0 {
			delete(g.indexes.nameIndex, key)
		}
	}

	// Remove from birth date index
	newDateIndex := make([]*dateIndexEntry, 0)
	for _, entry := range g.indexes.birthDateIndex {
		if entry.xrefID != xrefID {
			newDateIndex = append(newDateIndex, entry)
		}
	}
	g.indexes.birthDateIndex = newDateIndex

	// Remove from place index
	for key, xrefIDs := range g.indexes.placeIndex {
		g.indexes.placeIndex[key] = g.removeFromSlice(xrefIDs, xrefID)
		if len(g.indexes.placeIndex[key]) == 0 {
			delete(g.indexes.placeIndex, key)
		}
	}

	// Remove from sex index
	for key, xrefIDs := range g.indexes.sexIndex {
		g.indexes.sexIndex[key] = g.removeFromSlice(xrefIDs, xrefID)
		if len(g.indexes.sexIndex[key]) == 0 {
			delete(g.indexes.sexIndex, key)
		}
	}

	// Remove from boolean indexes
	delete(g.indexes.hasChildrenIndex, xrefID)
	delete(g.indexes.hasSpouseIndex, xrefID)
	delete(g.indexes.livingIndex, xrefID)
}

func (g *Graph) removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
