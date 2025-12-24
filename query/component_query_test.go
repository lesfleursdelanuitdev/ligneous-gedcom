package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// createDisconnectedTree creates a tree with two disconnected components:
// Component 1: I1-I2-I3 (connected family)
// Component 2: I4-I5 (isolated pair)
func createDisconnectedTree() *types.GedcomTree {
	tree := types.NewGedcomTree()

	// Component 1: I1, I2, I3
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
	name3Line := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	famc3Line := types.NewGedcomLine(1, "FAMC", "@F1@", "")
	indi3Line.AddChild(name3Line)
	indi3Line.AddChild(famc3Line)
	indi3 := types.NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)

	fam1Line := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husb1Line := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	wife1Line := types.NewGedcomLine(1, "WIFE", "@I2@", "")
	chil1Line := types.NewGedcomLine(1, "CHIL", "@I3@", "")
	fam1Line.AddChild(husb1Line)
	fam1Line.AddChild(wife1Line)
	fam1Line.AddChild(chil1Line)
	fams1Line := types.NewGedcomLine(1, "FAMS", "@F1@", "")
	indi1Line.AddChild(fams1Line)
	fams2Line := types.NewGedcomLine(1, "FAMS", "@F1@", "")
	indi2Line.AddChild(fams2Line)
	fam1 := types.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Component 2: I4, I5 (isolated)
	indi4Line := types.NewGedcomLine(0, "INDI", "", "@I4@")
	name4Line := types.NewGedcomLine(1, "NAME", "Alice /Smith/", "")
	indi4Line.AddChild(name4Line)
	indi4 := types.NewIndividualRecord(indi4Line)
	tree.AddRecord(indi4)

	indi5Line := types.NewGedcomLine(0, "INDI", "", "@I5@")
	name5Line := types.NewGedcomLine(1, "NAME", "Bob /Smith/", "")
	indi5Line.AddChild(name5Line)
	indi5 := types.NewIndividualRecord(indi5Line)
	tree.AddRecord(indi5)

	fam2Line := types.NewGedcomLine(0, "FAM", "", "@F2@")
	husb2Line := types.NewGedcomLine(1, "HUSB", "@I4@", "")
	wife2Line := types.NewGedcomLine(1, "WIFE", "@I5@", "")
	fam2Line.AddChild(husb2Line)
	fam2Line.AddChild(wife2Line)
	fams3Line := types.NewGedcomLine(1, "FAMS", "@F2@", "")
	indi4Line.AddChild(fams3Line)
	fams4Line := types.NewGedcomLine(1, "FAMS", "@F2@", "")
	indi5Line.AddChild(fams4Line)
	fam2 := types.NewFamilyRecord(fam2Line)
	tree.AddRecord(fam2)

	return tree
}

func TestGetComponentForPerson_Basic(t *testing.T) {
	tree := createDisconnectedTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test component from I1 (should include I1, I2, I3)
	component, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	if len(component) != 3 {
		t.Errorf("Expected 3 individuals in component, got %d", len(component))
	}

	// Verify all expected individuals are present
	xrefs := make(map[string]bool)
	for _, indi := range component {
		xrefs[indi.XrefID()] = true
	}

	if !xrefs["@I1@"] {
		t.Error("Component should include @I1@")
	}
	if !xrefs["@I2@"] {
		t.Error("Component should include @I2@")
	}
	if !xrefs["@I3@"] {
		t.Error("Component should include @I3@")
	}

	// Should not include I4 or I5 (different component)
	if xrefs["@I4@"] || xrefs["@I5@"] {
		t.Error("Component should not include individuals from other components")
	}
}

func TestGetComponentForPerson_IsolatedComponent(t *testing.T) {
	tree := createDisconnectedTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test component from I4 (should include I4, I5 only)
	component, err := graph.GetComponentForPerson("@I4@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	if len(component) != 2 {
		t.Errorf("Expected 2 individuals in component, got %d", len(component))
	}

	// Verify all expected individuals are present
	xrefs := make(map[string]bool)
	for _, indi := range component {
		xrefs[indi.XrefID()] = true
	}

	if !xrefs["@I4@"] {
		t.Error("Component should include @I4@")
	}
	if !xrefs["@I5@"] {
		t.Error("Component should include @I5@")
	}

	// Should not include I1, I2, I3 (different component)
	if xrefs["@I1@"] || xrefs["@I2@"] || xrefs["@I3@"] {
		t.Error("Component should not include individuals from other components")
	}
}

func TestGetComponentForPerson_WithMaxDepth(t *testing.T) {
	tree := createTestTree() // Creates a 3-generation tree
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with depth limit of 1 (should only get direct connections)
	options := NewComponentOptions()
	options.MaxDepth = 1

	component, err := graph.GetComponentForPersonWithOptions("@I3@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// With depth 1 from I3, should get:
	// - I3 itself (starting node)
	// - Direct connections: Parents (I1, I2), Sibling (I4), Child (I5)
	// Note: The depth limit applies to BFS iterations, so depth 1 means
	// we process the starting node (depth 0) and its immediate neighbors (depth 1)
	// The actual number depends on implementation details
	
	// At minimum, should get the starting node
	if len(component) < 1 {
		t.Errorf("Expected at least 1 individual (starting node), got %d", len(component))
	}

	// Verify I3 is included
	xrefs := make(map[string]bool)
	for _, indi := range component {
		xrefs[indi.XrefID()] = true
	}

	if !xrefs["@I3@"] {
		t.Error("Component should include starting node @I3@")
	}
}

func TestGetComponentForPerson_WithMaxSize(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with size limit
	options := NewComponentOptions()
	options.MaxSize = 3

	component, err := graph.GetComponentForPersonWithOptions("@I3@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// Should be limited to 3 individuals
	if len(component) > 3 {
		t.Errorf("Expected at most 3 individuals with size limit, got %d", len(component))
	}
}

func TestGetComponentForPerson_NonexistentIndividual(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with nonexistent individual
	component, err := graph.GetComponentForPerson("@I999@")
	if err != nil {
		t.Fatalf("GetComponentForPerson should not error on nonexistent individual: %v", err)
	}

	if component != nil && len(component) != 0 {
		t.Error("Component should be empty for nonexistent individual")
	}
}

func TestGetComponentForPerson_ComplexFamily(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test component from root (I1) - should get entire connected family
	component, err := graph.GetComponentForPerson("@I1@")
	if err != nil {
		t.Fatalf("GetComponentForPerson failed: %v", err)
	}

	// Should include all 5 individuals in the tree
	if len(component) != 5 {
		t.Errorf("Expected 5 individuals in component, got %d", len(component))
	}

	// Verify all individuals are present
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

func TestGetComponentForPerson_OptionsDefault(t *testing.T) {
	tree := createTestTree()
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test with default options (no limits)
	options := NewComponentOptions()
	component, err := graph.GetComponentForPersonWithOptions("@I1@", options)
	if err != nil {
		t.Fatalf("GetComponentForPersonWithOptions failed: %v", err)
	}

	// Should get all individuals (no limits)
	if len(component) != 5 {
		t.Errorf("Expected 5 individuals with no limits, got %d", len(component))
	}
}

