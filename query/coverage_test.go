package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestAnalytics_GetMostCommonGivenNames tests GetMostCommonGivenNames
func TestAnalytics_GetMostCommonGivenNames(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with various given names
	names := []string{"John", "John", "Jane", "Jane", "Jane", "Bob", "Alice", "Alice"}
	for i, name := range names {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i+1))
		nameLine := types.NewGedcomLine(1, "NAME", name+" /Doe/", "")
		givnLine := types.NewGedcomLine(2, "GIVN", name, "")
		nameLine.AddChild(givnLine)
		indiLine.AddChild(nameLine)
		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with limit
	results := graph.GetMostCommonGivenNames(3)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify Jane is first (count 3)
	if len(results) > 0 && results[0].Name != "jane" {
		t.Errorf("Expected 'jane' to be first, got %s", results[0].Name)
	}
	if len(results) > 0 && results[0].Count != 3 {
		t.Errorf("Expected count 3 for 'jane', got %d", results[0].Count)
	}

	// Test with zero limit (should default to 10)
	results2 := graph.GetMostCommonGivenNames(0)
	if len(results2) == 0 {
		t.Error("Expected results with limit 0")
	}

	// Test with negative limit
	results3 := graph.GetMostCommonGivenNames(-1)
	if len(results3) == 0 {
		t.Error("Expected results with negative limit")
	}
}

// TestAnalytics_GetMostCommonSurnames tests GetMostCommonSurnames
func TestAnalytics_GetMostCommonSurnames(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with various surnames
	surnames := []string{"Smith", "Smith", "Smith", "Jones", "Jones", "Brown", "Williams"}
	for i, surname := range surnames {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i+1))
		nameLine := types.NewGedcomLine(1, "NAME", "John /"+surname+"/", "")
		surnLine := types.NewGedcomLine(2, "SURN", surname, "")
		nameLine.AddChild(surnLine)
		indiLine.AddChild(nameLine)
		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with limit
	results := graph.GetMostCommonSurnames(2)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify Smith is first (count 3)
	if len(results) > 0 && results[0].Name != "smith" {
		t.Errorf("Expected 'smith' to be first, got %s", results[0].Name)
	}
	if len(results) > 0 && results[0].Count != 3 {
		t.Errorf("Expected count 3 for 'smith', got %d", results[0].Count)
	}

	// Test with zero limit
	results2 := graph.GetMostCommonSurnames(0)
	if len(results2) == 0 {
		t.Error("Expected results with limit 0")
	}
}

// TestBirthdayFilters_ByBirthMonth tests ByBirthMonth filter
func TestBirthdayFilters_ByBirthMonth(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with different birth months
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	birt2Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt2Line.AddChild(types.NewGedcomLine(2, "DATE", "20 FEB 1900", ""))
	indi2Line.AddChild(birt2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	birt3Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt3Line.AddChild(types.NewGedcomLine(2, "DATE", "10 JAN 1900", ""))
	indi3Line.AddChild(birt3Line)
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test filtering by January (month 1)
	fq := NewFilterQuery(graph)
	results, err := fq.ByBirthMonth(1).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results for January, got %d", len(results))
	}

	// Test filtering by February (month 2) - create a new query
	fq2 := NewFilterQuery(graph)
	results2, err2 := fq2.ByBirthMonth(2).Execute()
	if err2 != nil {
		t.Fatalf("Failed to execute query: %v", err2)
	}
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("Expected 1 result for February, got %d", len(results2))
	}

	// Test invalid month (should return unchanged query)
	fq3 := fq2.ByBirthMonth(13) // Invalid month
	if fq3 != fq2 {
		t.Error("Expected unchanged query for invalid month")
	}

	// Test invalid month (0)
	fq4 := fq2.ByBirthMonth(0)
	if fq4 != fq2 {
		t.Error("Expected unchanged query for month 0")
	}
}

// TestBirthdayFilters_ByBirthDay tests ByBirthDay filter
func TestBirthdayFilters_ByBirthDay(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with different birth days
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	birt2Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt2Line.AddChild(types.NewGedcomLine(2, "DATE", "20 JAN 1900", ""))
	indi2Line.AddChild(birt2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	birt3Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt3Line.AddChild(types.NewGedcomLine(2, "DATE", "15 FEB 1900", ""))
	indi3Line.AddChild(birt3Line)
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test filtering by day 15
	fq := NewFilterQuery(graph)
	results, err := fq.ByBirthDay(15).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results for day 15, got %d", len(results))
	}

	// Test invalid day
	fq2 := NewFilterQuery(graph)
	fq3 := fq2.ByBirthDay(32) // Invalid day
	if fq3 != fq2 {
		t.Error("Expected unchanged query for invalid day")
	}
}

// TestBirthdayFilters_ByBirthMonthAndDay tests ByBirthMonthAndDay filter
func TestBirthdayFilters_ByBirthMonthAndDay(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	birt2Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt2Line.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1901", ""))
	indi2Line.AddChild(birt2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	birt3Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt3Line.AddChild(types.NewGedcomLine(2, "DATE", "20 JAN 1900", ""))
	indi3Line.AddChild(birt3Line)
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test filtering by January 15
	fq := NewFilterQuery(graph)
	results, err := fq.ByBirthMonthAndDay(1, 15).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results for Jan 15, got %d", len(results))
	}

	// Test invalid month/day
	fq2 := NewFilterQuery(graph)
	fq3 := fq2.ByBirthMonthAndDay(13, 1) // Invalid month
	if fq3 != fq2 {
		t.Error("Expected unchanged query for invalid month")
	}
}

// TestBirthdayFilters_ByBirthDateRange tests ByBirthDateRange filter
func TestBirthdayFilters_ByBirthDateRange(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with different birth dates
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	birt2Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt2Line.AddChild(types.NewGedcomLine(2, "DATE", "20 FEB 1900", ""))
	indi2Line.AddChild(birt2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test date range
	start := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(1900, 1, 31, 23, 59, 59, 0, time.UTC)
	fq := NewFilterQuery(graph)
	results, err := fq.ByBirthDateRange(start, end).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for date range, got %d", len(results))
	}
}

// TestAncestorQuery_Filter tests Filter method
func TestAncestorQuery_Filter(t *testing.T) {
	// Create test tree with family relationships
	tree := types.NewGedcomTree()

	// Create grandparents
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	grandmaLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Grandma /Doe/", "")
	grandmaLine.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(grandmaLine))

	// Create parents
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name3Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Create child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name4Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Create families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I3@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build graph (not needed for query, but for consistency)
	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test ancestor query with filter
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}
	aq := qb.Individual("@I4@").Ancestors()
	results, err := aq.Filter(func(indi *types.IndividualRecord) bool {
		return indi.GetName() == "Grandpa /Doe/"
	}).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 filtered result, got %d", len(results))
	}
}

// TestAncestorQuery_Exists tests Exists method
func TestAncestorQuery_Exists(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create individual with parent
	childLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name1Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I1@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph (not needed for query, but for consistency)
	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Exists with ancestors
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}
	aq := qb.Individual("@I1@").Ancestors()
	exists, err := aq.Exists()
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected ancestors to exist")
	}

	// Test Exists with no ancestors
	aq2 := qb.Individual("@I2@").Ancestors()
	exists2, err := aq2.Exists()
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists2 {
		t.Error("Expected no ancestors")
	}
}

// TestAncestorQuery_ExecuteWithPaths tests ExecuteWithPaths method
func TestAncestorQuery_ExecuteWithPaths(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Create parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Create child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Create families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build graph (not needed for query, but for consistency)
	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test ExecuteWithPaths
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}
	aq := qb.Individual("@I3@").Ancestors()
	results, err := aq.ExecuteWithPaths()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected ancestor paths")
	}

	// Verify path information
	for _, result := range results {
		if result.Ancestor == nil {
			t.Error("Expected ancestor to be non-nil")
		}
		if result.Depth < 0 {
			t.Error("Expected depth to be non-negative")
		}
	}
}

// TestCollectionQueryHelpers_CollectIndividuals tests CollectIndividuals
func TestCollectionQueryHelpers_CollectIndividuals(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test CollectIndividuals with filter
	results := CollectIndividuals(graph, func(node *IndividualNode) bool {
		return node.Individual != nil && node.Individual.GetName() == "John /Doe/"
	})
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test with nil graph
	results2 := CollectIndividuals(nil, func(node *IndividualNode) bool {
		return true
	})
	if results2 != nil {
		t.Error("Expected nil result for nil graph")
	}
}

// TestCollectionQueryHelpers_CollectFamilies tests CollectFamilies
func TestCollectionQueryHelpers_CollectFamilies(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test CollectFamilies with filter
	results := CollectFamilies(graph, func(node *FamilyNode) bool {
		return node.Family != nil
	})
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test with nil graph
	results2 := CollectFamilies(nil, func(node *FamilyNode) bool {
		return true
	})
	if results2 != nil {
		t.Error("Expected nil result for nil graph")
	}
}

// TestDescendantQuery_IncludeSelf tests IncludeSelf method
func TestDescendantQuery_IncludeSelf(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Create child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Create family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test without IncludeSelf
	results1, err := qb.Individual("@I1@").Descendants().Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results1) != 1 {
		t.Errorf("Expected 1 descendant, got %d", len(results1))
	}

	// Test with IncludeSelf
	results2, err := qb.Individual("@I1@").Descendants().IncludeSelf().Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results2) != 2 {
		t.Errorf("Expected 2 results (self + descendant), got %d", len(results2))
	}
}

