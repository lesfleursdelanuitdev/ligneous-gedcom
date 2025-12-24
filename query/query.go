package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// QueryBuilder is the entry point for building queries on a GEDCOM graph.
type QueryBuilder struct {
	graph *Graph
}

// NewQuery creates a new query builder from a GEDCOM tree.
// It builds the graph automatically (eager loading).
func NewQuery(tree *types.GedcomTree) (*QueryBuilder, error) {
	graph, err := BuildGraph(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	return &QueryBuilder{
		graph: graph,
	}, nil
}

// NewQueryLazy creates a new query builder with lazy loading enabled.
// Nodes and edges are loaded on-demand when accessed.
// This is more memory-efficient for large datasets.
func NewQueryLazy(tree *types.GedcomTree) (*QueryBuilder, error) {
	graph, err := BuildGraphLazy(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to build lazy graph: %w", err)
	}

	return &QueryBuilder{
		graph: graph,
	}, nil
}

// NewQueryFromGraph creates a new query builder from an existing graph.
func NewQueryFromGraph(graph *Graph) *QueryBuilder {
	return &QueryBuilder{
		graph: graph,
	}
}

// Individual starts a query from a specific individual.
func (qb *QueryBuilder) Individual(xrefID string) *IndividualQuery {
	return &IndividualQuery{
		xrefID: xrefID,
		graph:  qb.graph,
	}
}

// Individuals starts a query from multiple individuals.
func (qb *QueryBuilder) Individuals(xrefIDs ...string) *MultiIndividualQuery {
	return &MultiIndividualQuery{
		xrefIDs: xrefIDs,
		graph:   qb.graph,
	}
}

// AllIndividuals starts a query from all individuals.
func (qb *QueryBuilder) AllIndividuals() *MultiIndividualQuery {
	individuals := qb.graph.GetAllIndividuals()
	xrefIDs := make([]string, 0, len(individuals))
	for id := range individuals {
		xrefIDs = append(xrefIDs, id)
	}
	return &MultiIndividualQuery{
		xrefIDs: xrefIDs,
		graph:   qb.graph,
	}
}

// Filter starts a filter query on all individuals.
func (qb *QueryBuilder) Filter() *FilterQuery {
	return NewFilterQuery(qb.graph)
}

// Family starts a query from a family.
func (qb *QueryBuilder) Family(xrefID string) *FamilyQuery {
	return &FamilyQuery{
		xrefID: xrefID,
		graph:  qb.graph,
	}
}

// Graph returns the internal graph representation for advanced operations.
func (qb *QueryBuilder) Graph() *Graph {
	return qb.graph
}

// Metrics returns a GraphMetricsQuery for graph analytics.
func (qb *QueryBuilder) Metrics() *GraphMetricsQuery {
	return NewGraphMetricsQuery(qb.graph)
}

// Families returns a FamilyCollectionQuery for collection operations on families.
func (qb *QueryBuilder) Families() *FamilyCollectionQuery {
	return NewFamilyCollectionQuery(qb.graph)
}

// AllFamilies returns all family records in the tree.
// Deprecated: Use Families().All() instead.
func (qb *QueryBuilder) AllFamilies() ([]*types.FamilyRecord, error) {
	return qb.Families().All()
}

// Events returns an EventCollectionQuery for collection operations on events.
func (qb *QueryBuilder) Events() *EventCollectionQuery {
	return NewEventCollectionQuery(qb.graph)
}

// AllEvents returns all events from all individuals and families in the tree.
// Deprecated: Use Events().All() instead.
func (qb *QueryBuilder) AllEvents() ([]EventInfo, error) {
	return qb.Events().All()
}

// Names returns a NameCollectionQuery for collection operations on names.
func (qb *QueryBuilder) Names() *NameCollectionQuery {
	return NewNameCollectionQuery(qb.graph)
}

// Places returns a PlaceCollectionQuery for collection operations on places.
func (qb *QueryBuilder) Places() *PlaceCollectionQuery {
	return NewPlaceCollectionQuery(qb.graph)
}

// AllPlaces returns all unique places found in the tree.
// Places are extracted from birth, death, marriage, and other event locations.
// Deprecated: Use Places().All() instead.
func (qb *QueryBuilder) AllPlaces() ([]string, error) {
	return qb.Places().All()
}

// UniqueNames returns all unique names found in the tree.
// Returns a map with "given" and "surname" keys, each containing a slice of unique names.
// Uses GedcomName parsing for accurate extraction of name components.
// Deprecated: Use Names().Unique().By(NameUniqueByGiven).Execute() and Names().Unique().By(NameUniqueBySurname).Execute() instead.
func (qb *QueryBuilder) UniqueNames() (map[string][]string, error) {
	givenNamesSet := make(map[string]bool)
	surnamesSet := make(map[string]bool)

	// Get names from all individuals
	allIndividuals := qb.graph.GetAllIndividuals()
	for _, indiNode := range allIndividuals {
		if indiNode.Individual != nil {
			// Use GetNamesParsed() to get structured names
			names, err := indiNode.Individual.GetNamesParsed()
			if err != nil {
				// Continue with other individuals if parsing fails
				continue
			}

			for _, name := range names {
				if name != nil && name.IsValid() {
					// Extract given name
					if name.Given != "" {
						givenNamesSet[name.Given] = true
					}
					// Extract surname
					if name.Surname != "" {
						surnamesSet[name.Surname] = true
					}
				}
			}
		}
	}

	// Convert sets to slices
	givenNames := make([]string, 0, len(givenNamesSet))
	for name := range givenNamesSet {
		givenNames = append(givenNames, name)
	}

	surnames := make([]string, 0, len(surnamesSet))
	for name := range surnamesSet {
		surnames = append(surnames, name)
	}

	return map[string][]string{
		"given":   givenNames,
		"surname": surnames,
	}, nil
}

// IndividualQuery represents a query starting from a specific individual.
type IndividualQuery struct {
	xrefID string
	graph  *Graph
}

// Ancestors finds all ancestors of this individual.
func (iq *IndividualQuery) Ancestors() *AncestorQuery {
	return &AncestorQuery{
		startXrefID: iq.xrefID,
		graph:       iq.graph,
		options:     NewAncestorOptions(),
	}
}

// Descendants finds all descendants of this individual.
func (iq *IndividualQuery) Descendants() *DescendantQuery {
	return &DescendantQuery{
		startXrefID: iq.xrefID,
		graph:       iq.graph,
		options:     NewDescendantOptions(),
	}
}

// GetSubtree returns ancestors and descendants of this individual.
func (iq *IndividualQuery) GetSubtree() *SubtreeQuery {
	return &SubtreeQuery{
		startXrefID: iq.xrefID,
		graph:       iq.graph,
		options:     NewSubtreeOptions(),
	}
}

// Parents returns direct parents.
// Results are cached for repeated queries.
func (iq *IndividualQuery) Parents() ([]*types.IndividualRecord, error) {
	// Check cache
	cacheKey := makeCacheKey("parents", iq.xrefID)
	if cached, ok := iq.graph.cache.get(cacheKey); ok {
		if result, ok := cached.([]*types.IndividualRecord); ok {
			return result, nil
		}
	}

	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	// Compute parents from edges (no longer cached in node)
	parentNodes := node.getParentsFromEdges()
	parents := make([]*types.IndividualRecord, 0, len(parentNodes))
	for _, parentNode := range parentNodes {
		if parentNode.Individual != nil {
			parents = append(parents, parentNode.Individual)
		}
	}

	// Cache result
	iq.graph.cache.set(cacheKey, parents)

	return parents, nil
}

// Children returns direct children.
// Results are cached for repeated queries.
func (iq *IndividualQuery) Children() ([]*types.IndividualRecord, error) {
	// Check cache
	cacheKey := makeCacheKey("children", iq.xrefID)
	if cached, ok := iq.graph.cache.get(cacheKey); ok {
		if result, ok := cached.([]*types.IndividualRecord); ok {
			return result, nil
		}
	}

	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	// Compute children from edges (no longer cached in node)
	childNodes := node.getChildrenFromEdges()
	children := make([]*types.IndividualRecord, 0, len(childNodes))
	for _, childNode := range childNodes {
		if childNode.Individual != nil {
			children = append(children, childNode.Individual)
		}
	}

	// Cache result
	iq.graph.cache.set(cacheKey, children)

	return children, nil
}

// Siblings returns siblings (full and half).
func (iq *IndividualQuery) Siblings() ([]*types.IndividualRecord, error) {
	// Check cache
	cacheKey := makeCacheKey("siblings", iq.xrefID)
	if cached, ok := iq.graph.cache.get(cacheKey); ok {
		if result, ok := cached.([]*types.IndividualRecord); ok {
			return result, nil
		}
	}

	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	// Compute siblings from edges (no longer cached in node)
	siblingNodes := node.getSiblingsFromEdges()
	siblings := make([]*types.IndividualRecord, 0, len(siblingNodes))
	for _, siblingNode := range siblingNodes {
		if siblingNode.Individual != nil {
			siblings = append(siblings, siblingNode.Individual)
		}
	}

	// Cache result
	iq.graph.cache.set(cacheKey, siblings)

	return siblings, nil
}

// Spouses returns all spouses.
func (iq *IndividualQuery) Spouses() ([]*types.IndividualRecord, error) {
	// Check cache
	cacheKey := makeCacheKey("spouses", iq.xrefID)
	if cached, ok := iq.graph.cache.get(cacheKey); ok {
		if result, ok := cached.([]*types.IndividualRecord); ok {
			return result, nil
		}
	}

	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	// Compute spouses from edges (no longer cached in node)
	spouseNodes := node.getSpousesFromEdges()
	spouses := make([]*types.IndividualRecord, 0, len(spouseNodes))
	for _, spouseNode := range spouseNodes {
		if spouseNode.Individual != nil {
			spouses = append(spouses, spouseNode.Individual)
		}
	}

	// Cache result
	iq.graph.cache.set(cacheKey, spouses)

	return spouses, nil
}

// RelationshipTo finds relationship to another individual.
func (iq *IndividualQuery) RelationshipTo(otherXrefID string) *RelationshipQuery {
	return &RelationshipQuery{
		fromXrefID: iq.xrefID,
		toXrefID:   otherXrefID,
		graph:      iq.graph,
	}
}

// RelationshipTo returns the relationship result directly (convenience method).
func (iq *IndividualQuery) RelationshipToResult(otherXrefID string) (*RelationshipResult, error) {
	return iq.graph.CalculateRelationship(iq.xrefID, otherXrefID)
}

// PathTo finds path(s) to another individual.
func (iq *IndividualQuery) PathTo(otherXrefID string) *PathQuery {
	return &PathQuery{
		fromXrefID: iq.xrefID,
		toXrefID:   otherXrefID,
		graph:      iq.graph,
		options:    NewPathOptions(),
	}
}

// CommonAncestors finds common ancestors with another individual.
func (iq *IndividualQuery) CommonAncestors(otherXrefID string) ([]*types.IndividualRecord, error) {
	nodes, err := iq.graph.CommonAncestors(iq.xrefID, otherXrefID)
	if err != nil {
		return nil, err
	}

	records := make([]*types.IndividualRecord, 0, len(nodes))
	for _, node := range nodes {
		if node.Individual != nil {
			records = append(records, node.Individual)
		}
	}

	return records, nil
}

// Cousins finds cousins (configurable degree).
func (iq *IndividualQuery) Cousins(degree int) ([]*types.IndividualRecord, error) {
	// Get all individuals
	allIndividuals := iq.graph.GetAllIndividuals()
	cousins := make([]*types.IndividualRecord, 0)

	for _, otherNode := range allIndividuals {
		if otherNode.ID() == iq.xrefID {
			continue
		}

		result, err := iq.graph.CalculateRelationship(iq.xrefID, otherNode.ID())
		if err != nil {
			continue
		}

		if result.IsCollateral && result.Degree == degree && result.Removal == 0 {
			if otherNode.Individual != nil {
				cousins = append(cousins, otherNode.Individual)
			}
		}
	}

	return cousins, nil
}

// Uncles finds uncles/aunts.
func (iq *IndividualQuery) Uncles() ([]*types.IndividualRecord, error) {
	// Uncles are siblings of parents
	parents, err := iq.Parents()
	if err != nil {
		return nil, err
	}

	uncles := make([]*types.IndividualRecord, 0)
	for _, parent := range parents {
		parentQuery := iq.graph.GetIndividual(parent.XrefID())
		if parentQuery == nil {
			continue
		}

		siblingNodes := parentQuery.getSiblingsFromEdges()
		for _, siblingNode := range siblingNodes {
			// Exclude the parent itself
			if siblingNode.ID() != parent.XrefID() && siblingNode.Individual != nil {
				uncles = append(uncles, siblingNode.Individual)
			}
		}
	}

	return uncles, nil
}

// Nephews finds nephews/nieces.
func (iq *IndividualQuery) Nephews() ([]*types.IndividualRecord, error) {
	// Nephews are children of siblings
	siblings, err := iq.Siblings()
	if err != nil {
		return nil, err
	}

	nephews := make([]*types.IndividualRecord, 0)
	for _, sibling := range siblings {
		siblingNode := iq.graph.GetIndividual(sibling.XrefID())
		if siblingNode == nil {
			continue
		}

		childNodes := siblingNode.getChildrenFromEdges()
		for _, childNode := range childNodes {
			if childNode.Individual != nil {
				nephews = append(nephews, childNode.Individual)
			}
		}
	}

	return nephews, nil
}

// Grandparents returns grandparents.
func (iq *IndividualQuery) Grandparents() ([]*types.IndividualRecord, error) {
	parents, err := iq.Parents()
	if err != nil {
		return nil, err
	}

	grandparents := make([]*types.IndividualRecord, 0)
	for _, parent := range parents {
		parentNode := iq.graph.GetIndividual(parent.XrefID())
		if parentNode == nil {
			continue
		}

		grandparentNodes := parentNode.getParentsFromEdges()
		for _, grandparentNode := range grandparentNodes {
			if grandparentNode.Individual != nil {
				grandparents = append(grandparents, grandparentNode.Individual)
			}
		}
	}

	return grandparents, nil
}

// Grandchildren returns grandchildren.
func (iq *IndividualQuery) Grandchildren() ([]*types.IndividualRecord, error) {
	children, err := iq.Children()
	if err != nil {
		return nil, err
	}

	grandchildren := make([]*types.IndividualRecord, 0)
	for _, child := range children {
		childNode := iq.graph.GetIndividual(child.XrefID())
		if childNode == nil {
			continue
		}

		grandchildNodes := childNode.getChildrenFromEdges()
		for _, grandchildNode := range grandchildNodes {
			if grandchildNode.Individual != nil {
				grandchildren = append(grandchildren, grandchildNode.Individual)
			}
		}
	}

	return grandchildren, nil
}
