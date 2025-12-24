package validator

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// HeaderValidator validates Header (HEAD) record.
type HeaderValidator struct {
	*BaseValidator
	validTags map[string]bool
}

// NewHeaderValidator creates a new HeaderValidator.
func NewHeaderValidator(errorManager *types.ErrorManager) *HeaderValidator {
	validTags := map[string]bool{
		"GEDC": true, "CHAR": true, "SOUR": true, "DATE": true, "TIME": true, "FILE": true,
		"LANG": true, "SUBM": true, "SUBN": true, "COPR": true, "DEST": true, "NOTE": true,
	}

	return &HeaderValidator{
		BaseValidator: NewBaseValidator(errorManager),
		validTags:     validTags,
	}
}

// Validate validates the header record.
func (hv *HeaderValidator) Validate(tree *types.GedcomTree) error {
	header := tree.GetHeader()
	if header == nil {
		hv.AddError(types.SeveritySevere,
			"Missing HEAD record",
			0,
			"Header Validation")
		return nil
	}

	hv.validateStructure(header)
	hv.validateGedc(header)
	return nil
}

// validateStructure validates the structure of the header.
func (hv *HeaderValidator) validateStructure(header types.Record) {
	firstLine := header.FirstLine()

	for _, lines := range firstLine.Children {
		for _, line := range lines {
			if !hv.validTags[line.Tag] && !isUserDefinedTag(line.Tag) {
				hv.AddError(types.SeverityWarning,
					"HEAD: Invalid tag "+line.Tag,
					line.LineNumber,
					"Header Validation")
			}
		}
	}
}

// validateGedc validates the GEDC (GEDCOM) structure.
func (hv *HeaderValidator) validateGedc(header types.Record) {
	gedcLines := header.GetLines("GEDC")
	if len(gedcLines) == 0 {
		hv.AddError(types.SeveritySevere,
			"HEAD: Missing GEDC tag",
			header.FirstLine().LineNumber,
			"Header Validation")
		return
	}

	// Validate VERS under GEDC
	versValue := header.GetValue("GEDC.VERS")
	if versValue == "" {
		hv.AddError(types.SeverityWarning,
			"HEAD: Missing GEDC.VERS",
			header.FirstLine().LineNumber,
			"Header Validation")
	}
}
