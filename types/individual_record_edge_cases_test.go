package types

import (
	"testing"
)

// TestIndividualRecord_GetNamesParsed_ErrorHandling tests error handling in GetNamesParsed.
func TestIndividualRecord_GetNamesParsed_ErrorHandling(t *testing.T) {
	// Test with name that fails to parse (should continue with other names)
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	
	// Valid name
	name1Line := NewGedcomLine(1, "NAME", "John /Doe/", "")
	name1Line.AddChild(NewGedcomLine(2, "GIVN", "John", ""))
	name1Line.AddChild(NewGedcomLine(2, "SURN", "Doe", ""))
	line.AddChild(name1Line)
	
	// Potentially problematic name (empty)
	name2Line := NewGedcomLine(1, "NAME", "", "")
	line.AddChild(name2Line)
	
	record := NewIndividualRecord(line)
	names, err := record.GetNamesParsed()
	if err != nil {
		t.Fatalf("GetNamesParsed() returned error: %v", err)
	}
	// Should still return at least one valid name
	if len(names) == 0 {
		t.Error("GetNamesParsed() should return valid names even if some fail to parse")
	}
}

// TestIndividualRecord_Events tests the Events method.
func TestIndividualRecord_Events(t *testing.T) {
	// Test with multiple events
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	birtLine.AddChild(NewGedcomLine(2, "DATE", "1 Jan 1900", ""))
	line.AddChild(birtLine)
	
	deatLine := NewGedcomLine(1, "DEAT", "", "")
	deatLine.AddChild(NewGedcomLine(2, "DATE", "1 Jan 2000", ""))
	line.AddChild(deatLine)
	
	record := NewIndividualRecord(line)
	events := record.Events()
	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}
	
	// Test with no events
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	events2 := record2.Events()
	if len(events2) != 0 {
		t.Errorf("Expected 0 events, got %d", len(events2))
	}
}

// TestIndividualRecord_EventsByType_AllTypes tests EventsByType with various event types.
func TestIndividualRecord_EventsByType_AllTypes(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	
	// Add various event types
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	birtLine.AddChild(NewGedcomLine(2, "DATE", "1 Jan 1900", ""))
	line.AddChild(birtLine)
	
	bapmLine := NewGedcomLine(1, "BAPM", "", "")
	bapmLine.AddChild(NewGedcomLine(2, "DATE", "1 Feb 1900", ""))
	line.AddChild(bapmLine)
	
	buriLine := NewGedcomLine(1, "BURI", "", "")
	buriLine.AddChild(NewGedcomLine(2, "DATE", "1 Jan 2000", ""))
	line.AddChild(buriLine)
	
	record := NewIndividualRecord(line)
	
	// Test each event type
	births := record.EventsByType(EventTypeBirth)
	if len(births) != 1 {
		t.Errorf("Expected 1 birth event, got %d", len(births))
	}
	
	baptisms := record.EventsByType(EventTypeBaptism)
	if len(baptisms) != 1 {
		t.Errorf("Expected 1 baptism event, got %d", len(baptisms))
	}
	
	burials := record.EventsByType(EventTypeBurial)
	if len(burials) != 1 {
		t.Errorf("Expected 1 burial event, got %d", len(burials))
	}
	
	// Test with non-existent type
	deaths := record.EventsByType(EventTypeDeath)
	if len(deaths) != 0 {
		t.Errorf("Expected 0 death events, got %d", len(deaths))
	}
}

// TestIndividualRecord_FamilyWithSpouse_NilSpouse tests FamilyWithSpouse with nil spouse.
func TestIndividualRecord_FamilyWithSpouse_NilSpouse(t *testing.T) {
	// Create a tree
	tree := NewGedcomTree()
	
	indiLine := NewGedcomLine(0, "INDI", "", "@I1@")
	indi := NewIndividualRecord(indiLine)
	tree.AddRecord(indi)
	
	// Test with nil spouse
	foundFamily, err := indi.FamilyWithSpouse(nil)
	if err == nil {
		t.Error("FamilyWithSpouse() should return error when spouse is nil")
	}
	if foundFamily != nil {
		t.Error("FamilyWithSpouse() should return nil when spouse is nil")
	}
}

