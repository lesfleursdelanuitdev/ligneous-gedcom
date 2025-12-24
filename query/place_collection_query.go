package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// PlaceUniqueBy specifies what makes a place unique.
type PlaceUniqueBy string

const (
	PlaceUniqueByFullString PlaceUniqueBy = "full_string" // By full place string
	PlaceUniqueByCity       PlaceUniqueBy = "city"       // By city component
	PlaceUniqueByState      PlaceUniqueBy = "state"       // By state component
	PlaceUniqueByCountry    PlaceUniqueBy = "country"     // By country component
	PlaceUniqueByCityState  PlaceUniqueBy = "city_state" // By city + state
)

// PlaceCollectionQuery provides collection operations on places.
type PlaceCollectionQuery struct {
	graph       *Graph
	uniqueBy    PlaceUniqueBy
	fromBirth   bool
	fromDeath   bool
	fromMarriage bool
	fromEvents  bool
}

// NewPlaceCollectionQuery creates a new PlaceCollectionQuery.
func NewPlaceCollectionQuery(graph *Graph) *PlaceCollectionQuery {
	return &PlaceCollectionQuery{
		graph:        graph,
		uniqueBy:     PlaceUniqueByFullString, // Default: by full string
		fromBirth:    true,                     // Default: include all
		fromDeath:    true,
		fromMarriage: true,
		fromEvents:  true,
	}
}

// All returns all places from all events.
func (pcq *PlaceCollectionQuery) All() ([]string, error) {
	if pcq.graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	placesSet := make(map[string]bool)

	// Get places from all individuals
	err := ForEachIndividual(pcq.graph, func(indiNode *IndividualNode) error {
		// Birth place
		if pcq.fromBirth {
			if place := indiNode.Individual.GetBirthPlace(); place != "" {
				placesSet[place] = true
			}
		}
		// Death place
		if pcq.fromDeath {
			if place := indiNode.Individual.GetDeathPlace(); place != "" {
				placesSet[place] = true
			}
		}
		// Other event places
		if pcq.fromEvents {
			events := indiNode.Individual.GetEvents()
			for _, event := range events {
				if place, ok := event["place"].(string); ok && place != "" {
					placesSet[place] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Get places from all families
	err = ForEachFamily(pcq.graph, func(famNode *FamilyNode) error {
		// Marriage place
		if pcq.fromMarriage {
			if place := famNode.Family.GetMarriagePlace(); place != "" {
				placesSet[place] = true
			}
		}
		// Divorce place
		if place := famNode.Family.GetDivorcePlace(); place != "" {
			placesSet[place] = true
		}
		// Other event places
		if pcq.fromEvents {
			events := famNode.Family.GetEvents()
			for _, event := range events {
				if place, ok := event["place"].(string); ok && place != "" {
					placesSet[place] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert set to slice
	places := make([]string, 0, len(placesSet))
	for place := range placesSet {
		places = append(places, place)
	}

	return places, nil
}

// Unique returns unique places based on criteria.
func (pcq *PlaceCollectionQuery) Unique() *PlaceCollectionQuery {
	return pcq
}

// By specifies uniqueness criteria.
func (pcq *PlaceCollectionQuery) By(criteria PlaceUniqueBy) *PlaceCollectionQuery {
	pcq.uniqueBy = criteria
	return pcq
}

// FromBirth only includes birth places.
func (pcq *PlaceCollectionQuery) FromBirth() *PlaceCollectionQuery {
	pcq.fromBirth = true
	pcq.fromDeath = false
	pcq.fromMarriage = false
	pcq.fromEvents = false
	return pcq
}

// FromDeath only includes death places.
func (pcq *PlaceCollectionQuery) FromDeath() *PlaceCollectionQuery {
	pcq.fromDeath = true
	pcq.fromBirth = false
	pcq.fromMarriage = false
	pcq.fromEvents = false
	return pcq
}

// FromMarriage only includes marriage places.
func (pcq *PlaceCollectionQuery) FromMarriage() *PlaceCollectionQuery {
	pcq.fromMarriage = true
	pcq.fromBirth = false
	pcq.fromDeath = false
	pcq.fromEvents = false
	return pcq
}

// Count returns the number of unique places.
func (pcq *PlaceCollectionQuery) Count() (int, error) {
	results, err := pcq.Execute()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}

// Execute runs the query and returns unique places based on criteria.
func (pcq *PlaceCollectionQuery) Execute() ([]string, error) {
	// Get all places
	allPlaces, err := pcq.All()
	if err != nil {
		return nil, err
	}

	// Apply uniqueness logic
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, placeStr := range allPlaces {
		var key string

		switch pcq.uniqueBy {
		case PlaceUniqueByFullString:
			key = placeStr
		case PlaceUniqueByCity, PlaceUniqueByState, PlaceUniqueByCountry, PlaceUniqueByCityState:
			// Parse place to extract components
			place, err := types.ParsePlace(placeStr)
			if err != nil || place == nil {
				// If parsing fails, use full string
				key = placeStr
			} else {
				switch pcq.uniqueBy {
				case PlaceUniqueByCity:
					key = place.City
				case PlaceUniqueByState:
					key = place.State
				case PlaceUniqueByCountry:
					key = place.Country
				case PlaceUniqueByCityState:
					key = fmt.Sprintf("%s|%s", place.City, place.State)
				}
			}
		default:
			key = placeStr
		}

		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}

	return result, nil
}

