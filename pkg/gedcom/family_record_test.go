package gedcom

import (
	"testing"
)

func TestFamilyRecord_GetHusbandAndWife(t *testing.T) {
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := NewGedcomLine(1, "WIFE", "@I2@", "")
	line.AddChild(husbLine)
	line.AddChild(wifeLine)

	record := NewFamilyRecord(line)
	if record.GetHusband() != "@I1@" {
		t.Errorf("Expected husband '@I1@', got %q", record.GetHusband())
	}
	if record.GetWife() != "@I2@" {
		t.Errorf("Expected wife '@I2@', got %q", record.GetWife())
	}
}

func TestFamilyRecord_GetChildren(t *testing.T) {
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	chil1 := NewGedcomLine(1, "CHIL", "@I3@", "")
	chil2 := NewGedcomLine(1, "CHIL", "@I4@", "")
	line.AddChild(chil1)
	line.AddChild(chil2)

	record := NewFamilyRecord(line)
	children := record.GetChildren()

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
	if children[0] != "@I3@" || children[1] != "@I4@" {
		t.Errorf("Expected children ['@I3@', '@I4@'], got %v", children)
	}
}

func TestFamilyRecord_GetMarriageData(t *testing.T) {
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	marrLine := NewGedcomLine(1, "MARR", "", "")
	dateLine := NewGedcomLine(2, "DATE", "Dec 1859", "")
	placLine := NewGedcomLine(2, "PLAC", "Rapid City", "")
	sourLine := NewGedcomLine(2, "SOUR", "@S1@", "")
	marrLine.AddChild(dateLine)
	marrLine.AddChild(placLine)
	marrLine.AddChild(sourLine)
	line.AddChild(marrLine)

	record := NewFamilyRecord(line)
	marriageData := record.GetMarriageData()

	if marriageData["date"] != "Dec 1859" {
		t.Errorf("Expected marriage date 'Dec 1859', got %q", marriageData["date"])
	}
	if marriageData["place"] != "Rapid City" {
		t.Errorf("Expected marriage place 'Rapid City', got %q", marriageData["place"])
	}
}

func TestFamilyRecord_GetDivorceData(t *testing.T) {
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	divLine := NewGedcomLine(1, "DIV", "", "")
	dateLine := NewGedcomLine(2, "DATE", "Jan 1980", "")
	placLine := NewGedcomLine(2, "PLAC", "New York", "")
	sourLine := NewGedcomLine(2, "SOUR", "@S2@", "")
	divLine.AddChild(dateLine)
	divLine.AddChild(placLine)
	divLine.AddChild(sourLine)
	line.AddChild(divLine)

	record := NewFamilyRecord(line)

	if record.GetDivorceDate() != "Jan 1980" {
		t.Errorf("Expected divorce date 'Jan 1980', got %q", record.GetDivorceDate())
	}

	if record.GetDivorcePlace() != "New York" {
		t.Errorf("Expected divorce place 'New York', got %q", record.GetDivorcePlace())
	}

	divorceData := record.GetDivorceData()
	if divorceData["date"] != "Jan 1980" {
		t.Errorf("Expected divorce date 'Jan 1980', got %q", divorceData["date"])
	}
	if divorceData["place"] != "New York" {
		t.Errorf("Expected divorce place 'New York', got %q", divorceData["place"])
	}
}

func TestFamilyRecord_GetEvents(t *testing.T) {
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	
	// Add marriage event
	marrLine := NewGedcomLine(1, "MARR", "", "")
	marrDate := NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	marrPlace := NewGedcomLine(2, "PLAC", "New York", "")
	marrLine.AddChild(marrDate)
	marrLine.AddChild(marrPlace)
	line.AddChild(marrLine)

	// Add divorce event
	divLine := NewGedcomLine(1, "DIV", "", "")
	divDate := NewGedcomLine(2, "DATE", "1 Jan 1950", "")
	divLine.AddChild(divDate)
	line.AddChild(divLine)

	record := NewFamilyRecord(line)
	events := record.GetEvents()

	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}

	// Check for MARR event
	foundMarr := false
	for _, event := range events {
		if event["type"] == "MARR" {
			foundMarr = true
			if event["date"] != "1 Jan 1900" {
				t.Errorf("Expected MARR date '1 Jan 1900', got %q", event["date"])
			}
		}
	}
	if !foundMarr {
		t.Error("Expected to find MARR event")
	}
}

func TestFamilyRecord_GetNotes(t *testing.T) {
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	note1 := NewGedcomLine(1, "NOTE", "@N1@", "")
	note2 := NewGedcomLine(1, "NOTE", "@N2@", "")
	line.AddChild(note1)
	line.AddChild(note2)

	record := NewFamilyRecord(line)
	notes := record.GetNotes()

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}
	if notes[0] != "@N1@" || notes[1] != "@N2@" {
		t.Errorf("Expected notes ['@N1@', '@N2@'], got %v", notes)
	}
}

func TestFamilyRecord_GetSources(t *testing.T) {
	line := NewGedcomLine(0, "FAM", "", "@F1@")
	sour1 := NewGedcomLine(1, "SOUR", "@S1@", "")
	sour2 := NewGedcomLine(1, "SOUR", "@S2@", "")
	line.AddChild(sour1)
	line.AddChild(sour2)

	record := NewFamilyRecord(line)
	sources := record.GetSources()

	if len(sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(sources))
	}
	if sources[0] != "@S1@" || sources[1] != "@S2@" {
		t.Errorf("Expected sources ['@S1@', '@S2@'], got %v", sources)
	}
}


