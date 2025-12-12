package gedcom

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

	events := make([]map[string]interface{}, 0)
	for _, tag := range eventTags {
		eventLines := fr.GetLines(tag)
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

// GetNotes returns all note xrefs.
func (fr *FamilyRecord) GetNotes() []string {
	return fr.GetValues("NOTE")
}

// GetSources returns all source xrefs.
func (fr *FamilyRecord) GetSources() []string {
	return fr.GetValues("SOUR")
}

