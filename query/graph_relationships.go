package query

import (
	"fmt"
)

// GetSpouses returns all spouses of an individual by XREF ID.
// This is a convenience method that wraps node.Spouses().
// Returns an error if the individual is not found.
func (g *Graph) GetSpouses(xrefID string) ([]*IndividualNode, error) {
	node := g.GetIndividual(xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", xrefID)
	}
	return node.Spouses(), nil
}

// GetChildren returns all children of an individual by XREF ID.
// This is a convenience method that wraps node.Children().
// Returns an error if the individual is not found.
func (g *Graph) GetChildren(xrefID string) ([]*IndividualNode, error) {
	node := g.GetIndividual(xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", xrefID)
	}
	return node.Children(), nil
}

// GetParents returns all parents of an individual by XREF ID.
// This is a convenience method that wraps node.Parents().
// Returns an error if the individual is not found.
func (g *Graph) GetParents(xrefID string) ([]*IndividualNode, error) {
	node := g.GetIndividual(xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", xrefID)
	}
	return node.Parents(), nil
}

// GetSiblings returns all siblings of an individual by XREF ID.
// This is a convenience method that wraps node.Siblings().
// Returns an error if the individual is not found.
func (g *Graph) GetSiblings(xrefID string) ([]*IndividualNode, error) {
	node := g.GetIndividual(xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", xrefID)
	}
	return node.Siblings(), nil
}

// GetFamilyHusband returns the husband of a family by XREF ID.
// This is a convenience method that wraps familyNode.Husband().
// Returns an error if the family is not found.
func (g *Graph) GetFamilyHusband(familyXrefID string) (*IndividualNode, error) {
	node := g.GetFamily(familyXrefID)
	if node == nil {
		return nil, fmt.Errorf("family %s not found", familyXrefID)
	}
	return node.Husband(), nil
}

// GetFamilyWife returns the wife of a family by XREF ID.
// This is a convenience method that wraps familyNode.Wife().
// Returns an error if the family is not found.
func (g *Graph) GetFamilyWife(familyXrefID string) (*IndividualNode, error) {
	node := g.GetFamily(familyXrefID)
	if node == nil {
		return nil, fmt.Errorf("family %s not found", familyXrefID)
	}
	return node.Wife(), nil
}

// GetFamilyChildren returns all children of a family by XREF ID.
// This is a convenience method that wraps familyNode.Children().
// Returns an error if the family is not found.
func (g *Graph) GetFamilyChildren(familyXrefID string) ([]*IndividualNode, error) {
	node := g.GetFamily(familyXrefID)
	if node == nil {
		return nil, fmt.Errorf("family %s not found", familyXrefID)
	}
	return node.Children(), nil
}


