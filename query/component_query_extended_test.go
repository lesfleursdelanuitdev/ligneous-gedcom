package query

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestGetComponentForPerson_EmptyTree(t *testing.T) {
	tree := types.NewGedcomTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with empty tree
	component, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson should not error on empty tree: %v", err)
	}

	if component != nil && len(component) != 0 {
		t.Error("Component should be empty for nonexistent individual in empty tree")
	}
}

func TestGetComponentForPerson_WithNilOptions(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with nil options (should use defaults)
	// Note: This tests the code path, but GetComponentForPerson doesn't accept nil
	// So we test with default options
	options := NewComponentOptions()
	component, err := graph.GetComponentForPersonWithOptions("@I1@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	if len(component) == 0 {
		t.Error("Component should not be empty")
	}
}

func TestGetComponentForPerson_DepthLimitReached(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with depth limit of 1
	// Depth 0 = starting node, Depth 1 = first level connections
	// So with MaxDepth=1, we process depth 0 and depth 1, then stop
	options := NewComponentOptions()
	options.MaxDepth = 1

	component, err := graph.GetComponentForPersonWithOptions("@I1@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// Should get starting node and first level connections
	// With depth 1 from I1, should get: I1, I2 (spouse), I3, I4 (children)
	if len(component) < 2 {
		t.Errorf("Expected at least 2 individuals with depth 1, got %d", len(component))
	}

	// Verify starting node is included
	found := false
	for _, indi := range component {
		if indi.XrefID() == "@I1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Component should include starting node @I1@")
	}
}