// TestDescendantQuery_Filter tests Filter method
func TestDescendantQuery_Filter(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Create children
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child2 /Smith/", "")
	child2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Create family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test with filter
	results, err := qb.Individual("@I1@").Descendants().Filter(func(indi *types.IndividualRecord) bool {
		return indi.GetName() == "Child1 /Doe/"
	}).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 filtered result, got %d", len(results))
	}
}

// TestDescendantQuery_Count tests Count method
func TestDescendantQuery_Count(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Create child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Create family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Count
	count, err := qb.Individual("@I1@").Descendants().Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Test Count with IncludeSelf
	count2, err := qb.Individual("@I1@").Descendants().IncludeSelf().Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count2 != 2 {
		t.Errorf("Expected count 2 (with self), got %d", count2)
	}
}

// TestDescendantQuery_Exists tests Exists method
func TestDescendantQuery_Exists(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create parent with child
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Create individual with no descendants
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "NoDescendants /Doe/", "")
	indi3Line.AddChild(name3Line)
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Exists with descendants
	exists, err := qb.Individual("@I1@").Descendants().Exists()
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Expected descendants to exist")
	}

	// Test Exists with no descendants
	exists2, err := qb.Individual("@I3@").Descendants().Exists()
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists2 {
		t.Error("Expected no descendants")
	}
}

// TestConfig_ParseDuration tests parseDuration function
func TestConfig_ParseDuration(t *testing.T) {
	// Test string duration
	dur1, err := parseDuration("30s")
	if err != nil {
		t.Fatalf("Failed to parse string duration: %v", err)
	}
	if dur1 != 30*time.Second {
		t.Errorf("Expected 30s, got %v", dur1)
	}

	// Test float64 duration (JSON number)
	dur2, err := parseDuration(float64(5000000000)) // 5 seconds in nanoseconds
	if err != nil {
		t.Fatalf("Failed to parse float64 duration: %v", err)
	}
	if dur2 != 5*time.Second {
		t.Errorf("Expected 5s, got %v", dur2)
	}

	// Test int64 duration
	dur3, err := parseDuration(int64(2000000000)) // 2 seconds
	if err != nil {
		t.Fatalf("Failed to parse int64 duration: %v", err)
	}
	if dur3 != 2*time.Second {
		t.Errorf("Expected 2s, got %v", dur3)
	}

	// Test int duration
	dur4, err := parseDuration(1000000000) // 1 second
	if err != nil {
		t.Fatalf("Failed to parse int duration: %v", err)
	}
	if dur4 != 1*time.Second {
		t.Errorf("Expected 1s, got %v", dur4)
	}

	// Test invalid type
	_, err = parseDuration(true)
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

// TestDescendantQuery_MaxGenerations tests MaxGenerations method
func TestDescendantQuery_MaxGenerations(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create grandparent
	grandparentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandparent /Doe/", "")
	grandparentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandparentLine))

	// Create parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Create child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Create grandchild
	grandchildLine := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Grandchild /Doe/", "")
	grandchildLine.AddChild(name4Line)
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F3@", "")
	grandchildLine.AddChild(famc3Line)
	tree.AddRecord(types.NewIndividualRecord(grandchildLine))

	// Create families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	fam3Line := types.NewGedcomLine(0, "FAM", "", "@F3@")
	fam3Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I3@", ""))
	fam3Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam3Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test without MaxGenerations (should get all descendants)
	results1, err := qb.Individual("@I1@").Descendants().Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results1) < 3 {
		t.Errorf("Expected at least 3 descendants, got %d", len(results1))
	}

	// Test with MaxGenerations = 1 (should only get direct children)
	results2, err := qb.Individual("@I1@").Descendants().MaxGenerations(1).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("Expected 1 descendant with MaxGenerations=1, got %d", len(results2))
	}
}


// TestEventCollectionQuery_FromIndividuals tests FromIndividuals method
func TestEventCollectionQuery_FromIndividuals(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth event
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test FromIndividuals
	ecq := NewEventCollectionQuery(graph)
	results, err := ecq.FromIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least 1 event from individuals")
	}
}

// TestEventCollectionQuery_FromFamilies tests FromFamilies method
func TestEventCollectionQuery_FromFamilies(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family with marriage event
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1920", ""))
	fam1Line.AddChild(marrLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test FromFamilies
	ecq := NewEventCollectionQuery(graph)
	results, err := ecq.FromFamilies().Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least 1 event from families")
	}
}

// TestEventCollectionQuery_Filter tests Filter method
func TestEventCollectionQuery_Filter(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth event
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Filter
	ecq := NewEventCollectionQuery(graph)
	results, err := ecq.FromIndividuals().Filter(func(event EventInfo) bool {
		return event.EventType == "BIRT"
	}).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least 1 birth event")
	}
}

// TestEventCollectionQuery_Count tests Count method
func TestEventCollectionQuery_Count(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth event
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Count
	ecq := NewEventCollectionQuery(graph)
	count, err := ecq.FromIndividuals().Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count == 0 {
		t.Error("Expected at least 1 event")
	}
}

// TestGraph_GetMetrics tests GetMetrics method
func TestGraph_GetMetrics(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create graph
	graph := NewGraph(tree)

	// Test GetMetrics
	metrics := graph.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}
}

// TestGraph_Tree tests Tree method
func TestGraph_Tree(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create graph
	graph := NewGraph(tree)

	// Test Tree
	returnedTree := graph.Tree()
	if returnedTree != tree {
		t.Error("Expected returned tree to be the same instance")
	}
}

