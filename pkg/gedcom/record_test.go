package gedcom

import "testing"

func TestBaseRecord_FirstLine(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	record := NewBaseRecord(line)

	firstLine := record.FirstLine()
	if firstLine != line {
		t.Error("FirstLine should return the original line")
	}
}

func TestBaseRecord_GetValue(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := NewGedcomLine(1, "NAME", "John /Doe/", "")
	line.AddChild(nameLine)

	record := NewBaseRecord(line)

	// Get direct value
	if record.GetValue("") != "" {
		t.Error("Empty selector should return line value")
	}

	// Get child value
	name := record.GetValue("NAME")
	if name != "John /Doe/" {
		t.Errorf("Expected 'John /Doe/', got %q", name)
	}
}

func TestBaseRecord_GetValues(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	note1 := NewGedcomLine(1, "NOTE", "Note 1", "")
	note2 := NewGedcomLine(1, "NOTE", "Note 2", "")
	line.AddChild(note1)
	line.AddChild(note2)

	record := NewBaseRecord(line)

	values := record.GetValues("NOTE")
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
	if values[0] != "Note 1" || values[1] != "Note 2" {
		t.Errorf("Expected ['Note 1', 'Note 2'], got %v", values)
	}
}

func TestBaseRecord_GetLines(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	note1 := NewGedcomLine(1, "NOTE", "Note 1", "")
	note2 := NewGedcomLine(1, "NOTE", "Note 2", "")
	line.AddChild(note1)
	line.AddChild(note2)

	record := NewBaseRecord(line)

	lines := record.GetLines("NOTE")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
	if lines[0] != note1 || lines[1] != note2 {
		t.Error("GetLines returned incorrect lines")
	}
}


