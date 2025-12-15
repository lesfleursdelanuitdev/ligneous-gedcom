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

func TestIndividualRecord_GetNames(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	name1 := NewGedcomLine(1, "NAME", "John /Doe/", "")
	name2 := NewGedcomLine(1, "NAME", "Johnny /Doe/", "")
	line.AddChild(name1)
	line.AddChild(name2)

	record := NewIndividualRecord(line)
	names := record.GetNames()

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
	if names[0] != "John /Doe/" || names[1] != "Johnny /Doe/" {
		t.Errorf("Expected names ['John /Doe/', 'Johnny /Doe/'], got %v", names)
	}
}

func TestIndividualRecord_GetSex(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	sexLine := NewGedcomLine(1, "SEX", "M", "")
	line.AddChild(sexLine)

	record := NewIndividualRecord(line)
	if record.GetSex() != "M" {
		t.Errorf("Expected sex 'M', got %q", record.GetSex())
	}
}

func TestIndividualRecord_GetDeathData(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	deatLine := NewGedcomLine(1, "DEAT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 2000", "")
	placLine := NewGedcomLine(2, "PLAC", "New York", "")
	sourLine := NewGedcomLine(2, "SOUR", "@S1@", "")
	deatLine.AddChild(dateLine)
	deatLine.AddChild(placLine)
	deatLine.AddChild(sourLine)
	line.AddChild(deatLine)

	record := NewIndividualRecord(line)

	if record.GetDeathDate() != "1 Jan 2000" {
		t.Errorf("Expected death date '1 Jan 2000', got %q", record.GetDeathDate())
	}

	if record.GetDeathPlace() != "New York" {
		t.Errorf("Expected death place 'New York', got %q", record.GetDeathPlace())
	}

	deathData := record.GetDeathData()
	if deathData["date"] != "1 Jan 2000" {
		t.Errorf("Expected death date '1 Jan 2000', got %q", deathData["date"])
	}
	if deathData["place"] != "New York" {
		t.Errorf("Expected death place 'New York', got %q", deathData["place"])
	}
}

func TestIndividualRecord_GetFamiliesAsChild(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	famc1 := NewGedcomLine(1, "FAMC", "@F1@", "")
	famc2 := NewGedcomLine(1, "FAMC", "@F2@", "")
	line.AddChild(famc1)
	line.AddChild(famc2)

	record := NewIndividualRecord(line)
	families := record.GetFamiliesAsChild()

	if len(families) != 2 {
		t.Errorf("Expected 2 families, got %d", len(families))
	}
	if families[0] != "@F1@" || families[1] != "@F2@" {
		t.Errorf("Expected families ['@F1@', '@F2@'], got %v", families)
	}
}

func TestIndividualRecord_GetOccupation(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	occuLine := NewGedcomLine(1, "OCCU", "Engineer", "")
	line.AddChild(occuLine)

	record := NewIndividualRecord(line)
	occupation := record.GetOccupation()

	if occupation != "Engineer" {
		t.Errorf("Expected occupation 'Engineer', got %q", occupation)
	}
}

func TestIndividualRecord_GetAttributes(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	
	// Add RESI (residence) attribute
	resiLine := NewGedcomLine(1, "RESI", "", "")
	resiDate := NewGedcomLine(2, "DATE", "1900-1950", "")
	resiPlace := NewGedcomLine(2, "PLAC", "New York", "")
	resiLine.AddChild(resiDate)
	resiLine.AddChild(resiPlace)
	line.AddChild(resiLine)

	record := NewIndividualRecord(line)
	attributes := record.GetAttributes()

	if len(attributes) < 1 {
		t.Errorf("Expected at least 1 attribute, got %d", len(attributes))
	}

	foundResi := false
	for _, attr := range attributes {
		if attr["type"] == "RESI" {
			foundResi = true
			if attr["date"] != "1900-1950" {
				t.Errorf("Expected RESI date '1900-1950', got %q", attr["date"])
			}
		}
	}
	if !foundResi {
		t.Error("Expected to find RESI attribute")
	}
}

func TestIndividualRecord_GetNotes(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	note1 := NewGedcomLine(1, "NOTE", "@N1@", "")
	note2 := NewGedcomLine(1, "NOTE", "@N2@", "")
	line.AddChild(note1)
	line.AddChild(note2)

	record := NewIndividualRecord(line)
	notes := record.GetNotes()

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}
	if notes[0] != "@N1@" || notes[1] != "@N2@" {
		t.Errorf("Expected notes ['@N1@', '@N2@'], got %v", notes)
	}
}

func TestIndividualRecord_GetSources(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	sour1 := NewGedcomLine(1, "SOUR", "@S1@", "")
	sour2 := NewGedcomLine(1, "SOUR", "@S2@", "")
	line.AddChild(sour1)
	line.AddChild(sour2)

	record := NewIndividualRecord(line)
	sources := record.GetSources()

	if len(sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(sources))
	}
	if sources[0] != "@S1@" || sources[1] != "@S2@" {
		t.Errorf("Expected sources ['@S1@', '@S2@'], got %v", sources)
	}
}

