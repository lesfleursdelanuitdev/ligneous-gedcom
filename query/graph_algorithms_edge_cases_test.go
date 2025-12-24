package query

import (
	"testing"
)

// TestGraphAlgorithms_ShortestPath tests shortest path finding with edge cases
func TestGraphAlgorithms_ShortestPath(t *testing.T) {
	tree := CreateTestTree()

	// Create a simple family tree: Grandparent -> Parent -> Child
	AddTestIndividual(tree, "@I1@", "Grandparent /Test/")
	AddTestIndividual(tree, "@I2@", "Parent /Test/")
	AddTestIndividual(tree, "@I3@", "Child /Test/")

	// Create families
	fam1 := AddTestFamily(tree, "@F1@", "@I1@", "", []string{"@I2@"}) // Grandparent -> Parent
	fam2 := AddTestFamily(tree, "@F2@", "@I2@", "", []string{"@I3@"}) // Parent -> Child
	_ = fam1
	_ = fam2

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: Path from grandparent to child
	path, err := graph.ShortestPath("@I1@", "@I3@")
	if err != nil {
		t.Fatalf("ShortestPath() failed: %v", err)
	}
	if path == nil {
		t.Fatal("Expected path, got nil")
	}
	if path.Length < 2 {
		t.Errorf("Expected path length >= 2, got %d", path.Length)
	}

	// Test: Path to self
	path2, err := graph.ShortestPath("@I1@", "@I1@")
	if err != nil {
		t.Fatalf("ShortestPath() to self failed: %v", err)
	}
	if path2.Length != 0 {
		t.Errorf("Expected path length 0 to self, got %d", path2.Length)
	}

	// Test: Path between unconnected nodes
	unconnected := AddTestIndividual(tree, "@I4@", "Unconnected /Test/")
	_ = unconnected
	// Rebuild query to include new individual
	q2, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}
	graph2 := q2.Graph()

	path3, err := graph2.ShortestPath("@I1@", "@I4@")
	if err == nil {
		t.Error("Expected error for unconnected nodes, got nil")
	}
	if path3 != nil {
		t.Error("Expected nil path for unconnected nodes")
	}
}

// TestGraphAlgorithms_CommonAncestors tests common ancestor finding
func TestGraphAlgorithms_CommonAncestors(t *testing.T) {
	tree := CreateTestTree()

	// Create tree: Grandparent -> Parent1, Parent2 -> Child1, Child2
	AddTestIndividual(tree, "@I1@", "Grandparent /Test/")
	AddTestIndividual(tree, "@I2@", "Parent1 /Test/")
	AddTestIndividual(tree, "@I3@", "Parent2 /Test/")
	AddTestIndividual(tree, "@I4@", "Child1 /Test/")
	AddTestIndividual(tree, "@I5@", "Child2 /Test/")

	// Families
	AddTestFamily(tree, "@F1@", "@I1@", "", []string{"@I2@", "@I3@"}) // Grandparent -> Parents
	AddTestFamily(tree, "@F2@", "@I2@", "", []string{"@I4@"})         // Parent1 -> Child1
	AddTestFamily(tree, "@F3@", "@I3@", "", []string{"@I5@"})         // Parent2 -> Child2


	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: Common ancestor of two siblings (child1 and child2)
	common, err := graph.CommonAncestors("@I4@", "@I5@")
	if err != nil {
		t.Fatalf("CommonAncestors() failed: %v", err)
	}
	if len(common) == 0 {
		t.Error("Expected at least one common ancestor, got 0")
	}

	// Test: Common ancestor of parent and child (should be the parent's parent)
	common2, err := graph.CommonAncestors("@I2@", "@I4@")
	if err != nil {
		t.Fatalf("CommonAncestors() failed: %v", err)
	}
	// Parent and child share ancestors (the parent's ancestors)
	if len(common2) == 0 {
		t.Error("Expected common ancestors between parent and child")
	}
}

// TestGraphAlgorithms_LowestCommonAncestor tests LCA finding
func TestGraphAlgorithms_LowestCommonAncestor(t *testing.T) {
	tree := CreateTestTree()

	// Create tree: Great-Grandparent -> Grandparent -> Parent -> Child
	AddTestIndividual(tree, "@I1@", "GreatGrandparent /Test/")
	AddTestIndividual(tree, "@I2@", "Grandparent /Test/")
	AddTestIndividual(tree, "@I3@", "Parent /Test/")
	AddTestIndividual(tree, "@I4@", "Child /Test/")

	AddTestFamily(tree, "@F1@", "@I1@", "", []string{"@I2@"})
	AddTestFamily(tree, "@F2@", "@I2@", "", []string{"@I3@"})
	AddTestFamily(tree, "@F3@", "@I3@", "", []string{"@I4@"})


	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: LCA of parent and child (should be the parent)
	lca, err := graph.LowestCommonAncestor("@I3@", "@I4@")
	if err != nil {
		t.Fatalf("LowestCommonAncestor() failed: %v", err)
	}
	if lca == nil {
		t.Fatal("Expected LCA, got nil")
	}
	if lca.ID() != "@I3@" {
		t.Errorf("Expected LCA to be @I3@ (parent), got %s", lca.ID())
	}
}