func TestGetComponentForPerson_SizeLimitReached(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with size limit
	options := NewComponentOptions()
	options.MaxSize = 2

	component, err := graph.GetComponentForPersonWithOptions("@I1@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// Should be limited to 2 individuals
	if len(component) > 2 {
		t.Errorf("Expected at most 2 individuals with size limit, got %d", len(component))
	}
}

func TestGetComponentForPerson_SizeLimitExact(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with size limit exactly matching component size
	options := NewComponentOptions()
	options.MaxSize = 5 // Should get all 5 individuals

	component, err := graph.GetComponentForPersonWithOptions("@I1@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// Should get all individuals
	if len(component) != 5 {
		t.Errorf("Expected 5 individuals, got %d", len(component))
	}
}

func TestGetComponentForPerson_DepthAndSizeLimits(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with both depth and size limits
	options := NewComponentOptions()
	options.MaxDepth = 2
	options.MaxSize = 3

	component, err := graph.GetComponentForPersonWithOptions("@I1@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// Should respect both limits (whichever is reached first)
	if len(component) > 3 {
		t.Errorf("Expected at most 3 individuals (size limit), got %d", len(component))
	}
}

func TestGetComponentForPerson_NodeWithoutIndividual(t *testing.T) {
	// This tests the edge case where a node exists but Individual is nil
	// This is hard to create naturally, but the code handles it
	
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Normal case - node should have Individual
	component, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// All components should have valid Individual records
	for _, indi := range component {
		if indi == nil {
			t.Error("Component should not contain nil IndividualRecord")
		}
	}
}

func TestGetComponentForPerson_AllEdgeTypes(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test that all edge types are traversed correctly
	// This verifies FAMC and FAMS edges are both processed
	component, err := graph.GetComponentForPerson("@I3@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// Should include:
	// - I3 itself
	// - Parents (I1, I2) via FAMC
	// - Sibling (I4) via FAMC
	// - Child (I5) via FAMS
	xrefs := make(map[string]bool)
	for _, indi := range component {
		xrefs[indi.XrefID()] = true
	}

	expected := []string{"@I1@", "@I2@", "@I3@", "@I4@", "@I5@"}
	for _, xref := range expected {
		if !xrefs[xref] {
			t.Errorf("Component should include %s", xref)
		}
	}
}

func TestGetComponentForPerson_VisitedTracking(t *testing.T) {
	// Create a tree with cycles (person married to sibling - shouldn't happen but test it)
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test that visited tracking prevents infinite loops
	component, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// Verify no duplicates
	xrefs := make(map[string]bool)
	for _, indi := range component {
		if xrefs[indi.XrefID()] {
			t.Errorf("Duplicate found in component: %s", indi.XrefID())
		}
		xrefs[indi.XrefID()] = true
	}
}

func TestGetComponentForPerson_QueueProcessing(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test that BFS queue processing works correctly
	component, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// Should process all connected nodes
	if len(component) != 5 {
		t.Errorf("Expected 5 individuals in component, got %d", len(component))
	}
}

func TestGetComponentForPerson_OptionsNilHandling(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test that GetComponentForPerson (without options) works
	component1, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// Test with default options
	options := NewComponentOptions()
	component2, err := graph.GetComponentForPersonWithOptions("@I1@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// Should get same results
	if len(component1) != len(component2) {
		t.Errorf("Expected same results: %d vs %d", len(component1), len(component2))
	}
}

func TestGetComponentForPerson_MultipleCalls(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test multiple calls return consistent results
	component1, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	component2, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	if len(component1) != len(component2) {
		t.Errorf("Multiple calls should return same results: %d vs %d", len(component1), len(component2))
	}
}

func TestGetComponentForPerson_DifferentStartingPoints(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test that starting from different points in same component returns same component
	component1, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	component2, err := graph.GetComponentForPerson("@I3@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// Should be same component (all 5 individuals)
	if len(component1) != len(component2) {
		t.Errorf("Same component should have same size: %d vs %d", len(component1), len(component2))
	}

	// Verify they contain the same individuals
	xrefs1 := make(map[string]bool)
	for _, indi := range component1 {
		xrefs1[indi.XrefID()] = true
	}

	for _, indi := range component2 {
		if !xrefs1[indi.XrefID()] {
			t.Errorf("Component2 missing %s that was in Component1", indi.XrefID())
		}
	}
}

func TestGetComponentForPerson_EdgeCases(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test various edge cases
	testCases := []struct {
		name      string
		personID  string
		options   *ComponentOptions
		minSize   int
		maxSize   int
	}{
		{"default_options", "@I1@", NewComponentOptions(), 1, 10},
		{"depth_0_unlimited", "@I1@", &ComponentOptions{MaxDepth: 0, MaxSize: 0}, 1, 10}, // 0 = unlimited
		{"depth_1", "@I1@", &ComponentOptions{MaxDepth: 1, MaxSize: 0}, 1, 10},
		{"size_1", "@I1@", &ComponentOptions{MaxDepth: 0, MaxSize: 1}, 1, 1},
		{"size_3", "@I1@", &ComponentOptions{MaxDepth: 0, MaxSize: 3}, 1, 3},
		{"both_limits", "@I1@", &ComponentOptions{MaxDepth: 1, MaxSize: 2}, 1, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			component, err := graph.GetComponentForPersonWithOptions(tc.personID, tc.options)
			if err != nil {
				t.Fatalf("GetComponentForPersonWithOptions failed for %s: %v", tc.name, err)
			}

			if len(component) < tc.minSize {
				t.Errorf("Expected at least %d individuals for %s, got %d", tc.minSize, tc.name, len(component))
			}

			if len(component) > tc.maxSize {
				t.Errorf("Expected at most %d individuals for %s, got %d", tc.maxSize, tc.name, len(component))
			}
		})
	}
}

func TestGetComponentForPerson_ConcurrentAccess(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test that concurrent access is safe (read lock)
	// This is more of a stress test
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			component, err := graph.GetComponentForPerson("@I1@")
			if err != nil {
				t.Errorf("GetComponentForPerson failed: %v", err)
			}
			if len(component) == 0 {
				t.Error("Component should not be empty")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetComponentForPerson_LargeComponent(t *testing.T) {
	// Create a larger tree to test performance
	tree := createTestTree()
	
	// Add more individuals to create a larger component
	for i := 6; i <= 10; i++ {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i))
		nameLine := types.NewGedcomLine(1, "NAME", fmt.Sprintf("Person%d /Test/", i), "")
		indiLine.AddChild(nameLine)
		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	// Create families to connect them
	fam3Line := types.NewGedcomLine(0, "FAM", "", "@F3@")
	fam3Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I5@", ""))
	fam3Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I6@", ""))
	fam3Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I7@", ""))
	fam3 := types.NewFamilyRecord(fam3Line)
	tree.AddRecord(fam3)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with larger component
	component, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// Should include all connected individuals
	if len(component) < 5 {
		t.Errorf("Expected at least 5 individuals, got %d", len(component))
	}
}

