package validator

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestParallelIndividualValidator_ValidateReferences(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with invalid FAMS reference format
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	famsLine := types.NewGedcomLine(1, "FAMS", "INVALID_FORMAT", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(famsLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid FAMS reference format INVALID_FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMS reference format")
	}
}

func TestParallelIndividualValidator_ValidateReferences_FAMC(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with invalid FAMC reference format
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	famcLine := types.NewGedcomLine(1, "FAMC", "INVALID_FORMAT", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(famcLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid FAMC reference format INVALID_FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMC reference format")
	}
}

func TestParallelIndividualValidator_ValidateReferences_MultipleFAMS(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with multiple FAMS references
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	fams1 := types.NewGedcomLine(1, "FAMS", "@F1@", "")
	fams2 := types.NewGedcomLine(1, "FAMS", "@F2@", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(fams1)
	indiLine.AddChild(fams2)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should validate format (but not existence in parallel validator)
	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid FAMS reference format @F1@" {
			t.Error("Should not error for valid FAMS format")
		}
	}
}

func TestParallelIndividualValidator_ValidateStructure_InvalidTag(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with invalid tag
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	invalidLine := types.NewGedcomLine(1, "INVALID_TAG", "value", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(invalidLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid tag INVALID_TAG" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid tag")
	}
}

func TestParallelIndividualValidator_ValidateStructure_UserDefinedTag(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with user-defined tag (should not error)
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	userTag := types.NewGedcomLine(1, "_CUSTOM", "value", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(userTag)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid tag _CUSTOM" {
			t.Error("Should not error for user-defined tags")
		}
	}
}

func TestParallelIndividualValidator_ValidateStructure_MissingRequiredTag(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual without NAME (required)
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	sexLine := types.NewGedcomLine(1, "SEX", "M", "")
	indiLine.AddChild(sexLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Missing required tag NAME" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing NAME tag")
	}
}

func TestParallelIndividualValidator_ConcurrentValidation(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewParallelIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create many individuals to test concurrent processing
	for i := 1; i <= 50; i++ {
		xrefID := fmt.Sprintf("@I%d@", i)
		indiLine := types.NewGedcomLine(0, "INDI", "", xrefID)
		nameLine := types.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", i), "")
		indiLine.AddChild(nameLine)
		indiRecord := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indiRecord)
	}

	err := validator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should complete without panics
	individuals := tree.GetAllIndividuals()
	if len(individuals) != 50 {
		t.Errorf("Expected 50 individuals, got %d", len(individuals))
	}

	t.Logf("Successfully validated %d individuals concurrently", len(individuals))
}