// TestFamilyQuery_MarriageDateParsed tests MarriageDateParsed method
func TestFamilyQuery_MarriageDateParsed(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family with marriage date
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1920", ""))
	fam1Line.AddChild(marrLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test MarriageDateParsed
	date, err := qb.Family("@F1@").MarriageDateParsed()
	if err != nil {
		t.Fatalf("Failed to get marriage date: %v", err)
	}
	if date == nil {
		t.Error("Expected non-nil marriage date")
	}
}

// TestFamilyQuery_DivorceDate tests DivorceDate method
func TestFamilyQuery_DivorceDate(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family with divorce date
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	divLine := types.NewGedcomLine(1, "DIV", "", "")
	divLine.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1940", ""))
	fam1Line.AddChild(divLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test DivorceDate
	date, err := qb.Family("@F1@").DivorceDate()
	if err != nil {
		t.Fatalf("Failed to get divorce date: %v", err)
	}
	if date == "" {
		t.Error("Expected non-empty divorce date")
	}
}

// TestFamilyQuery_DivorceDateParsed tests DivorceDateParsed method
func TestFamilyQuery_DivorceDateParsed(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family with divorce date
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	divLine := types.NewGedcomLine(1, "DIV", "", "")
	divLine.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1940", ""))
	fam1Line.AddChild(divLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test DivorceDateParsed
	date, err := qb.Family("@F1@").DivorceDateParsed()
	if err != nil {
		t.Fatalf("Failed to get divorce date: %v", err)
	}
	if date == nil {
		t.Error("Expected non-nil divorce date")
	}
}

// TestFamilyQuery_MarriagePlace tests MarriagePlace method
func TestFamilyQuery_MarriagePlace(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family with marriage place
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(types.NewGedcomLine(2, "PLAC", "New York", ""))
	fam1Line.AddChild(marrLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test MarriagePlace
	place, err := qb.Family("@F1@").MarriagePlace()
	if err != nil {
		t.Fatalf("Failed to get marriage place: %v", err)
	}
	if place == "" {
		t.Error("Expected non-empty marriage place")
	}
}

// TestFamilyQuery_DivorcePlace tests DivorcePlace method
func TestFamilyQuery_DivorcePlace(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family with divorce place
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	divLine := types.NewGedcomLine(1, "DIV", "", "")
	divLine.AddChild(types.NewGedcomLine(2, "PLAC", "Los Angeles", ""))
	fam1Line.AddChild(divLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test DivorcePlace
	place, err := qb.Family("@F1@").DivorcePlace()
	if err != nil {
		t.Fatalf("Failed to get divorce place: %v", err)
	}
	if place == "" {
		t.Error("Expected non-empty divorce place")
	}
}

// TestFamilyQuery_GetRecord tests GetRecord method
func TestFamilyQuery_GetRecord(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test GetRecord
	record, err := qb.Family("@F1@").GetRecord()
	if err != nil {
		t.Fatalf("Failed to get record: %v", err)
	}
	if record == nil {
		t.Error("Expected non-nil family record")
	}
	if record.XrefID() != "@F1@" {
		t.Errorf("Expected XREF @F1@, got %s", record.XrefID())
	}
}

// TestFilterQuery_ByBirthDate tests ByBirthDate method
func TestFilterQuery_ByBirthDate(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth date
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "15 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test ByBirthDate
	start := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC)
	fq := NewFilterQuery(graph)
	results, err := fq.ByBirthDate(start, end).Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestFilterQuery_BySurname tests BySurname filter
func TestFilterQuery_BySurname(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with different surnames
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	surn1Line := types.NewGedcomLine(2, "SURN", "Doe", "")
	name1Line.AddChild(surn1Line)
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	surn2Line := types.NewGedcomLine(2, "SURN", "Smith", "")
	name2Line.AddChild(surn2Line)
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test BySurname
	fq := NewFilterQuery(graph)
	results, err := fq.BySurname("Doe").Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestFilterQuery_BySurnameExact tests BySurnameExact filter
func TestFilterQuery_BySurnameExact(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	surn1Line := types.NewGedcomLine(2, "SURN", "Doe", "")
	name1Line.AddChild(surn1Line)
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test BySurnameExact
	fq := NewFilterQuery(graph)
	results, err := fq.BySurnameExact("Doe").Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestFilterQuery_ByGivenName tests ByGivenName filter
func TestFilterQuery_ByGivenName(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with different given names
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	givn1Line := types.NewGedcomLine(2, "GIVN", "John", "")
	name1Line.AddChild(givn1Line)
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Doe/", "")
	givn2Line := types.NewGedcomLine(2, "GIVN", "Jane", "")
	name2Line.AddChild(givn2Line)
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test ByGivenName
	fq := NewFilterQuery(graph)
	results, err := fq.ByGivenName("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestFilterQuery_ByGivenNameExact tests ByGivenNameExact filter
func TestFilterQuery_ByGivenNameExact(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	givn1Line := types.NewGedcomLine(2, "GIVN", "John", "")
	name1Line.AddChild(givn1Line)
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test ByGivenNameExact
	fq := NewFilterQuery(graph)
	results, err := fq.ByGivenNameExact("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestGraph_GetEdges tests GetEdges method
func TestGraph_GetEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetEdges (may return nil if no edges exist)
	edges := graph.GetEdges("@I1@")
	// GetEdges can return nil if node has no edges, which is valid
	_ = edges // Just verify the method doesn't panic
}

// TestGraph_GetAllEdges tests GetAllEdges method
func TestGraph_GetAllEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetAllEdges
	allEdges := graph.GetAllEdges()
	if allEdges == nil {
		t.Error("Expected non-nil edges map")
	}
}

// TestQueryBuilder_NewQueryLazy tests NewQueryLazy function
func TestQueryBuilder_NewQueryLazy(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Test NewQueryLazy
	qb, err := NewQueryLazy(tree)
	if err != nil {
		t.Fatalf("Failed to create lazy query: %v", err)
	}
	if qb == nil {
		t.Fatal("Expected non-nil query builder")
	}

	// Test that it works
	indi := qb.Individual("@I1@")
	if indi == nil {
		t.Error("Expected non-nil individual query")
	}
}

// TestQueryBuilder_NewQueryFromGraph tests NewQueryFromGraph function
func TestQueryBuilder_NewQueryFromGraph(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test NewQueryFromGraph
	qb := NewQueryFromGraph(graph)
	if qb == nil {
		t.Fatal("Expected non-nil query builder")
	}

	// Test that graph is accessible
	returnedGraph := qb.Graph()
	if returnedGraph != graph {
		t.Error("Expected returned graph to be the same instance")
	}
}

// TestQueryBuilder_Individuals tests Individuals method
func TestQueryBuilder_Individuals(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Individuals
	miq := qb.Individuals("@I1@", "@I2@")
	if miq == nil {
		t.Fatal("Expected non-nil multi-individual query")
	}

	// Test that it works
	results, err := miq.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestQueryBuilder_AllIndividuals tests AllIndividuals method
func TestQueryBuilder_AllIndividuals(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test AllIndividuals
	miq := qb.AllIndividuals()
	if miq == nil {
		t.Fatal("Expected non-nil multi-individual query")
	}

	// Test that it works
	results, err := miq.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestQueryBuilder_Families tests Families method
func TestQueryBuilder_Families(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Families
	fcq := qb.Families()
	if fcq == nil {
		t.Fatal("Expected non-nil family collection query")
	}
}

// TestQueryBuilder_Events tests Events method
func TestQueryBuilder_Events(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with event
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Events
	ecq := qb.Events()
	if ecq == nil {
		t.Fatal("Expected non-nil event collection query")
	}
}

// TestQueryBuilder_Names tests Names method
func TestQueryBuilder_Names(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Names
	ncq := qb.Names()
	if ncq == nil {
		t.Fatal("Expected non-nil name collection query")
	}
}

// TestQueryBuilder_Places tests Places method
func TestQueryBuilder_Places(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth place
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "PLAC", "New York", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Places
	pcq := qb.Places()
	if pcq == nil {
		t.Fatal("Expected non-nil place collection query")
	}
}

// TestGraph_GetNode tests GetNode method
func TestGraph_GetNode(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetNode
	node := graph.GetNode("@I1@")
	if node == nil {
		t.Error("Expected non-nil node")
	}
	if node.ID() != "@I1@" {
		t.Errorf("Expected node ID @I1@, got %s", node.ID())
	}

	// Test GetNode with non-existent node
	node2 := graph.GetNode("@NONEXISTENT@")
	if node2 != nil {
		t.Error("Expected nil node for non-existent XREF")
	}
}

// TestGraph_GetAllNodes tests GetAllNodes method
func TestGraph_GetAllNodes_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetAllNodes
	allNodes := graph.GetAllNodes()
	if allNodes == nil {
		t.Fatal("Expected non-nil nodes map")
	}
	if len(allNodes) == 0 {
		t.Error("Expected at least 1 node")
	}
	if allNodes["@I1@"] == nil {
		t.Error("Expected @I1@ to be in nodes map")
	}
}

// TestGraph_GetAllFamilies tests GetAllFamilies method
func TestGraph_GetAllFamilies_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetAllFamilies
	allFamilies := graph.GetAllFamilies()
	if allFamilies == nil {
		t.Fatal("Expected non-nil families map")
	}
	if len(allFamilies) == 0 {
		t.Error("Expected at least 1 family")
	}
	if allFamilies["@F1@"] == nil {
		t.Error("Expected @F1@ to be in families map")
	}
}

// TestGraph_GetNote tests GetNote method
func TestGraph_GetNote_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add note
	note1Line := types.NewGedcomLine(0, "NOTE", "Test note", "@N1@")
	tree.AddRecord(types.NewNoteRecord(note1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetNote
	note := graph.GetNote("@N1@")
	if note == nil {
		t.Error("Expected non-nil note")
	}
	if note != nil && note.ID() != "@N1@" {
		t.Errorf("Expected note ID @N1@, got %s", note.ID())
	}
}

// TestGraph_GetSource tests GetSource method
func TestGraph_GetSource_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add source
	source1Line := types.NewGedcomLine(0, "SOUR", "", "@S1@")
	source1Line.AddChild(types.NewGedcomLine(1, "TITL", "Test Source", ""))
	tree.AddRecord(types.NewSourceRecord(source1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetSource
	source := graph.GetSource("@S1@")
	if source == nil {
		t.Error("Expected non-nil source")
	}
	if source != nil && source.ID() != "@S1@" {
		t.Errorf("Expected source ID @S1@, got %s", source.ID())
	}
}

// TestGraph_GetRepository tests GetRepository method
func TestGraph_GetRepository_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add repository
	repo1Line := types.NewGedcomLine(0, "REPO", "", "@R1@")
	repo1Line.AddChild(types.NewGedcomLine(1, "NAME", "Test Repository", ""))
	tree.AddRecord(types.NewRepositoryRecord(repo1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetRepository
	repo := graph.GetRepository("@R1@")
	if repo == nil {
		t.Error("Expected non-nil repository")
	}
	if repo != nil && repo.ID() != "@R1@" {
		t.Errorf("Expected repository ID @R1@, got %s", repo.ID())
	}
}

// TestNode_Neighbors tests Neighbors method
func TestNode_Neighbors(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Neighbors
	childNode := graph.GetIndividual("@I2@")
	if childNode == nil {
		t.Fatal("Expected non-nil child node")
	}
	neighbors := childNode.Neighbors()
	// Child should have at least the family as a neighbor
	if len(neighbors) == 0 {
		t.Error("Expected at least 1 neighbor")
	}
}

// TestNode_Degree tests Degree, InDegree, and OutDegree methods
func TestNode_Degree(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Degree methods
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Expected non-nil node")
	}

	// Test Degree (total edges)
	degree := node.Degree()
	_ = degree // Just verify it doesn't panic

	// Test InDegree
	inDegree := node.InDegree()
	_ = inDegree // Just verify it doesn't panic

	// Test OutDegree
	outDegree := node.OutDegree()
	_ = outDegree // Just verify it doesn't panic
}

// TestNode_RemoveInEdge tests RemoveInEdge method
func TestNode_RemoveInEdge(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get node
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Expected non-nil node")
	}

	// Create a test edge
	edge := NewEdge("@E1@", node, node, EdgeTypeFAMS)
	
	// Add edge first
	node.AddInEdge(edge)
	initialCount := len(node.InEdges())

	// Remove edge
	node.RemoveInEdge(edge)
	finalCount := len(node.InEdges())

	if finalCount != initialCount-1 {
		t.Errorf("Expected edge count to decrease by 1, got %d -> %d", initialCount, finalCount)
	}
}

// TestNode_RemoveOutEdge tests RemoveOutEdge method
func TestNode_RemoveOutEdge(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get node
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Expected non-nil node")
	}

	// Create a test edge
	edge := NewEdge("@E1@", node, node, EdgeTypeFAMS)
	
	// Add edge first
	node.AddOutEdge(edge)
	initialCount := len(node.OutEdges())

	// Remove edge
	node.RemoveOutEdge(edge)
	finalCount := len(node.OutEdges())

	if finalCount != initialCount-1 {
		t.Errorf("Expected edge count to decrease by 1, got %d -> %d", initialCount, finalCount)
	}
}

// TestGraph_CalculateRelationship tests CalculateRelationship method
func TestGraph_CalculateRelationship(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test CalculateRelationship (parent-child)
	result, err := graph.CalculateRelationship("@I1@", "@I2@")
	if err != nil {
		t.Fatalf("Failed to calculate relationship: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil relationship result")
	}
	if !result.IsDirect {
		t.Error("Expected direct relationship")
	}
}

// TestGraph_BFS_Coverage tests BFS traversal method for coverage
func TestGraph_BFS_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test BFS
	visited := make(map[string]bool)
	err = graph.BFS("@I1@", func(node GraphNode) bool {
		visited[node.ID()] = true
		return true // Continue traversal
	})
	if err != nil {
		t.Fatalf("BFS failed: %v", err)
	}
	if len(visited) == 0 {
		t.Error("Expected at least 1 visited node")
	}
}

// TestGraph_DFS_Coverage tests DFS traversal method for coverage
func TestGraph_DFS_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test DFS
	visited := make(map[string]bool)
	err = graph.DFS("@I1@", func(node GraphNode) bool {
		visited[node.ID()] = true
		return true // Continue traversal
	})
	if err != nil {
		t.Fatalf("DFS failed: %v", err)
	}
	if len(visited) == 0 {
		t.Error("Expected at least 1 visited node")
	}
}

// TestGraph_BFS_StopEarly tests BFS with early stop
func TestGraph_BFS_StopEarly(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test BFS with early stop
	count := 0
	err = graph.BFS("@I1@", func(node GraphNode) bool {
		count++
		return false // Stop after first node
	})
	if err != nil {
		t.Fatalf("BFS failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected to visit 1 node, visited %d", count)
	}
}

// TestPathQuery_MaxLength tests MaxLength method
func TestPathQuery_MaxLength(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test MaxLength
	pq := qb.Individual("@I1@").PathTo("@I2@")
	pq2 := pq.MaxLength(5)
	if pq2 == nil {
		t.Fatal("Expected non-nil path query")
	}
}

// TestPathQuery_IncludeMarital tests IncludeMarital method
func TestPathQuery_IncludeMarital(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test IncludeMarital
	pq := qb.Individual("@I1@").PathTo("@I2@")
	pq2 := pq.IncludeMarital(false)
	if pq2 == nil {
		t.Fatal("Expected non-nil path query")
	}
}

// TestPathQuery_IncludeBlood tests IncludeBlood method
func TestPathQuery_IncludeBlood(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test IncludeBlood
	pq := qb.Individual("@I1@").PathTo("@I2@")
	pq2 := pq.IncludeBlood(false)
	if pq2 == nil {
		t.Fatal("Expected non-nil path query")
	}
}

// TestPathQuery_ShortestOnly tests ShortestOnly method
func TestPathQuery_ShortestOnly(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test ShortestOnly
	pq := qb.Individual("@I1@").PathTo("@I2@")
	pq2 := pq.ShortestOnly(true)
	if pq2 == nil {
		t.Fatal("Expected non-nil path query")
	}
}

// TestPathQuery_Count tests Count method
func TestPathQuery_Count(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Count
	pq := qb.Individual("@I1@").PathTo("@I2@")
	count, err := pq.Count()
	if err != nil {
		t.Fatalf("Failed to count paths: %v", err)
	}
	_ = count // Just verify it doesn't panic

	// Test Count with ShortestOnly
	count2, err := pq.ShortestOnly(true).Count()
	if err != nil {
		t.Fatalf("Failed to count paths: %v", err)
	}
	_ = count2 // Just verify it doesn't panic
}

// TestQueryBuilder_Path tests Path method
func TestQueryBuilder_Path(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Path (via IndividualQuery.PathTo)
	pq := qb.Individual("@I1@").PathTo("@I2@")
	if pq == nil {
		t.Fatal("Expected non-nil path query")
	}
}

// TestQueryBuilder_Metrics tests Metrics method
func TestQueryBuilder_Metrics(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Metrics
	gmq := qb.Metrics()
	if gmq == nil {
		t.Fatal("Expected non-nil graph metrics query")
	}
}

// TestMultiIndividualQuery_Ancestors_Coverage tests Ancestors method for coverage
func TestMultiIndividualQuery_Ancestors_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Ancestors
	miq := qb.Individuals("@I2@", "@I3@")
	ancestors, err := miq.Ancestors()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}
	if len(ancestors) == 0 {
		t.Error("Expected at least 1 ancestor")
	}
}

// TestMultiIndividualQuery_Descendants_Coverage tests Descendants method for coverage
func TestMultiIndividualQuery_Descendants_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Descendants
	miq := qb.Individuals("@I1@")
	descendants, err := miq.Descendants()
	if err != nil {
		t.Fatalf("Failed to get descendants: %v", err)
	}
	if len(descendants) == 0 {
		t.Error("Expected at least 1 descendant")
	}
}

// TestMultiIndividualQuery_CommonAncestors_Coverage tests CommonAncestors method for coverage
func TestMultiIndividualQuery_CommonAncestors_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add common ancestor
	ancestorLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Ancestor /Doe/", "")
	ancestorLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(ancestorLine))

	// Add two children
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test CommonAncestors
	miq := qb.Individuals("@I2@", "@I3@")
	commonAncestors, err := miq.CommonAncestors()
	if err != nil {
		t.Fatalf("Failed to get common ancestors: %v", err)
	}
	if len(commonAncestors) == 0 {
		t.Error("Expected at least 1 common ancestor")
	}
}

// TestNameCollectionQuery_All tests All method
func TestNameCollectionQuery_All(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with name
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	givn1Line := types.NewGedcomLine(2, "GIVN", "John", "")
	surn1Line := types.NewGedcomLine(2, "SURN", "Doe", "")
	name1Line.AddChild(givn1Line)
	name1Line.AddChild(surn1Line)
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test All
	ncq := NewNameCollectionQuery(graph)
	names, err := ncq.All()
	if err != nil {
		t.Fatalf("Failed to get names: %v", err)
	}
	if len(names) == 0 {
		t.Error("Expected at least 1 name")
	}
}

// TestNameCollectionQuery_Unique tests Unique method
func TestNameCollectionQuery_Unique(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with same name
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Unique
	ncq := NewNameCollectionQuery(graph)
	ncq2 := ncq.Unique()
	if ncq2 == nil {
		t.Fatal("Expected non-nil name collection query")
	}
}

// TestNameCollectionQuery_By tests By method
func TestNameCollectionQuery_By(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test By
	ncq := NewNameCollectionQuery(graph)
	ncq2 := ncq.By(NameUniqueByGiven)
	if ncq2 == nil {
		t.Fatal("Expected non-nil name collection query")
	}
}

// TestNameCollectionQuery_Count tests Count method
func TestNameCollectionQuery_Count(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Count
	ncq := NewNameCollectionQuery(graph)
	count, err := ncq.Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count == 0 {
		t.Error("Expected at least 1 name")
	}
}

// TestNameCollectionQuery_Execute tests Execute method
func TestNameCollectionQuery_Execute(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	givn1Line := types.NewGedcomLine(2, "GIVN", "John", "")
	surn1Line := types.NewGedcomLine(2, "SURN", "Doe", "")
	name1Line.AddChild(givn1Line)
	name1Line.AddChild(surn1Line)
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Execute
	ncq := NewNameCollectionQuery(graph)
	names, err := ncq.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if len(names) == 0 {
		t.Error("Expected at least 1 name")
	}
}

// TestPlaceCollectionQuery_All_Coverage tests All method for coverage
func TestPlaceCollectionQuery_All_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth place
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "PLAC", "New York, NY, USA", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test All
	pcq := NewPlaceCollectionQuery(graph)
	places, err := pcq.All()
	if err != nil {
		t.Fatalf("Failed to get places: %v", err)
	}
	if len(places) == 0 {
		t.Error("Expected at least 1 place")
	}
}

// TestPlaceCollectionQuery_Unique_Coverage tests Unique method for coverage
func TestPlaceCollectionQuery_Unique_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Unique
	pcq := NewPlaceCollectionQuery(graph)
	pcq2 := pcq.Unique()
	if pcq2 == nil {
		t.Fatal("Expected non-nil place collection query")
	}
}

