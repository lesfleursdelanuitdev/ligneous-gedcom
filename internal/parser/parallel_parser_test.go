package parser

import (
	"os"
	"testing"
)

func TestParallelHierarchicalParser_Parse(t *testing.T) {
	parser := NewParallelHierarchicalParser()

	// Create a simple test file content
	testContent := `0 HEAD
1 GEDC
2 VERS 5.5.5
0 @I1@ INDI
1 NAME John /Doe/
1 SEX M
0 @F1@ FAM
1 HUSB @I1@
0 TRLR
`

	// Write to temp file
	tmpFile := "/tmp/test_parallel.ged"
	err := writeTestFile(tmpFile, testContent)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	// Parse
	tree, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify structure
	individuals := tree.GetAllIndividuals()
	if len(individuals) != 1 {
		t.Errorf("Expected 1 individual, got %d", len(individuals))
	}

	families := tree.GetAllFamilies()
	if len(families) != 1 {
		t.Errorf("Expected 1 family, got %d", len(families))
	}

	// Verify individual
	indi := individuals["@I1@"]
	if indi == nil {
		t.Fatal("Individual @I1@ not found")
	}
	if indi.GetValue("NAME") != "John /Doe/" {
		t.Errorf("Expected name 'John /Doe/', got '%s'", indi.GetValue("NAME"))
	}
}

func BenchmarkParse_Sequential(b *testing.B) {
	file := "/apps/family-tree/gedcom/gracis.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		b.Skipf("File not found: %s", file)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewHierarchicalParser()
		_, err := parser.Parse(file)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

func BenchmarkParse_Parallel(b *testing.B) {
	file := "/apps/family-tree/gedcom/gracis.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		b.Skipf("File not found: %s", file)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParallelHierarchicalParser()
		_, err := parser.Parse(file)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

// Helper function
func writeTestFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

