package gedcom

import "testing"

func TestSourceRecord_AllMethods(t *testing.T) {
	sourceLine := NewGedcomLine(0, "SOUR", "", "@S1@")
	
	titleLine := NewGedcomLine(1, "TITL", "Source Title", "")
	abbrLine := NewGedcomLine(1, "ABBR", "SRC1", "")
	repoLine := NewGedcomLine(1, "REPO", "@R1@", "")
	
	sourceLine.AddChild(titleLine)
	sourceLine.AddChild(abbrLine)
	sourceLine.AddChild(repoLine)

	source := NewSourceRecord(sourceLine)

	if source.GetTitle() != "Source Title" {
		t.Errorf("Expected title 'Source Title', got %q", source.GetTitle())
	}

	if source.GetAbbreviation() != "SRC1" {
		t.Errorf("Expected abbreviation 'SRC1', got %q", source.GetAbbreviation())
	}

	if source.GetRepository() != "@R1@" {
		t.Errorf("Expected repository '@R1@', got %q", source.GetRepository())
	}
}