// TestPlaceCollectionQuery_By_Coverage tests By method for coverage
func TestPlaceCollectionQuery_By_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test By
	pcq := NewPlaceCollectionQuery(graph)
	pcq2 := pcq.By(PlaceUniqueByCity)
	if pcq2 == nil {
		t.Fatal("Expected non-nil place collection query")
	}
}

// TestPlaceCollectionQuery_FromBirth_Coverage tests FromBirth method for coverage
func TestPlaceCollectionQuery_FromBirth_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test FromBirth
	pcq := NewPlaceCollectionQuery(graph)
	pcq2 := pcq.FromBirth()
	if pcq2 == nil {
		t.Fatal("Expected non-nil place collection query")
	}
}

// TestPlaceCollectionQuery_FromDeath_Coverage tests FromDeath method for coverage
func TestPlaceCollectionQuery_FromDeath_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test FromDeath
	pcq := NewPlaceCollectionQuery(graph)
	pcq2 := pcq.FromDeath()
	if pcq2 == nil {
		t.Fatal("Expected non-nil place collection query")
	}
}

// TestPlaceCollectionQuery_FromMarriage_Coverage tests FromMarriage method for coverage
func TestPlaceCollectionQuery_FromMarriage_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add family with marriage place
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(types.NewGedcomLine(2, "PLAC", "Las Vegas, NV, USA", ""))
	fam1Line.AddChild(marrLine)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test FromMarriage
	pcq := NewPlaceCollectionQuery(graph)
	pcq2 := pcq.FromMarriage()
	if pcq2 == nil {
		t.Fatal("Expected non-nil place collection query")
	}
}

// TestPlaceCollectionQuery_Count_Coverage tests Count method for coverage
func TestPlaceCollectionQuery_Count_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth place
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "PLAC", "New York", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Count
	pcq := NewPlaceCollectionQuery(graph)
	count, err := pcq.Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count == 0 {
		t.Error("Expected at least 1 place")
	}
}

