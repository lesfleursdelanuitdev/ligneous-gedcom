package gedcom

import (
	"testing"
)

func TestIndividualRecord_GetName(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	line.AddChild(nameLine)

	record := NewIndividualRecord(line)
	if record.GetName() != "John /Doe/" {
		t.Errorf("Expected name 'John /Doe/', got %q", record.GetName())
	}
}

func TestIndividualRecord_GetGivenNameAndSurname(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	givnLine := NewGedcomLine(2, "GIVN", "John", "")
	surnLine := NewGedcomLine(2, "SURN", "Doe", "")
	nameLine.AddChild(givnLine)
	nameLine.AddChild(surnLine)
	line.AddChild(nameLine)

	record := NewIndividualRecord(line)
	if record.GetGivenName() != "John" {
		t.Errorf("Expected given name 'John', got %q", record.GetGivenName())
	}
	if record.GetSurname() != "Doe" {
		t.Errorf("Expected surname 'Doe', got %q", record.GetSurname())
	}
}

func TestIndividualRecord_GetBirthData(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	placLine := NewGedcomLine(2, "PLAC", "New York", "")
	sourLine := NewGedcomLine(2, "SOUR", "@S1@", "")
	birtLine.AddChild(dateLine)
	birtLine.AddChild(placLine)
	birtLine.AddChild(sourLine)
	line.AddChild(birtLine)

	record := NewIndividualRecord(line)
	birthData := record.GetBirthData()

	if birthData["date"] != "1 Jan 1900" {
		t.Errorf("Expected birth date '1 Jan 1900', got %q", birthData["date"])
	}
	if birthData["place"] != "New York" {
		t.Errorf("Expected birth place 'New York', got %q", birthData["place"])
	}
	sources, ok := birthData["sources"].([]string)
	if !ok || len(sources) != 1 || sources[0] != "@S1@" {
		t.Errorf("Expected sources ['@S1@'], got %v", sources)
	}
}

func TestIndividualRecord_GetFamiliesAsSpouse(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	fams1 := NewGedcomLine(1, "FAMS", "@F1@", "")
	fams2 := NewGedcomLine(1, "FAMS", "@F2@", "")
	line.AddChild(fams1)
	line.AddChild(fams2)

	record := NewIndividualRecord(line)
	families := record.GetFamiliesAsSpouse()

	if len(families) != 2 {
		t.Errorf("Expected 2 families, got %d", len(families))
	}
	if families[0] != "@F1@" || families[1] != "@F2@" {
		t.Errorf("Expected families ['@F1@', '@F2@'], got %v", families)
	}
}

func TestIndividualRecord_GetEvents(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	placLine := NewGedcomLine(2, "PLAC", "New York", "")
	birtLine.AddChild(dateLine)
	birtLine.AddChild(placLine)
	line.AddChild(birtLine)

	deatLine := NewGedcomLine(1, "DEAT", "", "")
	deatDateLine := NewGedcomLine(2, "DATE", "1 Jan 2000", "")
	deatLine.AddChild(deatDateLine)
	line.AddChild(deatLine)

	record := NewIndividualRecord(line)
	events := record.GetEvents()

	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}

	// Find BIRT event
	foundBirt := false
	for _, event := range events {
		if event["type"] == "BIRT" {
			foundBirt = true
			if event["date"] != "1 Jan 1900" {
				t.Errorf("Expected BIRT date '1 Jan 1900', got %q", event["date"])
			}
			if event["place"] != "New York" {
				t.Errorf("Expected BIRT place 'New York', got %q", event["place"])
			}
		}
	}
	if !foundBirt {
		t.Error("Expected to find BIRT event")
	}
}

