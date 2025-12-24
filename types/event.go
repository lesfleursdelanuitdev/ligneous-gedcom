package types

// Event represents a unified event structure that can represent any GEDCOM event type.
// This includes standard events (BIRT, DEAT, MARR, etc.) and custom events (EVEN with TYPE).
type Event struct {
	// Type identifies the event type (BIRT, DEAT, MARR, or custom type for EVEN)
	Type EventType

	// CustomType is the actual type name for custom events (EVEN with TYPE sub-tag).
	// For standard events, this will be empty.
	CustomType string

	// Date is the structured date associated with the event.
	Date *DateNode

	// Place is the structured place associated with the event.
	Place *PlaceNode

	// Sources contains source citations for this event.
	Sources []string

	// Notes contains note references for this event.
	Notes []string

	// Value is the raw value of the event tag (usually empty for most events).
	Value string

	// OriginalLine is the original GedcomLine that this event was parsed from.
	// This allows access to any additional sub-tags not explicitly represented.
	OriginalLine *GedcomLine
}

// IsCustom returns true if this is a custom event (EVEN tag with TYPE sub-tag).
func (e *Event) IsCustom() bool {
	return e.Type == EventTypeCustom && e.CustomType != ""
}

// EffectiveType returns the actual event type.
// For custom events, returns the CustomType; otherwise returns the Type.
func (e *Event) EffectiveType() string {
	if e.IsCustom() {
		return e.CustomType
	}
	return e.Type.String()
}

// HasDate returns true if the event has a date.
func (e *Event) HasDate() bool {
	return e.Date != nil && e.Date.IsValid()
}

// HasPlace returns true if the event has a place.
func (e *Event) HasPlace() bool {
	return e.Place != nil && e.Place.IsValid()
}

// String returns a string representation of the event.
func (e *Event) String() string {
	if e.IsCustom() {
		return e.CustomType
	}
	return e.Type.String()
}

