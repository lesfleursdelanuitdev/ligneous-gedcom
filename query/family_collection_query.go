package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// FamilyUniqueBy specifies what makes a family unique.
type FamilyUniqueBy string

const (
	FamilyUniqueByXref          FamilyUniqueBy = "xref"           // By family XREF (all unique)
	FamilyUniqueByHusband       FamilyUniqueBy = "husband"        // By husband XREF
	FamilyUniqueByWife          FamilyUniqueBy = "wife"           // By wife XREF
	FamilyUniqueByChildren      FamilyUniqueBy = "children"       // By number of children
	FamilyUniqueByMarriageDate  FamilyUniqueBy = "marriage_date"  // By marriage date
	FamilyUniqueByMarriagePlace FamilyUniqueBy = "marriage_place" // By marriage place
	FamilyUniqueByHusbandWife   FamilyUniqueBy = "husband_wife"   // By husband+wife combination
)

// FamilyFilter represents a filter function for families.
type FamilyFilter func(*types.FamilyRecord) bool

// FamilyCollectionQuery provides collection operations on families.
type FamilyCollectionQuery struct {
	graph    *Graph
	uniqueBy FamilyUniqueBy
	filters  []FamilyFilter
}

// NewFamilyCollectionQuery creates a new FamilyCollectionQuery.
func NewFamilyCollectionQuery(graph *Graph) *FamilyCollectionQuery {
	return &FamilyCollectionQuery{
		graph:    graph,
		uniqueBy: FamilyUniqueByXref, // Default: all families are unique
		filters:  make([]FamilyFilter, 0),
	}
}

// All returns all families in the tree.
func (fcq *FamilyCollectionQuery) All() ([]*types.FamilyRecord, error) {
	if fcq.graph == nil {
		return nil, fmt.Errorf("graph is nil")
	}
	familyNodes := fcq.graph.GetAllFamilies()
	families := make([]*types.FamilyRecord, 0, len(familyNodes))
	for _, node := range familyNodes {
		if node.Family != nil {
			// Apply filters
			matches := true
			for _, filter := range fcq.filters {
				if !filter(node.Family) {
					matches = false
					break
				}
			}
			if matches {
				families = append(families, node.Family)
			}
		}
	}
	return families, nil
}

// Unique returns unique families based on criteria.
func (fcq *FamilyCollectionQuery) Unique() *FamilyCollectionQuery {
	// This is a builder method - actual uniqueness is applied in Execute()
	return fcq
}

// By specifies uniqueness criteria.
func (fcq *FamilyCollectionQuery) By(criteria FamilyUniqueBy) *FamilyCollectionQuery {
	fcq.uniqueBy = criteria
	return fcq
}

// Filter adds a filter condition.
func (fcq *FamilyCollectionQuery) Filter(fn FamilyFilter) *FamilyCollectionQuery {
	fcq.filters = append(fcq.filters, fn)
	return fcq
}

// Count returns the number of families.
func (fcq *FamilyCollectionQuery) Count() (int, error) {
	results, err := fcq.Execute()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}

// Execute runs the query and returns results with uniqueness applied.
func (fcq *FamilyCollectionQuery) Execute() ([]*types.FamilyRecord, error) {
	// Get all families (with filters applied)
	allFamilies, err := fcq.All()
	if err != nil {
		return nil, err
	}

	// Apply uniqueness logic
	if fcq.uniqueBy == FamilyUniqueByXref {
		// All families are unique by XREF (default)
		return allFamilies, nil
	}

	return fcq.applyUniqueness(allFamilies), nil
}

// applyUniqueness applies uniqueness logic based on uniqueBy criteria.
func (fcq *FamilyCollectionQuery) applyUniqueness(families []*types.FamilyRecord) []*types.FamilyRecord {
	switch fcq.uniqueBy {
	case FamilyUniqueByHusband:
		return fcq.uniqueByHusband(families)
	case FamilyUniqueByWife:
		return fcq.uniqueByWife(families)
	case FamilyUniqueByChildren:
		return fcq.uniqueByChildren(families)
	case FamilyUniqueByMarriageDate:
		return fcq.uniqueByMarriageDate(families)
	case FamilyUniqueByMarriagePlace:
		return fcq.uniqueByMarriagePlace(families)
	case FamilyUniqueByHusbandWife:
		return fcq.uniqueByHusbandWife(families)
	default:
		// Default: all unique
		return families
	}
}

// uniqueByHusband returns families with unique husband XREFs.
func (fcq *FamilyCollectionQuery) uniqueByHusband(families []*types.FamilyRecord) []*types.FamilyRecord {
	return applyUniquenessByStringKey(families, func(fam *types.FamilyRecord) string {
		return fam.GetHusband()
	}, true) // Skip empty
}

// uniqueByWife returns families with unique wife XREFs.
func (fcq *FamilyCollectionQuery) uniqueByWife(families []*types.FamilyRecord) []*types.FamilyRecord {
	return applyUniquenessByStringKey(families, func(fam *types.FamilyRecord) string {
		return fam.GetWife()
	}, true) // Skip empty
}

// uniqueByChildren returns families with unique numbers of children.
func (fcq *FamilyCollectionQuery) uniqueByChildren(families []*types.FamilyRecord) []*types.FamilyRecord {
	return applyUniquenessByIntKey(families, func(fam *types.FamilyRecord) int {
		return len(fam.GetChildren())
	})
}

// uniqueByMarriageDate returns families with unique marriage dates.
func (fcq *FamilyCollectionQuery) uniqueByMarriageDate(families []*types.FamilyRecord) []*types.FamilyRecord {
	return applyUniquenessByStringKey(families, func(fam *types.FamilyRecord) string {
		return fam.GetMarriageDate()
	}, true) // Skip empty
}

// uniqueByMarriagePlace returns families with unique marriage places.
func (fcq *FamilyCollectionQuery) uniqueByMarriagePlace(families []*types.FamilyRecord) []*types.FamilyRecord {
	return applyUniquenessByStringKey(families, func(fam *types.FamilyRecord) string {
		return fam.GetMarriagePlace()
	}, true) // Skip empty
}

// uniqueByHusbandWife returns families with unique husband+wife combinations.
func (fcq *FamilyCollectionQuery) uniqueByHusbandWife(families []*types.FamilyRecord) []*types.FamilyRecord {
	seen := make(map[string]bool)
	result := make([]*types.FamilyRecord, 0)
	for _, fam := range families {
		husband := fam.GetHusband()
		wife := fam.GetWife()
		key := fmt.Sprintf("%s|%s", husband, wife)
		if !seen[key] {
			seen[key] = true
			result = append(result, fam)
		}
	}
	return result
}
