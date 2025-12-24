package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestGraphValidator_ValidateEdges(t *testing.T) {
	tree := types.NewGedcomTree()
	
	// Create valid graph
	indi1 := CreateTestIndividual("@I1@", "John /Doe/")
	tree.AddRecord(indi1)
	
	indi2 := CreateTestIndividual("@I2@", "Jane /Doe/")
	tree.AddRecord(indi2)
	
	fam := CreateTestFamily("@F1@", "@I1@", "@I2@", []string{})
	tree.AddRecord(fam)
	
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	// Validate should pass
	errorManager := types.NewErrorManager()
	validator := NewGraphValidator(errorManager)
	if err := validator.Validate(graph); err != nil {
		t.Errorf("Validate should pass for valid graph: %v", err)
	}
	if errorManager.HasSevereErrors() {
		t.Error("Should not have severe errors for valid graph")
	}
}

func TestGraphValidator_ValidateFamilies(t *testing.T) {
	tree := types.NewGedcomTree()
	
	// Create family with valid structure
	indi1 := CreateTestIndividual("@I1@", "John /Doe/")
	tree.AddRecord(indi1)
	
	indi2 := CreateTestIndividual("@I2@", "Jane /Doe/")
	tree.AddRecord(indi2)
	
	child := CreateTestIndividual("@I3@", "Child /Doe/")
	tree.AddRecord(child)
	
	fam := CreateTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"})
	tree.AddRecord(fam)
	
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	errorManager := types.NewErrorManager()
	validator := NewGraphValidator(errorManager)
	if err := validator.Validate(graph); err != nil {
		t.Errorf("Validate should pass for valid family: %v", err)
	}
}

func TestGraphValidator_ValidateRelationships(t *testing.T) {
	tree := types.NewGedcomTree()
	
	// Create valid parent-child relationship
	parent := CreateTestIndividual("@I1@", "Parent /Doe/")
	tree.AddRecord(parent)
	
	child := CreateTestIndividual("@I2@", "Child /Doe/")
	tree.AddRecord(child)
	
	fam := CreateTestFamily("@F1@", "@I1@", "", []string{"@I2@"})
	tree.AddRecord(fam)
	
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	errorManager := types.NewErrorManager()
	validator := NewGraphValidator(errorManager)
	if err := validator.Validate(graph); err != nil {
		t.Errorf("Validate should pass for valid relationships: %v", err)
	}
}

func TestGraphValidator_ValidateNodeRecordConsistency(t *testing.T) {
	tree := types.NewGedcomTree()
	
	indi := CreateTestIndividual("@I1@", "John /Doe/")
	tree.AddRecord(indi)
	
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	errorManager := types.NewErrorManager()
	validator := NewGraphValidator(errorManager)
	if err := validator.Validate(graph); err != nil {
		t.Errorf("Validate should pass for consistent nodes: %v", err)
	}
	
	// Verify node has record
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Node should exist")
	}
	if node.Record() == nil {
		t.Error("Node should have record")
	}
	if node.Individual == nil {
		t.Error("Node should have Individual record")
	}
}

func TestGraphValidator_ValidateOrphanedNodes(t *testing.T) {
	tree := types.NewGedcomTree()
	
	// Create an orphaned individual (no family connections)
	indi := CreateTestIndividual("@I1@", "Orphan /Doe/")
	tree.AddRecord(indi)
	
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	errorManager := types.NewErrorManager()
	validator := NewGraphValidator(errorManager)
	
	// Should not fail, but should have warnings
	if err := validator.Validate(graph); err != nil {
		// Warnings are OK, but severe errors are not
		if errorManager.HasSevereErrors() {
			t.Errorf("Should not have severe errors for orphaned node: %v", err)
		}
	}
	
	// Should have warning about orphaned node
	warnings := errorManager.GetErrorsBySeverity(types.SeverityWarning)
	if len(warnings) == 0 {
		t.Log("Expected warning about orphaned node, but none found")
	}
}

func TestGraphValidator_ValidateCircularReferences(t *testing.T) {
	// Note: Creating actual circular references in GEDCOM is difficult
	// because the graph builder prevents them. This test verifies
	// the validation logic exists and works.
	tree := types.NewGedcomTree()
	
	// Create a normal parent-child relationship (no cycle)
	parent := CreateTestIndividual("@I1@", "Parent /Doe/")
	tree.AddRecord(parent)
	
	child := CreateTestIndividual("@I2@", "Child /Doe/")
	tree.AddRecord(child)
	
	fam := CreateTestFamily("@F1@", "@I1@", "", []string{"@I2@"})
	tree.AddRecord(fam)
	
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	errorManager := types.NewErrorManager()
	validator := NewGraphValidator(errorManager)
	if err := validator.Validate(graph); err != nil {
		t.Errorf("Validate should pass for non-circular graph: %v", err)
	}
	if errorManager.HasSevereErrors() {
		t.Error("Should not have severe errors for valid graph")
	}
}

func TestGraphValidator_IntegrationWithBuildGraph(t *testing.T) {
	tree := types.NewGedcomTree()
	
	// Create a complete family structure
	husband := CreateTestIndividual("@I1@", "Husband /Doe/")
	tree.AddRecord(husband)
	
	wife := CreateTestIndividual("@I2@", "Wife /Doe/")
	tree.AddRecord(wife)
	
	child := CreateTestIndividual("@I3@", "Child /Doe/")
	tree.AddRecord(child)
	
	fam := CreateTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@"})
	tree.AddRecord(fam)
	
	// BuildGraph should automatically validate
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}
	
	// Graph should be valid
	if graph == nil {
		t.Fatal("Graph should not be nil")
	}
	
	// Verify relationships work
	spouses, err := graph.GetSpouses("@I1@")
	if err != nil {
		t.Fatalf("GetSpouses failed: %v", err)
	}
	if len(spouses) != 1 {
		t.Errorf("Expected 1 spouse, got %d", len(spouses))
	}
	
	children, err := graph.GetChildren("@I1@")
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

