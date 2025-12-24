package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHierarchicalParser_GetTree(t *testing.T) {
	parser := NewHierarchicalParser()

	// Create a temporary GEDCOM file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.ged")
	gedcomContent := "0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 TRLR\n"
	os.WriteFile(tmpFile, []byte(gedcomContent), 0644)

	tree, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test GetTree
	retrievedTree := parser.GetTree()
	if retrievedTree == nil {
		t.Error("GetTree() returned nil")
	}

	// Verify it's the same tree
	if retrievedTree != tree {
		t.Error("GetTree() returned different tree instance")
	}

	// Verify tree content
	header := retrievedTree.GetHeader()
	if header == nil {
		t.Error("Expected header in tree")
	}
}

func TestBasicParser_GetTree(t *testing.T) {
	parser := NewBasicParser()

	// Create a temporary GEDCOM file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.ged")
	gedcomContent := "0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 TRLR\n"
	os.WriteFile(tmpFile, []byte(gedcomContent), 0644)

	tree, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test GetTree
	retrievedTree := parser.GetTree()
	if retrievedTree == nil {
		t.Error("GetTree() returned nil")
	}

	// Verify it's the same tree
	if retrievedTree != tree {
		t.Error("GetTree() returned different tree instance")
	}

	// Verify tree content
	header := retrievedTree.GetHeader()
	if header == nil {
		t.Error("Expected header in tree")
	}
}

func TestHierarchicalParser_ErrorMethods(t *testing.T) {
	parser := NewHierarchicalParser()

	// Initially no errors
	if parser.HasErrors() {
		t.Error("Expected no errors initially")
	}
	if parser.HasSevereErrors() {
		t.Error("Expected no severe errors initially")
	}

	errors := parser.GetErrors()
	if len(errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errors))
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

	// Check error manager
	errorManager := parser.GetErrorManager()
	if errorManager == nil {
		t.Error("Expected error manager")
	}

	// After parsing, may have errors
	errors = parser.GetErrors()
	hasErrors := parser.HasErrors()
	hasSevereErrors := parser.HasSevereErrors()

	// Log for debugging
	t.Logf("Errors: %d, HasErrors: %v, HasSevereErrors: %v", len(errors), hasErrors, hasSevereErrors)
}

func TestHierarchicalParser_EmptyInput(t *testing.T) {
	parser := NewHierarchicalParser()

	// Create empty file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.ged")
	os.WriteFile(tmpFile, []byte{}, 0644)

	tree, err := parser.Parse(tmpFile)
	if err == nil {
		t.Error("Expected error for empty input")
	}
	if tree != nil {
		t.Error("Expected nil tree for empty input")
	}
}

func TestHierarchicalParser_OnlyTrailer(t *testing.T) {
	parser := NewHierarchicalParser()

	// Create file with only trailer
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "trailer.ged")
	os.WriteFile(tmpFile, []byte("0 TRLR\n"), 0644)

	tree, err := parser.Parse(tmpFile)
	// May or may not error depending on implementation
	if err == nil {
		// If no error, tree should exist
		if tree == nil {
			t.Error("Expected tree even with only trailer")
		}
	}
}

func TestHierarchicalParser_MultipleRecords(t *testing.T) {
	parser := NewHierarchicalParser()

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
0 TRLR
`
	os.WriteFile(tmpFile, []byte(gedcomContent), 0644)

	tree, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify multiple individuals
	individuals := tree.GetAllIndividuals()
	if len(individuals) != 2 {
		t.Errorf("Expected 2 individuals, got %d", len(individuals))
	}

	// Verify GetTree returns same tree
	retrievedTree := parser.GetTree()
	if retrievedTree != tree {
		t.Error("GetTree() returned different tree")
	}
}
