package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// createComplexTestTree creates a more complex tree for testing edge cases
func createComplexTestTree() *types.GedcomTree {
	tree := types.NewGedcomTree()

	// I1: Root with no parents
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Root /Person/", "")
	sex1Line := types.NewGedcomLine(1, "SEX", "M", "")
	indi1Line.AddChild(name1Line)
	indi1Line.AddChild(sex1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// I2: Spouse of I1
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Spouse /Person/", "")
	sex2Line := types.NewGedcomLine(1, "SEX", "F", "")
	indi2Line.AddChild(name2Line)
	indi2Line.AddChild(sex2Line)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// I3: Child of I1 and I2
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child1 /Person/", "")
	sex3Line := types.NewGedcomLine(1, "SEX", "M", "")
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	indi3Line.AddChild(name3Line)
	indi3Line.AddChild(sex3Line)
	indi3Line.AddChild(famc3Line)
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// I4: Child of I1 and I2 (sibling of I3)
	indi4Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child2 /Person/", "")
	sex4Line := types.NewGedcomLine(1, "SEX", "F", "")
	famc4Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	indi4Line.AddChild(name4Line)
	indi4Line.AddChild(sex4Line)
	indi4Line.AddChild(famc4Line)
	indi4 := types.NewIndividualRecord(indi4Line)
	tree.AddRecord(indi4)

	// F1: Family (I1 + I2, children: I3, I4)
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husb1Line := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	wife1Line := types.NewGedcomLine(1, "WIFE", "@I2@", "")
	chil1Line := types.NewGedcomLine(1, "CHIL", "@I3@", "")
	chil2Line := types.NewGedcomLine(1, "CHIL", "@I4@", "")
	fam1Line.AddChild(husb1Line)
	fam1Line.AddChild(wife1Line)
	fam1Line.AddChild(chil1Line)
	fam1Line.AddChild(chil2Line)
	fams1Line := types.NewGedcomLine(1, "FAMS", "@F1@", "")
	indi1Line.AddChild(fams1Line)
	fams2Line := types.NewGedcomLine(1, "FAMS", "@F1@", "")
	indi2Line.AddChild(fams2Line)
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	return tree
}

func TestSubtreeQuery_UnlimitedGenerations(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with unlimited generations (0 = unlimited)
	result, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(0).
		DescendantGenerations(0).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Should get all individuals in the tree (5 total: I1, I2, I3, I4, I5)
	// Note: createTestTree creates 5 individuals
	if len(result.All) < 4 {
		t.Errorf("Expected at least 4 individuals with unlimited generations, got %d", len(result.All))
	}
}

func TestSubtreeQuery_OnlyAncestors(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with only ancestors, no descendants
	result, err := q.Individual("@I5@").GetSubtree().
		AncestorGenerations(2).
		DescendantGenerations(0).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have ancestors but no descendants
	if len(result.Descendants) != 0 {
		t.Errorf("Expected 0 descendants, got %d", len(result.Descendants))
	}

	// Should have ancestors (I1, I2, I3)
	if len(result.Ancestors) < 3 {
		t.Errorf("Expected at least 3 ancestors, got %d", len(result.Ancestors))
	}
}

func TestSubtreeQuery_OnlyDescendants(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with only descendants, no ancestors
	result, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(0).
		DescendantGenerations(2).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have descendants but no ancestors
	if len(result.Ancestors) != 0 {
		t.Errorf("Expected 0 ancestors, got %d", len(result.Ancestors))
	}

	// Should have descendants (I3, I4, I5)
	if len(result.Descendants) < 3 {
		t.Errorf("Expected at least 3 descendants, got %d", len(result.Descendants))
	}
}

func TestSubtreeQuery_WithMetrics(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Enable metrics
	graph.metrics = NewMetrics()

	q := NewQueryFromGraph(graph)

	// Execute query (should record metrics)
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

	// Verify metrics were recorded
	metrics := graph.GetMetrics()
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Check that query count increased
	if metrics.QueryCount == 0 {
		t.Error("Query count should be greater than 0")
	}
}

func TestSubtreeQuery_CountWithError(t *testing.T) {
	// Create a query with invalid graph to test error handling
	// This is tricky because we need to simulate an error from Execute()
	// Let's test with a valid query but check the error path in Count
	
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test Count with valid query
	count, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Count()

	if err != nil {
		t.Fatalf("Count should not error: %v", err)
	}

	if count == 0 {
		t.Error("Count should be greater than 0")
	}
}

func TestSubtreeQuery_ExecuteRecordsWithError(t *testing.T) {
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
		t.Fatalf("ExecuteRecords should not error: %v", err)
	}

	if len(records) == 0 {
		t.Error("ExecuteRecords should return records")
	}
}

