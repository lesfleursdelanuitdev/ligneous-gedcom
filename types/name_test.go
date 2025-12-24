package types

import (
	"testing"
)

func TestParseName_SimpleName(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	if name == nil {
		t.Fatal("Expected name to be non-nil")
	}

	if !name.IsParsed {
		t.Error("Expected name to be parsed")
	}

	if name.Given != "John" {
		t.Errorf("Expected Given 'John', got %q", name.Given)
	}

	if name.Surname != "Doe" {
		t.Errorf("Expected Surname 'Doe', got %q", name.Surname)
	}

	if name.FullName() != "John Doe" {
		t.Errorf("Expected FullName 'John Doe', got %q", name.FullName())
	}
}

func TestParseName_WithSubTags(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	givnLine := NewGedcomLine(2, "GIVN", "John", "")
	surnLine := NewGedcomLine(2, "SURN", "Doe", "")
	npfxLine := NewGedcomLine(2, "NPFX", "Dr.", "")
	nsfxLine := NewGedcomLine(2, "NSFX", "Jr.", "")
	nameLine.AddChild(givnLine)
	nameLine.AddChild(surnLine)
	nameLine.AddChild(npfxLine)
	nameLine.AddChild(nsfxLine)

	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	if name.Prefix != "Dr." {
		t.Errorf("Expected Prefix 'Dr.', got %q", name.Prefix)
	}

	if name.Given != "John" {
		t.Errorf("Expected Given 'John', got %q", name.Given)
	}

	if name.Surname != "Doe" {
		t.Errorf("Expected Surname 'Doe', got %q", name.Surname)
	}

	if name.Suffix != "Jr." {
		t.Errorf("Expected Suffix 'Jr.', got %q", name.Suffix)
	}

	expectedFull := "Dr. John Doe Jr."
	if name.FullName() != expectedFull {
		t.Errorf("Expected FullName %q, got %q", expectedFull, name.FullName())
	}
}

func TestParseName_WithSurnamePrefix(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "John /van der Berg/", "")
	spfxLine := NewGedcomLine(2, "SPFX", "van der", "")
	surnLine := NewGedcomLine(2, "SURN", "Berg", "")
	nameLine.AddChild(spfxLine)
	nameLine.AddChild(surnLine)

	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	if name.SurnamePrefix != "van der" {
		t.Errorf("Expected SurnamePrefix 'van der', got %q", name.SurnamePrefix)
	}

	if name.Surname != "Berg" {
		t.Errorf("Expected Surname 'Berg', got %q", name.Surname)
	}

	if name.GetFullSurname() != "van der Berg" {
		t.Errorf("Expected GetFullSurname 'van der Berg', got %q", name.GetFullSurname())
	}
}

func TestParseName_WithNickname(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	nickLine := NewGedcomLine(2, "NICK", "Johnny", "")
	nameLine.AddChild(nickLine)

	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	if name.Nickname != "Johnny" {
		t.Errorf("Expected Nickname 'Johnny', got %q", name.Nickname)
	}

	if !name.HasNickname() {
		t.Error("Expected HasNickname() to return true")
	}
}

func TestParseName_WithType(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	typeLine := NewGedcomLine(2, "TYPE", "married", "")
	nameLine.AddChild(typeLine)

	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	if name.Type != NameTypeMarried {
		t.Errorf("Expected Type NameTypeMarried, got %q", name.Type)
	}
}

func TestParseName_MultipleTypes(t *testing.T) {
	tests := []struct {
		typeValue string
		expected  NameType
	}{
		{"birth", NameTypeBirth},
		{"married", NameTypeMarried},
		{"aka", NameTypeAka},
		{"religious", NameTypeReligious},
		{"other", NameTypeOther},
		{"unknown", NameTypeUnknown},
		{"", NameTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.typeValue, func(t *testing.T) {
			nameLine := NewGedcomLine(1, "NAME", "Test /Name/", "")
			if tt.typeValue != "" {
				typeLine := NewGedcomLine(2, "TYPE", tt.typeValue, "")
				nameLine.AddChild(typeLine)
			}

			name, err := ParseName(nameLine)
			if err != nil {
				t.Fatalf("ParseName failed: %v", err)
			}

			if name.Type != tt.expected {
				t.Errorf("Expected Type %q, got %q", tt.expected, name.Type)
			}
		})
	}
}

func TestParseName_ComplexName(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "Dr. John van der Berg Jr.", "")
	// Parse from value since no sub-tags
	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	// Should parse prefix, given, suffix from value
	// Note: This is a heuristic parse, may not be perfect
	if name.Given == "" && name.Surname == "" {
		t.Error("Expected at least given or surname to be parsed")
	}
}

func TestParseName_UnstructuredName(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "John Doe", "")
	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	// Should parse as unstructured (last word = surname)
	if name.Surname == "" && name.Given == "" {
		t.Error("Expected at least given or surname to be parsed")
	}
}

