package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestIndividualNode_Spouses tests the Spouses method on IndividualNode.
func TestIndividualNode_Spouses(t *testing.T) {
	// Create a tree with individuals and families
	tree := types.NewGedcomTree()

	// Create individual 1
	indi1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1 := types.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Create individual 2 (spouse)
	indi2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	indi2Line.AddChild(types.NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	indi2 := types.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Create family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	family := types.NewFamilyRecord(famLine)
	tree.AddRecord(family)

	// Add FAMS to individual 1
	indi1Line.AddChild(types.NewGedcomLine(1, "FAMS", "@F1@", ""))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph() failed: %v", err)
	}

	// Test Spouses via node method
	node1 := graph.GetIndividual("@I1@")
	if node1 == nil {
		t.Fatal("GetIndividual() returned nil")
	}

	spouses := node1.Spouses()
	if len(spouses) != 1 {
		t.Errorf("Expected 1 spouse, got %d", len(spouses))
	}
	if spouses[0].ID() != "@I2@" {
		t.Errorf("Expected spouse @I2@, got %s", spouses[0].ID())
	}

	// Test Spouses via graph convenience method
	spouses2, err := graph.GetSpouses("@I1@")
	if err != nil {
		t.Fatalf("GetSpouses() returned error: %v", err)
	}
	if len(spouses2) != 1 {
		t.Errorf("Expected 1 spouse, got %d", len(spouses2))
	}
	if spouses2[0].ID() != "@I2@" {
		t.Errorf("Expected spouse @I2@, got %s", spouses2[0].ID())
	}

	// Test with individual not in graph
	_, err = graph.GetSpouses("@I99@")
	if err == nil {
		t.Error("GetSpouses() should return error when individual not found")
	}
}

// TestIndividualNode_Children tests the Children method on IndividualNode.
func TestIndividualNode_Children(t *testing.T) {
	// Create a tree with individuals and families
	tree := types.NewGedcomTree()

	// Create parent
	parentLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	parentLine.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	parent := types.NewIndividualRecord(parentLine)
	tree.AddRecord(parent)

	// Create children
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	child1Line.AddChild(types.NewGedcomLine(1, "NAME", "Child1 /Doe/", ""))
	child1 := types.NewIndividualRecord(child1Line)
	tree.AddRecord(child1)

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	child2Line.AddChild(types.NewGedcomLine(1, "NAME", "Child2 /Doe/", ""))
	child2 := types.NewIndividualRecord(child2Line)
	tree.AddRecord(child2)

	// Create family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	family := types.NewFamilyRecord(famLine)
	tree.AddRecord(family)

	// Add FAMS to parent
	parentLine.AddChild(types.NewGedcomLine(1, "FAMS", "@F1@", ""))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph() failed: %v", err)
	}

	// Test Children via node method
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("GetIndividual() returned nil")
	}

	children := node.Children()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}

	// Test Children via graph convenience method
	children2, err := graph.GetChildren("@I1@")
	if err != nil {
		t.Fatalf("GetChildren() returned error: %v", err)
	}
	if len(children2) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children2))
	}
}

// TestIndividualNode_Parents tests the Parents method on IndividualNode.
func TestIndividualNode_Parents(t *testing.T) {
	// Create a tree with individuals and families
	tree := types.NewGedcomTree()

	// Create parents
	parent1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	parent1Line.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	parent1 := types.NewIndividualRecord(parent1Line)
	tree.AddRecord(parent1)

	parent2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	parent2Line.AddChild(types.NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	parent2 := types.NewIndividualRecord(parent2Line)
	tree.AddRecord(parent2)

	// Create child
	childLine := types.NewGedcomLine(0, "INDI", "", "@I3@")
	childLine.AddChild(types.NewGedcomLine(1, "NAME", "Child /Doe/", ""))
	child := types.NewIndividualRecord(childLine)
	tree.AddRecord(child)

	// Create family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	family := types.NewFamilyRecord(famLine)
	tree.AddRecord(family)

	// Add FAMC to child
	childLine.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph() failed: %v", err)
	}

	// Test Parents via node method
	node := graph.GetIndividual("@I3@")
	if node == nil {
		t.Fatal("GetIndividual() returned nil")
	}

	parents := node.Parents()
	if len(parents) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(parents))
	}

	// Test Parents via graph convenience method
	parents2, err := graph.GetParents("@I3@")
	if err != nil {
		t.Fatalf("GetParents() returned error: %v", err)
	}
	if len(parents2) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(parents2))
	}
}

// TestIndividualNode_Siblings tests the Siblings method on IndividualNode.
func TestIndividualNode_Siblings(t *testing.T) {
	// Create a tree with individuals and families
	tree := types.NewGedcomTree()

	// Create parents
	parent1Line := types.NewGedcomLine(0, "INDI", "", "@I1@")
	parent1 := types.NewIndividualRecord(parent1Line)
	tree.AddRecord(parent1)

	parent2Line := types.NewGedcomLine(0, "INDI", "", "@I2@")
	parent2 := types.NewIndividualRecord(parent2Line)
	tree.AddRecord(parent2)

	// Create siblings
	child1Line := types.NewGedcomLine(0, "INDI", "", "@I3@")
	child1 := types.NewIndividualRecord(child1Line)
	tree.AddRecord(child1)

	child2Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	child2 := types.NewIndividualRecord(child2Line)
	tree.AddRecord(child2)

	child3Line := types.NewGedcomLine(0, "INDI", "", "@I5@")
	child3 := types.NewIndividualRecord(child3Line)
	tree.AddRecord(child3)

	// Create family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I3@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I4@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I5@", ""))
	family := types.NewFamilyRecord(famLine)
	tree.AddRecord(family)

	// Add FAMC to children
	child1Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	child2Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))
	child3Line.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph() failed: %v", err)
	}

	// Test Siblings via node method (should not include self)
	node := graph.GetIndividual("@I3@")
	if node == nil {
		t.Fatal("GetIndividual() returned nil")
	}

	siblings := node.Siblings()
	if len(siblings) != 2 {
		t.Errorf("Expected 2 siblings, got %d", len(siblings))
	}
	// Verify self is not included
	for _, sib := range siblings {
		if sib.ID() == "@I3@" {
			t.Error("Siblings() should not include self")
		}
	}

	// Test Siblings via graph convenience method
	siblings2, err := graph.GetSiblings("@I3@")
	if err != nil {
		t.Fatalf("GetSiblings() returned error: %v", err)
	}
	if len(siblings2) != 2 {
		t.Errorf("Expected 2 siblings, got %d", len(siblings2))
	}
}


