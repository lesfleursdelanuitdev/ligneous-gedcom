package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// GraphValidator validates graph integrity and edge consistency.
// This validator checks that the graph structure is correct after building.
type GraphValidator struct {
	errorManager *types.ErrorManager
}

// NewGraphValidator creates a new GraphValidator.
func NewGraphValidator(errorManager *types.ErrorManager) *GraphValidator {
	return &GraphValidator{
		errorManager: errorManager,
	}
}

// Validate validates the graph structure and returns an error if validation fails.
// Checks edge consistency, family structure, and relationship integrity.
func (gv *GraphValidator) Validate(graph *Graph) error {
	// Validate edge consistency
	if err := gv.validateEdges(graph); err != nil {
		return err
	}

	// Validate family structure
	if err := gv.validateFamilies(graph); err != nil {
		return err
	}

	// Validate relationship integrity
	if err := gv.validateRelationships(graph); err != nil {
		return err
	}

	// Validate node-record consistency
	if err := gv.validateNodeRecordConsistency(graph); err != nil {
		return err
	}

	// Validate for circular references
	if err := gv.validateCircularReferences(graph); err != nil {
		return err
	}

	// Validate for orphaned nodes (warnings only)
	gv.validateOrphanedNodes(graph)

	return nil
}

// validateEdges checks that all edges have valid From/To nodes and proper edge types.
func (gv *GraphValidator) validateEdges(graph *Graph) error {
	graph.mu.RLock()
	defer graph.mu.RUnlock()

	for edgeID, edge := range graph.edges {
		// Check edge has From node
		if edge.From == nil {
			gv.errorManager.AddError(
				types.SeveritySevere,
				fmt.Sprintf("Edge %s has nil From node", edgeID),
				0,
				"Graph Validation",
			)
			continue
		}

		// Check edge has To node
		if edge.To == nil {
			gv.errorManager.AddError(
				types.SeveritySevere,
				fmt.Sprintf("Edge %s has nil To node", edgeID),
				0,
				"Graph Validation",
			)
			continue
		}

		// Check edge type is valid
		if edge.EdgeType == "" {
			gv.errorManager.AddError(
				types.SeveritySevere,
				fmt.Sprintf("Edge %s has empty EdgeType", edgeID),
				0,
				"Graph Validation",
			)
			continue
		}

		// Check From node has this edge in its OutEdges
		fromOutEdges := edge.From.OutEdges()
		found := false
		for _, outEdge := range fromOutEdges {
			if outEdge.ID == edge.ID {
				found = true
				break
			}
		}
		if !found {
			gv.errorManager.AddError(
				types.SeverityWarning,
				fmt.Sprintf("Edge %s not found in From node's OutEdges", edgeID),
				0,
				"Graph Validation",
			)
		}

		// Check To node has this edge in its InEdges
		toInEdges := edge.To.InEdges()
		found = false
		for _, inEdge := range toInEdges {
			if inEdge.ID == edge.ID {
				found = true
				break
			}
		}
		if !found {
			gv.errorManager.AddError(
				types.SeverityWarning,
				fmt.Sprintf("Edge %s not found in To node's InEdges", edgeID),
				0,
				"Graph Validation",
			)
		}
	}

	if gv.errorManager.HasSevereErrors() {
		return fmt.Errorf("severe edge validation errors found")
	}

	return nil
}