func TestParseName_EmptyName(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "", "")
	name, err := ParseName(nameLine)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	if name != nil && name.IsValid() {
		t.Error("Expected name to be invalid")
	}
}

func TestParseName_NilLine(t *testing.T) {
	name, err := ParseName(nil)
	if err == nil {
		t.Error("Expected error for nil line")
	}

	if name != nil {
		t.Error("Expected name to be nil")
	}
}

func TestParseName_WrongTag(t *testing.T) {
	line := NewGedcomLine(1, "BIRT", "test", "")
	name, err := ParseName(line)
	if err == nil {
		t.Error("Expected error for wrong tag")
	}

	if name != nil {
		t.Error("Expected name to be nil")
	}
}

func TestGedcomName_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		surname  string
		expected bool
	}{
		{"both", "John", "Doe", true},
		{"given only", "John", "", true},
		{"surname only", "", "Doe", true},
		{"neither", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := &GedcomName{
				Given:    tt.given,
				Surname:  tt.surname,
				IsParsed: true,
			}

			if name.IsValid() != tt.expected {
				t.Errorf("Expected IsValid() %v, got %v", tt.expected, name.IsValid())
			}
		})
	}
}

func TestGedcomName_HasPrefix(t *testing.T) {
	name := &GedcomName{Prefix: "Dr."}
	if !name.HasPrefix() {
		t.Error("Expected HasPrefix() to return true")
	}

	name.Prefix = ""
	if name.HasPrefix() {
		t.Error("Expected HasPrefix() to return false")
	}
}

func TestGedcomName_HasSuffix(t *testing.T) {
	name := &GedcomName{Suffix: "Jr."}
	if !name.HasSuffix() {
		t.Error("Expected HasSuffix() to return true")
	}

	name.Suffix = ""
	if name.HasSuffix() {
		t.Error("Expected HasSuffix() to return false")
	}
}

