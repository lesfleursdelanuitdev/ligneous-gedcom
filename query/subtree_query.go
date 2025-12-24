package query

import (
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// SubtreeOptions holds configuration for subtree queries.
type SubtreeOptions struct {
	AncestorGenerations   int                                 // Max ancestor generations (0 = unlimited)
	DescendantGenerations int                                 // Max descendant generations (0 = unlimited)
	IncludeSelf           bool                                // Include starting individual
	IncludeSiblings       bool                                // Include siblings
	IncludeSpouses        bool                                // Include spouses
	Filter                func(*types.IndividualRecord) bool // Custom filter function
}

// NewSubtreeOptions creates new SubtreeOptions with defaults.
func NewSubtreeOptions() *SubtreeOptions {
	return &SubtreeOptions{
		AncestorGenerations:   0, // Unlimited
		DescendantGenerations: 0, // Unlimited
		IncludeSelf:           true,
		IncludeSiblings:       false,
		IncludeSpouses:        false,
		Filter:                nil,
	}
}

// SubtreeQuery represents a query for a subtree (ancestors + descendants).
type SubtreeQuery struct {
	startXrefID string
	graph       *Graph
	options     *SubtreeOptions
}

// AncestorGenerations limits the depth of ancestor search.
func (sq *SubtreeQuery) AncestorGenerations(n int) *SubtreeQuery {
	sq.options.AncestorGenerations = n
	return sq
}

// DescendantGenerations limits the depth of descendant search.
func (sq *SubtreeQuery) DescendantGenerations(n int) *SubtreeQuery {
	sq.options.DescendantGenerations = n
	return sq
}

// IncludeSelf includes the starting individual in results.
func (sq *SubtreeQuery) IncludeSelf() *SubtreeQuery {
	sq.options.IncludeSelf = true
	return sq
}

// ExcludeSelf excludes the starting individual from results.
func (sq *SubtreeQuery) ExcludeSelf() *SubtreeQuery {
	sq.options.IncludeSelf = false
	return sq
}

// IncludeSiblings includes siblings in results.
func (sq *SubtreeQuery) IncludeSiblings() *SubtreeQuery {
	sq.options.IncludeSiblings = true
	return sq
}

// IncludeSpouses includes spouses in results.
func (sq *SubtreeQuery) IncludeSpouses() *SubtreeQuery {
	sq.options.IncludeSpouses = true
	return sq
}

// Filter applies a custom filter function to results.
func (sq *SubtreeQuery) Filter(fn func(*types.IndividualRecord) bool) *SubtreeQuery {
	sq.options.Filter = fn
	return sq
}

// SubtreeResult represents the result of a subtree query.
type SubtreeResult struct {
	Root        *types.IndividualRecord   // The starting individual
	Ancestors   []*types.IndividualRecord // Ancestors (upward)
	Descendants []*types.IndividualRecord // Descendants (downward)
	Siblings    []*types.IndividualRecord // Siblings (if included)
	Spouses     []*types.IndividualRecord // Spouses (if included)
	All         []*types.IndividualRecord // Combined list (deduplicated)
}

// Execute runs the query and returns subtree results.
func (sq *SubtreeQuery) Execute() (*SubtreeResult, error) {
	// Record metrics if available
	start := time.Now()
	defer func() {
		if sq.graph.metrics != nil {
			duration := time.Since(start)
			sq.graph.metrics.RecordQuery(duration)
		}
	}()

	// Get root individual
	rootNode := sq.graph.GetIndividual(sq.startXrefID)
	if rootNode == nil {
		return nil, nil
	}

	result := &SubtreeResult{}

	// Get root record
	if rootNode.Individual != nil {
		result.Root = rootNode.Individual
	}

	// Get ancestors
	if sq.options.AncestorGenerations != 0 || sq.options.AncestorGenerations == 0 {
		ancestorQuery := &AncestorQuery{
			startXrefID: sq.startXrefID,
			graph:       sq.graph,
			options: &AncestorOptions{
				MaxGenerations: sq.options.AncestorGenerations,
				IncludeSelf:    false, // Don't include self in ancestors
				Filter:         sq.options.Filter,
				Order:          OrderBFS,
			},
		}
		ancestors, err := ancestorQuery.Execute()
		if err != nil {
			return nil, err
		}
		result.Ancestors = ancestors
	}

	// Get descendants
	if sq.options.DescendantGenerations != 0 || sq.options.DescendantGenerations == 0 {
		descendantQuery := &DescendantQuery{
			startXrefID: sq.startXrefID,
			graph:       sq.graph,
			options: &DescendantOptions{
				MaxGenerations: sq.options.DescendantGenerations,
				IncludeSelf:    false, // Don't include self in descendants
				Filter:         sq.options.Filter,
				Order:          OrderBFS,
			},
		}
		descendants, err := descendantQuery.Execute()
		if err != nil {
			return nil, err
		}
		result.Descendants = descendants
	}

	// Get siblings if requested
	if sq.options.IncludeSiblings {
		siblingNodes := rootNode.getSiblingsFromEdges()
		for _, siblingNode := range siblingNodes {
			if siblingNode.Individual != nil {
				// Apply filter if provided
				if sq.options.Filter == nil || sq.options.Filter(siblingNode.Individual) {
					result.Siblings = append(result.Siblings, siblingNode.Individual)
				}
			}
		}
	}

	// Get spouses if requested
	if sq.options.IncludeSpouses {
		spouseNodes := rootNode.getSpousesFromEdges()
		for _, spouseNode := range spouseNodes {
			if spouseNode.Individual != nil {
				// Apply filter if provided
				if sq.options.Filter == nil || sq.options.Filter(spouseNode.Individual) {
					result.Spouses = append(result.Spouses, spouseNode.Individual)
				}
			}
		}
	}

	// Build combined list (deduplicated)
	seen := make(map[string]bool)
	all := make([]*types.IndividualRecord, 0)

	// Add root if requested
	if sq.options.IncludeSelf && result.Root != nil {
		if sq.options.Filter == nil || sq.options.Filter(result.Root) {
			all = append(all, result.Root)
			seen[result.Root.XrefID()] = true
		}
	}

	// Add ancestors
	for _, ancestor := range result.Ancestors {
		if !seen[ancestor.XrefID()] {
			all = append(all, ancestor)
			seen[ancestor.XrefID()] = true
		}
	}

	// Add descendants
	for _, descendant := range result.Descendants {
		if !seen[descendant.XrefID()] {
			all = append(all, descendant)
			seen[descendant.XrefID()] = true
		}
	}

	// Add siblings
	for _, sibling := range result.Siblings {
		if !seen[sibling.XrefID()] {
			all = append(all, sibling)
			seen[sibling.XrefID()] = true
		}
	}

	// Add spouses
	for _, spouse := range result.Spouses {
		if !seen[spouse.XrefID()] {
			all = append(all, spouse)
			seen[spouse.XrefID()] = true
		}
	}

	result.All = all

	return result, nil
}

// Count returns the total number of individuals in the subtree.
func (sq *SubtreeQuery) Count() (int, error) {
	result, err := sq.Execute()
	if err != nil {
		return 0, err
	}
	return len(result.All), nil
}

// ExecuteRecords returns just the combined list of records (convenience method).
func (sq *SubtreeQuery) ExecuteRecords() ([]*types.IndividualRecord, error) {
	result, err := sq.Execute()
	if err != nil {
		return nil, err
	}
	return result.All, nil
}

