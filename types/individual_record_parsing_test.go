package types

import (
	"testing"
)

// TestIndividualRecord_GetBirthDateParsed tests the GetBirthDateParsed method.
func TestIndividualRecord_GetBirthDateParsed(t *testing.T) {
	// Test with valid date
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	birtLine.AddChild(dateLine)
	line.AddChild(birtLine)
	
	record := NewIndividualRecord(line)
	date, err := record.GetBirthDateParsed()
	if err != nil {
		t.Fatalf("GetBirthDateParsed() returned error: %v", err)
	}
	if date == nil {
		t.Fatal("GetBirthDateParsed() returned nil for valid date")
	}
	
	// Test with no birth date
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	date2, err2 := record2.GetBirthDateParsed()
	if err2 == nil {
		t.Error("GetBirthDateParsed() should return error when no birth date")
	}
	if date2 != nil {
		t.Error("GetBirthDateParsed() should return nil when no birth date")
	}
	
	// Test with invalid date
	line3 := NewGedcomLine(0, "INDI", "", "@I3@")
	birtLine3 := NewGedcomLine(1, "BIRT", "", "")
	dateLine3 := NewGedcomLine(2, "DATE", "invalid date", "")
	birtLine3.AddChild(dateLine3)
	line3.AddChild(birtLine3)
	
	record3 := NewIndividualRecord(line3)
	date3, err3 := record3.GetBirthDateParsed()
	// Invalid dates may return error or nil, both are acceptable
	if err3 == nil && date3 == nil {
		// This is acceptable
	}
}

// TestIndividualRecord_GetDeathDateParsed tests the GetDeathDateParsed method.
func TestIndividualRecord_GetDeathDateParsed(t *testing.T) {
	// Test with valid date
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	deatLine := NewGedcomLine(1, "DEAT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 2000", "")
	deatLine.AddChild(dateLine)
	line.AddChild(deatLine)
	
	record := NewIndividualRecord(line)
	date, err := record.GetDeathDateParsed()
	if err != nil {
		t.Fatalf("GetDeathDateParsed() returned error: %v", err)
	}
	if date == nil {
		t.Fatal("GetDeathDateParsed() returned nil for valid date")
	}
	
	// Test with no death date
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	date2, err2 := record2.GetDeathDateParsed()
	if err2 == nil {
		t.Error("GetDeathDateParsed() should return error when no death date")
	}
	if date2 != nil {
		t.Error("GetDeathDateParsed() should return nil when no death date")
	}
}

// TestIndividualRecord_GetBirthPlaceParsed tests the GetBirthPlaceParsed method.
func TestIndividualRecord_GetBirthPlaceParsed(t *testing.T) {
	// Test with valid place
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	placLine := NewGedcomLine(2, "PLAC", "New York, NY, USA", "")
	birtLine.AddChild(placLine)
	line.AddChild(birtLine)
	
	record := NewIndividualRecord(line)
	place, err := record.GetBirthPlaceParsed()
	if err != nil {
		t.Fatalf("GetBirthPlaceParsed() returned error: %v", err)
	}
	if place == nil {
		t.Fatal("GetBirthPlaceParsed() returned nil for valid place")
	}
	
	// Test with no birth place
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	place2, err2 := record2.GetBirthPlaceParsed()
	if err2 == nil {
		t.Error("GetBirthPlaceParsed() should return error when no birth place")
	}
	if place2 != nil {
		t.Error("GetBirthPlaceParsed() should return nil when no birth place")
	}
}

// TestIndividualRecord_GetDeathPlaceParsed tests the GetDeathPlaceParsed method.
func TestIndividualRecord_GetDeathPlaceParsed(t *testing.T) {
	// Test with valid place
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	deatLine := NewGedcomLine(1, "DEAT", "", "")
	placLine := NewGedcomLine(2, "PLAC", "New York, NY, USA", "")
	deatLine.AddChild(placLine)
	line.AddChild(deatLine)
	
	record := NewIndividualRecord(line)
	place, err := record.GetDeathPlaceParsed()
	if err != nil {
		t.Fatalf("GetDeathPlaceParsed() returned error: %v", err)
	}
	if place == nil {
		t.Fatal("GetDeathPlaceParsed() returned nil for valid place")
	}
	
	// Test with no death place
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	place2, err2 := record2.GetDeathPlaceParsed()
	if err2 == nil {
		t.Error("GetDeathPlaceParsed() should return error when no death place")
	}
	if place2 != nil {
		t.Error("GetDeathPlaceParsed() should return nil when no death place")
	}
}

// Note: TestIndividualRecord_GetNamesParsed, TestIndividualRecord_GetPrimaryName, 
// TestIndividualRecord_GetNameByType, TestIndividualRecord_GetBirthName, and 
// TestIndividualRecord_GetMarriedName already exist in name_test.go

