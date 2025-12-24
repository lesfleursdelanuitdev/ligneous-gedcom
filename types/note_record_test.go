package types

import "testing"

func TestNoteRecord_GetText(t *testing.T) {
	// Note with direct value
	noteLine := NewGedcomLine(0, "NOTE", "This is a note", "@N1@")
	note := NewNoteRecord(noteLine)

	text := note.GetText()
	if text != "This is a note" {
		t.Errorf("Expected 'This is a note', got %q", text)
	}

	// Note with empty value
	noteLine2 := NewGedcomLine(0, "NOTE", "", "@N2@")
	note2 := NewNoteRecord(noteLine2)

	text2 := note2.GetText()
	if text2 != "" {
		t.Errorf("Expected empty string, got %q", text2)
	}
}