// TestIndividualRecord_FamilyWithUnknownSpouse_MultipleFamilies tests with multiple families.
func TestIndividualRecord_FamilyWithUnknownSpouse_MultipleFamilies(t *testing.T) {
	// Create a tree
	tree := NewGedcomTree()
	
	indiLine := NewGedcomLine(0, "INDI", "", "@I1@")
	indi := NewIndividualRecord(indiLine)
	tree.AddRecord(indi)
	
	// Create family with unknown spouse
	fam1Line := NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	family1 := NewFamilyRecord(fam1Line)
	tree.AddRecord(family1)
	
	// Create family with known spouse
	indi2Line := NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)
	
	fam2Line := NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam2Line.AddChild(NewGedcomLine(1, "WIFE", "@I2@", ""))
	family2 := NewFamilyRecord(fam2Line)
	tree.AddRecord(family2)
	
	// Add FAMS to individual
	indiLine.AddChild(NewGedcomLine(1, "FAMS", "@F1@", ""))
	indiLine.AddChild(NewGedcomLine(1, "FAMS", "@F2@", ""))
	
	// Test FamilyWithUnknownSpouse (should find F1)
	foundFamily, err := indi.FamilyWithUnknownSpouse()
	if err != nil {
		t.Fatalf("FamilyWithUnknownSpouse() returned error: %v", err)
	}
	if foundFamily == nil {
		t.Fatal("FamilyWithUnknownSpouse() should find family with unknown spouse")
	}
	if foundFamily.XrefID() != "@F1@" {
		t.Errorf("Expected family @F1@, got %s", foundFamily.XrefID())
	}
}

// TestIndividualRecord_SpouseChildren_UnknownSpouse tests SpouseChildren with unknown spouse.
func TestIndividualRecord_SpouseChildren_UnknownSpouse(t *testing.T) {
	// Create a tree
	tree := NewGedcomTree()
	
	indiLine := NewGedcomLine(0, "INDI", "", "@I1@")
	indi := NewIndividualRecord(indiLine)
	tree.AddRecord(indi)
	
	// Create child
	childLine := NewGedcomLine(0, "INDI", "", "@I3@")
	child := NewIndividualRecord(childLine)
	tree.AddRecord(child)
	
	// Create family with unknown spouse
	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(NewGedcomLine(1, "CHIL", "@I3@", ""))
	family := NewFamilyRecord(famLine)
	tree.AddRecord(family)
	
	// Add FAMS to individual
	indiLine.AddChild(NewGedcomLine(1, "FAMS", "@F1@", ""))
	
	// NOTE: SpouseChildren() method has been removed.
	// Use the graph package for relationship queries instead.
	// This test is kept for reference but does not test the removed method.
	t.Skip("SpouseChildren() method has been removed - use graph package instead")
}

// TestIndividualRecord_SpouseChildren_MultipleSpouses tests SpouseChildren with multiple spouses.
func TestIndividualRecord_SpouseChildren_MultipleSpouses(t *testing.T) {
	// Create a tree
	tree := NewGedcomTree()
	
	indi1Line := NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)
	
	indi2Line := NewGedcomLine(0, "INDI", "", "@I2@")
	indi2 := NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)
	
	indi3Line := NewGedcomLine(0, "INDI", "", "@I3@")
	indi3 := NewIndividualRecord(indi3Line)
	tree.AddRecord(indi3)
	
	// Create children
	child1Line := NewGedcomLine(0, "INDI", "", "@I4@")
	child1 := NewIndividualRecord(child1Line)
	tree.AddRecord(child1)
	
	child2Line := NewGedcomLine(0, "INDI", "", "@I5@")
	child2 := NewIndividualRecord(child2Line)
	tree.AddRecord(child2)
	
	// Create family 1 with spouse 2
	fam1Line := NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(NewGedcomLine(1, "WIFE", "@I2@", ""))
	fam1Line.AddChild(NewGedcomLine(1, "CHIL", "@I4@", ""))
	family1 := NewFamilyRecord(fam1Line)
	tree.AddRecord(family1)
	
	// Create family 2 with spouse 3
	fam2Line := NewGedcomLine(0, "FAM", "", "@F2@")
	fam2Line.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam2Line.AddChild(NewGedcomLine(1, "WIFE", "@I3@", ""))
	fam2Line.AddChild(NewGedcomLine(1, "CHIL", "@I5@", ""))
	family2 := NewFamilyRecord(fam2Line)
	tree.AddRecord(family2)
	
	// Add FAMS to individual 1
	indi1Line.AddChild(NewGedcomLine(1, "FAMS", "@F1@", ""))
	indi1Line.AddChild(NewGedcomLine(1, "FAMS", "@F2@", ""))
	
	// NOTE: SpouseChildren() method has been removed.
	// Use the graph package for relationship queries instead.
	// This test is kept for reference but does not test the removed method.
	t.Skip("SpouseChildren() method has been removed - use graph package instead")
}

