package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestHeaderValidator_ValidateGedc_MissingGEDC(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create header without GEDC
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Missing GEDC tag" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for missing GEDC tag")
	}
}

func TestHeaderValidator_ValidateGedc_MissingVERS(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create header with GEDC but without VERS
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	gedcLine := types.NewGedcomLine(1, "GEDC", "", "")
	headerLine.AddChild(gedcLine)
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Missing GEDC.VERS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for missing GEDC.VERS")
	}
}

func TestHeaderValidator_ValidateStructure_InvalidTag(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create header with invalid tag
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	invalidLine := types.NewGedcomLine(1, "INVALID_TAG", "value", "")
	headerLine.AddChild(invalidLine)
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Invalid tag INVALID_TAG" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning for invalid header tag")
	}
}

func TestHeaderValidator_ValidateStructure_UserDefinedTag(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create header with user-defined tag (should not error)
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	userTag := types.NewGedcomLine(1, "_CUSTOM", "value", "")
	gedcLine := types.NewGedcomLine(1, "GEDC", "", "")
	versLine := types.NewGedcomLine(2, "VERS", "5.5.5", "")
	gedcLine.AddChild(versLine)
	headerLine.AddChild(userTag)
	headerLine.AddChild(gedcLine)
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Header Validation" && err.Message == "HEAD: Invalid tag _CUSTOM" {
			t.Error("Should not error for user-defined tags")
		}
	}
}

func TestHeaderValidator_ValidateGedc_ValidStructure(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewHeaderValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create header with valid GEDC structure
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	gedcLine := types.NewGedcomLine(1, "GEDC", "", "")
	versLine := types.NewGedcomLine(2, "VERS", "5.5.5", "")
	formLine := types.NewGedcomLine(2, "FORM", "LINEAGE-LINKED", "")
	gedcLine.AddChild(versLine)
	gedcLine.AddChild(formLine)
	headerLine.AddChild(gedcLine)
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	for _, err := range errors {
		if err.Context == "Header Validation" && (err.Message == "HEAD: Missing GEDC tag" || err.Message == "HEAD: Missing GEDC.VERS") {
			t.Error("Should not error for valid GEDC structure")
		}
	}
}


