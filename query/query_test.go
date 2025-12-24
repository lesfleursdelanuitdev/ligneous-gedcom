package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestQueryBuilder_Individual(t *testing.T) {
	tree := types.NewGedcomTree()

	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	indiQuery := query.Individual("@I1@")
	if indiQuery == nil {
		t.Fatal("Expected IndividualQuery to be created")
	}

	if indiQuery.xrefID != "@I1@" {
		t.Errorf("Expected xrefID @I1@, got %s", indiQuery.xrefID)
	}
}

func TestIndividualQuery_Parents(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create parent
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	parents, err := query.Individual("@I2@").Parents()
	if err != nil {
		t.Fatalf("Failed to get parents: %v", err)
	}

	if len(parents) == 0 {
		t.Error("Expected at least one parent")
	}

	found := false
	for _, parent := range parents {
		if parent.XrefID() == "@I1@" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected @I1@ to be a parent")
	}
}

func TestIndividualQuery_Ancestors(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create grandparent
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create parent
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create child
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family 1: I1 is parent of I2
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3
	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2 := types.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	ancestors, err := query.Individual("@I3@").Ancestors().Execute()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	if len(ancestors) < 2 {
		t.Errorf("Expected at least 2 ancestors, got %d", len(ancestors))
	}

	// Check that I1 and I2 are ancestors
	foundI1 := false
	foundI2 := false
	for _, ancestor := range ancestors {
		if ancestor.XrefID() == "@I1@" {
			foundI1 = true
		}
		if ancestor.XrefID() == "@I2@" {
			foundI2 = true
		}
	}

	if !foundI1 {
		t.Error("Expected @I1@ to be an ancestor")
	}
	if !foundI2 {
		t.Error("Expected @I2@ to be an ancestor")
	}
}

func TestAncestorQuery_MaxGenerations(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create 3-generation family
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family 1: I1 is parent of I2
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3
	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2 := types.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Limit to 1 generation
	ancestors, err := query.Individual("@I3@").Ancestors().MaxGenerations(1).Execute()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	// Should only get I2 (parent), not I1 (grandparent)
	if len(ancestors) != 1 {
		t.Errorf("Expected 1 ancestor with MaxGenerations(1), got %d", len(ancestors))
	}

	if len(ancestors) > 0 && ancestors[0].XrefID() != "@I2@" {
		t.Errorf("Expected @I2@, got %s", ancestors[0].XrefID())
	}
}

func TestIndividualQuery_Descendants(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create parent
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create grandchild
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F2@", ""))
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family 1: I1 is parent of I2
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Family 2: I2 is parent of I3
	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(types.NewGedcomLine(1, "HUSB", "@I2@", ""))
	fam2Line.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	fam2 := types.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	descendants, err := query.Individual("@I1@").Descendants().Execute()
	if err != nil {
		t.Fatalf("Failed to get descendants: %v", err)
	}

	if len(descendants) < 2 {
		t.Errorf("Expected at least 2 descendants, got %d", len(descendants))
	}
}

func TestIndividualQuery_RelationshipTo(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create parent
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	result, err := query.Individual("@I1@").RelationshipTo("@I2@").Execute()
	if err != nil {
		t.Fatalf("Failed to calculate relationship: %v", err)
	}

	if !result.IsDirect {
		t.Error("Expected direct relationship")
	}

	if result.RelationshipType != "parent" {
		t.Errorf("Expected relationship type 'parent', got '%s'", result.RelationshipType)
	}
}

func TestIndividualQuery_PathTo(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create parent
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	path, err := query.Individual("@I1@").PathTo("@I2@").Shortest()
	if err != nil {
		t.Fatalf("Failed to find path: %v", err)
	}

	if path == nil {
		t.Fatal("Expected path to be found")
	}

	if path.Length == 0 {
		t.Error("Expected path length > 0")
	}
}

