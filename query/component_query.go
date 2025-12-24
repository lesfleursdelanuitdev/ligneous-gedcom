package query

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// ComponentOptions holds configuration for component queries.
type ComponentOptions struct {
	MaxDepth int // Limit traversal depth (0 = unlimited)
	MaxSize  int // Limit result size (0 = unlimited)
}

// NewComponentOptions creates new ComponentOptions with defaults.
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{
		MaxDepth: 0, // Unlimited
		MaxSize:  0, // Unlimited
	}
}

// GetComponentForPerson returns all individuals in the connected component
// containing the specified person.
func (g *Graph) GetComponentForPerson(personID string) ([]*types.IndividualRecord, error) {
	return g.GetComponentForPersonWithOptions(personID, NewComponentOptions())
}

// GetComponentForPersonWithOptions returns connected component with options.
func (g *Graph) GetComponentForPersonWithOptions(personID string, options *ComponentOptions) ([]*types.IndividualRecord, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Get starting node
	startNode := g.GetIndividual(personID)
	if startNode == nil {
		return nil, nil
	}

	// Use BFS to find all connected individuals
	visited := make(map[string]bool)
	component := make([]*types.IndividualRecord, 0)
	queue := []*IndividualNode{startNode}
	visited[personID] = true
	depth := 0

	for len(queue) > 0 {
		// Check depth limit
		if options.MaxDepth > 0 && depth > options.MaxDepth {
			break
		}

		// Check size limit
		if options.MaxSize > 0 && len(component) >= options.MaxSize {
			break
		}

		current := queue[0]
		queue = queue[1:]

		// Add current individual to component
		if current.Individual != nil {
			component = append(component, current.Individual)
		}

		// Find connected individuals through all edges
		connectedIndividuals := make(map[string]*IndividualNode)

		// Through FAMC edges (parents and siblings)
		for _, edge := range current.OutEdges() {
			if edge.EdgeType == EdgeTypeFAMC && edge.Family != nil {
				famNode := edge.Family
				husband := famNode.getHusbandFromEdges()
				if husband != nil {
					connectedIndividuals[husband.ID()] = husband
				}
				wife := famNode.getWifeFromEdges()
				if wife != nil {
					connectedIndividuals[wife.ID()] = wife
				}
				// Also add siblings (other children of same family)
				children := famNode.getChildrenFromEdges()
				for _, child := range children {
					if child.ID() != current.ID() {
						connectedIndividuals[child.ID()] = child
					}
				}
			}
		}

		// Through FAMS edges (spouses and children)
		for _, edge := range current.OutEdges() {
			if edge.EdgeType == EdgeTypeFAMS && edge.Family != nil {
				famNode := edge.Family
				husband := famNode.getHusbandFromEdges()
				if husband != nil && husband.ID() != current.ID() {
					connectedIndividuals[husband.ID()] = husband
				}
				wife := famNode.getWifeFromEdges()
				if wife != nil && wife.ID() != current.ID() {
					connectedIndividuals[wife.ID()] = wife
				}
				children := famNode.getChildrenFromEdges()
				for _, child := range children {
					connectedIndividuals[child.ID()] = child
				}
			}
		}

		// Add connected individuals to queue
		for neighborID, neighborNode := range connectedIndividuals {
			if !visited[neighborID] {
				visited[neighborID] = true
				queue = append(queue, neighborNode)
			}
		}

		depth++
	}

	return component, nil
}

