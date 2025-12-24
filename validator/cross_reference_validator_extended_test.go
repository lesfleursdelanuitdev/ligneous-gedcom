package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestCrossReferenceValidator_ValidateIndividualReferences_FAMC(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with non-existent FAMC reference
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	famcLine := types.NewGedcomLine(1, "FAMC", "@F999@", "") // Non-existent family
	indiLine.AddChild(nameLine)
	indiLine.AddChild(famcLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference: @F999@ in INDI record @I1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMC cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateIndividualReferences_MultipleFAMS(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create individual with multiple FAMS references (one valid, one invalid)
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	fams1 := types.NewGedcomLine(1, "FAMS", "@F1@", "")
	fams2 := types.NewGedcomLine(1, "FAMS", "@F999@", "") // Non-existent
	indiLine.AddChild(nameLine)
	indiLine.AddChild(fams1)
	indiLine.AddChild(fams2)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Create valid family
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famRecord := types.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference: @F999@ in INDI record @I1@" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid FAMS cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateFamilyReferences_WIFE(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create family with non-existent WIFE reference
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := types.NewGedcomLine(1, "WIFE", "@I999@", "") // Non-existent
	famLine.AddChild(husbLine)
	famLine.AddChild(wifeLine)
	famRecord := types.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Create valid husband
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

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
		t.Error("Expected error for invalid WIFE cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateFamilyReferences_CHIL(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create family with non-existent CHIL reference
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	chil1 := types.NewGedcomLine(1, "CHIL", "@I2@", "")      // Valid
	chil2 := types.NewGedcomLine(1, "CHIL", "@I999@", "")   // Non-existent
	famLine.AddChild(husbLine)
	famLine.AddChild(chil1)
	famLine.AddChild(chil2)
	famRecord := types.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	// Create valid husband and one child
	husbIndiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	husbNameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	husbIndiLine.AddChild(husbNameLine)
	husbIndiRecord := types.NewIndividualRecord(husbIndiLine)
	tree.AddRecord(husbIndiRecord)

	chilIndiLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	chilNameLine := types.NewGedcomLine(1, "NAME", "Child /Doe/", "")
	chilIndiLine.AddChild(chilNameLine)
	chilIndiRecord := types.NewIndividualRecord(chilIndiLine)
	tree.AddRecord(chilIndiRecord)

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
		t.Error("Expected error for invalid CHIL cross-reference")
	}
}

func TestCrossReferenceValidator_ValidateXrefIDs_InvalidFormat(t *testing.T) {
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
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference ID: INVALID_XREF in INDI record" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid xref ID format")
	}
}

func TestCrossReferenceValidator_ValidateXrefIDs_FamilyInvalidFormat(t *testing.T) {
	errorManager := types.NewErrorManager()
	validator := NewCrossReferenceValidator(errorManager)

	tree := types.NewGedcomTree()

	// Create family with invalid xref format
	famLine := types.NewGedcomLine(0, "FAM", "", "INVALID_XREF")
	husbLine := types.NewGedcomLine(1, "HUSB", "@I1@", "")
	famLine.AddChild(husbLine)
	famRecord := types.NewFamilyRecord(famLine)
	tree.AddRecord(famRecord)

	validator.Validate(tree)

	errors := errorManager.Errors()
	found := false
	for _, err := range errors {
		if err.Context == "Cross-Reference Validation" && err.Message == "Invalid cross-reference ID: INVALID_XREF in FAM record" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error for invalid family xref ID format")
	}
}