func TestFilterQuery_ByName(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Jane /Smith/", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByName("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_BySex(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "SEX", "M", ""))
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "SEX", "F", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().BySex("M").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestFilterQuery_HasChildren(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create parent
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create individual without children
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().HasChildren().Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_MultipleFilters(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1Line.AddChild(types.NewGedcomLine(1, "SEX", "M", ""))
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	indi2Line.AddChild(types.NewGedcomLine(1, "SEX", "F", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Filter by name "John" AND sex "M"
	results, err := query.Filter().ByName("John").BySex("M").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Filter by name "John" AND sex "F" (should return 0)
	results2, err := query.Filter().ByName("John").BySex("F").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results2) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results2))
	}
}

func TestAncestorQuery_IncludeSelf(t *testing.T) {
	tree := types.NewGedcomTree()

	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	ancestors, err := query.Individual("@I1@").Ancestors().IncludeSelf().Execute()
	if err != nil {
		t.Fatalf("Failed to get ancestors: %v", err)
	}

	found := false
	for _, ancestor := range ancestors {
		if ancestor.XrefID() == "@I1@" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected @I1@ to be included when IncludeSelf() is called")
	}
}

func TestPathQuery_Shortest(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	path, err := query.Individual("@I1@").PathTo("@I2@").Shortest()
	if err != nil {
		t.Fatalf("Failed to find shortest path: %v", err)
	}

	if path == nil {
		t.Fatal("Expected path to be found")
	}

	if path.Length == 0 {
		t.Error("Expected path length > 0")
	}
}

func TestPathQuery_All(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	paths, err := query.Individual("@I1@").PathTo("@I2@").All()
	if err != nil {
		t.Fatalf("Failed to find paths: %v", err)
	}

	if len(paths) == 0 {
		t.Error("Expected at least one path")
	}
}

func TestAncestorQuery_Count(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	count, err := query.Individual("@I2@").Ancestors().Count()
	if err != nil {
		t.Fatalf("Failed to count ancestors: %v", err)
	}

	if count == 0 {
		t.Error("Expected at least one ancestor")
	}
}