func TestGedcomName_GetFullSurname(t *testing.T) {
	tests := []struct {
		name           string
		surnamePrefix  string
		surname        string
		expected       string
	}{
		{"both", "van der", "Berg", "van der Berg"},
		{"prefix only", "van der", "", "van der"},
		{"surname only", "", "Berg", "Berg"},
		{"neither", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := &GedcomName{
				SurnamePrefix: tt.surnamePrefix,
				Surname:       tt.surname,
			}

			result := name.GetFullSurname()
			if result != tt.expected {
				t.Errorf("Expected GetFullSurname() %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIndividualRecord_GetNamesParsed(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	name1 := NewGedcomLine(1, "NAME", "John /Doe/", "")
	name2 := NewGedcomLine(1, "NAME", "Johnny /Doe/", "")
	line.AddChild(name1)
	line.AddChild(name2)

	record := NewIndividualRecord(line)
	names, err := record.GetNamesParsed()
	if err != nil {
		t.Fatalf("GetNamesParsed failed: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	if names[0].Given != "John" {
		t.Errorf("Expected first name Given 'John', got %q", names[0].Given)
	}

	if names[1].Given != "Johnny" {
		t.Errorf("Expected second name Given 'Johnny', got %q", names[1].Given)
	}
}

func TestIndividualRecord_GetPrimaryName(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	line.AddChild(nameLine)

	record := NewIndividualRecord(line)
	name, err := record.GetPrimaryName()
	if err != nil {
		t.Fatalf("GetPrimaryName failed: %v", err)
	}

	if name == nil {
		t.Fatal("Expected name to be non-nil")
	}

	if name.Given != "John" {
		t.Errorf("Expected Given 'John', got %q", name.Given)
	}

	if name.Surname != "Doe" {
		t.Errorf("Expected Surname 'Doe', got %q", name.Surname)
	}
}

func TestIndividualRecord_GetPrimaryName_NoName(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	record := NewIndividualRecord(line)

	name, err := record.GetPrimaryName()
	if err == nil {
		t.Error("Expected error when no name found")
	}

	if name != nil {
		t.Error("Expected name to be nil")
	}
}

func TestIndividualRecord_GetNameByType(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	
	// Birth name
	birthNameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	birthTypeLine := NewGedcomLine(2, "TYPE", "birth", "")
	birthNameLine.AddChild(birthTypeLine)
	line.AddChild(birthNameLine)

	// Married name
	marriedNameLine := NewGedcomLine(1, "NAME", "John /Smith/", "")
	marriedTypeLine := NewGedcomLine(2, "TYPE", "married", "")
	marriedNameLine.AddChild(marriedTypeLine)
	line.AddChild(marriedNameLine)

	record := NewIndividualRecord(line)

	// Get birth name
	birthName, err := record.GetNameByType(NameTypeBirth)
	if err != nil {
		t.Fatalf("GetNameByType failed: %v", err)
	}

	if birthName.Surname != "Doe" {
		t.Errorf("Expected birth name Surname 'Doe', got %q", birthName.Surname)
	}

	// Get married name
	marriedName, err := record.GetNameByType(NameTypeMarried)
	if err != nil {
		t.Fatalf("GetNameByType failed: %v", err)
	}

	if marriedName.Surname != "Smith" {
		t.Errorf("Expected married name Surname 'Smith', got %q", marriedName.Surname)
	}

	// Get non-existent type
	akaName, err := record.GetNameByType(NameTypeAka)
	if err == nil {
		t.Error("Expected error when name type not found")
	}

	if akaName != nil {
		t.Error("Expected name to be nil")
	}
}

func TestIndividualRecord_GetBirthName(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	
	// Birth name
	birthNameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	birthTypeLine := NewGedcomLine(2, "TYPE", "birth", "")
	birthNameLine.AddChild(birthTypeLine)
	line.AddChild(birthNameLine)

	record := NewIndividualRecord(line)

	birthName, err := record.GetBirthName()
	if err != nil {
		t.Fatalf("GetBirthName failed: %v", err)
	}

	if birthName.Surname != "Doe" {
		t.Errorf("Expected birth name Surname 'Doe', got %q", birthName.Surname)
	}
}

func TestIndividualRecord_GetBirthName_Fallback(t *testing.T) {
	// No birth name, should fallback to primary
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	line.AddChild(nameLine)

	record := NewIndividualRecord(line)

	birthName, err := record.GetBirthName()
	if err != nil {
		t.Fatalf("GetBirthName failed: %v", err)
	}

	// Should fallback to primary name
	if birthName.Surname != "Doe" {
		t.Errorf("Expected fallback Surname 'Doe', got %q", birthName.Surname)
	}
}

func TestIndividualRecord_GetMarriedName(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	
	marriedNameLine := NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	marriedTypeLine := NewGedcomLine(2, "TYPE", "married", "")
	marriedNameLine.AddChild(marriedTypeLine)
	line.AddChild(marriedNameLine)

	record := NewIndividualRecord(line)

	marriedName, err := record.GetMarriedName()
	if err != nil {
		t.Fatalf("GetMarriedName failed: %v", err)
	}

	if marriedName.Surname != "Smith" {
		t.Errorf("Expected married name Surname 'Smith', got %q", marriedName.Surname)
	}
}

func TestIndividualRecord_GetMarriedName_NotFound(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	line.AddChild(nameLine)

	record := NewIndividualRecord(line)

	marriedName, err := record.GetMarriedName()
	if err == nil {
		t.Error("Expected error when married name not found")
	}

	if marriedName != nil {
		t.Error("Expected name to be nil")
	}
}

func TestParseName_AllComponents(t *testing.T) {
	nameLine := NewGedcomLine(1, "NAME", "Dr. John /Doe/ Jr.", "")
	npfxLine := NewGedcomLine(2, "NPFX", "Dr.", "")
	givnLine := NewGedcomLine(2, "GIVN", "John", "")
	surnLine := NewGedcomLine(2, "SURN", "Doe", "")
	nsfxLine := NewGedcomLine(2, "NSFX", "Jr.", "")
	nickLine := NewGedcomLine(2, "NICK", "Johnny", "")
	spfxLine := NewGedcomLine(2, "SPFX", "van", "")
	typeLine := NewGedcomLine(2, "TYPE", "birth", "")
	
	nameLine.AddChild(npfxLine)
	nameLine.AddChild(givnLine)
	nameLine.AddChild(surnLine)
	nameLine.AddChild(nsfxLine)
	nameLine.AddChild(nickLine)
	nameLine.AddChild(spfxLine)
	nameLine.AddChild(typeLine)

	name, err := ParseName(nameLine)
	if err != nil {
		t.Fatalf("ParseName failed: %v", err)
	}

	if name.Prefix != "Dr." {
		t.Errorf("Expected Prefix 'Dr.', got %q", name.Prefix)
	}

	if name.Given != "John" {
		t.Errorf("Expected Given 'John', got %q", name.Given)
	}

	if name.Nickname != "Johnny" {
		t.Errorf("Expected Nickname 'Johnny', got %q", name.Nickname)
	}

	if name.SurnamePrefix != "van" {
		t.Errorf("Expected SurnamePrefix 'van', got %q", name.SurnamePrefix)
	}

	if name.Surname != "Doe" {
		t.Errorf("Expected Surname 'Doe', got %q", name.Surname)
	}

	if name.Suffix != "Jr." {
		t.Errorf("Expected Suffix 'Jr.', got %q", name.Suffix)
	}

	if name.Type != NameTypeBirth {
		t.Errorf("Expected Type NameTypeBirth, got %q", name.Type)
	}

	expectedFull := "Dr. John van Doe Jr."
	if name.FullName() != expectedFull {
		t.Errorf("Expected FullName %q, got %q", expectedFull, name.FullName())
	}
}