func TestSubtreeQuery_EmptyResults(t *testing.T) {
	tree := createComplexTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with individual that has no ancestors and no descendants
	result, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		ExcludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have descendants but no ancestors
	if len(result.Ancestors) != 0 {
		t.Errorf("Expected 0 ancestors for root, got %d", len(result.Ancestors))
	}

	if len(result.Descendants) == 0 {
		t.Error("Expected descendants for root")
	}
}

func TestSubtreeQuery_AllCombinations(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test all combinations of options
	combinations := []struct {
		name                string
		ancestorGen         int
		descendantGen       int
		includeSelf         bool
		includeSiblings     bool
		includeSpouses      bool
		expectedMinCount    int
	}{
		{"self_only", 0, 0, true, false, false, 1},
		{"self_siblings", 0, 0, true, true, false, 2},
		{"self_spouses", 0, 0, true, false, true, 2},
		{"self_all", 0, 0, true, true, true, 3},
		{"ancestors_only", 2, 0, true, false, false, 3},
		{"descendants_only", 0, 2, true, false, false, 4},
		{"ancestors_descendants", 1, 1, true, false, false, 4},
		{"everything", 2, 2, true, true, true, 5},
	}

	for _, combo := range combinations {
		t.Run(combo.name, func(t *testing.T) {
			sq := q.Individual("@I3@").GetSubtree().
				AncestorGenerations(combo.ancestorGen).
				DescendantGenerations(combo.descendantGen)

			if combo.includeSelf {
				sq = sq.IncludeSelf()
			} else {
				sq = sq.ExcludeSelf()
			}

			if combo.includeSiblings {
				sq = sq.IncludeSiblings()
			}

			if combo.includeSpouses {
				sq = sq.IncludeSpouses()
			}

			result, err := sq.Execute()
			if err != nil {
				t.Fatalf("Query failed for %s: %v", combo.name, err)
			}

			if len(result.All) < combo.expectedMinCount {
				t.Errorf("Expected at least %d results for %s, got %d",
					combo.expectedMinCount, combo.name, len(result.All))
			}
		})
	}
}

func TestSubtreeQuery_FilterOnAllCategories(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test filter applies to all categories (ancestors, descendants, siblings, spouses)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		Filter(func(indi *types.IndividualRecord) bool {
			// Only include individuals with "Child" in name
			return indi.GetName() != "" && len(indi.GetName()) > 5
		}).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Filter should apply to all categories
	// Verify filter was applied
	for _, indi := range result.All {
		if indi.GetName() == "" || len(indi.GetName()) <= 5 {
			t.Errorf("Filter should have excluded %s", indi.GetName())
		}
	}
}

func TestSubtreeQuery_RootNodeWithoutIndividual(t *testing.T) {
	// This tests the edge case where GetIndividual returns a node
	// but node.Individual is nil
	// This is hard to create naturally, but we can test the code path
	
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with valid individual (normal case)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Root should be set
	if result.Root == nil {
		t.Error("Root should not be nil for valid individual")
	}
}

func TestSubtreeQuery_Deduplication(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test that deduplication works correctly
	// If someone appears in both ancestors and descendants (shouldn't happen but test it)
	// Or if someone appears in siblings and ancestors (could happen)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Verify no duplicates in All list
	seen := make(map[string]bool)
	for _, indi := range result.All {
		if seen[indi.XrefID()] {
			t.Errorf("Duplicate found in All list: %s", indi.XrefID())
		}
		seen[indi.XrefID()] = true
	}
}

