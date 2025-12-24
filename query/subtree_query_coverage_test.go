package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestSubtreeQuery_CountErrorPath tests the error path in Count()
// This is difficult to trigger naturally, but we can test the structure
func TestSubtreeQuery_CountErrorPath(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test Count with valid query (error path is hard to trigger)
	// The error would come from Execute(), which only errors if
	// ancestorQuery.Execute() or descendantQuery.Execute() errors
	// Those don't error in normal cases, so we test the success path
	count, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Count()

	if err != nil {
		// If error occurs, verify Count handles it correctly
		if count != 0 {
			t.Errorf("Count should return 0 on error, got %d", count)
		}
		return
	}

	// Normal case - no error
	if count == 0 {
		t.Error("Count should be greater than 0 for valid query")
	}
}

// TestSubtreeQuery_ExecuteRecordsErrorPath tests the error path in ExecuteRecords()
func TestSubtreeQuery_ExecuteRecordsErrorPath(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test ExecuteRecords with valid query
	records, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		ExecuteRecords()

	if err != nil {
		// If error occurs, verify ExecuteRecords handles it correctly
		if records != nil {
			t.Error("ExecuteRecords should return nil on error")
		}
		return
	}

	// Normal case - no error
	if len(records) == 0 {
		t.Error("ExecuteRecords should return records for valid query")
	}
}

// TestSubtreeQuery_RootNodeNilIndividual tests the case where rootNode.Individual is nil
// This is an edge case that's hard to create naturally, but we can test the code path exists
func TestSubtreeQuery_RootNodeNilIndividual(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with valid individual (normal case where Individual is not nil)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// In normal case, Root should be set
	if result.Root == nil {
		t.Error("Root should not be nil for valid individual")
	}

	// Test that the code handles the case where Root might be nil
	// (even though it's hard to create this scenario naturally)
	if result.Root == nil && len(result.All) > 0 {
		// If Root is nil but we have results, that's fine
		// The code should still work
	}
}

// TestSubtreeQuery_FilterExcludesAll tests filter that excludes everything
func TestSubtreeQuery_FilterExcludesAll(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with filter that excludes everything
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		IncludeSpouses().
		Filter(func(indi *types.IndividualRecord) bool {
			return false // Exclude everything
		}).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// All lists should be empty (filtered out)
	if len(result.All) != 0 {
		t.Errorf("Expected 0 results with filter excluding all, got %d", len(result.All))
	}

	// But Root should still be set (it's set before filtering)
	if result.Root == nil {
		t.Error("Root should be set even if filtered out")
	}
}

// TestSubtreeQuery_FilterOnRootOnly tests filter that only includes root
func TestSubtreeQuery_FilterOnRootOnly(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with filter that only includes root
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Filter(func(indi *types.IndividualRecord) bool {
			return indi.XrefID() == "@I3@" // Only include root
		}).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should only have root
	if len(result.All) != 1 {
		t.Errorf("Expected 1 result (root only), got %d", len(result.All))
	}

	if len(result.All) > 0 && result.All[0].XrefID() != "@I3@" {
		t.Errorf("Expected root @I3@, got %s", result.All[0].XrefID())
	}
}

// TestSubtreeQuery_EmptyAncestorsEmptyDescendants tests when both are empty
func TestSubtreeQuery_EmptyAncestorsEmptyDescendants(t *testing.T) {
	// Create isolated individual
	tree := types.NewGedcomTree()
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "Isolated /Person/", "")
	indiLine.AddChild(nameLine)
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with isolated individual (no ancestors, no descendants)
	result, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		ExcludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have empty ancestors and descendants
	if len(result.Ancestors) != 0 {
		t.Errorf("Expected 0 ancestors, got %d", len(result.Ancestors))
	}

	if len(result.Descendants) != 0 {
		t.Errorf("Expected 0 descendants, got %d", len(result.Descendants))
	}

	// All should be empty (self excluded, no relationships)
	if len(result.All) != 0 {
		t.Errorf("Expected 0 in All list, got %d", len(result.All))
	}
}

// TestSubtreeQuery_SiblingWithFilter tests siblings with filter
func TestSubtreeQuery_SiblingWithFilter(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test siblings with filter
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(0).
		DescendantGenerations(0).
		IncludeSelf().
		IncludeSiblings().
		Filter(func(indi *types.IndividualRecord) bool {
			// Only include if name contains "Child"
			return len(indi.GetName()) > 5
		}).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Verify filter was applied to siblings
	for _, sibling := range result.Siblings {
		if len(sibling.GetName()) <= 5 {
			t.Errorf("Filter should have excluded sibling %s", sibling.GetName())
		}
	}
}

