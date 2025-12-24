package types

import "testing"

func TestHeaderRecord_AllMethods(t *testing.T) {
	headerLine := NewGedcomLine(0, "HEAD", "", "")
	
	// Add GEDC structure
	gedcLine := NewGedcomLine(1, "GEDC", "", "")
	versLine := NewGedcomLine(2, "VERS", "5.5.5", "")
	formLine := NewGedcomLine(2, "FORM", "LINEAGE-LINKED", "")
	gedcLine.AddChild(versLine)
	gedcLine.AddChild(formLine)
	headerLine.AddChild(gedcLine)

	// Add CHAR
	charLine := NewGedcomLine(1, "CHAR", "UTF-8", "")
	headerLine.AddChild(charLine)

	// Add SOUR structure
	sourLine := NewGedcomLine(1, "SOUR", "", "")
	sourName := NewGedcomLine(2, "NAME", "MyApp", "")
	sourVers := NewGedcomLine(2, "VERS", "1.0.0", "")
	sourCorp := NewGedcomLine(2, "CORP", "MyCorp", "")
	sourLine.AddChild(sourName)
	sourLine.AddChild(sourVers)
	sourLine.AddChild(sourCorp)
	headerLine.AddChild(sourLine)

	// Add other fields
	submLine := NewGedcomLine(1, "SUBM", "@U1@", "")
	fileLine := NewGedcomLine(1, "FILE", "test.ged", "")
	langLine := NewGedcomLine(1, "LANG", "English", "")
	dateLine := NewGedcomLine(1, "DATE", "01 Jan 2024", "")
	headerLine.AddChild(submLine)
	headerLine.AddChild(fileLine)
	headerLine.AddChild(langLine)
	headerLine.AddChild(dateLine)

	header := NewHeaderRecord(headerLine)

	// Test all methods
	if header.GetGedcomVersion() != "5.5.5" {
		t.Errorf("Expected version '5.5.5', got %q", header.GetGedcomVersion())
	}

	if header.GetGedcomForm() != "LINEAGE-LINKED" {
		t.Errorf("Expected form 'LINEAGE-LINKED', got %q", header.GetGedcomForm())
	}

	if header.GetCharacterEncoding() != "UTF-8" {
		t.Errorf("Expected encoding 'UTF-8', got %q", header.GetCharacterEncoding())
	}

	if header.GetSourceName() != "MyApp" {
		t.Errorf("Expected source name 'MyApp', got %q", header.GetSourceName())
	}

	if header.GetSourceVersion() != "1.0.0" {
		t.Errorf("Expected source version '1.0.0', got %q", header.GetSourceVersion())
	}

	if header.GetSourceCorporation() != "MyCorp" {
		t.Errorf("Expected source corp 'MyCorp', got %q", header.GetSourceCorporation())
	}

	if header.GetSubmissionXref() != "@U1@" {
		t.Errorf("Expected submitter '@U1@', got %q", header.GetSubmissionXref())
	}

	if header.GetFile() != "test.ged" {
		t.Errorf("Expected file 'test.ged', got %q", header.GetFile())
	}

	if header.GetLanguage() != "English" {
		t.Errorf("Expected language 'English', got %q", header.GetLanguage())
	}

	if header.GetDate() != "01 Jan 2024" {
		t.Errorf("Expected date '01 Jan 2024', got %q", header.GetDate())
	}
}


