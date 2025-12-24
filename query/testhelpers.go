package query

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// CreateTestTree creates a new empty GEDCOM tree for testing.
func CreateTestTree() *types.GedcomTree {
	return types.NewGedcomTree()
}

// CreateTestIndividual creates a test individual record with the given XREF and name.
func CreateTestIndividual(xref, name string) *types.IndividualRecord {
	indiLine := types.NewGedcomLine(0, "INDI", "", xref)
	if name != "" {
		indiLine.AddChild(types.NewGedcomLine(1, "NAME", name, ""))
	}
	return types.NewIndividualRecord(indiLine)
}

// CreateTestIndividualWithName creates a test individual with a structured name.
func CreateTestIndividualWithName(xref, name string) *types.IndividualRecord {
	indiLine := types.NewGedcomLine(0, "INDI", "", xref)
	nameLine := types.NewGedcomLine(1, "NAME", name, "")
	indiLine.AddChild(nameLine)
	return types.NewIndividualRecord(indiLine)
}

// CreateTestIndividualWithBirth creates a test individual with birth information.
func CreateTestIndividualWithBirth(xref, name, birthDate, birthPlace string) *types.IndividualRecord {
	indiLine := types.NewGedcomLine(0, "INDI", "", xref)
	if name != "" {
		indiLine.AddChild(types.NewGedcomLine(1, "NAME", name, ""))
	}
	if birthDate != "" || birthPlace != "" {
		birtLine := types.NewGedcomLine(1, "BIRT", "", "")
		if birthDate != "" {
			birtLine.AddChild(types.NewGedcomLine(2, "DATE", birthDate, ""))
		}
		if birthPlace != "" {
			birtLine.AddChild(types.NewGedcomLine(2, "PLAC", birthPlace, ""))
		}
		indiLine.AddChild(birtLine)
	}
	return types.NewIndividualRecord(indiLine)
}

// CreateTestFamily creates a test family record.
func CreateTestFamily(xref string, husband, wife string, children []string) *types.FamilyRecord {
	famLine := types.NewGedcomLine(0, "FAM", "", xref)
	if husband != "" {
		famLine.AddChild(types.NewGedcomLine(1, "HUSB", husband, ""))
	}
	if wife != "" {
		famLine.AddChild(types.NewGedcomLine(1, "WIFE", wife, ""))
	}
	for _, child := range children {
		famLine.AddChild(types.NewGedcomLine(1, "CHIL", child, ""))
	}
	return types.NewFamilyRecord(famLine)
}

// CreateTestFamilyWithMarriage creates a test family with marriage information.
func CreateTestFamilyWithMarriage(xref, husband, wife string, marriageDate, marriagePlace string) *types.FamilyRecord {
	famLine := types.NewGedcomLine(0, "FAM", "", xref)
	if husband != "" {
		famLine.AddChild(types.NewGedcomLine(1, "HUSB", husband, ""))
	}
	if wife != "" {
		famLine.AddChild(types.NewGedcomLine(1, "WIFE", wife, ""))
	}
	if marriageDate != "" || marriagePlace != "" {
		marrLine := types.NewGedcomLine(1, "MARR", "", "")
		if marriageDate != "" {
			marrLine.AddChild(types.NewGedcomLine(2, "DATE", marriageDate, ""))
		}
		if marriagePlace != "" {
			marrLine.AddChild(types.NewGedcomLine(2, "PLAC", marriagePlace, ""))
		}
		famLine.AddChild(marrLine)
	}
	return types.NewFamilyRecord(famLine)
}

// CreateTestQuery creates a QueryBuilder from a test tree.
func CreateTestQuery(tree *types.GedcomTree) (*QueryBuilder, error) {
	return NewQuery(tree)
}

// CreateTestGraph creates a Graph from a test tree.
func CreateTestGraph(tree *types.GedcomTree) (*Graph, error) {
	return BuildGraph(tree)
}

// AddTestIndividual adds an individual to a tree and returns it.
func AddTestIndividual(tree *types.GedcomTree, xref, name string) *types.IndividualRecord {
	indi := CreateTestIndividual(xref, name)
	tree.AddRecord(indi)
	return indi
}

// AddTestFamily adds a family to a tree and returns it.
func AddTestFamily(tree *types.GedcomTree, xref string, husband, wife string, children []string) *types.FamilyRecord {
	fam := CreateTestFamily(xref, husband, wife, children)
	tree.AddRecord(fam)
	return fam
}

