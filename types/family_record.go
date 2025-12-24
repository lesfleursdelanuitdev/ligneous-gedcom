package types

import "fmt"

// FamilyRecord represents a Family (FAM) record with domain-specific methods.
type FamilyRecord struct {
	*BaseRecord
}

// NewFamilyRecord creates a new FamilyRecord from a GedcomLine.
func NewFamilyRecord(line *GedcomLine) *FamilyRecord {
	return &FamilyRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetHusband returns the husband's xref (HUSB).
func (fr *FamilyRecord) GetHusband() string {
	return fr.GetValue("HUSB")
}

// GetWife returns the wife's xref (WIFE).
func (fr *FamilyRecord) GetWife() string {
	return fr.GetValue("WIFE")
}

// GetChildren returns all children xrefs (CHIL).
func (fr *FamilyRecord) GetChildren() []string {
	return fr.GetValues("CHIL")
}

// GetMarriageDate returns the marriage date.
func (fr *FamilyRecord) GetMarriageDate() string {
	return fr.GetValue("MARR.DATE")
}

// GetMarriagePlace returns the marriage place.
func (fr *FamilyRecord) GetMarriagePlace() string {
	return fr.GetValue("MARR.PLAC")
}

// GetMarriageData returns a map with marriage date, place, and sources.
func (fr *FamilyRecord) GetMarriageData() map[string]interface{} {
	return map[string]interface{}{
		"date":    fr.GetMarriageDate(),
		"place":   fr.GetMarriagePlace(),
		"sources": fr.GetValues("MARR.SOUR"),
	}
}

// GetDivorceDate returns the divorce date.
func (fr *FamilyRecord) GetDivorceDate() string {
	return fr.GetValue("DIV.DATE")
}

// GetDivorcePlace returns the divorce place.
func (fr *FamilyRecord) GetDivorcePlace() string {
	return fr.GetValue("DIV.PLAC")
}

// GetDivorceData returns a map with divorce date, place, and sources.
func (fr *FamilyRecord) GetDivorceData() map[string]interface{} {
	return map[string]interface{}{
		"date":    fr.GetDivorceDate(),
		"place":   fr.GetDivorcePlace(),
		"sources": fr.GetValues("DIV.SOUR"),
	}
}

// GetEvents returns all family event records (MARR, DIV, ANUL, etc.).
func (fr *FamilyRecord) GetEvents() []map[string]interface{} {
	eventTags := []string{"MARR", "DIV", "ANUL", "CENS", "DIVF", "ENGA", "MARB", "MARC", "MARL", "MARS"}
	return extractEvents(fr, eventTags)
}

// GetNotes returns all note xrefs.
// Delegates to BaseRecord.GetNotes().
func (fr *FamilyRecord) GetNotes() []string {
	return fr.BaseRecord.GetNotes()
}

// GetSources returns all source xrefs.
// Delegates to BaseRecord.GetSources().
func (fr *FamilyRecord) GetSources() []string {
	return fr.BaseRecord.GetSources()
}

// GetMarriageDateParsed returns the marriage date as a parsed GedcomDate.
// Returns error if date string is empty or cannot be parsed.
func (fr *FamilyRecord) GetMarriageDateParsed() (*GedcomDate, error) {
	dateStr := fr.GetMarriageDate()
	if dateStr == "" {
		return nil, fmt.Errorf("no marriage date found")
	}
	return ParseDate(dateStr)
}

// GetDivorceDateParsed returns the divorce date as a parsed GedcomDate.
// Returns nil without error if date string is empty.
// Returns error only if date string cannot be parsed.
func (fr *FamilyRecord) GetDivorceDateParsed() (*GedcomDate, error) {
	dateStr := fr.GetDivorceDate()
	if dateStr == "" {
		return nil, nil
	}
	return ParseDate(dateStr)
}

// GetMarriagePlaceParsed returns the marriage place as a parsed GedcomPlace.
// Returns error if place string is empty or cannot be parsed.
func (fr *FamilyRecord) GetMarriagePlaceParsed() (*GedcomPlace, error) {
	placeStr := fr.GetMarriagePlace()
	if placeStr == "" {
		return nil, fmt.Errorf("no marriage place found")
	}
	return ParsePlace(placeStr)
}

// GetDivorcePlaceParsed returns the divorce place as a parsed GedcomPlace.
// Returns nil without error if place string is empty.
// Returns error only if place string cannot be parsed.
func (fr *FamilyRecord) GetDivorcePlaceParsed() (*GedcomPlace, error) {
	placeStr := fr.GetDivorcePlace()
	if placeStr == "" {
		return nil, nil
	}
	return ParsePlace(placeStr)
}

// ============================================================================
// Phase 2: Enhanced FamilyRecord Relationship Methods
// ============================================================================

// GetHusbandRecord returns the husband's IndividualRecord.
// Returns nil if no husband is specified or the husband record is not found.
// Returns an error if the record is not part of a tree.
func (fr *FamilyRecord) GetHusbandRecord() (*IndividualRecord, error) {
	husbandXref := fr.GetHusband()
	if husbandXref == "" {
		return nil, nil // No husband, but not an error
	}

	tree := fr.getTree()
	if tree == nil {
		return nil, fmt.Errorf("family record is not part of a tree")
	}

	husbandRecord := tree.GetIndividual(husbandXref)
	if husbandRecord == nil {
		return nil, nil // Husband not found, but not an error
	}

	husband, ok := husbandRecord.(*IndividualRecord)
	if !ok {
		return nil, fmt.Errorf("husband record %s is not an IndividualRecord", husbandXref)
	}

	return husband, nil
}

// GetWifeRecord returns the wife's IndividualRecord.
// Returns nil if no wife is specified or the wife record is not found.
// Returns an error if the record is not part of a tree.
func (fr *FamilyRecord) GetWifeRecord() (*IndividualRecord, error) {
	wifeXref := fr.GetWife()
	if wifeXref == "" {
		return nil, nil // No wife, but not an error
	}

	tree := fr.getTree()
	if tree == nil {
		return nil, fmt.Errorf("family record is not part of a tree")
	}

	wifeRecord := tree.GetIndividual(wifeXref)
	if wifeRecord == nil {
		return nil, nil // Wife not found, but not an error
	}

	wife, ok := wifeRecord.(*IndividualRecord)
	if !ok {
		return nil, fmt.Errorf("wife record %s is not an IndividualRecord", wifeXref)
	}

	return wife, nil
}

// GetChildrenRecords returns all children's IndividualRecords.
// Returns an empty slice if no children are specified or children records are not found.
// Returns an error if the record is not part of a tree.
func (fr *FamilyRecord) GetChildrenRecords() ([]*IndividualRecord, error) {
	childXrefs := fr.GetChildren()
	if len(childXrefs) == 0 {
		return []*IndividualRecord{}, nil
	}

	tree := fr.getTree()
	if tree == nil {
		return nil, fmt.Errorf("family record is not part of a tree")
	}

	children := make([]*IndividualRecord, 0, len(childXrefs))
	for _, childXref := range childXrefs {
		childRecord := tree.GetIndividual(childXref)
		if childRecord == nil {
			continue // Skip missing children
		}

		child, ok := childRecord.(*IndividualRecord)
		if !ok {
			continue // Skip non-individual records
		}

		children = append(children, child)
	}

	return children, nil
}

// GetSpouses returns both husband and wife as IndividualRecords.
// Returns an empty slice if neither spouse is specified.
// Returns an error if the record is not part of a tree.
func (fr *FamilyRecord) GetSpouses() ([]*IndividualRecord, error) {
	spouses := make([]*IndividualRecord, 0, 2)

	husband, err := fr.GetHusbandRecord()
	if err != nil {
		return nil, err
	}
	if husband != nil {
		spouses = append(spouses, husband)
	}

	wife, err := fr.GetWifeRecord()
	if err != nil {
		return nil, err
	}
	if wife != nil {
		spouses = append(spouses, wife)
	}

	return spouses, nil
}

// HasChild checks if the specified individual is a child of this family.
// Returns an error if the record is not part of a tree.
func (fr *FamilyRecord) HasChild(individual *IndividualRecord) bool {
	if individual == nil {
		return false
	}

	childXrefs := fr.GetChildren()
	for _, childXref := range childXrefs {
		if childXref == individual.XrefID() {
			return true
		}
	}

	return false
}