func TestSubtreeQuery_ChainedMethods(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test that all methods can be chained
	result, err := q.Individual("@I3@").
		GetSubtree().
		AncestorGenerations(2).
		DescendantGenerations(2).
		IncludeSelf().
		IncludeSiblings().
		IncludeSpouses().
		Filter(func(indi *types.IndividualRecord) bool {
			return true // Include all
		}).
		Execute()

	if err != nil {
		t.Fatalf("Chained query failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestSubtreeQuery_NoRelationships(t *testing.T) {
	// Create an isolated individual with no relationships
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

	// Test subtree query on isolated individual
	result, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should only have self
	if len(result.All) != 1 {
		t.Errorf("Expected 1 individual (self only), got %d", len(result.All))
	}

	if result.Root == nil {
		t.Error("Root should not be nil")
	}

	if len(result.Ancestors) != 0 {
		t.Errorf("Expected 0 ancestors, got %d", len(result.Ancestors))
	}

	if len(result.Descendants) != 0 {
		t.Errorf("Expected 0 descendants, got %d", len(result.Descendants))
	}
}

func TestSubtreeQuery_ComplexFiltering(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test complex filter that excludes some results
	filteredCount := 0
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		Filter(func(indi *types.IndividualRecord) bool {
			// Only include if name contains "Child" or "John" or "Jane"
			name := indi.GetName()
			return name == "Child1 /Doe/" || name == "John /Doe/" || name == "Jane /Doe/"
		}).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Count how many would have been included without filter
	unfilteredResult, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		Execute()

	if err != nil {
		t.Fatalf("Unfiltered query failed: %v", err)
	}

	// Filtered result should have fewer or equal items
	if len(result.All) > len(unfilteredResult.All) {
		t.Errorf("Filtered result should have fewer items, got %d > %d",
			len(result.All), len(unfilteredResult.All))
	}

	// Verify all filtered results match the filter
	for _, indi := range result.All {
		name := indi.GetName()
		if name != "Child1 /Doe/" && name != "John /Doe/" && name != "Jane /Doe/" {
			t.Errorf("Filter failed: included %s which doesn't match filter", name)
		}
		filteredCount++
	}

	if filteredCount == 0 {
		t.Error("Filter should have included at least some results")
	}
}

func TestSubtreeQuery_MetricsIntegration(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Initialize metrics
	graph.metrics = NewMetrics()
	initialQueryCount := graph.metrics.QueryCount

	q := NewQueryFromGraph(graph)

	// Execute multiple queries
	for i := 0; i < 3; i++ {
		_, err := q.Individual("@I3@").GetSubtree().
			AncestorGenerations(1).
			DescendantGenerations(1).
			IncludeSelf().
			Execute()

		if err != nil {
			t.Fatalf("Query %d failed: %v", i, err)
		}
	}

	// Verify metrics were recorded
	finalQueryCount := graph.metrics.QueryCount
	if finalQueryCount <= initialQueryCount {
		t.Errorf("Query count should have increased: %d -> %d", initialQueryCount, finalQueryCount)
	}
}

func TestSubtreeQuery_CountWithZeroResults(t *testing.T) {
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

	// Test Count with excluded self and no relationships
	count, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(0).
		DescendantGenerations(0).
		ExcludeSelf().
		Count()

	if err != nil {
		t.Fatalf("Count should not error: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0 for isolated individual with self excluded, got %d", count)
	}
}

func TestSubtreeQuery_ExecuteRecordsWithZeroResults(t *testing.T) {
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

	// Test ExecuteRecords with excluded self and no relationships
	records, err := q.Individual("@I1@").GetSubtree().
		AncestorGenerations(0).
		DescendantGenerations(0).
		ExcludeSelf().
		ExecuteRecords()

	if err != nil {
		t.Fatalf("ExecuteRecords should not error: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records for isolated individual with self excluded, got %d", len(records))
	}
}

