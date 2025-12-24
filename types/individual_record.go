package types

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
	return extractEvents(ir, eventTags)
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
// Delegates to BaseRecord.GetNotes().
func (ir *IndividualRecord) GetNotes() []string {
	return ir.BaseRecord.GetNotes()
}

// GetSources returns all source xrefs.
// Delegates to BaseRecord.GetSources().
func (ir *IndividualRecord) GetSources() []string {
	return ir.BaseRecord.GetSources()
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

// GetNamesParsed returns all names as parsed GedcomName objects.
// Supports multiple NAME records per individual (GEDCOM 5.5.1).
// Returns empty slice if no names found.
func (ir *IndividualRecord) GetNamesParsed() ([]*GedcomName, error) {
	nameLines := ir.GetLines("NAME")
	if len(nameLines) == 0 {
		return []*GedcomName{}, nil
	}

	names := make([]*GedcomName, 0, len(nameLines))
	for _, nameLine := range nameLines {
		name, err := ParseName(nameLine)
		if err != nil {
			// Continue parsing other names even if one fails
			continue
		}
		if name != nil {
			names = append(names, name)
		}
	}

	return names, nil
}

// GetPrimaryName returns the first (primary) name as a parsed GedcomName.
// Returns nil if no names found.
func (ir *IndividualRecord) GetPrimaryName() (*GedcomName, error) {
	nameLines := ir.GetLines("NAME")
	if len(nameLines) == 0 {
		return nil, fmt.Errorf("no name found")
	}

	return ParseName(nameLines[0])
}

// GetNameByType returns a name of the specified type.
// Returns nil if no name of that type is found.
func (ir *IndividualRecord) GetNameByType(nameType NameType) (*GedcomName, error) {
	names, err := ir.GetNamesParsed()
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		if name.Type == nameType {
			return name, nil
		}
	}

	return nil, fmt.Errorf("no name of type %s found", nameType)
}

// GetBirthName returns the birth name (TYPE birth) if available.
// Falls back to primary name if no birth name is found.
func (ir *IndividualRecord) GetBirthName() (*GedcomName, error) {
	birthName, err := ir.GetNameByType(NameTypeBirth)
	if err == nil && birthName != nil {
		return birthName, nil
	}

	// Fallback to primary name
	return ir.GetPrimaryName()
}

// GetMarriedName returns the married name (TYPE married) if available.
// Returns nil if no married name is found.
func (ir *IndividualRecord) GetMarriedName() (*GedcomName, error) {
	return ir.GetNameByType(NameTypeMarried)
}

// ============================================================================
// Structured Node Methods (Hybrid Approach)
// These methods return structured node types similar to elliotchance,
// while keeping the existing string-based methods above.
// ============================================================================

// Name returns the primary name as a structured NameNode.
// Returns nil if no name is found.
func (ir *IndividualRecord) Name() *NameNode {
	nameLines := ir.GetLines("NAME")
	if len(nameLines) == 0 {
		return nil
	}
	return NewNameNodeFromLine(nameLines[0])
}

// Names returns all names as structured NameNodes.
// Returns empty slice if no names found.
func (ir *IndividualRecord) Names() []*NameNode {
	nameLines := ir.GetLines("NAME")
	if len(nameLines) == 0 {
		return nil
	}

	names := make([]*NameNode, 0, len(nameLines))
	for _, nameLine := range nameLines {
		if nameNode := NewNameNodeFromLine(nameLine); nameNode != nil {
			names = append(names, nameNode)
		}
	}
	return names
}

// Birth returns the birth event as a structured Event.
// Returns nil if no birth event is found.
func (ir *IndividualRecord) Birth() *Event {
	birthLines := ir.GetLines("BIRT")
	if len(birthLines) == 0 {
		return nil
	}

	event, err := ParseEvent(birthLines[0])
	if err != nil || event == nil {
		return nil
	}
	return event
}