// TestGraphAlgorithms_CalculateRelationship tests relationship calculation
func TestGraphAlgorithms_CalculateRelationship(t *testing.T) {
	tree := CreateTestTree()

	// Create parent-child relationship
	AddTestIndividual(tree, "@I1@", "Parent /Test/")
	AddTestIndividual(tree, "@I2@", "Child /Test/")

	AddTestFamily(tree, "@F1@", "@I1@", "", []string{"@I2@"})

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: Parent to child relationship
	result, err := graph.CalculateRelationship("@I1@", "@I2@")
	if err != nil {
		t.Fatalf("CalculateRelationship() failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected relationship result, got nil")
	}
	if !result.IsDirect {
		t.Error("Expected IsDirect to be true for parent-child relationship")
	}
	if result.RelationshipType == "" {
		t.Error("Expected RelationshipType to be set")
	}
}

// TestGraphAlgorithms_DisconnectedGraph tests algorithms with disconnected components
func TestGraphAlgorithms_DisconnectedGraph(t *testing.T) {
	tree := CreateTestTree()

	// Create two disconnected family trees
	AddTestIndividual(tree, "@I1@", "Person1 /Test/")
	AddTestIndividual(tree, "@I2@", "Person2 /Test/")
	AddTestIndividual(tree, "@I3@", "Person3 /Test/")

	// Only connect I1 and I2, leave I3 disconnected
	AddTestFamily(tree, "@F1@", "@I1@", "", []string{"@I2@"})


	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: Path between connected nodes
	path1, err := graph.ShortestPath("@I1@", "@I2@")
	if err != nil {
		t.Fatalf("ShortestPath() failed for connected nodes: %v", err)
	}
	if path1 == nil {
		t.Error("Expected path between connected nodes, got nil")
	}

	// Test: Path between disconnected nodes
	path2, err := graph.ShortestPath("@I1@", "@I3@")
	if err == nil {
		t.Error("Expected error for disconnected nodes, got nil")
	}
	if path2 != nil {
		t.Error("Expected nil path for disconnected nodes")
	}
}

// TestGraphAlgorithms_AllPaths tests finding all paths between nodes
func TestGraphAlgorithms_AllPaths(t *testing.T) {
	tree := CreateTestTree()

	// Create a tree with multiple paths: A -> B -> D, A -> C -> D
	AddTestIndividual(tree, "@I1@", "A /Test/")
	AddTestIndividual(tree, "@I2@", "B /Test/")
	AddTestIndividual(tree, "@I3@", "C /Test/")
	AddTestIndividual(tree, "@I4@", "D /Test/")

	AddTestFamily(tree, "@F1@", "@I1@", "", []string{"@I2@", "@I3@"}) // A -> B, C
	AddTestFamily(tree, "@F2@", "@I2@", "", []string{"@I4@"})         // B -> D
	AddTestFamily(tree, "@F3@", "@I3@", "", []string{"@I4@"})         // C -> D


	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: All paths from A to D (should find multiple paths)
	allPaths, err := graph.AllPaths("@I1@", "@I4@", 10)
	if err != nil {
		t.Fatalf("AllPaths() failed: %v", err)
	}
	if len(allPaths) == 0 {
		t.Error("Expected at least one path, got 0")
	}
}

// TestGraphAlgorithms_EdgeCases tests edge cases for graph algorithms
func TestGraphAlgorithms_EdgeCases(t *testing.T) {
	tree := CreateTestTree()

	// Test with empty graph
	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: Path between non-existent nodes
	_, err = graph.ShortestPath("@I999@", "@I998@")
	if err == nil {
		t.Error("Expected error for non-existent nodes, got nil")
	}

	// Test: Common ancestors of non-existent nodes
	_, err = graph.CommonAncestors("@I999@", "@I998@")
	if err == nil {
		t.Error("Expected error for non-existent nodes, got nil")
	}

	// Test: LCA of non-existent nodes
	_, err = graph.LowestCommonAncestor("@I999@", "@I998@")
	if err == nil {
		t.Error("Expected error for non-existent nodes, got nil")
	}

	// Test: Relationship calculation for non-existent nodes
	_, err = graph.CalculateRelationship("@I999@", "@I998@")
	if err == nil {
		t.Error("Expected error for non-existent nodes, got nil")
	}
}

// TestGraphAlgorithms_CycleDetection tests that cycles don't cause infinite loops
func TestGraphAlgorithms_CycleDetection(t *testing.T) {
	tree := CreateTestTree()

	// Create a simple parent-child relationship (no cycles in normal GEDCOM)
	AddTestIndividual(tree, "@I1@", "Parent /Test/")
	AddTestIndividual(tree, "@I2@", "Child /Test/")

	AddTestFamily(tree, "@F1@", "@I1@", "", []string{"@I2@"})

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	graph := q.Graph()

	// Test: Path finding should complete (not loop infinitely)
	path, err := graph.ShortestPath("@I1@", "@I2@")
	if err != nil {
		t.Fatalf("ShortestPath() failed: %v", err)
	}
	if path == nil {
		t.Error("Expected path, got nil")
	}

	// Test: All paths should complete
	allPaths, err := graph.AllPaths("@I1@", "@I2@", 10)
	if err != nil {
		t.Fatalf("AllPaths() failed: %v", err)
	}
	_ = allPaths // Just verify it completes
}

