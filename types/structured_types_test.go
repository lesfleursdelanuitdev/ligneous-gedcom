package types

import (
	"testing"
)

func TestEventType_IsCustom(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  bool
	}{
		{EventTypeBirth, false},
		{EventTypeDeath, false},
		{EventTypeCustom, true},
		{EventTypeMarriage, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			result := tt.eventType.IsCustom()
			if result != tt.expected {
				t.Errorf("IsCustom() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvent_IsCustom(t *testing.T) {
	// Standard event
	standardEvent := &Event{
		Type: EventTypeBirth,
	}
	if standardEvent.IsCustom() {
		t.Errorf("Standard event should not be custom")
	}

	// Custom event
	customEvent := &Event{
		Type:       EventTypeCustom,
		CustomType: "Military Service",
	}
	if !customEvent.IsCustom() {
		t.Errorf("Custom event should be identified as custom")
	}
}

func TestEvent_EffectiveType(t *testing.T) {
	// Standard event
	standardEvent := &Event{
		Type: EventTypeBirth,
	}
	if standardEvent.EffectiveType() != "BIRT" {
		t.Errorf("EffectiveType() = %s, want BIRT", standardEvent.EffectiveType())
	}

	// Custom event
	customEvent := &Event{
		Type:       EventTypeCustom,
		CustomType: "Military Service",
	}
	if customEvent.EffectiveType() != "Military Service" {
		t.Errorf("EffectiveType() = %s, want Military Service", customEvent.EffectiveType())
	}
}

func TestPlaceNode_NewPlaceNode(t *testing.T) {
	tests := []struct {
		input    string
		expected struct {
			Name    string
			County  string
			State   string
			Country string
		}
	}{
		{
			"New York, New York, New York, USA",
			struct {
				Name    string
				County  string
				State   string
				Country string
			}{"New York", "New York", "New York", "USA"},
		},
		{
			"London, England",
			struct {
				Name    string
				County  string
				State   string
				Country string
			}{"London", "England", "", ""},
		},
		{
			"Paris",
			struct {
				Name    string
				County  string
				State   string
				Country string
			}{"Paris", "", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pn := NewPlaceNode(tt.input)
			if pn.Name != tt.expected.Name {
				t.Errorf("Name = %s, want %s", pn.Name, tt.expected.Name)
			}
			if pn.County != tt.expected.County {
				t.Errorf("County = %s, want %s", pn.County, tt.expected.County)
			}
			if pn.State != tt.expected.State {
				t.Errorf("State = %s, want %s", pn.State, tt.expected.State)
			}
			if pn.Country != tt.expected.Country {
				t.Errorf("Country = %s, want %s", pn.Country, tt.expected.Country)
			}
		})
	}
}

func TestNameNode_Methods(t *testing.T) {
	name := &GedcomName{
		Given:   "John",
		Surname: "Doe",
		Prefix:  "Mr.",
		Suffix:  "Jr.",
		Type:    NameTypeBirth,
		IsParsed: true,
	}

	nameNode := NewNameNode(name)

	if nameNode.GivenName() != "John" {
		t.Errorf("GivenName() = %s, want John", nameNode.GivenName())
	}
	if nameNode.Surname() != "Doe" {
		t.Errorf("Surname() = %s, want Doe", nameNode.Surname())
	}
	if nameNode.Prefix() != "Mr." {
		t.Errorf("Prefix() = %s, want Mr.", nameNode.Prefix())
	}
	if nameNode.Suffix() != "Jr." {
		t.Errorf("Suffix() = %s, want Jr.", nameNode.Suffix())
	}
	if nameNode.Type() != NameTypeBirth {
		t.Errorf("Type() = %s, want %s", nameNode.Type(), NameTypeBirth)
	}
}

func TestDateNode_NewDateNode(t *testing.T) {
	dn := NewDateNode("15 JAN 1800")
	if dn == nil {
		t.Fatal("NewDateNode returned nil")
	}

	if !dn.IsValid() {
		t.Errorf("Date should be valid")
	}

	if dn.String() == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestParseEvent_StandardEvent(t *testing.T) {
	// Create a test BIRT event line
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	birtLine.AddChild(NewGedcomLine(2, "DATE", "15 JAN 1800", ""))
	birtLine.AddChild(NewGedcomLine(2, "PLAC", "New York, New York, USA", ""))

	event, err := ParseEvent(birtLine)
	if err != nil {
		t.Fatalf("ParseEvent failed: %v", err)
	}

	if event.Type != EventTypeBirth {
		t.Errorf("Type = %s, want %s", event.Type, EventTypeBirth)
	}

	if event.Date == nil || !event.Date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if event.Place == nil || !event.Place.IsValid() {
		t.Errorf("Place should be valid")
	}
}

func TestParseEvent_CustomEvent(t *testing.T) {
	// Create a test EVEN event line with TYPE
	evenLine := NewGedcomLine(1, "EVEN", "", "")
	evenLine.AddChild(NewGedcomLine(2, "TYPE", "Military Service", ""))
	evenLine.AddChild(NewGedcomLine(2, "DATE", "1942", ""))
	evenLine.AddChild(NewGedcomLine(2, "PLAC", "Fort Benning, Georgia, USA", ""))

	event, err := ParseEvent(evenLine)
	if err != nil {
		t.Fatalf("ParseEvent failed: %v", err)
	}

	if event.Type != EventTypeCustom {
		t.Errorf("Type = %s, want %s", event.Type, EventTypeCustom)
	}

	if event.CustomType != "Military Service" {
		t.Errorf("CustomType = %s, want Military Service", event.CustomType)
	}

	if !event.IsCustom() {
		t.Errorf("Event should be identified as custom")
	}

	if event.EffectiveType() != "Military Service" {
		t.Errorf("EffectiveType() = %s, want Military Service", event.EffectiveType())
	}
}

func TestIndividualRecord_StructuredMethods(t *testing.T) {
	// Create a test individual with name and birth event
	indiLine := NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(NewGedcomLine(1, "NAME", "John /Doe/", ""))
	birtLine := NewGedcomLine(1, "BIRT", "", "")
	birtLine.AddChild(NewGedcomLine(2, "DATE", "15 JAN 1800", ""))
	birtLine.AddChild(NewGedcomLine(2, "PLAC", "New York, New York, USA", ""))
	indiLine.AddChild(birtLine)

	indi := NewIndividualRecord(indiLine)

	// Test Name() method
	nameNode := indi.Name()
	if nameNode == nil {
		t.Fatal("Name() returned nil")
	}
	if nameNode.GivenName() == "" {
		t.Errorf("GivenName() should not be empty")
	}

	// Test Birth() method
	birth := indi.Birth()
	if birth == nil {
		t.Fatal("Birth() returned nil")
	}
	if birth.Type != EventTypeBirth {
		t.Errorf("Birth event type = %s, want %s", birth.Type, EventTypeBirth)
	}

	// Test BirthDate() method
	birthDate := indi.BirthDate()
	if birthDate == nil {
		t.Fatal("BirthDate() returned nil")
	}
	if !birthDate.IsValid() {
		t.Errorf("Birth date should be valid")
	}

	// Test BirthPlace() method
	birthPlace := indi.BirthPlace()
	if birthPlace == nil {
		t.Fatal("BirthPlace() returned nil")
	}
	if !birthPlace.IsValid() {
		t.Errorf("Birth place should be valid")
	}
}

func TestIndividualRecord_CustomEvents(t *testing.T) {
	// Create a test individual with custom event
	indiLine := NewGedcomLine(0, "INDI", "", "@I1@")
	evenLine := NewGedcomLine(1, "EVEN", "", "")
	evenLine.AddChild(NewGedcomLine(2, "TYPE", "Military Service", ""))
	evenLine.AddChild(NewGedcomLine(2, "DATE", "1942", ""))
	indiLine.AddChild(evenLine)

	indi := NewIndividualRecord(indiLine)

	// Test CustomEvents() method
	customEvents := indi.CustomEvents()
	if len(customEvents) != 1 {
		t.Fatalf("CustomEvents() returned %d events, want 1", len(customEvents))
	}

	if customEvents[0].CustomType != "Military Service" {
		t.Errorf("CustomType = %s, want Military Service", customEvents[0].CustomType)
	}

	// Test CustomEventsByType() method
	militaryEvents := indi.CustomEventsByType("Military Service")
	if len(militaryEvents) != 1 {
		t.Fatalf("CustomEventsByType() returned %d events, want 1", len(militaryEvents))
	}
}