// Death returns the death event as a structured Event.
// Returns nil if no death event is found.
func (ir *IndividualRecord) Death() *Event {
	deathLines := ir.GetLines("DEAT")
	if len(deathLines) == 0 {
		return nil
	}

	event, err := ParseEvent(deathLines[0])
	if err != nil || event == nil {
		return nil
	}
	return event
}

// Births returns all birth events as structured Events.
// Returns empty slice if no birth events found.
func (ir *IndividualRecord) Births() []*Event {
	return FilterEventsByType(ExtractEvents(ir), EventTypeBirth)
}

// Deaths returns all death events as structured Events.
// Returns empty slice if no death events found.
func (ir *IndividualRecord) Deaths() []*Event {
	return FilterEventsByType(ExtractEvents(ir), EventTypeDeath)
}

// Events returns all events as structured Events.
// Includes standard events (BIRT, DEAT, etc.) and custom events (EVEN).
func (ir *IndividualRecord) Events() []*Event {
	return ExtractEvents(ir)
}

// EventsByType returns all events of the specified type.
// For custom events, use FilterCustomEvents instead.
func (ir *IndividualRecord) EventsByType(eventType EventType) []*Event {
	return FilterEventsByType(ExtractEvents(ir), eventType)
}

// CustomEvents returns all custom events (EVEN tags).
func (ir *IndividualRecord) CustomEvents() []*Event {
	events := ExtractEvents(ir)
	custom := make([]*Event, 0)
	for _, event := range events {
		if event.IsCustom() {
			custom = append(custom, event)
		}
	}
	return custom
}

// CustomEventsByType returns custom events with the specified custom type name.
func (ir *IndividualRecord) CustomEventsByType(customType string) []*Event {
	return FilterCustomEvents(ExtractEvents(ir), customType)
}

// Baptism returns the baptism event as a structured Event.
func (ir *IndividualRecord) Baptism() *Event {
	baptismLines := ir.GetLines("BAPM")
	if len(baptismLines) == 0 {
		return nil
	}

	event, err := ParseEvent(baptismLines[0])
	if err != nil || event == nil {
		return nil
	}
	return event
}

// Burial returns the burial event as a structured Event.
func (ir *IndividualRecord) Burial() *Event {
	burialLines := ir.GetLines("BURI")
	if len(burialLines) == 0 {
		return nil
	}

	event, err := ParseEvent(burialLines[0])
	if err != nil || event == nil {
		return nil
	}
	return event
}

// Baptisms returns all baptism events.
func (ir *IndividualRecord) Baptisms() []*Event {
	return FilterEventsByType(ExtractEvents(ir), EventTypeBaptism)
}

// Burials returns all burial events.
func (ir *IndividualRecord) Burials() []*Event {
	return FilterEventsByType(ExtractEvents(ir), EventTypeBurial)
}

// BirthDate returns the birth date as a structured DateNode.
// Returns nil if no birth date is found.
func (ir *IndividualRecord) BirthDate() *DateNode {
	birth := ir.Birth()
	if birth == nil || birth.Date == nil {
		return nil
	}
	return birth.Date
}

// DeathDate returns the death date as a structured DateNode.
// Returns nil if no death date is found.
func (ir *IndividualRecord) DeathDate() *DateNode {
	death := ir.Death()
	if death == nil || death.Date == nil {
		return nil
	}
	return death.Date
}

// BirthPlace returns the birth place as a structured PlaceNode.
// Returns nil if no birth place is found.
func (ir *IndividualRecord) BirthPlace() *PlaceNode {
	birth := ir.Birth()
	if birth == nil || birth.Place == nil {
		return nil
	}
	return birth.Place
}

// DeathPlace returns the death place as a structured PlaceNode.
// Returns nil if no death place is found.
func (ir *IndividualRecord) DeathPlace() *PlaceNode {
	death := ir.Death()
	if death == nil || death.Place == nil {
		return nil
	}
	return death.Place
}

