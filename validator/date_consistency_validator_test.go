package validator

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestDateConsistencyValidator_DeathBeforeBirth(t *testing.T) {
	// Create individual with death before birth (Severe error)
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := types.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := types.NewGedcomLine(2, "DATE", "15 JAN 1900", "")
	deatLine := types.NewGedcomLine(1, "DEAT", "", "")
	deatDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before birth!

	birtLine.AddChild(birtDateLine)
	deatLine.AddChild(deatDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(deatLine)

	indi := types.NewIndividualRecord(indiLine)
	tree := types.NewGedcomTree()
	tree.AddRecord(indi)

	errorManager := types.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	// Should have at least one severe error
	severeFound := false
	for _, err := range errors {
		if err.Severity == types.SeveritySevere {
			severeFound = true
			if err.Message == "" {
				t.Error("Error message should not be empty")
			}
		}
	}

	if !severeFound {
		t.Error("Expected severe error for death before birth")
	}
}

func TestDateConsistencyValidator_MarriageBeforeBirth(t *testing.T) {
	// Create individual and family with marriage before birth
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := types.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := types.NewGedcomLine(2, "DATE", "15 JAN 1900", "")
	birtLine.AddChild(birtDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(types.NewGedcomLine(1, "FAMS", "@F1@", ""))

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before birth!
	marrLine.AddChild(marrDateLine)
	famLine.AddChild(marrLine)
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))

	indi := types.NewIndividualRecord(indiLine)
	fam := types.NewFamilyRecord(famLine)

	tree := types.NewGedcomTree()
	tree.AddRecord(indi)
	tree.AddRecord(fam)

	errorManager := types.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	severeFound := false
	for _, err := range errors {
		if err.Severity == types.SeveritySevere {
			severeFound = true
		}
	}

	if !severeFound {
		t.Error("Expected severe error for marriage before birth")
	}
}

func TestDateConsistencyValidator_YoungMarriageAge(t *testing.T) {
	// Create individual with very young marriage age (Warning)
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := types.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1900", "")
	birtLine.AddChild(birtDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(types.NewGedcomLine(1, "FAMS", "@F1@", ""))

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1905", "") // Age 5!
	marrLine.AddChild(marrDateLine)
	famLine.AddChild(marrLine)
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))

	indi := types.NewIndividualRecord(indiLine)
	fam := types.NewFamilyRecord(famLine)

	tree := types.NewGedcomTree()
	tree.AddRecord(indi)
	tree.AddRecord(fam)

	errorManager := types.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()
	config.MinMarriageAge = 12

	errors := validator.Validate(tree, config)

	warningFound := false
	for _, err := range errors {
		if err.Severity == types.SeverityWarning {
			warningFound = true
		}
	}

	if !warningFound {
		t.Error("Expected warning for very young marriage age")
	}
}

func TestDateConsistencyValidator_OldParentAge(t *testing.T) {
	// Create family with very old parent (Warning)
	fatherLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	fatherBirtLine := types.NewGedcomLine(1, "BIRT", "", "")
	fatherBirtDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1800", "")
	fatherBirtLine.AddChild(fatherBirtDateLine)
	fatherLine.AddChild(fatherBirtLine)

	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	childBirtLine := types.NewGedcomLine(1, "BIRT", "", "")
	childBirtDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1900", "") // Father would be 100
	childBirtLine.AddChild(childBirtDateLine)
	childLine.AddChild(childBirtLine)
	childLine.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))

	father := types.NewIndividualRecord(fatherLine)
	child := types.NewIndividualRecord(childLine)
	fam := types.NewFamilyRecord(famLine)

	tree := types.NewGedcomTree()
	tree.AddRecord(father)
	tree.AddRecord(child)
	tree.AddRecord(fam)

	errorManager := types.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()
	config.MaxParentAge = 80

	errors := validator.Validate(tree, config)

	warningFound := false
	for _, err := range errors {
		if err.Severity == types.SeverityWarning {
			warningFound = true
		}
	}

	if !warningFound {
		t.Error("Expected warning for very old parent age")
	}
}

