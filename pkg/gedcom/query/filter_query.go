package query

import (
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

// Filter represents a filter function for individuals.
type Filter func(*gedcom.IndividualRecord) bool

// FilterQuery represents a query with filtering capabilities.
type FilterQuery struct {
	graph   *Graph
	filters []Filter

	// Indexed filter state
	nameFilter        string
	nameExactFilter   string
	nameStartsFilter  string
	nameEndsFilter    string
	birthDateStart    *time.Time
	birthDateEnd      *time.Time
	birthPlaceFilter  string
	sexFilter         string
	hasChildrenFilter *bool
	hasSpouseFilter   *bool
	livingFilter      *bool
}

// NewFilterQuery creates a new FilterQuery.
func NewFilterQuery(graph *Graph) *FilterQuery {
	return &FilterQuery{
		graph:   graph,
		filters: make([]Filter, 0),
	}
}

// Where adds a filter condition.
func (fq *FilterQuery) Where(filter Filter) *FilterQuery {
	fq.filters = append(fq.filters, filter)
	return fq
}

// Count returns the number of matching individuals.
func (fq *FilterQuery) Count() (int, error) {
	results, err := fq.Execute()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}