// ============================================================================
// Phase 1: Direct Relationship Methods
// ============================================================================
// NOTE: Relationship methods (Spouses, Children, Parents, Siblings, SpouseChildren)
// have been removed. Use the graph package for relationship queries:
//   - graph.GetSpouses(xrefID) or node.Spouses()
//   - graph.GetChildren(xrefID) or node.Children()
//   - graph.GetParents(xrefID) or node.Parents()
//   - graph.GetSiblings(xrefID) or node.Siblings()
// Relationship queries should go through the graph, not records.

// Families returns all families this individual is part of (as spouse OR child).
// Traverses families directly without requiring the query package.
// Returns an error if the record is not part of a tree.
func (ir *IndividualRecord) Families() ([]*FamilyRecord, error) {
	tree := ir.getTree()
	if tree == nil {
		return nil, fmt.Errorf("individual record is not part of a tree")
	}

	// Get families as spouse and as child
	famsAsSpouse := ir.GetFamiliesAsSpouse()
	famsAsChild := ir.GetFamiliesAsChild()

	// Combine and deduplicate
	famXrefs := make(map[string]bool)
	for _, xref := range famsAsSpouse {
		famXrefs[xref] = true
	}
	for _, xref := range famsAsChild {
		famXrefs[xref] = true
	}

	// Convert to FamilyRecord slice
	families := make([]*FamilyRecord, 0, len(famXrefs))
	for xref := range famXrefs {
		record := tree.GetFamily(xref)
		if fr, ok := record.(*FamilyRecord); ok {
			families = append(families, fr)
		}
	}

	return families, nil
}

// ============================================================================
// Phase 2: Family Lookup Methods
// ============================================================================

// FamilyWithSpouse finds the family record where this individual is married to the specified spouse.
// Returns nil if no such family exists.
// Returns an error if the record is not part of a tree.
func (ir *IndividualRecord) FamilyWithSpouse(spouse *IndividualRecord) (*FamilyRecord, error) {
	if spouse == nil {
		return nil, fmt.Errorf("spouse cannot be nil")
	}

	tree := ir.getTree()
	if tree == nil {
		return nil, fmt.Errorf("individual record is not part of a tree")
	}

	// Get all families where this individual is a spouse
	famXrefs := ir.GetFamiliesAsSpouse()
	for _, famXref := range famXrefs {
		famRecord := tree.GetFamily(famXref)
		if famRecord == nil {
			continue
		}

		fr, ok := famRecord.(*FamilyRecord)
		if !ok {
			continue
		}

		// Check if the spouse matches
		husbandXref := fr.GetHusband()
		wifeXref := fr.GetWife()

		// Check if this individual is husband and spouse is wife, or vice versa
		if (ir.XrefID() == husbandXref && spouse.XrefID() == wifeXref) ||
			(ir.XrefID() == wifeXref && spouse.XrefID() == husbandXref) {
			return fr, nil
		}
	}

	return nil, nil // No family found, but not an error
}

// FamilyWithUnknownSpouse finds a family record where this individual is a spouse
// but the other spouse is unknown (nil).
// Returns nil if no such family exists.
// Returns an error if the record is not part of a tree.
func (ir *IndividualRecord) FamilyWithUnknownSpouse() (*FamilyRecord, error) {
	tree := ir.getTree()
	if tree == nil {
		return nil, fmt.Errorf("individual record is not part of a tree")
	}

	// Get all families where this individual is a spouse
	famXrefs := ir.GetFamiliesAsSpouse()
	for _, famXref := range famXrefs {
		famRecord := tree.GetFamily(famXref)
		if famRecord == nil {
			continue
		}

		fr, ok := famRecord.(*FamilyRecord)
		if !ok {
			continue
		}

		husbandXref := fr.GetHusband()
		wifeXref := fr.GetWife()

		// Check if this individual is husband and wife is missing, or vice versa
		if (ir.XrefID() == husbandXref && wifeXref == "") ||
			(ir.XrefID() == wifeXref && husbandXref == "") {
			return fr, nil
		}
	}

	return nil, nil // No family found, but not an error
}

// ============================================================================
// Phase 3: SpouseChildren Pattern
// ============================================================================
// NOTE: SpouseChildren method has been removed. Use the graph package for
// relationship queries. Relationship queries should go through the graph, not records.