func TestDateConsistencyValidator_MissingBirthDate(t *testing.T) {
	// Create individual without birth date (Info)
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "John /Doe/", ""))

	indi := types.NewIndividualRecord(indiLine)
	tree := types.NewGedcomTree()
	tree.AddRecord(indi)

	errorManager := types.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	infoFound := false
	for _, err := range errors {
		if err.Severity == types.SeverityInfo {
			infoFound = true
		}
	}

	if !infoFound {
		t.Error("Expected info-level error for missing birth date")
	}
}

func TestDateConsistencyValidator_DivorceBeforeMarriage(t *testing.T) {
	// Create family with divorce before marriage (Severe)
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1900", "")
	divLine := types.NewGedcomLine(1, "DIV", "", "")
	divDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before marriage!

	marrLine.AddChild(marrDateLine)
	divLine.AddChild(divDateLine)
	famLine.AddChild(marrLine)
	famLine.AddChild(divLine)

	fam := types.NewFamilyRecord(famLine)
	tree := types.NewGedcomTree()
	tree.AddRecord(fam)

	errorManager := types.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	severeFound := false
	for _, err := range errors {
		if err.Severity == types.SeveritySevere {
			severeFound = true
		}
	}

	if !severeFound {
		t.Error("Expected severe error for divorce before marriage")
	}
}

func TestDateConsistencyValidator_ChildBeforeParentBirth(t *testing.T) {
	// Create family where child is born before parent (Severe)
	fatherLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	fatherBirtLine := types.NewGedcomLine(1, "BIRT", "", "")
	fatherBirtDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1900", "")
	fatherBirtLine.AddChild(fatherBirtDateLine)
	fatherLine.AddChild(fatherBirtLine)

	childLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	childBirtLine := types.NewGedcomLine(1, "BIRT", "", "")
	childBirtDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before father!
	childBirtLine.AddChild(childBirtDateLine)
	childLine.AddChild(childBirtLine)
	childLine.AddChild(types.NewGedcomLine(1, "FAMC", "@F1@", ""))

	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "CHIL", "@I2@", ""))

	father := types.NewIndividualRecord(fatherLine)
	child := types.NewIndividualRecord(childLine)
	fam := types.NewFamilyRecord(famLine)

	tree := types.NewGedcomTree()
	tree.AddRecord(father)
	tree.AddRecord(child)
	tree.AddRecord(fam)

	errorManager := types.NewErrorManager()
	validator := NewDateConsistencyValidator(errorManager)
	config := NewValidationConfig()

	errors := validator.Validate(tree, config)

	severeFound := false
	for _, err := range errors {
		if err.Severity == types.SeveritySevere {
			severeFound = true
		}
	}

	if !severeFound {
		t.Error("Expected severe error for child born before parent")
	}
}

func TestAdvancedValidator_Integration(t *testing.T) {
	// Test the AdvancedValidator with DateConsistencyValidator
	errorManager := types.NewErrorManager()
	advancedValidator := NewAdvancedValidator(errorManager)

	// Add date consistency rule
	dateRule := NewDateConsistencyValidator(errorManager)
	advancedValidator.AddRule(dateRule)

	// Create a tree with date issues
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := types.NewGedcomLine(1, "BIRT", "", "")
	birtDateLine := types.NewGedcomLine(2, "DATE", "15 JAN 1900", "")
	deatLine := types.NewGedcomLine(1, "DEAT", "", "")
	deatDateLine := types.NewGedcomLine(2, "DATE", "1 JAN 1800", "") // Before birth!

	birtLine.AddChild(birtDateLine)
	deatLine.AddChild(deatDateLine)
	indiLine.AddChild(birtLine)
	indiLine.AddChild(deatLine)

	indi := types.NewIndividualRecord(indiLine)
	tree := types.NewGedcomTree()
	tree.AddRecord(indi)

	// Validate
	err := advancedValidator.Validate(tree)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Check that errors were collected
	errors := errorManager.Errors()
	if len(errors) == 0 {
		t.Error("Expected errors to be collected")
	}
}
