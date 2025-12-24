package types

import (
	"testing"
)

func TestIndividualRecord_DatePlaceParsing(t *testing.T) {
	// Create an individual record with birth date and place
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "15 JAN 1800", "")
	placLine := NewGedcomLine(2, "PLAC", "Rapid City, South Dakota", "")

	birtLine.AddChild(dateLine)
	birtLine.AddChild(placLine)

	indiLine := NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(birtLine)

	record := NewIndividualRecord(indiLine)

	// Test parsed date
	date, err := record.GetBirthDateParsed()
	if err != nil {
		t.Fatalf("GetBirthDateParsed() failed: %v", err)
	}

	if !date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if date.Year != 1800 {
		t.Errorf("Year = %d, want 1800", date.Year)
	}

	if date.Month != 1 {
		t.Errorf("Month = %d, want 1", date.Month)
	}

	if date.Day != 15 {
		t.Errorf("Day = %d, want 15", date.Day)
	}

	iso := date.ToISO8601()
	if iso != "1800-01-15" {
		t.Errorf("ToISO8601() = %q, want %q", iso, "1800-01-15")
	}

	// Test parsed place
	place, err := record.GetBirthPlaceParsed()
	if err != nil {
		t.Fatalf("GetBirthPlaceParsed() failed: %v", err)
	}

	if !place.IsValid() {
		t.Errorf("Place should be valid")
	}

	if place.City != "Rapid City" {
		t.Errorf("City = %q, want %q", place.City, "Rapid City")
	}

	if place.State != "South Dakota" {
		t.Errorf("State = %q, want %q", place.State, "South Dakota")
	}
}

func TestFamilyRecord_DatePlaceParsing(t *testing.T) {
	// Create a family record with marriage date and place
	marrLine := NewGedcomLine(1, "MARR", "", "")
	dateLine := NewGedcomLine(2, "DATE", "ABT 1850", "")
	placLine := NewGedcomLine(2, "PLAC", "New York, NY, USA", "")

	marrLine.AddChild(dateLine)
	marrLine.AddChild(placLine)

	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(marrLine)

	record := NewFamilyRecord(famLine)

	// Test parsed date
	date, err := record.GetMarriageDateParsed()
	if err != nil {
		t.Fatalf("GetMarriageDateParsed() failed: %v", err)
	}

	if !date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if date.Type != DateTypeAbout {
		t.Errorf("Type = %s, want %s", date.Type, DateTypeAbout)
	}

	if date.Year != 1850 {
		t.Errorf("Year = %d, want 1850", date.Year)
	}

	// Test parsed place
	place, err := record.GetMarriagePlaceParsed()
	if err != nil {
		t.Fatalf("GetMarriagePlaceParsed() failed: %v", err)
	}

	if !place.IsValid() {
		t.Errorf("Place should be valid")
	}

	if place.City != "New York" {
		t.Errorf("City = %q, want %q", place.City, "New York")
	}

	if place.State != "NY" {
		t.Errorf("State = %q, want %q", place.State, "NY")
	}

	if place.Country != "USA" {
		t.Errorf("Country = %q, want %q", place.Country, "USA")
	}
}
