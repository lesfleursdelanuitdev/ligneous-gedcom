package types

// EventNode represents a structured wrapper around an Event.
// Provides a node-like interface similar to elliotchance's approach.
type EventNode struct {
	*Event
}

// NewEventNode creates a new EventNode from an Event.
func NewEventNode(event *Event) *EventNode {
	if event == nil {
		return nil
	}
	return &EventNode{Event: event}
}

// Dates returns the date nodes associated with this event.
// Returns a slice containing the single DateNode if present.
func (en *EventNode) Dates() []*DateNode {
	if en == nil || en.Date == nil {
		return nil
	}
	return []*DateNode{en.Date}
}

// DateRange returns the date range for this event.
// Returns a zero DateRange if no date is present.
func (en *EventNode) DateRange() DateRange {
	if en == nil || en.Date == nil {
		return NewZeroDateRange()
	}
	return en.Date.DateRange
}

// Years returns the years value of the event's date.
// Returns 0 if no date is present.
func (en *EventNode) Years() float64 {
	if en == nil || en.Date == nil {
		return 0
	}
	return en.Date.Years()
}

// Equals compares two event nodes for equality.
// Two events are equal if they have the same type and date.
func (en *EventNode) Equals(other *EventNode) bool {
	if en == nil || other == nil {
		return en == other
	}

	if en.Type != other.Type {
		return false
	}

	if en.IsCustom() && en.CustomType != other.CustomType {
		return false
	}

	// Compare dates if both have them
	if en.Date != nil && other.Date != nil {
		return en.Date.Equals(other.Date)
	}

	return en.Date == other.Date
}