// validateFamilies checks that families have valid structure (HUSB/WIFE/CHIL edges).
func (gv *GraphValidator) validateFamilies(graph *Graph) error {
	graph.mu.RLock()
	defer graph.mu.RUnlock()

	for _, familyNode := range graph.families {
		if familyNode == nil {
			continue
		}

		hasHusband := false
		hasWife := false
		childCount := 0

		// Check family edges
		for _, edge := range familyNode.OutEdges() {
			switch edge.EdgeType {
			case EdgeTypeHUSB:
				hasHusband = true
				if edge.To == nil {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Family %s has HUSB edge with nil To node", familyNode.ID()),
						0,
						"Graph Validation",
					)
				} else if _, ok := edge.To.(*IndividualNode); !ok {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Family %s has HUSB edge pointing to non-Individual node", familyNode.ID()),
						0,
						"Graph Validation",
					)
				}
			case EdgeTypeWIFE:
				hasWife = true
				if edge.To == nil {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Family %s has WIFE edge with nil To node", familyNode.ID()),
						0,
						"Graph Validation",
					)
				} else if _, ok := edge.To.(*IndividualNode); !ok {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Family %s has WIFE edge pointing to non-Individual node", familyNode.ID()),
						0,
						"Graph Validation",
					)
				}
			case EdgeTypeCHIL:
				childCount++
				if edge.To == nil {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Family %s has CHIL edge with nil To node", familyNode.ID()),
						0,
						"Graph Validation",
					)
				} else if _, ok := edge.To.(*IndividualNode); !ok {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Family %s has CHIL edge pointing to non-Individual node", familyNode.ID()),
						0,
						"Graph Validation",
					)
				}
			}
		}

		// Warn if family has no parents (might be valid for some cases)
		if !hasHusband && !hasWife {
			gv.errorManager.AddError(
				types.SeverityWarning,
				fmt.Sprintf("Family %s has no HUSB or WIFE edges", familyNode.ID()),
				0,
				"Graph Validation",
			)
		}
	}

	if gv.errorManager.HasSevereErrors() {
		return fmt.Errorf("severe family validation errors found")
	}

	return nil
}

// validateRelationships checks relationship consistency (bidirectional edges, etc.).
func (gv *GraphValidator) validateRelationships(graph *Graph) error {
	graph.mu.RLock()
	defer graph.mu.RUnlock()

	// Check FAMC/FAMS bidirectional consistency
	for _, individualNode := range graph.individuals {
		if individualNode == nil {
			continue
		}

		// Check FAMC edges have corresponding FAMS edges on family
		for _, edge := range individualNode.OutEdges() {
			if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
				// Individual should have FAMC edge to family
				// Family should have CHIL edge to individual
				found := false
				for _, famEdge := range edge.Family.OutEdges() {
					if famEdge.EdgeType == EdgeTypeCHIL && famEdge.To != nil && famEdge.To.ID() == individualNode.ID() {
						found = true
						break
					}
				}
				if !found {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Individual %s has FAMC edge to family %s, but family lacks CHIL edge to individual", individualNode.ID(), edge.Family.ID()),
						0,
						"Graph Validation",
					)
				}
			}

			// Check FAMS edges have corresponding HUSB/WIFE edges on family
			if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
				found := false
				for _, famEdge := range edge.Family.OutEdges() {
					if (famEdge.EdgeType == EdgeTypeHUSB || famEdge.EdgeType == EdgeTypeWIFE) &&
						famEdge.To != nil && famEdge.To.ID() == individualNode.ID() {
						found = true
						break
					}
				}
				if !found {
					gv.errorManager.AddError(
						types.SeveritySevere,
						fmt.Sprintf("Individual %s has FAMS edge to family %s, but family lacks HUSB/WIFE edge to individual", individualNode.ID(), edge.Family.ID()),
						0,
						"Graph Validation",
					)
				}
			}
		}
	}

	if gv.errorManager.HasSevereErrors() {
		return fmt.Errorf("severe relationship validation errors found")
	}

	return nil
}

