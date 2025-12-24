package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// NameUniqueBy specifies what makes a name unique.
type NameUniqueBy string

const (
	NameUniqueByFullName     NameUniqueBy = "full_name"      // By full name string
	NameUniqueByGiven        NameUniqueBy = "given"          // By given name
	NameUniqueBySurname      NameUniqueBy = "surname"        // By surname
	NameUniqueByGivenSurname NameUniqueBy = "given_surname" // By given + surname
	NameUniqueBySurnameGiven NameUniqueBy = "surname_given" // By surname + given
)

// NameInfo represents a name with context.
type NameInfo struct {
	IndividualXref string
	Name           *types.GedcomName
	IsPrimary      bool
}

// NameCollectionQuery provides collection operations on names.
type NameCollectionQuery struct {
	graph    *Graph
	uniqueBy NameUniqueBy
}

// NewNameCollectionQuery creates a new NameCollectionQuery.
func NewNameCollectionQuery(graph *Graph) *NameCollectionQuery {
	return &NameCollectionQuery{
		graph:    graph,
		uniqueBy: NameUniqueByFullName, // Default: by full name
	}
}

// All returns all names from all individuals.
func (ncq *NameCollectionQuery) All() ([]NameInfo, error) {
	if ncq.graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	allNames := make([]NameInfo, 0)

	// Get names from all individuals
	err := ForEachIndividual(ncq.graph, func(indiNode *IndividualNode) error {
		names, err := indiNode.Individual.GetNamesParsed()
		if err != nil {
			return nil // Continue iteration
		}

		for i, name := range names {
			if name != nil && name.IsValid() {
				allNames = append(allNames, NameInfo{
					IndividualXref: indiNode.ID(),
					Name:           name,
					IsPrimary:      i == 0,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allNames, nil
}

// Unique returns unique names based on criteria.
func (ncq *NameCollectionQuery) Unique() *NameCollectionQuery {
	return ncq
}

// By specifies uniqueness criteria.
func (ncq *NameCollectionQuery) By(criteria NameUniqueBy) *NameCollectionQuery {
	ncq.uniqueBy = criteria
	return ncq
}

// Count returns the number of unique names.
func (ncq *NameCollectionQuery) Count() (int, error) {
	results, err := ncq.Execute()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}

// Execute runs the query and returns unique name strings based on criteria.
func (ncq *NameCollectionQuery) Execute() ([]string, error) {
	// Get all names
	allNames, err := ncq.All()
	if err != nil {
		return nil, err
	}

	// Apply uniqueness logic
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, nameInfo := range allNames {
		if nameInfo.Name == nil {
			continue
		}

		var key string
		switch ncq.uniqueBy {
		case NameUniqueByFullName:
			key = nameInfo.Name.FullName()
		case NameUniqueByGiven:
			key = nameInfo.Name.Given
		case NameUniqueBySurname:
			key = nameInfo.Name.Surname
		case NameUniqueByGivenSurname:
			key = fmt.Sprintf("%s|%s", nameInfo.Name.Given, nameInfo.Name.Surname)
		case NameUniqueBySurnameGiven:
			key = fmt.Sprintf("%s|%s", nameInfo.Name.Surname, nameInfo.Name.Given)
		default:
			key = nameInfo.Name.FullName()
		}

		if key != "" && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}

	return result, nil
}

