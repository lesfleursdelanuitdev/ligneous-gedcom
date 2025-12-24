package types

import (
	"testing"
)

// TestIndividualRecord_Baptism tests the Baptism method.
func TestIndividualRecord_Baptism(t *testing.T) {
	// Test with baptism event
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	bapmLine := NewGedcomLine(1, "BAPM", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	placLine := NewGedcomLine(2, "PLAC", "Church", "")
	bapmLine.AddChild(dateLine)
	bapmLine.AddChild(placLine)
	line.AddChild(bapmLine)
	
	record := NewIndividualRecord(line)
	baptism := record.Baptism()
	if baptism == nil {
		t.Fatal("Baptism() returned nil for valid baptism event")
	}
	if baptism.Type != EventTypeBaptism {
		t.Errorf("Expected EventTypeBaptism, got %v", baptism.Type)
	}
	
	// Test with no baptism
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	baptism2 := record2.Baptism()
	if baptism2 != nil {
		t.Error("Baptism() should return nil when no baptism event")
	}
}

// TestIndividualRecord_Burial tests the Burial method.
func TestIndividualRecord_Burial(t *testing.T) {
	// Test with burial event
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	buriLine := NewGedcomLine(1, "BURI", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 2000", "")
	placLine := NewGedcomLine(2, "PLAC", "Cemetery", "")
	buriLine.AddChild(dateLine)
	buriLine.AddChild(placLine)
	line.AddChild(buriLine)
	
	record := NewIndividualRecord(line)
	burial := record.Burial()
	if burial == nil {
		t.Fatal("Burial() returned nil for valid burial event")
	}
	if burial.Type != EventTypeBurial {
		t.Errorf("Expected EventTypeBurial, got %v", burial.Type)
	}
	
	// Test with no burial
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	burial2 := record2.Burial()
	if burial2 != nil {
		t.Error("Burial() should return nil when no burial event")
	}
}

// TestIndividualRecord_Baptisms tests the Baptisms method.
func TestIndividualRecord_Baptisms(t *testing.T) {
	// Test with multiple baptism events
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	bapm1Line := NewGedcomLine(1, "BAPM", "", "")
	bapm1Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 1900", ""))
	line.AddChild(bapm1Line)
	
	bapm2Line := NewGedcomLine(1, "BAPM", "", "")
	bapm2Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 1901", ""))
	line.AddChild(bapm2Line)
	
	record := NewIndividualRecord(line)
	baptisms := record.Baptisms()
	if len(baptisms) != 2 {
		t.Errorf("Expected 2 baptisms, got %d", len(baptisms))
	}
	
	// Test with no baptisms
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	baptisms2 := record2.Baptisms()
	if len(baptisms2) != 0 {
		t.Errorf("Expected 0 baptisms, got %d", len(baptisms2))
	}
}

// TestIndividualRecord_Burials tests the Burials method.
func TestIndividualRecord_Burials(t *testing.T) {
	// Test with multiple burial events
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	buri1Line := NewGedcomLine(1, "BURI", "", "")
	buri1Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 2000", ""))
	line.AddChild(buri1Line)
	
	buri2Line := NewGedcomLine(1, "BURI", "", "")
	buri2Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 2001", ""))
	line.AddChild(buri2Line)
	
	record := NewIndividualRecord(line)
	burials := record.Burials()
	if len(burials) != 2 {
		t.Errorf("Expected 2 burials, got %d", len(burials))
	}
	
	// Test with no burials
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	burials2 := record2.Burials()
	if len(burials2) != 0 {
		t.Errorf("Expected 0 burials, got %d", len(burials2))
	}
}

// TestIndividualRecord_BirthDate tests the BirthDate method.
func TestIndividualRecord_BirthDate(t *testing.T) {
	// Test with birth date
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 1900", "")
	birtLine.AddChild(dateLine)
	line.AddChild(birtLine)
	
	record := NewIndividualRecord(line)
	birthDate := record.BirthDate()
	if birthDate == nil {
		t.Fatal("BirthDate() returned nil for valid birth date")
	}
	
	// Test with no birth date
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	birthDate2 := record2.BirthDate()
	if birthDate2 != nil {
		t.Error("BirthDate() should return nil when no birth date")
	}
}

// TestIndividualRecord_DeathDate tests the DeathDate method.
func TestIndividualRecord_DeathDate(t *testing.T) {
	// Test with death date
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	deatLine := NewGedcomLine(1, "DEAT", "", "")
	dateLine := NewGedcomLine(2, "DATE", "1 Jan 2000", "")
	deatLine.AddChild(dateLine)
	line.AddChild(deatLine)
	
	record := NewIndividualRecord(line)
	deathDate := record.DeathDate()
	if deathDate == nil {
		t.Fatal("DeathDate() returned nil for valid death date")
	}
	
	// Test with no death date
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	deathDate2 := record2.DeathDate()
	if deathDate2 != nil {
		t.Error("DeathDate() should return nil when no death date")
	}
}

// TestIndividualRecord_BirthPlace tests the BirthPlace method.
func TestIndividualRecord_BirthPlace(t *testing.T) {
	// Test with birth place
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	placLine := NewGedcomLine(2, "PLAC", "New York, NY, USA", "")
	birtLine.AddChild(placLine)
	line.AddChild(birtLine)
	
	record := NewIndividualRecord(line)
	birthPlace := record.BirthPlace()
	if birthPlace == nil {
		t.Fatal("BirthPlace() returned nil for valid birth place")
	}
	
	// Test with no birth place
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	birthPlace2 := record2.BirthPlace()
	if birthPlace2 != nil {
		t.Error("BirthPlace() should return nil when no birth place")
	}
}

// TestIndividualRecord_DeathPlace tests the DeathPlace method.
func TestIndividualRecord_DeathPlace(t *testing.T) {
	// Test with death place
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	deatLine := NewGedcomLine(1, "DEAT", "", "")
	placLine := NewGedcomLine(2, "PLAC", "New York, NY, USA", "")
	deatLine.AddChild(placLine)
	line.AddChild(deatLine)
	
	record := NewIndividualRecord(line)
	deathPlace := record.DeathPlace()
	if deathPlace == nil {
		t.Fatal("DeathPlace() returned nil for valid death place")
	}
	
	// Test with no death place
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	deathPlace2 := record2.DeathPlace()
	if deathPlace2 != nil {
		t.Error("DeathPlace() should return nil when no death place")
	}
}

// TestIndividualRecord_Names tests the Names method (structured).
func TestIndividualRecord_Names(t *testing.T) {
	// Test with multiple names
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := NewGedcomLine(1, "NAME", "John /Doe/", "")
	line.AddChild(name1Line)
	
	name2Line := NewGedcomLine(1, "NAME", "Johnny /Doe/", "")
	line.AddChild(name2Line)
	
	record := NewIndividualRecord(line)
	names := record.Names()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
	
	// Test with no names
	line2 := NewGedcomLine(0, "INDI", "", "@I2@")
	record2 := NewIndividualRecord(line2)
	names2 := record2.Names()
	if names2 != nil && len(names2) != 0 {
		t.Errorf("Expected nil or empty slice, got %d names", len(names2))
	}
}

// TestIndividualRecord_Births tests the Births method.
func TestIndividualRecord_Births(t *testing.T) {
	// Test with multiple birth events
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birt1Line := NewGedcomLine(1, "BIRT", "", "")
	birt1Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 1900", ""))
	line.AddChild(birt1Line)
	
	birt2Line := NewGedcomLine(1, "BIRT", "", "")
	birt2Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 1901", ""))
	line.AddChild(birt2Line)
	
	record := NewIndividualRecord(line)
	births := record.Births()
	if len(births) != 2 {
		t.Errorf("Expected 2 births, got %d", len(births))
	}
}

// TestIndividualRecord_Deaths tests the Deaths method.
func TestIndividualRecord_Deaths(t *testing.T) {
	// Test with multiple death events
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	deat1Line := NewGedcomLine(1, "DEAT", "", "")
	deat1Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 2000", ""))
	line.AddChild(deat1Line)
	
	deat2Line := NewGedcomLine(1, "DEAT", "", "")
	deat2Line.AddChild(NewGedcomLine(2, "DATE", "1 Jan 2001", ""))
	line.AddChild(deat2Line)
	
	record := NewIndividualRecord(line)
	deaths := record.Deaths()
	if len(deaths) != 2 {
		t.Errorf("Expected 2 deaths, got %d", len(deaths))
	}
}

// TestIndividualRecord_EventsByType tests the EventsByType method.
func TestIndividualRecord_EventsByType(t *testing.T) {
	// Test with multiple event types
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	birtLine.AddChild(NewGedcomLine(2, "DATE", "1 Jan 1900", ""))
	line.AddChild(birtLine)
	
	deatLine := NewGedcomLine(1, "DEAT", "", "")
	deatLine.AddChild(NewGedcomLine(2, "DATE", "1 Jan 2000", ""))
	line.AddChild(deatLine)
	
	record := NewIndividualRecord(line)
	births := record.EventsByType(EventTypeBirth)
	if len(births) != 1 {
		t.Errorf("Expected 1 birth event, got %d", len(births))
	}
	
	deaths := record.EventsByType(EventTypeDeath)
	if len(deaths) != 1 {
		t.Errorf("Expected 1 death event, got %d", len(deaths))
	}
}

// TestIndividualRecord_CustomEventsByType tests the CustomEventsByType method.
// Note: TestIndividualRecord_CustomEvents already exists in structured_types_test.go
func TestIndividualRecord_CustomEventsByType(t *testing.T) {
	// Test with custom events of specific type
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	even1Line := NewGedcomLine(1, "EVEN", "", "")
	type1Line := NewGedcomLine(2, "TYPE", "Graduation", "")
	even1Line.AddChild(type1Line)
	line.AddChild(even1Line)
	
	even2Line := NewGedcomLine(1, "EVEN", "", "")
	type2Line := NewGedcomLine(2, "TYPE", "Graduation", "")
	even2Line.AddChild(type2Line)
	line.AddChild(even2Line)
	
	even3Line := NewGedcomLine(1, "EVEN", "", "")
	type3Line := NewGedcomLine(2, "TYPE", "Award", "")
	even3Line.AddChild(type3Line)
	line.AddChild(even3Line)
	
	record := NewIndividualRecord(line)
	graduations := record.CustomEventsByType("Graduation")
	if len(graduations) != 2 {
		t.Errorf("Expected 2 graduation events, got %d", len(graduations))
	}
	
	awards := record.CustomEventsByType("Award")
	if len(awards) != 1 {
		t.Errorf("Expected 1 award event, got %d", len(awards))
	}
}

