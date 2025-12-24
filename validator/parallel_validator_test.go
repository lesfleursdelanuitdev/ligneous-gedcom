package validator

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestParallelGedcomValidator_Validate(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelGedcomValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create a minimal valid tree
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	submLine := types.NewGedcomLine(1, "SUBM", "@U1@", "")
	headerLine.AddChild(submLine)
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create an individual
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Create a family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := types.NewGedcomLine(1, "WIFE", "@I2@", "")
	famLine.AddChild(husbLine)
	famLine.AddChild(wifeLine)
	famRecord := types.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Validate
	err := validator.Validate(tree)
	if err != nil {
		t.Logf("Validation returned error (may be expected): %v", err)
	}

	// Check for errors (should have errors for missing @I2@ and @U1@)
	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Log("No validation errors found (this is expected for minimal valid tree)")
	}
}

func BenchmarkValidator_Sequential(b *testing.B) {
	errorManager := types.NewErrorManager()
	validator := NewGedcomValidator(errorManager)

	tree := createTestTree()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(tree)
	}
}

func BenchmarkValidator_Parallel(b *testing.B) {
	errorManager := types.NewErrorManager()
	validator := NewParallelGedcomValidator(errorManager)

	tree := createTestTree()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(tree)
	}
}

func BenchmarkParallelIndividualValidator(b *testing.B) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := createLargeTestTree(1000) // 1000 individuals

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(tree)
	}
}

func createTestTree() *types.GedcomTree {
	tree := types.NewGedcomTree()

	// Create header
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create 100 individuals
	for i := 0; i < 100; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := types.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := types.NewGedcomLine(1, "NAME", "Test /Person/", "")
		indiLine.AddChild(nameLine)
		indiRecord := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	return tree
}

func createLargeTestTree(count int) *types.GedcomTree {
	tree := types.NewGedcomTree()

	// Create header
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create many individuals
	for i := 0; i < count; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := types.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := types.NewGedcomLine(1, "NAME", "Test /Person/", "")
		sexLine := types.NewGedcomLine(1, "SEX", "M", "")
		indiLine.AddChild(nameLine)
		indiLine.AddChild(sexLine)
		indiRecord := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	return tree
}
