package types

import "fmt"

// ParseEvent parses an event from a GedcomLine.
// Handles both standard events (BIRT, DEAT, etc.) and custom events (EVEN with TYPE).
func ParseEvent(eventLine *GedcomLine) (*Event, error) {
	if eventLine == nil {
		return nil, fmt.Errorf("event line is nil")
	}

	event := &Event{
		Value:        eventLine.Value,
		OriginalLine: eventLine,
	}

	// Determine event type
	eventTag := eventLine.Tag
	event.Type = ParseEventType(eventTag)

	// Handle custom events (EVEN tag)
	if event.Type == EventTypeCustom {
		// For EVEN, the actual type is in the TYPE sub-tag
		typeLines := eventLine.GetLines("TYPE")
		if len(typeLines) > 0 && typeLines[0].Value != "" {
			event.CustomType = typeLines[0].Value
		} else {
			// EVEN without TYPE - use value or default
			if eventLine.Value != "" {
				event.CustomType = eventLine.Value
			} else {
				event.CustomType = "Unknown"
			}
		}
	}

	// Parse date
	dateLines := eventLine.GetLines("DATE")
	if len(dateLines) > 0 {
		event.Date = NewDateNodeFromLine(dateLines[0])
	}

	// Parse place
	placeLines := eventLine.GetLines("PLAC")
	if len(placeLines) > 0 {
		event.Place = NewPlaceNodeFromLine(placeLines[0])
	}

	// Parse sources
	sourceLines := eventLine.GetLines("SOUR")
	event.Sources = make([]string, 0, len(sourceLines))
	for _, sourceLine := range sourceLines {
		if sourceLine.XrefID != "" {
			event.Sources = append(event.Sources, sourceLine.XrefID)
		} else if sourceLine.Value != "" {
			event.Sources = append(event.Sources, sourceLine.Value)
		}
	}

	// Parse notes
	noteLines := eventLine.GetLines("NOTE")
	event.Notes = make([]string, 0, len(noteLines))
	for _, noteLine := range noteLines {
		if noteLine.XrefID != "" {
			event.Notes = append(event.Notes, noteLine.XrefID)
		} else if noteLine.Value != "" {
			event.Notes = append(event.Notes, noteLine.Value)
		}
	}

	return event, nil
}

// ExtractEvents extracts all events from an individual or family record.
// Returns events as structured Event objects.
func ExtractEvents(record Record) []*Event {
	if record == nil {
		return nil
	}

	events := make([]*Event, 0)

	// Standard event tags
	eventTags := []string{
		"BIRT", "DEAT", "BURI", "CREM", "CHR", "BAPM", "BARM", "BASM",
		"BLES", "CHRA", "CONF", "FCOM", "ORDN", "NATU", "EMIG", "IMMI",
		"CENS", "PROB", "WILL", "GRAD", "RETI", "RESI", "OCCU", "EDUC",
		"MARR", "DIV", "ANUL", "MARB", "MARC", "MARL", "MARS", "ENGA",
		"CAST", "DSCR", "NATI", "PROP", "RELI", "TITL", "EVEN",
	}

	for _, tag := range eventTags {
		eventLines := record.GetLines(tag)
		for _, eventLine := range eventLines {
			event, err := ParseEvent(eventLine)
			if err == nil && event != nil {
				events = append(events, event)
			}
		}
	}

	return events
}

// FilterEventsByType filters events by event type.
func FilterEventsByType(events []*Event, eventType EventType) []*Event {
	if events == nil {
		return nil
	}

	filtered := make([]*Event, 0)
	for _, event := range events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// FilterCustomEvents filters events by custom type name.
func FilterCustomEvents(events []*Event, customType string) []*Event {
	if events == nil {
		return nil
	}

	filtered := make([]*Event, 0)
	for _, event := range events {
		if event.IsCustom() && event.CustomType == customType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

