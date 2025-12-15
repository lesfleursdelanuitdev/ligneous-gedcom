package gedcom

import "fmt"

// IndividualRecord represents an Individual (INDI) record with domain-specific methods.
type IndividualRecord struct {
	*BaseRecord
}

// NewIndividualRecord creates a new IndividualRecord from a GedcomLine.
func NewIndividualRecord(line *GedcomLine) *IndividualRecord {
	return &IndividualRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetName returns the primary name value.
// Returns the first NAME value found.
func (ir *IndividualRecord) GetName() string {
	return ir.GetValue("NAME")
}

// GetNames returns all name values (multiple NAME tags allowed).
func (ir *IndividualRecord) GetNames() []string {
	return ir.GetValues("NAME")
}

// GetGivenName returns the given name from the first NAME record.
func (ir *IndividualRecord) GetGivenName() string {
	return ir.GetValue("NAME.GIVN")
}

// GetSurname returns the surname from the first NAME record.
func (ir *IndividualRecord) GetSurname() string {
	return ir.GetValue("NAME.SURN")
}

// GetSex returns the sex value (M, F, U, etc.).
func (ir *IndividualRecord) GetSex() string {
	return ir.GetValue("SEX")
}

// GetBirthDate returns the birth date.
func (ir *IndividualRecord) GetBirthDate() string {
	return ir.GetValue("BIRT.DATE")
}

// GetBirthPlace returns the birth place.
func (ir *IndividualRecord) GetBirthPlace() string {
	return ir.GetValue("BIRT.PLAC")
}

// GetDeathDate returns the death date.
func (ir *IndividualRecord) GetDeathDate() string {
	return ir.GetValue("DEAT.DATE")
}

// GetDeathPlace returns the death place.
func (ir *IndividualRecord) GetDeathPlace() string {
	return ir.GetValue("DEAT.PLAC")
}

// GetBirthData returns a map with birth date, place, and sources.
func (ir *IndividualRecord) GetBirthData() map[string]interface{} {
	return map[string]interface{}{
		"date":    ir.GetBirthDate(),
		"place":   ir.GetBirthPlace(),
		"sources": ir.GetValues("BIRT.SOUR"),
	}
}

// GetDeathData returns a map with death date, place, and sources.
func (ir *IndividualRecord) GetDeathData() map[string]interface{} {
	return map[string]interface{}{
		"date":    ir.GetDeathDate(),
		"place":   ir.GetDeathPlace(),
		"sources": ir.GetValues("DEAT.SOUR"),
	}
}

// GetFamiliesAsSpouse returns all family xrefs where this individual is a spouse (FAMS).
func (ir *IndividualRecord) GetFamiliesAsSpouse() []string {
	return ir.GetValues("FAMS")
}

// GetFamiliesAsChild returns all family xrefs where this individual is a child (FAMC).
func (ir *IndividualRecord) GetFamiliesAsChild() []string {
	return ir.GetValues("FAMC")
}

// GetOccupation returns the occupation value.
func (ir *IndividualRecord) GetOccupation() string {
	return ir.GetValue("OCCU")
}

// GetEvents returns all event records (BIRT, DEAT, MARR, etc.).
// Returns a slice of maps with type, date, place, and description.
func (ir *IndividualRecord) GetEvents() []map[string]interface{} {
	eventTags := []string{"BIRT", "DEAT", "BURI", "CREM", "CHR", "BAPM", "BARM", "BASM",
		"BLES", "CHRA", "CONF", "FCOM", "ORDN", "NATU", "EMIG", "IMMI", "CENS",
		"PROB", "WILL", "GRAD", "RETI", "EVEN", "MARR", "DIV", "RESI", "OCCU", "EDUC"}

	events := make([]map[string]interface{}, 0)
	for _, tag := range eventTags {
		eventLines := ir.GetLines(tag)
		for _, line := range eventLines {
			event := map[string]interface{}{
				"type":        tag,
				"date":        line.GetValue("DATE"),
				"place":       line.GetValue("PLAC"),
				"description": line.Value,
			}
			events = append(events, event)
		}
	}
	return events
}

// GetAttributes returns all attribute records (CAST, DSCR, EDUC, NATI, OCCU, PROP, RELI, RESI, TITL).
func (ir *IndividualRecord) GetAttributes() []map[string]interface{} {
	attributeTags := []string{"CAST", "DSCR", "EDUC", "NATI", "OCCU", "PROP", "RELI", "RESI", "TITL"}

	attributes := make([]map[string]interface{}, 0)
	for _, tag := range attributeTags {
		attributeLines := ir.GetLines(tag)
		for _, line := range attributeLines {
			attribute := map[string]interface{}{
				"type":  tag,
				"value": line.Value,
				"date":  line.GetValue("DATE"),
				"place": line.GetValue("PLAC"),
			}
			attributes = append(attributes, attribute)
		}
	}
	return attributes
}

// GetNotes returns all note xrefs.
func (ir *IndividualRecord) GetNotes() []string {
	return ir.GetValues("NOTE")
}

// GetSources returns all source xrefs.
func (ir *IndividualRecord) GetSources() []string {
	return ir.GetValues("SOUR")
}

// GetBirthDateParsed returns the birth date as a parsed GedcomDate.
// Returns error if date string is empty or cannot be parsed.
func (ir *IndividualRecord) GetBirthDateParsed() (*GedcomDate, error) {
	dateStr := ir.GetBirthDate()
	if dateStr == "" {
		return nil, fmt.Errorf("no birth date found")
	}
	return ParseDate(dateStr)
}

// GetDeathDateParsed returns the death date as a parsed GedcomDate.
// Returns error if date string is empty or cannot be parsed.
func (ir *IndividualRecord) GetDeathDateParsed() (*GedcomDate, error) {
	dateStr := ir.GetDeathDate()
	if dateStr == "" {
		return nil, fmt.Errorf("no death date found")
	}
	return ParseDate(dateStr)
}

// GetBirthPlaceParsed returns the birth place as a parsed GedcomPlace.
// Returns error if place string is empty or cannot be parsed.
func (ir *IndividualRecord) GetBirthPlaceParsed() (*GedcomPlace, error) {
	placeStr := ir.GetBirthPlace()
	if placeStr == "" {
		return nil, fmt.Errorf("no birth place found")
	}
	return ParsePlace(placeStr)
}

// GetDeathPlaceParsed returns the death place as a parsed GedcomPlace.
// Returns error if place string is empty or cannot be parsed.
func (ir *IndividualRecord) GetDeathPlaceParsed() (*GedcomPlace, error) {
	placeStr := ir.GetDeathPlace()
	if placeStr == "" {
		return nil, fmt.Errorf("no death place found")
	}
	return ParsePlace(placeStr)
}