// validateNodeRecordConsistency checks that all nodes reference valid records.
func (gv *GraphValidator) validateNodeRecordConsistency(graph *Graph) error {
	graph.mu.RLock()
	defer graph.mu.RUnlock()

	// Check individual nodes
	for _, node := range graph.individuals {
		if node == nil {
			continue
		}
		if node.Individual == nil {
			gv.errorManager.AddError(
				types.SeveritySevere,
				fmt.Sprintf("IndividualNode %s has nil Individual record", node.ID()),
				0,
				"Graph Validation",
			)
		} else if node.Record() == nil {
			gv.errorManager.AddError(
				types.SeveritySevere,
				fmt.Sprintf("IndividualNode %s has nil Record()", node.ID()),
				0,
				"Graph Validation",
			)
		}
	}

	// Check family nodes
	for _, node := range graph.families {
		if node == nil {
			continue
		}
		if node.Family == nil {
			gv.errorManager.AddError(
				types.SeveritySevere,
				fmt.Sprintf("FamilyNode %s has nil Family record", node.ID()),
				0,
				"Graph Validation",
			)
		} else if node.Record() == nil {
			gv.errorManager.AddError(
				types.SeveritySevere,
				fmt.Sprintf("FamilyNode %s has nil Record()", node.ID()),
				0,
				"Graph Validation",
			)
		}
	}

	if gv.errorManager.HasSevereErrors() {
		return fmt.Errorf("severe node-record consistency errors found")
	}

	return nil
}

// validateCircularReferences checks for impossible relationship cycles.
// For example, an individual cannot be their own ancestor.
func (gv *GraphValidator) validateCircularReferences(graph *Graph) error {
	graph.mu.RLock()
	defer graph.mu.RUnlock()

	// Check for cycles in parent-child relationships
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for _, node := range graph.individuals {
		if node == nil {
			continue
		}
		if !visited[node.ID()] {
			if gv.hasCycle(node, visited, recursionStack, graph) {
				gv.errorManager.AddError(
					types.SeveritySevere,
					fmt.Sprintf("Circular reference detected involving individual %s", node.ID()),
					0,
					"Graph Validation",
				)
			}
		}
	}

	if gv.errorManager.HasSevereErrors() {
		return fmt.Errorf("severe circular reference errors found")
	}

	return nil
}

// hasCycle uses DFS to detect cycles in parent-child relationships.
func (gv *GraphValidator) hasCycle(node *IndividualNode, visited, recursionStack map[string]bool, graph *Graph) bool {
	nodeID := node.ID()
	visited[nodeID] = true
	recursionStack[nodeID] = true

	// Check all parent relationships (FAMC edges)
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
			// Get parents from family
			for _, famEdge := range edge.Family.OutEdges() {
				if (famEdge.EdgeType == EdgeTypeHUSB || famEdge.EdgeType == EdgeTypeWIFE) && famEdge.To != nil {
					parent, ok := famEdge.To.(*IndividualNode)
					if !ok {
						continue
					}
					parentID := parent.ID()

					// If parent is in recursion stack, we found a cycle
					if recursionStack[parentID] {
						return true
					}

					// If parent not visited, recurse
					if !visited[parentID] {
						if gv.hasCycle(parent, visited, recursionStack, graph) {
							return true
						}
					}
				}
			}
		}
	}

	recursionStack[nodeID] = false
	return false
}

// validateOrphanedNodes checks for nodes without any edges (warnings only).
// Orphaned nodes might be valid in some cases (e.g., unlinked individuals).
func (gv *GraphValidator) validateOrphanedNodes(graph *Graph) {
	graph.mu.RLock()
	defer graph.mu.RUnlock()

	// Check individual nodes
	for _, node := range graph.individuals {
		if node == nil {
			continue
		}
		outEdges := node.OutEdges()
		inEdges := node.InEdges()
		if len(outEdges) == 0 && len(inEdges) == 0 {
			gv.errorManager.AddError(
				types.SeverityWarning,
				fmt.Sprintf("Individual %s has no edges (orphaned node)", node.ID()),
				0,
				"Graph Validation",
			)
		}
	}

	// Check family nodes
	for _, node := range graph.families {
		if node == nil {
			continue
		}
		outEdges := node.OutEdges()
		if len(outEdges) == 0 {
			gv.errorManager.AddError(
				types.SeverityWarning,
				fmt.Sprintf("Family %s has no edges (orphaned node)", node.ID()),
				0,
				"Graph Validation",
			)
		}
	}
}

// GetErrorManager returns the error manager.
func (gv *GraphValidator) GetErrorManager() *types.ErrorManager {
	return gv.errorManager
}

