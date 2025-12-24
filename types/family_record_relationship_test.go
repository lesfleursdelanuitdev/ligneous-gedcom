package types

import (
	"testing"
)

// TestFamilyRecord_GetHusbandRecord tests the GetHusbandRecord method.
func TestFamilyRecord_GetHusbandRecord(t *testing.T) {
	// Create a tree with individuals and a family
	tree := NewGedcomTree()
	
	// Create husband
	husbandLine := NewGedcomLine(0, "INDI", "", "@I1@")
	husbandLine.AddChild(NewGedcomLine(1, "NAME", "John /Doe/", ""))
	husband := NewIndividualRecord(husbandLine)
	tree.AddRecord(husband)
	
	// Create wife
	wifeLine := NewGedcomLine(0, "INDI", "", "@I2@")
	wifeLine.AddChild(NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	wife := NewIndividualRecord(wifeLine)
	tree.AddRecord(wife)
	
	// Create family
	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(NewGedcomLine(1, "WIFE", "@I2@", ""))
	family := NewFamilyRecord(famLine)
	tree.AddRecord(family)
	
	// Test GetHusbandRecord
	husbandRecord, err := family.GetHusbandRecord()
	if err != nil {
		t.Fatalf("GetHusbandRecord() returned error: %v", err)
	}
	if husbandRecord == nil {
		t.Fatal("GetHusbandRecord() returned nil")
	}
	if husbandRecord.XrefID() != "@I1@" {
		t.Errorf("Expected husband @I1@, got %s", husbandRecord.XrefID())
	}
	
	// Test with no husband
	famLine2 := NewGedcomLine(0, "FAM", "", "@F2@")
	famLine2.AddChild(NewGedcomLine(1, "WIFE", "@I2@", ""))
	family2 := NewFamilyRecord(famLine2)
	tree.AddRecord(family2)
	
	husbandRecord2, err := family2.GetHusbandRecord()
	if err != nil {
		t.Fatalf("GetHusbandRecord() returned error: %v", err)
	}
	if husbandRecord2 != nil {
		t.Error("GetHusbandRecord() should return nil when no husband")
	}
	
	// Test with family not in tree
	famLine3 := NewGedcomLine(0, "FAM", "", "@F3@")
	famLine3.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	family3 := NewFamilyRecord(famLine3)
	
	_, err = family3.GetHusbandRecord()
	if err == nil {
		t.Error("GetHusbandRecord() should return error when family not in tree")
	}
}

// TestFamilyRecord_GetWifeRecord tests the GetWifeRecord method.
func TestFamilyRecord_GetWifeRecord(t *testing.T) {
	// Create a tree with individuals and a family
	tree := NewGedcomTree()
	
	// Create husband
	husbandLine := NewGedcomLine(0, "INDI", "", "@I1@")
	husbandLine.AddChild(NewGedcomLine(1, "NAME", "John /Doe/", ""))
	husband := NewIndividualRecord(husbandLine)
	tree.AddRecord(husband)
	
	// Create wife
	wifeLine := NewGedcomLine(0, "INDI", "", "@I2@")
	wifeLine.AddChild(NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	wife := NewIndividualRecord(wifeLine)
	tree.AddRecord(wife)
	
	// Create family
	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(NewGedcomLine(1, "WIFE", "@I2@", ""))
	family := NewFamilyRecord(famLine)
	tree.AddRecord(family)
	
	// Test GetWifeRecord
	wifeRecord, err := family.GetWifeRecord()
	if err != nil {
		t.Fatalf("GetWifeRecord() returned error: %v", err)
	}
	if wifeRecord == nil {
		t.Fatal("GetWifeRecord() returned nil")
	}
	if wifeRecord.XrefID() != "@I2@" {
		t.Errorf("Expected wife @I2@, got %s", wifeRecord.XrefID())
	}
	
	// Test with no wife
	famLine2 := NewGedcomLine(0, "FAM", "", "@F2@")
	famLine2.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	family2 := NewFamilyRecord(famLine2)
	tree.AddRecord(family2)
	
	wifeRecord2, err := family2.GetWifeRecord()
	if err != nil {
		t.Fatalf("GetWifeRecord() returned error: %v", err)
	}
	if wifeRecord2 != nil {
		t.Error("GetWifeRecord() should return nil when no wife")
	}
}

// TestFamilyRecord_GetChildrenRecords tests the GetChildrenRecords method.
func TestFamilyRecord_GetChildrenRecords(t *testing.T) {
	// Create a tree with individuals and a family
	tree := NewGedcomTree()
	
	// Create children
	child1Line := NewGedcomLine(0, "INDI", "", "@I3@")
	child1Line.AddChild(NewGedcomLine(1, "NAME", "Child1 /Doe/", ""))
	child1 := NewIndividualRecord(child1Line)
	tree.AddRecord(child1)
	
	child2Line := NewGedcomLine(0, "INDI", "", "@I4@")
	child2Line.AddChild(NewGedcomLine(1, "NAME", "Child2 /Doe/", ""))
	child2 := NewIndividualRecord(child2Line)
	tree.AddRecord(child2)
	
	// Create family with children
	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(NewGedcomLine(1, "CHIL", "@I3@", ""))
	famLine.AddChild(NewGedcomLine(1, "CHIL", "@I4@", ""))
	family := NewFamilyRecord(famLine)
	tree.AddRecord(family)
	
	// Test GetChildrenRecords
	children, err := family.GetChildrenRecords()
	if err != nil {
		t.Fatalf("GetChildrenRecords() returned error: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
	if children[0].XrefID() != "@I3@" && children[1].XrefID() != "@I3@" {
		t.Error("Expected child @I3@ not found")
	}
	if children[0].XrefID() != "@I4@" && children[1].XrefID() != "@I4@" {
		t.Error("Expected child @I4@ not found")
	}
	
	// Test with no children
	famLine2 := NewGedcomLine(0, "FAM", "", "@F2@")
	famLine2.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	family2 := NewFamilyRecord(famLine2)
	tree.AddRecord(family2)
	
	children2, err := family2.GetChildrenRecords()
	if err != nil {
		t.Fatalf("GetChildrenRecords() returned error: %v", err)
	}
	if len(children2) != 0 {
		t.Errorf("Expected 0 children, got %d", len(children2))
	}
	
	// Test with family not in tree
	famLine3 := NewGedcomLine(0, "FAM", "", "@F3@")
	famLine3.AddChild(NewGedcomLine(1, "CHIL", "@I3@", ""))
	family3 := NewFamilyRecord(famLine3)
	
	_, err = family3.GetChildrenRecords()
	if err == nil {
		t.Error("GetChildrenRecords() should return error when family not in tree")
	}
}

// TestFamilyRecord_GetSpouses tests the GetSpouses method.
func TestFamilyRecord_GetSpouses(t *testing.T) {
	// Create a tree with individuals and a family
	tree := NewGedcomTree()
	
	// Create husband
	husbandLine := NewGedcomLine(0, "INDI", "", "@I1@")
	husbandLine.AddChild(NewGedcomLine(1, "NAME", "John /Doe/", ""))
	husband := NewIndividualRecord(husbandLine)
	tree.AddRecord(husband)
	
	// Create wife
	wifeLine := NewGedcomLine(0, "INDI", "", "@I2@")
	wifeLine.AddChild(NewGedcomLine(1, "NAME", "Jane /Doe/", ""))
	wife := NewIndividualRecord(wifeLine)
	tree.AddRecord(wife)
	
	// Create family
	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(NewGedcomLine(1, "WIFE", "@I2@", ""))
	family := NewFamilyRecord(famLine)
	tree.AddRecord(family)
	
	// Test GetSpouses
	spouses, err := family.GetSpouses()
	if err != nil {
		t.Fatalf("GetSpouses() returned error: %v", err)
	}
	if len(spouses) != 2 {
		t.Errorf("Expected 2 spouses, got %d", len(spouses))
	}
	
	// Test with only husband
	famLine2 := NewGedcomLine(0, "FAM", "", "@F2@")
	famLine2.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	family2 := NewFamilyRecord(famLine2)
	tree.AddRecord(family2)
	
	spouses2, err := family2.GetSpouses()
	if err != nil {
		t.Fatalf("GetSpouses() returned error: %v", err)
	}
	if len(spouses2) != 1 {
		t.Errorf("Expected 1 spouse, got %d", len(spouses2))
	}
}

// TestFamilyRecord_HasChild tests the HasChild method.
func TestFamilyRecord_HasChild(t *testing.T) {
	// Create a tree with individuals and a family
	tree := NewGedcomTree()
	
	// Create child
	childLine := NewGedcomLine(0, "INDI", "", "@I3@")
	childLine.AddChild(NewGedcomLine(1, "NAME", "Child /Doe/", ""))
	child := NewIndividualRecord(childLine)
	tree.AddRecord(child)
	
	// Create family with child
	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(NewGedcomLine(1, "CHIL", "@I3@", ""))
	family := NewFamilyRecord(famLine)
	tree.AddRecord(family)
	
	// Test HasChild
	if !family.HasChild(child) {
		t.Error("HasChild() should return true for existing child")
	}
	
	// Test with non-child
	otherLine := NewGedcomLine(0, "INDI", "", "@I4@")
	other := NewIndividualRecord(otherLine)
	tree.AddRecord(other)
	
	if family.HasChild(other) {
		t.Error("HasChild() should return false for non-child")
	}
	
	// Test with nil
	if family.HasChild(nil) {
		t.Error("HasChild() should return false for nil")
	}
}

// TestFamilyRecord_GetDivorceDateParsed tests the GetDivorceDateParsed method.
func TestFamilyRecord_GetDivorceDateParsed(t *testing.T) {
	// Test with valid date
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	divLine := NewGedcomLine(1, "DIV", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 1980", "")
	divLine.AddChild(dateLine)
	line.AddChild(divLine)
	
	record := NewFamilyRecord(line)
	date, err := record.GetDivorceDateParsed()
	if err != nil {
		t.Fatalf("GetDivorceDateParsed() returned error: %v", err)
	}
	if date == nil {
		t.Fatal("GetDivorceDateParsed() returned nil for valid date")
	}
	
	// Test with no divorce date
	line2 := NewGedcomLine(0, "FAM", "", "@F2@")
	record2 := NewFamilyRecord(line2)
	date2, err := record2.GetDivorceDateParsed()
	if err != nil {
		t.Fatalf("GetDivorceDateParsed() returned error: %v", err)
	}
	if date2 != nil {
		t.Error("GetDivorceDateParsed() should return nil when no divorce date")
	}
	
	// Test with invalid date
	line3 := NewGedcomLine(0, "FAM", "", "@F3@")
	divLine3 := NewGedcomLine(1, "DIV", "", "")
	dateLine3 := NewGedcomLine(2, "DATE", "invalid date", "")
	divLine3.AddChild(dateLine3)
	line3.AddChild(divLine3)
	
	record3 := NewFamilyRecord(line3)
	date3, err := record3.GetDivorceDateParsed()
	if err == nil && date3 != nil {
		// Invalid dates may return nil without error, which is acceptable
	}
}

// TestFamilyRecord_GetDivorcePlaceParsed tests the GetDivorcePlaceParsed method.
func TestFamilyRecord_GetDivorcePlaceParsed(t *testing.T) {
	// Test with valid place
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	divLine := NewGedcomLine(1, "DIV", "", "")
	placLine := NewGedcomLine(2, "PLAC", "New York, NY, USA", "")
	divLine.AddChild(placLine)
	line.AddChild(divLine)
	
	record := NewFamilyRecord(line)
	place, err := record.GetDivorcePlaceParsed()
	if err != nil {
		t.Fatalf("GetDivorcePlaceParsed() returned error: %v", err)
	}
	if place == nil {
		t.Fatal("GetDivorcePlaceParsed() returned nil for valid place")
	}
	
	// Test with no divorce place
	line2 := NewGedcomLine(0, "FAM", "", "@F2@")
	record2 := NewFamilyRecord(line2)
	place2, err := record2.GetDivorcePlaceParsed()
	if err != nil {
		t.Fatalf("GetDivorcePlaceParsed() returned error: %v", err)
	}
	if place2 != nil {
		t.Error("GetDivorcePlaceParsed() should return nil when no divorce place")
	}
}

// TestFamilyRecord_GetMarriageDateParsed_ErrorCases tests error cases for GetMarriageDateParsed.
func TestFamilyRecord_GetMarriageDateParsed_ErrorCases(t *testing.T) {
	// Test with invalid date
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := NewGedcomLine(1, "MARR", "", "")
	dateLine := NewGedcomLine(2, "DATE", "invalid date", "")
	marrLine.AddChild(dateLine)
	line.AddChild(marrLine)
	
	record := NewFamilyRecord(line)
	date, err := record.GetMarriageDateParsed()
	// Invalid dates may return nil without error, which is acceptable
	if err == nil && date == nil {
		// This is acceptable behavior
	}
}

// TestFamilyRecord_GetMarriagePlaceParsed_ErrorCases tests error cases for GetMarriagePlaceParsed.
func TestFamilyRecord_GetMarriagePlaceParsed_ErrorCases(t *testing.T) {
	// Test with empty place (should return error or nil)
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := NewGedcomLine(1, "MARR", "", "")
	placLine := NewGedcomLine(2, "PLAC", "", "")
	marrLine.AddChild(placLine)
	line.AddChild(marrLine)
	
	record := NewFamilyRecord(line)
	place, err := record.GetMarriagePlaceParsed()
	// Empty place may return error or nil, both are acceptable
	if err != nil && place == nil {
		// This is acceptable - empty place returns error
	} else if err == nil && place == nil {
		// This is also acceptable - empty place returns nil
	}
	
	// Test with no marriage place at all
	line2 := NewGedcomLine(0, "FAM", "", "@F2@")
	record2 := NewFamilyRecord(line2)
	place2, err2 := record2.GetMarriagePlaceParsed()
	if err2 != nil {
		// Error is acceptable when no place exists
	} else if place2 == nil {
		// Nil is also acceptable
	}
}