func TestFilterQuery_Count(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	count, err := query.Filter().ByName("John").Count()
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	// Should be 0 since no individuals named "John"
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestFilterQuery_ByNameExact(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Smith/", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByNameExact("John /Doe/").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_ByNameStarts(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Jane /Smith/", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByNameStarts("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_ByNameEnds(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Jane /Smith/", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByNameEnds("/Doe/").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_NoChildren(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create parent
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create child
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create individual without children
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().NoChildren().Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 2 { // indi2 and indi3
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestFilterQuery_NoSpouse(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create individuals
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// Family (I1 and I2 are spouses)
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().NoSpouse().Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 { // Only I3
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I3@" {
		t.Errorf("Expected @I3@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_ByBirthYear(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1 := types.NewGedcomLine(1, "BIRT", "", "")
	birt1.AddChild(types.NewGedcomLine(2, "DATE", "1800", ""))
	indi1Line.AddChild(birt1)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	birt2 := types.NewGedcomLine(1, "BIRT", "", "")
	birt2.AddChild(types.NewGedcomLine(2, "DATE", "1850", ""))
	indi2Line.AddChild(birt2)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByBirthYear(1800).Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_ByBirthDateBefore(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1 := types.NewGedcomLine(1, "BIRT", "", "")
	birt1.AddChild(types.NewGedcomLine(2, "DATE", "1800", ""))
	indi1Line.AddChild(birt1)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	birt2 := types.NewGedcomLine(1, "BIRT", "", "")
	birt2.AddChild(types.NewGedcomLine(2, "DATE", "1850", ""))
	indi2Line.AddChild(birt2)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByBirthDateBefore(1825).Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I1@" {
		t.Errorf("Expected @I1@, got %s", results[0].XrefID())
	}
}

func TestFilterQuery_ByBirthDateAfter(t *testing.T) {
	tree := types.NewGedcomTree()

	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1 := types.NewGedcomLine(1, "BIRT", "", "")
	birt1.AddChild(types.NewGedcomLine(2, "DATE", "1800", ""))
	indi1Line.AddChild(birt1)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	birt2 := types.NewGedcomLine(1, "BIRT", "", "")
	birt2.AddChild(types.NewGedcomLine(2, "DATE", "1850", ""))
	indi2Line.AddChild(birt2)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	results, err := query.Filter().ByBirthDateAfter(1825).Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].XrefID() != "@I2@" {
		t.Errorf("Expected @I2@, got %s", results[0].XrefID())
	}
}

func TestQueryBuilder_AllFamilies(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create families
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	fam2 := types.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	families, err := query.AllFamilies()
	if err != nil {
		t.Fatalf("Failed to get all families: %v", err)
	}

	if len(families) != 2 {
		t.Errorf("Expected 2 families, got %d", len(families))
	}
}

func TestQueryBuilder_AllEvents(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create individual with birth event
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1 := types.NewGedcomLine(1, "BIRT", "", "")
	birt1.AddChild(types.NewGedcomLine(2, "DATE", "1800", ""))
	birt1.AddChild(types.NewGedcomLine(2, "PLAC", "New York", ""))
	indi1Line.AddChild(birt1)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create family with marriage event
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marr1 := types.NewGedcomLine(1, "MARR", "", "")
	marr1.AddChild(types.NewGedcomLine(2, "DATE", "1820", ""))
	marr1.AddChild(types.NewGedcomLine(2, "PLAC", "Boston", ""))
	fam1Line.AddChild(marr1)
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	events, err := query.AllEvents()
	if err != nil {
		t.Fatalf("Failed to get all events: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected at least one event")
	}
}

func TestQueryBuilder_AllPlaces(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create individual with birth place
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birt1 := types.NewGedcomLine(1, "BIRT", "", "")
	birt1.AddChild(types.NewGedcomLine(2, "PLAC", "New York", ""))
	indi1Line.AddChild(birt1)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create individual with death place
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	deat2 := types.NewGedcomLine(1, "DEAT", "", "")
	deat2.AddChild(types.NewGedcomLine(2, "PLAC", "Boston", ""))
	indi2Line.AddChild(deat2)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create family with marriage place
	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marr1 := types.NewGedcomLine(1, "MARR", "", "")
	marr1.AddChild(types.NewGedcomLine(2, "PLAC", "Philadelphia", ""))
	fam1Line.AddChild(marr1)
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	places, err := query.AllPlaces()
	if err != nil {
		t.Fatalf("Failed to get all places: %v", err)
	}

	if len(places) < 3 {
		t.Errorf("Expected at least 3 places, got %d", len(places))
	}

	// Check that all expected places are present
	placesMap := make(map[string]bool)
	for _, place := range places {
		placesMap[place] = true
	}

	if !placesMap["New York"] {
		t.Error("Expected 'New York' to be in places")
	}
	if !placesMap["Boston"] {
		t.Error("Expected 'Boston' to be in places")
	}
	if !placesMap["Philadelphia"] {
		t.Error("Expected 'Philadelphia' to be in places")
	}
}

func TestQueryBuilder_UniqueNames(t *testing.T) {
	tree := types.NewGedcomTree()

	// Create individuals with different names (using GEDCOM format with /Surname/)
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Doe/", "")
	indi2Line.AddChild(name2Line)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "John /Smith/", "")
	indi3Line.AddChild(name3Line)
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	query, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	names, err := query.UniqueNames()
	if err != nil {
		t.Fatalf("Failed to get unique names: %v", err)
	}

	// Check given names
	givenNames := names["given"]
	if len(givenNames) < 2 {
		t.Errorf("Expected at least 2 unique given names, got %d", len(givenNames))
	}

	// Check surnames
	surnames := names["surname"]
	if len(surnames) < 2 {
		t.Errorf("Expected at least 2 unique surnames, got %d", len(surnames))
	}

	// Verify specific names
	givenNamesMap := make(map[string]bool)
	for _, name := range givenNames {
		givenNamesMap[name] = true
	}

	surnamesMap := make(map[string]bool)
	for _, name := range surnames {
		surnamesMap[name] = true
	}

	if !givenNamesMap["John"] {
		t.Error("Expected 'John' to be in given names")
	}
	if !givenNamesMap["Jane"] {
		t.Error("Expected 'Jane' to be in given names")
	}
	if !surnamesMap["Doe"] {
		t.Error("Expected 'Doe' to be in surnames")
	}
	if !surnamesMap["Smith"] {
		t.Error("Expected 'Smith' to be in surnames")
	}
}
