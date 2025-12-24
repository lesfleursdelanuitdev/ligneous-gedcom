package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// createTestTree creates a simple family tree for testing:
// I1 (John) and I2 (Jane) are parents
// I3 (Child1) and I4 (Child2) are their children
// I5 (Grandchild) is child of I3
func createTestTree() *types.GedcomTree {
	tree := types.NewGedcomTree()

	// I1: John (parent)
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	sex1Line := types.NewGedcomLine(1, "SEX", "M", "")
	indi1Line.AddChild(name1Line)
	indi1Line.AddChild(sex1Line)
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// I2: Jane (parent)
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := types.NewGedcomLine(1, "NAME", "Jane /Doe/", "")
	sex2Line := types.NewGedcomLine(1, "SEX", "F", "")
	indi2Line.AddChild(name2Line)
	indi2Line.AddChild(sex2Line)
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// I3: Child1 (child of I1 and I2)
	indi3Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	name3Line := types.NewGedcomLine(1, "NAME", "Child1 /Doe/", "")
	sex3Line := types.NewGedcomLine(1, "SEX", "M", "")
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	indi3Line.AddChild(name3Line)
	indi3Line.AddChild(sex3Line)
	indi3Line.AddChild(famc3Line)
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	// I4: Child2 (child of I1 and I2)
	indi4Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Child2 /Doe/", "")
	sex4Line := types.NewGedcomLine(1, "SEX", "F", "")
	famc4Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	indi4Line.AddChild(name4Line)
	indi4Line.AddChild(sex4Line)
	indi4Line.AddChild(famc4Line)
	indi4 := types.NewIndividualRecord(indi4Line)
	tree.AddRecord(indi4)

	// I5: Grandchild (child of I3)
	indi5Line := types.NewGedcomLine(0, "INDI", "", "@I5@")
	name5Line := types.NewGedcomLine(1, "NAME", "Grandchild /Doe/", "")
	sex5Line := types.NewGedcomLine(1, "SEX", "M", "")
	famc5Line := types.NewGedcomLine(1, "FAMC", "@F2@", "")
	indi5Line.AddChild(name5Line)
	indi5Line.AddChild(sex5Line)
	indi5Line.AddChild(famc5Line)
	indi5 := types.NewIndividualRecord(indi5Line)
	tree.AddRecord(indi5)

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

	// F2: Family (I3, child: I5)
	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	husb2Line := types.NewGedcomLine(1, "HUSB", "@I3@", "")
	chil3Line := types.NewGedcomLine(1, "CHIL", "@I5@", "")
	fam2Line.AddChild(husb2Line)
	fam2Line.AddChild(chil3Line)
	fams3Line := types.NewGedcomLine(1, "FAMS", "@F2@", "")
	indi3Line.AddChild(fams3Line)
	fam2 := types.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	return tree
}

func TestSubtreeQuery_Basic(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test basic subtree query from I3 (should include I1, I2, I3, I4, I5)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Root == nil {
		t.Fatal("Root is nil")
	}

	if result.Root.XrefID() != "@I3@" {
		t.Errorf("Expected root @I3@, got %s", result.Root.XrefID())
	}

	// Should have ancestors (I1, I2)
	if len(result.Ancestors) != 2 {
		t.Errorf("Expected 2 ancestors, got %d", len(result.Ancestors))
	}

	// Should have descendants (I5)
	if len(result.Descendants) != 1 {
		t.Errorf("Expected 1 descendant, got %d", len(result.Descendants))
	}

	// Total should be: root (1) + ancestors (2) + descendants (1) = 4
	if len(result.All) != 4 {
		t.Errorf("Expected 4 total individuals, got %d", len(result.All))
	}
}

func TestSubtreeQuery_WithSiblings(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test subtree query with siblings from I3
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSiblings().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have sibling (I4)
	if len(result.Siblings) != 1 {
		t.Errorf("Expected 1 sibling, got %d", len(result.Siblings))
	}

	if result.Siblings[0].XrefID() != "@I4@" {
		t.Errorf("Expected sibling @I4@, got %s", result.Siblings[0].XrefID())
	}

	// Total should include sibling: root (1) + ancestors (2) + descendants (1) + siblings (1) = 5
	if len(result.All) != 5 {
		t.Errorf("Expected 5 total individuals (with sibling), got %d", len(result.All))
	}
}

func TestSubtreeQuery_WithSpouses(t *testing.T) {
	tree := createTestTree()

	// Add a spouse for I3
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

	// Test subtree query with spouses from I3
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		IncludeSpouses().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have spouse (I6)
	if len(result.Spouses) != 1 {
		t.Errorf("Expected 1 spouse, got %d", len(result.Spouses))
	}

	if result.Spouses[0].XrefID() != "@I6@" {
		t.Errorf("Expected spouse @I6@, got %s", result.Spouses[0].XrefID())
	}
}

func TestSubtreeQuery_GenerationLimits(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with limited ancestor generations (1 = only direct parents)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Should have 2 ancestors (I1 and I2, both direct parents)
	if len(result.Ancestors) != 2 {
		t.Errorf("Expected 2 ancestors with limit 1 (direct parents), got %d", len(result.Ancestors))
	}

	// Should still have descendants
	if len(result.Descendants) != 1 {
		t.Errorf("Expected 1 descendant, got %d", len(result.Descendants))
	}
}

func TestSubtreeQuery_ExcludeSelf(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test excluding self
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		ExcludeSelf().
		Execute()

	if err != nil {
		t.Fatalf("Subtree query failed: %v", err)
	}

	// Root should still be set
	if result.Root == nil {
		t.Fatal("Root should not be nil even when excluded")
	}

	// But root should not be in All list
	found := false
	for _, indi := range result.All {
		if indi.XrefID() == "@I3@" {
			found = true
			break
		}
	}

	if found {
		t.Error("Root should not be in All list when excluded")
	}

	// Total should be: ancestors (2) + descendants (1) = 3
	if len(result.All) != 3 {
		t.Errorf("Expected 3 total individuals (excluding self), got %d", len(result.All))
	}
}

func TestSubtreeQuery_Count(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test Count method
	count, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Count()

	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 4 {
		t.Errorf("Expected count 4, got %d", count)
	}
}

func TestSubtreeQuery_ExecuteRecords(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test ExecuteRecords convenience method
	records, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		ExecuteRecords()

	if err != nil {
		t.Fatalf("ExecuteRecords failed: %v", err)
	}

	if len(records) != 4 {
		t.Errorf("Expected 4 records, got %d", len(records))
	}
}

func TestSubtreeQuery_WithFilter(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with filter (only males)
	result, err := q.Individual("@I3@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		IncludeSelf().
		Filter(func(indi *types.IndividualRecord) bool {
			return indi.GetSex() == "M"
		}).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query with filter failed: %v", err)
	}

	// Should filter out females (I2, I4)
	// Expected: I1 (M), I3 (M), I5 (M) = 3
	if len(result.All) != 3 {
		t.Errorf("Expected 3 filtered records (males only), got %d", len(result.All))
	}
}

func TestSubtreeQuery_NonexistentIndividual(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	q := NewQueryFromGraph(graph)

	// Test with nonexistent individual
	result, err := q.Individual("@I999@").GetSubtree().
		AncestorGenerations(1).
		DescendantGenerations(1).
		Execute()

	if err != nil {
		t.Fatalf("Subtree query should not error on nonexistent individual: %v", err)
	}

	if result != nil {
		t.Error("Result should be nil for nonexistent individual")
	}
}
