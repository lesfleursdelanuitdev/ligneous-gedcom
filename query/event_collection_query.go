package query

import (
	"fmt"
)

// EventUniqueBy specifies what makes an event unique.
type EventUniqueBy string

const (
	EventUniqueByID        EventUniqueBy = "id"         // By event ID (all unique)
	EventUniqueByType      EventUniqueBy = "type"       // By event type
	EventUniqueByDate      EventUniqueBy = "date"       // By date
	EventUniqueByPlace     EventUniqueBy = "place"      // By place
	EventUniqueByTypeDate  EventUniqueBy = "type_date"   // By type + date
	EventUniqueByTypePlace EventUniqueBy = "type_place" // By type + place
	EventUniqueByOwner     EventUniqueBy = "owner"      // By owner (individual/family)
)

// EventFilter represents a filter function for events.
type EventFilter func(EventInfo) bool

// EventCollectionQuery provides collection operations on events.
type EventCollectionQuery struct {
	graph          *Graph
	uniqueBy       EventUniqueBy
	filters        []EventFilter
	fromIndividuals bool
	fromFamilies    bool
	eventTypes      []string
}

// NewEventCollectionQuery creates a new EventCollectionQuery.
func NewEventCollectionQuery(graph *Graph) *EventCollectionQuery {
	return &EventCollectionQuery{
		graph:           graph,
		uniqueBy:        EventUniqueByID, // Default: all events are unique
		filters:         make([]EventFilter, 0),
		fromIndividuals: true,  // Default: include both
		fromFamilies:    true,
		eventTypes:      make([]string, 0),
	}
}

