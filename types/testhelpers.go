package types

// CreateTestIndividual is a convenience function for creating test individuals.
// This is a simple wrapper that can be extended with more options if needed.
func CreateTestIndividual(xref, name string) *IndividualRecord {
	indiLine := NewGedcomLine(0, "INDI", "", xref)
	if name != "" {
		indiLine.AddChild(NewGedcomLine(1, "NAME", name, ""))
	}
	return NewIndividualRecord(indiLine)
}

// CreateTestFamily is a convenience function for creating test families.
func CreateTestFamily(xref string, husband, wife string, children []string) *FamilyRecord {
	famLine := NewGedcomLine(0, "FAM", "", xref)
	if husband != "" {
		famLine.AddChild(NewGedcomLine(1, "HUSB", husband, ""))
	}
	if wife != "" {
		famLine.AddChild(NewGedcomLine(1, "WIFE", wife, ""))
	}
	for _, child := range children {
		famLine.AddChild(NewGedcomLine(1, "CHIL", child, ""))
	}
	return NewFamilyRecord(famLine)
}

// CreateTestTree creates a new empty GEDCOM tree for testing.
func CreateTestTree() *GedcomTree {
	return NewGedcomTree()
}

