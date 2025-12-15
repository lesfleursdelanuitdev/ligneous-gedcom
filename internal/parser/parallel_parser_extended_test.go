package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParallelHierarchicalParser_GetTree(t *testing.T) {
	parser := NewParallelHierarchicalParser()

	// Create a temporary GEDCOM file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.ged")
	gedcomContent := "0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 TRLR\n"
	os.WriteFile(tmpFile, []byte(gedcomContent), 0644)

	tree, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test GetTree (if it exists)
	// Note: Parallel parser may not have GetTree, check implementation
	if tree == nil {
		t.Error("Expected tree from Parse")
	}

	// Verify tree content
	header := tree.GetHeader()
	if header == nil {
		t.Error("Expected header in tree")
	}
}

func TestParallelHierarchicalParser_ErrorMethods(t *testing.T) {
	parser := NewParallelHierarchicalParser()

	// Test error methods
	errorManager := parser.GetErrorManager()
	if errorManager == nil {
		t.Error("Expected error manager")
	}

	// Initially no errors
	hasErrors := parser.HasErrors()
	if hasErrors {
		t.Error("Expected no errors initially")
	}

	errors := parser.GetErrors()
	if len(errors) != 0 {
		t.Errorf("Expected 0 errors initially, got %d", len(errors))
	}

	// Parse invalid GEDCOM to generate errors
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.ged")
	gedcomContent := "0 HEAD\n1 INVALID_TAG\n0 TRLR\n"
	os.WriteFile(tmpFile, []byte(gedcomContent), 0644)

	_, err := parser.Parse(tmpFile)
	if err != nil {
		// Parser may return error or continue
	}

	// Check error methods after parsing
	hasErrors = parser.HasErrors()
	errors = parser.GetErrors()

	t.Logf("After parsing: HasErrors=%v, ErrorCount=%d", hasErrors, len(errors))
}

func TestParallelHierarchicalParser_MultipleRecords(t *testing.T) {
	parser := NewParallelHierarchicalParser()

	// Create file with multiple records
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "multi.ged")
	gedcomContent := `0 HEAD
1 GEDC
2 VERS 5.5.5
0 @I1@ INDI
1 NAME John /Doe/
0 @I2@ INDI
1 NAME Jane /Doe/
0 @F1@ FAM
1 HUSB @I1@
1 WIFE @I2@
0 TRLR
`
	os.WriteFile(tmpFile, []byte(gedcomContent), 0644)

	tree, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify multiple records
	individuals := tree.GetAllIndividuals()
	if len(individuals) != 2 {
		t.Errorf("Expected 2 individuals, got %d", len(individuals))
	}

	families := tree.GetAllFamilies()
	if len(families) != 1 {
		t.Errorf("Expected 1 family, got %d", len(families))
	}
}

func TestParallelHierarchicalParser_ErrorHandling(t *testing.T) {
	parser := NewParallelHierarchicalParser()

	// Create file with error conditions
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "error.ged")
	gedcomContent := `0 HEAD
1 GEDC
2 VERS 5.5.5
0 @I1@ INDI
1 NAME John /Doe/
1 INVALID_TAG value
0 TRLR
`
	os.WriteFile(tmpFile, []byte(gedcomContent), 0644)

	tree, err := parser.Parse(tmpFile)
	// Should continue parsing despite errors
	if err != nil && tree == nil {
		t.Log("Parser returned error (may be expected)")
	}

	// Check error manager
	errorManager := parser.GetErrorManager()
	if errorManager == nil {
		t.Error("Expected error manager")
	}

	errors := parser.GetErrors()
	hasErrors := parser.HasErrors()

	t.Logf("Error handling: HasErrors=%v, ErrorCount=%d", hasErrors, len(errors))
}