// All returns all events from all individuals and families.
func (ecq *EventCollectionQuery) All() ([]EventInfo, error) {
	if ecq.graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	allEvents := make([]EventInfo, 0)

	// Get events from individuals
	if ecq.fromIndividuals {
		err := ForEachIndividual(ecq.graph, func(indiNode *IndividualNode) error {
			events, err := (&IndividualQuery{xrefID: indiNode.ID(), graph: ecq.graph}).GetEvents()
			if err == nil {
				allEvents = append(allEvents, events...)
			}
			return nil // Continue iteration even if error
		})
		if err != nil {
			return nil, err
		}
	}

	// Get events from families
	if ecq.fromFamilies {
		err := ForEachFamily(ecq.graph, func(famNode *FamilyNode) error {
			events, err := (&FamilyQuery{xrefID: famNode.ID(), graph: ecq.graph}).GetEvents()
			if err == nil {
				allEvents = append(allEvents, events...)
			}
			return nil // Continue iteration even if error
		})
		if err != nil {
			return nil, err
		}
	}

	// Filter by event types
	if len(ecq.eventTypes) > 0 {
		filtered := make([]EventInfo, 0)
		for _, event := range allEvents {
			for _, eventType := range ecq.eventTypes {
				if event.EventType == eventType {
					filtered = append(filtered, event)
					break
				}
			}
		}
		allEvents = filtered
	}

	// Apply custom filters
	filtered := make([]EventInfo, 0)
	for _, event := range allEvents {
		matches := true
		for _, filter := range ecq.filters {
			if !filter(event) {
				matches = false
				break
			}
		}
		if matches {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

// Unique returns unique events based on criteria.
func (ecq *EventCollectionQuery) Unique() *EventCollectionQuery {
	return ecq
}

// By specifies uniqueness criteria.
func (ecq *EventCollectionQuery) By(criteria EventUniqueBy) *EventCollectionQuery {
	ecq.uniqueBy = criteria
	return ecq
}

// FromIndividuals only includes events from individuals.
func (ecq *EventCollectionQuery) FromIndividuals() *EventCollectionQuery {
	ecq.fromIndividuals = true
	ecq.fromFamilies = false
	return ecq
}

// FromFamilies only includes events from families.
func (ecq *EventCollectionQuery) FromFamilies() *EventCollectionQuery {
	ecq.fromFamilies = true
	ecq.fromIndividuals = false
	return ecq
}

// OfType filters by event type.
func (ecq *EventCollectionQuery) OfType(eventType string) *EventCollectionQuery {
	ecq.eventTypes = append(ecq.eventTypes, eventType)
	return ecq
}

// Filter adds a filter condition.
func (ecq *EventCollectionQuery) Filter(fn EventFilter) *EventCollectionQuery {
	ecq.filters = append(ecq.filters, fn)
	return ecq
}

// Count returns the number of events.
func (ecq *EventCollectionQuery) Count() (int, error) {
	results, err := ecq.Execute()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}

// Execute runs the query and returns results with uniqueness applied.
func (ecq *EventCollectionQuery) Execute() ([]EventInfo, error) {
	// Get all events (with filters applied)
	allEvents, err := ecq.All()
	if err != nil {
		return nil, err
	}

	// Apply uniqueness logic
	if ecq.uniqueBy == EventUniqueByID {
		// All events are unique by ID (default)
		return allEvents, nil
	}

	return ecq.applyUniqueness(allEvents), nil
}

// applyUniqueness applies uniqueness logic based on uniqueBy criteria.
func (ecq *EventCollectionQuery) applyUniqueness(events []EventInfo) []EventInfo {
	switch ecq.uniqueBy {
	case EventUniqueByType:
		return ecq.uniqueByType(events)
	case EventUniqueByDate:
		return ecq.uniqueByDate(events)
	case EventUniqueByPlace:
		return ecq.uniqueByPlace(events)
	case EventUniqueByTypeDate:
		return ecq.uniqueByTypeDate(events)
	case EventUniqueByTypePlace:
		return ecq.uniqueByTypePlace(events)
	case EventUniqueByOwner:
		return ecq.uniqueByOwner(events)
	default:
		// Default: all unique
		return events
	}
}

// uniqueByType returns events with unique types.
func (ecq *EventCollectionQuery) uniqueByType(events []EventInfo) []EventInfo {
	seen := make(map[string]bool)
	result := make([]EventInfo, 0)
	for _, event := range events {
		if event.EventType != "" && !seen[event.EventType] {
			seen[event.EventType] = true
			result = append(result, event)
		}
	}
	return result
}

// uniqueByDate returns events with unique dates.
func (ecq *EventCollectionQuery) uniqueByDate(events []EventInfo) []EventInfo {
	seen := make(map[string]bool)
	result := make([]EventInfo, 0)
	for _, event := range events {
		if event.Date != "" && !seen[event.Date] {
			seen[event.Date] = true
			result = append(result, event)
		}
	}
	return result
}

// uniqueByPlace returns events with unique places.
func (ecq *EventCollectionQuery) uniqueByPlace(events []EventInfo) []EventInfo {
	seen := make(map[string]bool)
	result := make([]EventInfo, 0)
	for _, event := range events {
		if event.Place != "" && !seen[event.Place] {
			seen[event.Place] = true
			result = append(result, event)
		}
	}
	return result
}

// uniqueByTypeDate returns events with unique type+date combinations.
func (ecq *EventCollectionQuery) uniqueByTypeDate(events []EventInfo) []EventInfo {
	return applyUniquenessByCompositeKey(events, func(event EventInfo) string {
		return fmt.Sprintf("%s|%s", event.EventType, event.Date)
	}, "|")
}

// uniqueByTypePlace returns events with unique type+place combinations.
func (ecq *EventCollectionQuery) uniqueByTypePlace(events []EventInfo) []EventInfo {
	return applyUniquenessByCompositeKey(events, func(event EventInfo) string {
		return fmt.Sprintf("%s|%s", event.EventType, event.Place)
	}, "|")
}

// uniqueByOwner returns events with unique owners (individual/family XREF).
func (ecq *EventCollectionQuery) uniqueByOwner(events []EventInfo) []EventInfo {
	return applyUniquenessByStringKey(events, func(event EventInfo) string {
		if event.Owner != nil {
			return event.Owner.ID()
		}
		return ""
	}, true) // Skip empty
}

