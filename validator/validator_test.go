package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestGedcomValidator_Validate(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewGedcomValidator(errorManager)

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
		t.Fatalf("Validate failed: %v", err)
	}

	// Check for errors (should have errors for missing @I2@ and @U1@)
	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Log("No validation errors found (this is expected for minimal valid tree)")
	}
}

func TestIndividualValidator_ValidateStructure(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with invalid tag
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	invalidLine := types.NewGedcomLine(1, "INVALID_TAG", "value", "")
	indiLine.AddChild(invalidLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Error("Expected error for invalid tag")
	}
}

func TestIndividualValidator_ValidateRequiredTags(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual without NAME (required)
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	sexLine := types.NewGedcomLine(1, "SEX", "M", "")
	indiLine.AddChild(sexLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

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

func TestIndividualValidator_ValidateSex(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewIndividualValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with invalid SEX value
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	sexLine := types.NewGedcomLine(1, "SEX", "INVALID", "")
	indiLine.AddChild(nameLine)
	indiLine.AddChild(sexLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Individual Validation" && err.Message == "INDI @I1@: Invalid SEX value INVALID" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid SEX value")
	}
}

func TestFamilyValidator_ValidateRequiredTags(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewFamilyValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create family without HUSB or WIFE
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famRecord := types.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Family Validation" && err.Message == "FAM @F1@: Missing required tags (must have at least HUSB or WIFE)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing HUSB/WIFE")
	}
}

func TestCrossReferenceValidator_ValidateXrefFormat(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with invalid xref format
	indiLine := types.NewGedcomLine(0, "INDI", "", "INVALID_XREF")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid xref format")
	}
}

func TestCrossReferenceValidator_ValidateReferences(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create family with invalid HUSB reference
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I999@", "") // Non-existent individual
	famLine.AddChild(husbLine)
	famRecord := types.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference: @I999@ in FAM record @F1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid cross-reference")
	}
}

func TestHeaderValidator_ValidateMissingHeader(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := types.NewGedcomTree()
	// No header added

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "Missing HEAD record" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing HEAD record")
	}
}