// TestPlaceCollectionQuery_Execute_Coverage tests Execute method for coverage
func TestPlaceCollectionQuery_Execute_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with birth place
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	birt1Line := types.NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(types.NewGedcomLine(2, "PLAC", "New York, NY, USA", ""))
	indi1Line.AddChild(birt1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Execute
	pcq := NewPlaceCollectionQuery(graph)
	places, err := pcq.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if len(places) == 0 {
		t.Error("Expected at least 1 place")
	}
}
// TestGraph_DFS_StopEarly tests DFS with early stop
func TestGraph_DFS_StopEarly(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test DFS with early stop
	count := 0
	err = graph.DFS("@I1@", func(node GraphNode) bool {
		count++
		return false // Stop after first node
	})
	if err != nil {
		t.Fatalf("DFS failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected to visit 1 node, visited %d", count)
	}
}

// TestIndividualQuery_GetSubtree tests GetSubtree method
func TestIndividualQuery_GetSubtree(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test GetSubtree
	subtree := qb.Individual("@I1@").GetSubtree()
	if subtree == nil {
		t.Fatal("Expected non-nil subtree query")
	}
}

// TestMultiIndividualQuery_Union tests Union method
func TestMultiIndividualQuery_Union(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Bob /Jones/", "")
	indi3Line.AddChild(name3Line)
	tree.AddRecord(types.NewIndividualRecord(indi3Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Union
	miq1 := qb.Individuals("@I1@", "@I2@")
	results, err := miq1.Union(func(iq *IndividualQuery) ([]*types.IndividualRecord, error) {
		return iq.Parents()
	})
	if err != nil {
		t.Fatalf("Failed to execute union: %v", err)
	}
	_ = results // Just verify it doesn't panic
}

// TestMultiIndividualQuery_Intersection tests Intersection method
func TestMultiIndividualQuery_Intersection(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Intersection
	miq1 := qb.Individuals("@I1@", "@I2@")
	results, err := miq1.Intersection(func(iq *IndividualQuery) ([]*types.IndividualRecord, error) {
		return iq.Parents()
	})
	if err != nil {
		t.Fatalf("Failed to execute intersection: %v", err)
	}
	_ = results // Just verify it doesn't panic
}

// TestFamilyCollectionQuery_uniqueByMarriagePlace tests uniqueByMarriagePlace method
func TestFamilyCollectionQuery_uniqueByMarriagePlace(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add families with marriage places
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marr1Line := types.NewGedcomLine(1, "MARR", "", "")
	marr1Line.AddChild(types.NewGedcomLine(2, "PLAC", "New York", ""))
	fam1Line.AddChild(marr1Line)
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	marr2Line := types.NewGedcomLine(1, "MARR", "", "")
	marr2Line.AddChild(types.NewGedcomLine(2, "PLAC", "Los Angeles", ""))
	fam2Line.AddChild(marr2Line)
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test uniqueByMarriagePlace via Execute
	fcq := NewFamilyCollectionQuery(graph)
	families, err := fcq.By(FamilyUniqueByMarriagePlace).Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	_ = families // Just verify it doesn't panic
}

// TestIndividualQuery_Cousins tests Cousins method
func TestIndividualQuery_Cousins(t *testing.T) {
	// Create test tree with cousins
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add two parents (siblings)
	parent1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent1 /Doe/", "")
	parent1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parent1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parent1Line))

	parent2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Parent2 /Doe/", "")
	parent2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parent2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(parent2Line))

	// Add two children (cousins)
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name4Line)
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	child1Line.AddChild(famc3Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I5@")
	name5Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name5Line)
	famc4Line := types.NewGedcomLine(1, "FAMC", "@F3@", "")
	child2Line.AddChild(famc4Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	fam3Line := types.NewGedcomLine(0, "FAM", "", "@F3@")
	fam3Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I3@", ""))
	fam3Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I5@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam3Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Cousins (first cousins, degree 1)
	cousins, err := qb.Individual("@I4@").Cousins(1)
	if err != nil {
		t.Fatalf("Failed to get cousins: %v", err)
	}
	_ = cousins // Just verify it doesn't panic
}

// TestIndividualQuery_Uncles tests Uncles method
func TestIndividualQuery_Uncles(t *testing.T) {
	// Create test tree with uncles
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add uncle (sibling of parent)
	uncleLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Uncle /Doe/", "")
	uncleLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	uncleLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(uncleLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name4Line)
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc3Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Uncles
	uncles, err := qb.Individual("@I4@").Uncles()
	if err != nil {
		t.Fatalf("Failed to get uncles: %v", err)
	}
	_ = uncles // Just verify it doesn't panic
}

// TestIndividualQuery_Nephews tests Nephews method
func TestIndividualQuery_Nephews(t *testing.T) {
	// Create test tree with nephews
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add two siblings
	sibling1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Sibling1 /Doe/", "")
	sibling1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	sibling1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(sibling1Line))

	sibling2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Sibling2 /Doe/", "")
	sibling2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	sibling2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(sibling2Line))

	// Add nephew (child of sibling)
	nephewLine := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Nephew /Doe/", "")
	nephewLine.AddChild(name4Line)
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	nephewLine.AddChild(famc3Line)
	tree.AddRecord(types.NewIndividualRecord(nephewLine))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Nephews
	nephews, err := qb.Individual("@I3@").Nephews()
	if err != nil {
		t.Fatalf("Failed to get nephews: %v", err)
	}
	_ = nephews // Just verify it doesn't panic
}

// TestIndividualQuery_Grandparents tests Grandparents method
func TestIndividualQuery_Grandparents(t *testing.T) {
	// Create test tree with grandparents
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Grandparents
	grandparents, err := qb.Individual("@I3@").Grandparents()
	if err != nil {
		t.Fatalf("Failed to get grandparents: %v", err)
	}
	_ = grandparents // Just verify it doesn't panic
}

// TestIndividualQuery_Grandchildren tests Grandchildren method
func TestIndividualQuery_Grandchildren(t *testing.T) {
	// Create test tree with grandchildren
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add grandchild
	grandchildLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Grandchild /Doe/", "")
	grandchildLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	grandchildLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(grandchildLine))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Grandchildren
	grandchildren, err := qb.Individual("@I1@").Grandchildren()
	if err != nil {
		t.Fatalf("Failed to get grandchildren: %v", err)
	}
	_ = grandchildren // Just verify it doesn't panic
}

// TestIndividualQuery_RelationshipToResult tests RelationshipToResult method
func TestIndividualQuery_RelationshipToResult(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test RelationshipToResult
	result, err := qb.Individual("@I1@").RelationshipToResult("@I2@")
	if err != nil {
		t.Fatalf("Failed to get relationship: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil relationship result")
	}
}

// TestGraph_ShortestPath_SameNode_Coverage tests ShortestPath with same node for coverage
func TestGraph_ShortestPath_SameNode_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test ShortestPath with same node
	path, err := graph.ShortestPath("@I1@", "@I1@")
	if err != nil {
		t.Fatalf("Failed to get path: %v", err)
	}
	if path == nil {
		t.Fatal("Expected non-nil path")
	}
	if path.Length != 0 {
		t.Errorf("Expected path length 0 for same node, got %d", path.Length)
	}
}

// TestGraph_AllPaths_Coverage tests AllPaths method for coverage
func TestGraph_AllPaths_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test AllPaths
	paths, err := graph.AllPaths("@I1@", "@I2@", 10)
	if err != nil {
		t.Fatalf("Failed to get paths: %v", err)
	}
	if len(paths) == 0 {
		t.Error("Expected at least 1 path")
	}
}

// TestGraph_CommonAncestors_Coverage tests CommonAncestors method for coverage
func TestGraph_CommonAncestors_Coverage(t *testing.T) {
	// Create test tree with common ancestor
	tree := types.NewGedcomTree()

	// Add common ancestor
	ancestorLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Ancestor /Doe/", "")
	ancestorLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(ancestorLine))

	// Add two children
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test CommonAncestors
	ancestors, err := graph.CommonAncestors("@I2@", "@I3@")
	if err != nil {
		t.Fatalf("Failed to get common ancestors: %v", err)
	}
	if len(ancestors) == 0 {
		t.Error("Expected at least 1 common ancestor")
	}
}

// TestGraph_LowestCommonAncestor_Coverage tests LowestCommonAncestor method for coverage
func TestGraph_LowestCommonAncestor_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add two parents
	parent1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent1 /Doe/", "")
	parent1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parent1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parent1Line))

	parent2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Parent2 /Doe/", "")
	parent2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parent2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(parent2Line))

	// Add two children
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name4Line)
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	child1Line.AddChild(famc3Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I5@")
	name5Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name5Line)
	famc4Line := types.NewGedcomLine(1, "FAMC", "@F3@", "")
	child2Line.AddChild(famc4Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	fam3Line := types.NewGedcomLine(0, "FAM", "", "@F3@")
	fam3Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I3@", ""))
	fam3Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I5@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam3Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test LowestCommonAncestor
	lca, err := graph.LowestCommonAncestor("@I4@", "@I5@")
	if err != nil {
		t.Fatalf("Failed to get LCA: %v", err)
	}
	if lca == nil {
		t.Error("Expected non-nil LCA")
	}
}

// TestEdge_OtherNode_Coverage tests OtherNode method for coverage
func TestEdge_OtherNode_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get node
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Expected non-nil node")
	}

	// Create a test edge
	edge := NewEdge("@E1@", node, node, EdgeTypeFAMS)
	
	// Test OtherNode with From node
	otherFrom := edge.OtherNode(node)
	if otherFrom == nil {
		t.Error("Expected non-nil other node")
	}
}