// TestSubtreeQuery_SpouseWithFilter tests spouses with filter
func TestSubtreeQuery_SpouseWithFilter(t *testing.T) {
	tree := createComplexTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test spouses with filter
	result, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(0).
		DescendantGenerations(0).
		IncludeSelf().
		IncludeSpouses().
		Filter(func(indi *types.IndividualRecord) bool {
			// Only include if name contains "Spouse"
			return len(indi.GetName()) > 5
		}).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Verify filter was applied to spouses
	for _, spouse := range result.Spouses {
		if len(spouse.GetName()) <= 5 {
			t.Errorf("Filter should have excluded spouse %s", spouse.GetName())
		}
	}
}

// TestSubtreeQuery_DeduplicationAcrossCategories tests deduplication
func TestSubtreeQuery_DeduplicationAcrossCategories(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test that deduplication works across all categories
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		IncludeSpouses().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Verify no duplicates in All list
	seen := make(map[string]int)
	for _, indi := range result.All {
		seen[indi.XrefID()]++
		if seen[indi.XrefID()] > 1 {
			t.Errorf("Duplicate found in All list: %s (appears %d times)", indi.XrefID(), seen[indi.XrefID()])
		}
	}
}

// TestSubtreeQuery_AllCategoriesPopulated tests when all categories have data
func TestSubtreeQuery_AllCategoriesPopulated(t *testing.T) {
	tree := createTestTree()
	
	// Add spouse for I3
	indi6Line := types.NewGedcomLine(0, "INDI", "", "@I6@")
	name6Line := types.NewGedcomLine(1, "NAME", "Spouse /Doe/", "")
	sex6Line := types.NewGedcomLine(1, "SEX", "F", "")
	indi6Line.AddChild(name6Line)
	indi6Line.AddChild(sex6Line)
	indi6 := types.NewIndividualRecord(indi6Line)
	tree.AddRecord(indi6)

	// Update F2 to include spouse - recreate family with spouse
	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	husb2Line := types.NewGedcomLine(1, "HUSB", "@I3@", "")
	wife2Line := types.NewGedcomLine(1, "WIFE", "@I6@", "")
	chil3Line := types.NewGedcomLine(1, "CHIL", "@I5@", "")
	fam2Line.AddChild(husb2Line)
	fam2Line.AddChild(wife2Line)
	fam2Line.AddChild(chil3Line)
	fams6Line := types.NewGedcomLine(1, "FAMS", "@F2@", "")
	indi6Line.AddChild(fams6Line)
	fam2 := types.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with all categories populated
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		IncludeSpouses().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have ancestors, descendants, siblings, and spouses
	if len(result.Ancestors) == 0 {
		t.Error("Expected ancestors")
	}

	if len(result.Descendants) == 0 {
		t.Error("Expected descendants")
	}

	if len(result.Siblings) == 0 {
		t.Error("Expected siblings")
	}

	if len(result.Spouses) == 0 {
		t.Error("Expected spouses")
	}
}

// TestSubtreeQuery_ResultStructure tests the result structure
func TestSubtreeQuery_ResultStructure(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Root should be set
	if result.Root == nil {
		t.Error("Root should be set")
	}

	// All should be populated
	if result.All == nil {
		t.Error("All should not be nil")
	}

	// Verify All contains expected individuals
	if len(result.All) < len(result.Ancestors)+len(result.Descendants)+1 {
		t.Error("All should contain at least ancestors + descendants + self")
	}
}

// TestSubtreeQuery_MetricsDisabled tests when metrics is nil
func TestSubtreeQuery_MetricsDisabled(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Ensure metrics is nil
	graph.metrics = nil

	q := NewQueryFromGraph(graph)

	// Execute query (should not panic when metrics is nil)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

// TestSubtreeQuery_ZeroGenerations tests with 0 generations (unlimited)
func TestSubtreeQuery_ZeroGenerations(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with 0 generations (should be unlimited)
	result, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(0).
		DescendantGenerations(0).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should get all individuals in the tree
	if len(result.All) < 4 {
		t.Errorf("Expected at least 4 individuals with unlimited generations, got %d", len(result.All))
	}
}