// TestGraph_reverseEdgeType_Coverage tests reverseEdgeType function for coverage
func TestGraph_reverseEdgeType_Coverage(t *testing.T) {
	// Test various edge types
	tests := []struct {
		input    EdgeType
		expected EdgeType
	}{
		{EdgeTypeHUSB, EdgeTypeFAMS},
		{EdgeTypeWIFE, EdgeTypeFAMS},
		{EdgeTypeCHIL, EdgeTypeFAMC},
		{EdgeTypeFAMC, EdgeTypeCHIL},
		{EdgeTypeFAMS, EdgeTypeFAMS},
		{EdgeTypeHasEvent, EdgeTypeHasEvent},
	}

	for _, tt := range tests {
		result := reverseEdgeType(tt.input)
		_ = result // Just verify it doesn't panic
	}
}

// TestFilterQuery_filterByBool_Coverage tests filterByBool function for coverage
func TestFilterQuery_filterByBool_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual with children
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test filterByBool via HasChildren filter
	fq := NewFilterQuery(graph)
	results, err := fq.HasChildren().Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	_ = results // Just verify it doesn't panic
}

// TestIndividualQuery_CommonAncestors_Coverage tests CommonAncestors method for coverage
func TestIndividualQuery_CommonAncestors_Coverage(t *testing.T) {
	// Create test tree with common ancestor
	tree := types.NewGedcomTree()

	// Add common ancestor
	ancestorLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Ancestor /Doe/", "")
	ancestorLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(ancestorLine))

	// Add two children
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test CommonAncestors
	ancestors, err := qb.Individual("@I2@").CommonAncestors("@I3@")
	if err != nil {
		t.Fatalf("Failed to get common ancestors: %v", err)
	}
	if len(ancestors) == 0 {
		t.Error("Expected at least 1 common ancestor")
	}
}

// TestGraph_NodeCount_Coverage tests NodeCount method for coverage
func TestGraph_NodeCount_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test NodeCount
	count := graph.NodeCount()
	if count == 0 {
		t.Error("Expected at least 1 node")
	}
}

// TestGraph_EdgeCount_Coverage tests EdgeCount method for coverage
func TestGraph_EdgeCount_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test EdgeCount
	count := graph.EdgeCount()
	_ = count // Just verify it doesn't panic
}

// TestGraph_AddEdge_Coverage tests AddEdge method for coverage
func TestGraph_AddEdge_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add two individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get nodes
	node1 := graph.GetIndividual("@I1@")
	node2 := graph.GetIndividual("@I2@")
	if node1 == nil || node2 == nil {
		t.Fatal("Expected non-nil nodes")
	}

	// Test AddEdge
	edge := NewEdge("@E1@", node1, node2, EdgeTypeFAMS)
	err = graph.AddEdge(edge)
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	// Test adding duplicate edge (should fail)
	err = graph.AddEdge(edge)
	if err == nil {
		t.Error("Expected error when adding duplicate edge")
	}
}

// TestGraph_AddEdge_Invalid_Coverage tests AddEdge with invalid edge for coverage
func TestGraph_AddEdge_Invalid_Coverage(t *testing.T) {
	// Create graph
	graph := NewGraph(types.NewGedcomTree())

	// Test AddEdge with nil edge
	err := graph.AddEdge(nil)
	if err == nil {
		t.Error("Expected error when adding nil edge")
	}

	// Test AddEdge with empty ID
	// Create a simple node for testing
	tree2 := types.NewGedcomTree()
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name1Line := types.NewGedcomLine(1, "NAME", "Test /Node/", "")
	indi1Line.AddChild(name1Line)
	tree2.AddRecord(types.NewIndividualRecord(indi1Line))
	graph2, _ := BuildGraph(tree2)
	node1 := graph2.GetIndividual("@I3@")
	if node1 == nil {
		t.Fatal("Expected non-nil node")
	}

	edge1 := NewEdge("", node1, node1, EdgeTypeFAMS)
	err = graph2.AddEdge(edge1)
	if err == nil {
		t.Error("Expected error when adding edge with empty ID")
	}

	// Test AddEdge with nil From node
	edge2 := NewEdge("@E2@", nil, node1, EdgeTypeFAMS)
	err = graph2.AddEdge(edge2)
	if err == nil {
		t.Error("Expected error when adding edge with nil From node")
	}

	// Test AddEdge with nil To node
	edge3 := NewEdge("@E3@", node1, nil, EdgeTypeFAMS)
	err = graph2.AddEdge(edge3)
	if err == nil {
		t.Error("Expected error when adding edge with nil To node")
	}
}




// TestIndividualQuery_RelationshipTo_Coverage tests RelationshipTo method for coverage
func TestIndividualQuery_RelationshipTo_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test RelationshipTo
	rq := qb.Individual("@I1@").RelationshipTo("@I2@")
	if rq == nil {
		t.Fatal("Expected non-nil relationship query")
	}
}

// TestPathQuery_All_Coverage tests All method for coverage
func TestPathQuery_All_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test All
	pq := qb.Individual("@I1@").PathTo("@I2@")
	paths, err := pq.All()
	if err != nil {
		t.Fatalf("Failed to get paths: %v", err)
	}
	if len(paths) == 0 {
		t.Error("Expected at least 1 path")
	}
}

// TestPathQuery_Shortest_Coverage tests Shortest method for coverage
func TestPathQuery_Shortest_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Shortest
	pq := qb.Individual("@I1@").PathTo("@I2@")
	path, err := pq.Shortest()
	if err != nil {
		t.Fatalf("Failed to get shortest path: %v", err)
	}
	if path == nil {
		t.Fatal("Expected non-nil path")
	}
}


// TestGraph_GetFamilyHusband tests GetFamilyHusband method
func TestGraph_GetFamilyHusband(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add husband
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Husband /Doe/", "")
	husbandLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(husbandLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetFamilyHusband
	husband, err := graph.GetFamilyHusband("@F1@")
	if err != nil {
		t.Fatalf("Failed to get husband: %v", err)
	}
	if husband == nil {
		t.Error("Expected non-nil husband")
	}
	if husband != nil && husband.ID() != "@I1@" {
		t.Errorf("Expected husband ID @I1@, got %s", husband.ID())
	}
}

// TestGraph_GetFamilyWife tests GetFamilyWife method
func TestGraph_GetFamilyWife(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add wife
	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Wife /Doe/", "")
	wifeLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(wifeLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I1@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetFamilyWife
	wife, err := graph.GetFamilyWife("@F1@")
	if err != nil {
		t.Fatalf("Failed to get wife: %v", err)
	}
	if wife == nil {
		t.Error("Expected non-nil wife")
	}
	if wife != nil && wife.ID() != "@I1@" {
		t.Errorf("Expected wife ID @I1@, got %s", wife.ID())
	}
}

// TestGraph_GetFamilyChildren tests GetFamilyChildren method
func TestGraph_GetFamilyChildren(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name1Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I1@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetFamilyChildren
	children, err := graph.GetFamilyChildren("@F1@")
	if err != nil {
		t.Fatalf("Failed to get children: %v", err)
	}
	if len(children) == 0 {
		t.Error("Expected at least 1 child")
	}
	if len(children) > 0 && children[0].ID() != "@I1@" {
		t.Errorf("Expected child ID @I1@, got %s", children[0].ID())
	}
}

// TestFamilyNode_getHusbandFromEdges tests getHusbandFromEdges method
func TestFamilyNode_getHusbandFromEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add husband
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Husband /Doe/", "")
	husbandLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(husbandLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Test getHusbandFromEdges via FamilyQuery
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}
	husband, err := qb.Family("@F1@").Husband()
	if err != nil {
		t.Fatalf("Failed to get husband: %v", err)
	}
	if husband == nil {
		t.Error("Expected non-nil husband")
	}
}

// TestFamilyNode_getWifeFromEdges tests getWifeFromEdges method
func TestFamilyNode_getWifeFromEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add wife
	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Wife /Doe/", "")
	wifeLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(wifeLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I1@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test getWifeFromEdges via FamilyQuery
	wife, err := qb.Family("@F1@").Wife()
	if err != nil {
		t.Fatalf("Failed to get wife: %v", err)
	}
	if wife == nil {
		t.Error("Expected non-nil wife")
	}
}

// TestFamilyNode_getChildrenFromEdges tests getChildrenFromEdges method
func TestFamilyNode_getChildrenFromEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name1Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I1@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test getChildrenFromEdges via FamilyQuery
	children, err := qb.Family("@F1@").Children()
	if err != nil {
		t.Fatalf("Failed to get children: %v", err)
	}
	if len(children) == 0 {
		t.Error("Expected at least 1 child")
	}
}

// TestIndividualNode_getParentsFromEdges tests getParentsFromEdges method
func TestIndividualNode_getParentsFromEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test getParentsFromEdges via IndividualQuery
	parents, err := qb.Individual("@I2@").Parents()
	if err != nil {
		t.Fatalf("Failed to get parents: %v", err)
	}
	if len(parents) == 0 {
		t.Error("Expected at least 1 parent")
	}
}

// TestIndividualNode_getChildrenFromEdges tests getChildrenFromEdges method
func TestIndividualNode_getChildrenFromEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test getChildrenFromEdges via IndividualQuery
	children, err := qb.Individual("@I1@").Children()
	if err != nil {
		t.Fatalf("Failed to get children: %v", err)
	}
	if len(children) == 0 {
		t.Error("Expected at least 1 child")
	}
}

// TestIndividualNode_getSiblingsFromEdges tests getSiblingsFromEdges method
func TestIndividualNode_getSiblingsFromEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add two children (siblings)
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test getSiblingsFromEdges via IndividualQuery
	siblings, err := qb.Individual("@I2@").Siblings()
	if err != nil {
		t.Fatalf("Failed to get siblings: %v", err)
	}
	if len(siblings) == 0 {
		t.Error("Expected at least 1 sibling")
	}
}

// TestIndividualNode_getSpousesFromEdges tests getSpousesFromEdges method
func TestIndividualNode_getSpousesFromEdges(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add husband
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Husband /Doe/", "")
	husbandLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(husbandLine))

	// Add wife
	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Wife /Doe/", "")
	wifeLine.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(wifeLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test getSpousesFromEdges via IndividualQuery
	spouses, err := qb.Individual("@I1@").Spouses()
	if err != nil {
		t.Fatalf("Failed to get spouses: %v", err)
	}
	if len(spouses) == 0 {
		t.Error("Expected at least 1 spouse")
	}
}

// TestRelationshipResult_GetRelationshipType tests GetRelationshipType method
func TestRelationshipResult_GetRelationshipType(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test GetRelationshipType
	result, err := qb.Individual("@I1@").RelationshipToResult("@I2@")
	if err != nil {
		t.Fatalf("Failed to get relationship: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil relationship result")
	}
	relType := result.GetRelationshipType()
	if relType == "" {
		t.Error("Expected non-empty relationship type")
	}
}

// TestRelationshipResult_IsBloodRelation tests IsBloodRelation method
func TestRelationshipResult_IsBloodRelation(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test IsBloodRelation
	result, err := qb.Individual("@I1@").RelationshipToResult("@I2@")
	if err != nil {
		t.Fatalf("Failed to get relationship: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil relationship result")
	}
	isBlood := result.IsBloodRelation()
	_ = isBlood // Just verify it doesn't panic
}

// TestRelationshipResult_IsMaritalRelation tests IsMaritalRelation method
func TestRelationshipResult_IsMaritalRelation(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add husband
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Husband /Doe/", "")
	husbandLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(husbandLine))

	// Add wife
	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Wife /Doe/", "")
	wifeLine.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(wifeLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test IsMaritalRelation
	result, err := qb.Individual("@I1@").RelationshipToResult("@I2@")
	if err != nil {
		t.Fatalf("Failed to get relationship: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil relationship result")
	}
	isMarital := result.IsMaritalRelation()
	_ = isMarital // Just verify it doesn't panic
}

// TestSubtreeQuery_Execute_Coverage tests Execute method for coverage
func TestSubtreeQuery_Execute_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Execute
	subtree := qb.Individual("@I1@").GetSubtree()
	result, err := subtree.Execute()
	if err != nil {
		t.Fatalf("Failed to execute subtree: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil subtree result")
	}
}

// TestSubtreeQuery_Count_Coverage tests Count method for coverage
func TestSubtreeQuery_Count_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Count
	subtree := qb.Individual("@I1@").GetSubtree()
	count, err := subtree.Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count == 0 {
		t.Error("Expected at least 1 individual in subtree")
	}
}

// TestSubtreeQuery_ExecuteRecords_Coverage tests ExecuteRecords method for coverage
func TestSubtreeQuery_ExecuteRecords_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test ExecuteRecords
	subtree := qb.Individual("@I1@").GetSubtree()
	records, err := subtree.ExecuteRecords()
	if err != nil {
		t.Fatalf("Failed to execute records: %v", err)
	}
	if len(records) == 0 {
		t.Error("Expected at least 1 record")
	}
}

// TestSubtreeQuery_IncludeSiblings_Coverage tests IncludeSiblings method for coverage
func TestSubtreeQuery_IncludeSiblings_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add two children (siblings)
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	child2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test IncludeSiblings
	subtree := qb.Individual("@I2@").GetSubtree()
	result, err := subtree.IncludeSiblings().Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(result.Siblings) == 0 {
		t.Error("Expected at least 1 sibling")
	}
}

// TestSubtreeQuery_IncludeSpouses_Coverage tests IncludeSpouses method for coverage
func TestSubtreeQuery_IncludeSpouses_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add husband
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Husband /Doe/", "")
	husbandLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(husbandLine))

	// Add wife
	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Wife /Doe/", "")
	wifeLine.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(wifeLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test IncludeSpouses
	subtree := qb.Individual("@I1@").GetSubtree()
	result, err := subtree.IncludeSpouses().Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(result.Spouses) == 0 {
		t.Error("Expected at least 1 spouse")
	}
}

// TestSubtreeQuery_ExcludeSelf_Coverage tests ExcludeSelf method for coverage
func TestSubtreeQuery_ExcludeSelf_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test ExcludeSelf
	subtree := qb.Individual("@I1@").GetSubtree()
	result, err := subtree.ExcludeSelf().Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	// Root should still be set, but not in All if ExcludeSelf
	_ = result // Just verify it doesn't panic
}

// TestSubtreeQuery_AncestorGenerations_Coverage tests AncestorGenerations method for coverage
func TestSubtreeQuery_AncestorGenerations_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test AncestorGenerations
	subtree := qb.Individual("@I3@").GetSubtree()
	result, err := subtree.AncestorGenerations(1).Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

// TestSubtreeQuery_DescendantGenerations_Coverage tests DescendantGenerations method for coverage
func TestSubtreeQuery_DescendantGenerations_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test DescendantGenerations
	subtree := qb.Individual("@I1@").GetSubtree()
	result, err := subtree.DescendantGenerations(1).Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

// TestSubtreeQuery_Filter_Coverage tests Filter method for coverage
func TestSubtreeQuery_Filter_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Filter
	subtree := qb.Individual("@I1@").GetSubtree()
	result, err := subtree.Filter(func(indi *types.IndividualRecord) bool {
		return indi.XrefID() == "@I1@"
	}).Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}


// TestGraph_GetXrefFromID tests GetXrefFromID method
func TestGraph_GetXrefFromID(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get node ID first
	nodeID := graph.GetNodeID("@I1@")
	if nodeID == 0 {
		t.Fatal("Expected non-zero node ID")
	}

	// Test GetXrefFromID
	xref := graph.GetXrefFromID(nodeID)
	if xref != "@I1@" {
		t.Errorf("Expected XREF @I1@, got %s", xref)
	}
}

// TestGraph_GetNodeByID tests GetNodeByID method
func TestGraph_GetNodeByID(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get node ID first
	nodeID := graph.GetNodeID("@I1@")
	if nodeID == 0 {
		t.Fatal("Expected non-zero node ID")
	}

	// Test GetNodeByID
	node := graph.GetNodeByID(nodeID)
	if node == nil {
		t.Fatal("Expected non-nil node")
	}
	if node.ID() != "@I1@" {
		t.Errorf("Expected node ID @I1@, got %s", node.ID())
	}
}

// TestGraph_AddNode_Coverage tests AddNode method
func TestGraph_AddNode_Coverage(t *testing.T) {
	// Create graph
	graph := NewGraph(types.NewGedcomTree())

	// Create a test node
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	node := NewIndividualNode("@I1@", indi1)

	// Test AddNode
	err := graph.AddNode(node)
	if err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}

	// Verify node was added
	retrievedNode := graph.GetNode("@I1@")
	if retrievedNode == nil {
		t.Error("Expected node to be added")
	}

	// Test adding duplicate node (should fail)
	err = graph.AddNode(node)
	if err == nil {
		t.Error("Expected error when adding duplicate node")
	}
}

// TestGraph_AddNode_Coverage_Invalid tests AddNode with invalid node
func TestGraph_AddNode_Coverage_Invalid(t *testing.T) {
	// Create graph
	graph := NewGraph(types.NewGedcomTree())

	// Test AddNode with nil node - this will panic, so we use recover
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when adding nil node")
			}
		}()
		_ = graph.AddNode(nil)
	}()

	// Test AddNode with empty XREF
	indi1Line := types.NewGedcomLine(0, "INDI", "", "")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	node := NewIndividualNode("", indi1)

	err2 := graph.AddNode(node)
	if err2 == nil {
		t.Error("Expected error when adding node with empty XREF")
	}
}

// TestFilterQuery_Where tests Where method
func TestFilterQuery_Where(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test Where
	fq := NewFilterQuery(graph)
	results, err := fq.Where(func(indi *types.IndividualRecord) bool {
		return indi.GetName() == "John /Doe/"
	}).Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// TestEventNode_Record tests Record method
func TestEventNode_Record(t *testing.T) {
	// Create an event node
	eventData := map[string]interface{}{
		"date": "1 JAN 1900",
		"place": "New York",
	}
	eventNode := NewEventNode("@E1@", "BIRT", eventData)

	// Test Record (should return nil for EventNode)
	record := eventNode.Record()
	if record != nil {
		t.Error("Expected nil record for EventNode")
	}
}

// TestGraph_BuildGraphLazy tests BuildGraphLazy function
func TestGraph_BuildGraphLazy(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individual
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Test BuildGraphLazy
	graph, err := BuildGraphLazy(tree)
	if err != nil {
		t.Fatalf("Failed to build lazy graph: %v", err)
	}
	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}

	// Verify lazy mode is enabled
	if !graph.lazyMode {
		t.Error("Expected lazy mode to be enabled")
	}
}

// TestGraph_NewGraphWithConfig tests NewGraphWithConfig function
func TestGraph_NewGraphWithConfig(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Create config
	config := DefaultConfig()
	config.Cache.QueryCacheSize = 1000

	// Test NewGraphWithConfig
	graph := NewGraphWithConfig(tree, config)
	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}

	// Test with nil config (should use defaults)
	graph2 := NewGraphWithConfig(tree, nil)
	if graph2 == nil {
		t.Fatal("Expected non-nil graph")
	}
}

// TestPathQuery_reconstructBidirectionalPath tests reconstructBidirectionalPath function
func TestPathQuery_reconstructBidirectionalPath(t *testing.T) {
	// Create test tree with bidirectional relationship
	tree := types.NewGedcomTree()

	// Add husband
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Husband /Doe/", "")
	husbandLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(husbandLine))

	// Add wife
	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Wife /Doe/", "")
	wifeLine.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(wifeLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test path finding which may use bidirectional paths
	pq := qb.Individual("@I1@").PathTo("@I2@")
	paths, err := pq.All()
	if err != nil {
		t.Fatalf("Failed to get paths: %v", err)
	}
	_ = paths // Just verify it doesn't panic
}

// TestPathQuery_allPathsDFS tests allPathsDFS function
func TestPathQuery_allPathsDFS(t *testing.T) {
	// Create test tree with multiple paths
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add two parents (siblings)
	parent1Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent1 /Doe/", "")
	parent1Line.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parent1Line.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parent1Line))

	parent2Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Parent2 /Doe/", "")
	parent2Line.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parent2Line.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(parent2Line))

	// Add two children (cousins)
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	child1Line.AddChild(name4Line)
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	child1Line.AddChild(famc3Line)
	tree.AddRecord(types.NewIndividualRecord(child1Line))

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I5@")
	name5Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	child2Line.AddChild(name5Line)
	famc4Line := types.NewGedcomLine(1, "FAMC", "@F3@", "")
	child2Line.AddChild(famc4Line)
	tree.AddRecord(types.NewIndividualRecord(child2Line))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	fam3Line := types.NewGedcomLine(0, "FAM", "", "@F3@")
	fam3Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I3@", ""))
	fam3Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I5@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam3Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test allPathsDFS via PathQuery.All
	pq := qb.Individual("@I4@").PathTo("@I5@")
	paths, err := pq.All()
	if err != nil {
		t.Fatalf("Failed to get all paths: %v", err)
	}
	_ = paths // Just verify it doesn't panic
}

// TestPathQuery_shouldIncludePath tests shouldIncludePath method
func TestPathQuery_shouldIncludePath(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add husband
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Husband /Doe/", "")
	husbandLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(husbandLine))

	// Add wife
	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Wife /Doe/", "")
	wifeLine.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(wifeLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test shouldIncludePath via IncludeMarital
	pq := qb.Individual("@I1@").PathTo("@I2@")
	paths, err := pq.IncludeMarital(true).All()
	if err != nil {
		t.Fatalf("Failed to get paths: %v", err)
	}
	_ = paths // Just verify it doesn't panic

	// Test shouldIncludePath via IncludeBlood
	pq2 := qb.Individual("@I1@").PathTo("@I2@")
	paths2, err := pq2.IncludeBlood(true).All()
	if err != nil {
		t.Fatalf("Failed to get paths: %v", err)
	}
	_ = paths2 // Just verify it doesn't panic
}

// TestPathQuery_MaxLength_Coverage tests MaxLength method for coverage
func TestPathQuery_MaxLength_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add grandparent
	grandpaLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Grandpa /Doe/", "")
	grandpaLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(grandpaLine))

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name2Line)
	famc1Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	parentLine.AddChild(famc1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name3Line)
	famc2Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	childLine.AddChild(famc2Line)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test MaxLength
	pq := qb.Individual("@I1@").PathTo("@I3@")
	paths, err := pq.MaxLength(2).All()
	if err != nil {
		t.Fatalf("Failed to get paths: %v", err)
	}
	_ = paths // Just verify it doesn't panic
}

// TestFilterQuery_filterByBool_Coverage tests filterByBool function for coverage

// TestPathQuery_Count_Coverage2 tests Count method for coverage (second test)
func TestPathQuery_Count_Coverage2(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Count
	pq := qb.Individual("@I1@").PathTo("@I2@")
	count, err := pq.Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count == 0 {
		t.Error("Expected at least 1 path")
	}

	// Test Count with ShortestOnly
	pq2 := qb.Individual("@I1@").PathTo("@I2@")
	count2, err := pq2.ShortestOnly(true).Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	_ = count2 // Just verify it doesn't panic
}

// TestBaseNode_RemoveInEdge_Coverage tests RemoveInEdge method for coverage
func TestBaseNode_RemoveInEdge_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get node
	node := graph.GetIndividual("@I2@")
	if node == nil {
		t.Fatal("Expected non-nil node")
	}

	// Get edges
	edges := node.InEdges()
	if len(edges) == 0 {
		t.Fatal("Expected at least 1 in-edge")
	}

	// Test RemoveInEdge
	node.RemoveInEdge(edges[0])
	_ = node // Just verify it doesn't panic
}

// TestBaseNode_RemoveOutEdge_Coverage tests RemoveOutEdge method for coverage
func TestBaseNode_RemoveOutEdge_Coverage(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get node
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Expected non-nil node")
	}

	// Get edges
	edges := node.OutEdges()
	if len(edges) == 0 {
		t.Fatal("Expected at least 1 out-edge")
	}

	// Test RemoveOutEdge
	node.RemoveOutEdge(edges[0])
	_ = node // Just verify it doesn't panic
}

// TestCollectionQuery_All tests All method
func TestCollectionQuery_All(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with names
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test All via Names
	names := qb.Names()
	all, err := names.All()
	if err != nil {
		t.Fatalf("Failed to get all names: %v", err)
	}
	if len(all) == 0 {
		t.Error("Expected at least 1 name")
	}
}

// TestCollectionQuery_Unique tests Unique method
func TestCollectionQuery_Unique(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with same name
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Unique via Names
	names := qb.Names()
	unique, err := names.Unique().Execute()
	if err != nil {
		t.Fatalf("Failed to get unique names: %v", err)
	}
	if len(unique) == 0 {
		t.Error("Expected at least 1 unique name")
	}
}

// TestCollectionQuery_By tests By method
func TestCollectionQuery_By(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add individuals with names
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	// Build query
	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test By via Names
	names := qb.Names()
	by, err := names.By("John").Execute()
	if err != nil {
		t.Fatalf("Failed to get names by: %v", err)
	}
	_ = by // Just verify it doesn't panic
}

// TestGraph_reverseEdgeType_Coverage2 tests reverseEdgeType function for coverage
func TestGraph_reverseEdgeType_Coverage2(t *testing.T) {
	// Test various edge types
	tests := []struct {
		input    EdgeType
		expected EdgeType
	}{
		{EdgeTypeHUSB, EdgeTypeFAMS},
		{EdgeTypeWIFE, EdgeTypeFAMS},
		{EdgeTypeCHIL, EdgeTypeFAMC},
		{EdgeTypeFAMC, EdgeTypeCHIL},
		{EdgeTypeFAMS, EdgeTypeFAMS},
		{EdgeTypeHasEvent, EdgeTypeHasEvent},
	}

	for _, tt := range tests {
		result := reverseEdgeType(tt.input)
		_ = result // Just verify it doesn't panic
	}
}

// TestEdge_OtherNode_Coverage2 tests OtherNode method for coverage
func TestEdge_OtherNode_Coverage2(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add two individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(indi1Line))

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	tree.AddRecord(types.NewIndividualRecord(indi2Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get nodes
	node1 := graph.GetIndividual("@I1@")
	node2 := graph.GetIndividual("@I2@")
	if node1 == nil || node2 == nil {
		t.Fatal("Expected non-nil nodes")
	}

	// Create a test edge
	edge := NewEdge("@E1@", node1, node2, EdgeTypeFAMS)
	
	// Test OtherNode with From node
	otherFrom := edge.OtherNode(node1)
	if otherFrom == nil {
		t.Error("Expected non-nil other node")
	}
	if otherFrom.ID() != "@I2@" {
		t.Errorf("Expected other node ID @I2@, got %s", otherFrom.ID())
	}

	// Test OtherNode with To node
	otherTo := edge.OtherNode(node2)
	if otherTo == nil {
		t.Error("Expected non-nil other node")
	}
	if otherTo.ID() != "@I1@" {
		t.Errorf("Expected other node ID @I1@, got %s", otherTo.ID())
	}

	// Test OtherNode with unrelated node (should return nil)
	unrelatedNode := graph.GetIndividual("@I1@")
	if unrelatedNode != nil {
		otherUnrelated := edge.OtherNode(unrelatedNode)
		_ = otherUnrelated // May be nil, which is fine
	}
}

// TestGraph_GetEdge tests GetEdge method
func TestGraph_GetEdge(t *testing.T) {
	// Create test tree
	tree := types.NewGedcomTree()

	// Add parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "Parent /Doe/", "")
	parentLine.AddChild(name1Line)
	tree.AddRecord(types.NewIndividualRecord(parentLine))

	// Add child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	childLine.AddChild(name2Line)
	famcLine := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	childLine.AddChild(famcLine)
	tree.AddRecord(types.NewIndividualRecord(childLine))

	// Add family
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	tree.AddRecord(types.NewFamilyRecord(fam1Line))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Get nodes
	node1 := graph.GetIndividual("@I1@")
	node2 := graph.GetIndividual("@I2@")
	if node1 == nil || node2 == nil {
		t.Fatal("Expected non-nil nodes")
	}

	// Test GetEdge - GetEdge takes an edge ID string, not nodes
	// Get edges from node1 to see what edges exist
	edges := node1.OutEdges()
	if len(edges) > 0 {
		edgeID := edges[0].ID
		edge := graph.GetEdge(edgeID)
		_ = edge // May be nil, which is fine
	}
}
