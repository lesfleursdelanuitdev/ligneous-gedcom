package types

import (
	"reflect"
	"testing"
)

func TestGedcomLine_ToGED(t *testing.T) {
	// Simple line
	line := NewGedcomLine(0, "HEAD", "", "")
	lines := line.ToGED()
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}
	if lines[0] != "0 HEAD" {
		t.Errorf("Expected '0 HEAD', got %q", lines[0])
	}

	// Line with value
	line2 := NewGedcomLine(1, "NAME", "John /Doe/", "")
	lines2 := line2.ToGED()
	if len(lines2) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines2))
	}
	if lines2[0] != "1 NAME John /Doe/" {
		t.Errorf("Expected '1 NAME John /Doe/', got %q", lines2[0])
	}

	// Line with xref
	line3 := NewGedcomLine(0, "INDI", "", "@I1@")
	lines3 := line3.ToGED()
	if len(lines3) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines3))
	}
	if lines3[0] != "0 @I1@ INDI" {
		t.Errorf("Expected '0 @I1@ INDI', got %q", lines3[0])
	}

	// Line with children
	parent := NewGedcomLine(0, "INDI", "", "@I1@")
	child1 := NewGedcomLine(1, "NAME", "John /Doe/", "")
	child2 := NewGedcomLine(1, "SEX", "M", "")
	parent.AddChild(child1)
	parent.AddChild(child2)

	lines4 := parent.ToGED()
	if len(lines4) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines4))
	}
	expected := []string{"0 @I1@ INDI", "1 NAME John /Doe/", "1 SEX M"}
	if !reflect.DeepEqual(lines4, expected) {
		t.Errorf("Expected %v, got %v", expected, lines4)
	}

	// Nested children
	parent2 := NewGedcomLine(0, "INDI", "", "@I2@")
	name := NewGedcomLine(1, "NAME", "Jane /Doe/", "")
	givn := NewGedcomLine(2, "GIVN", "Jane", "")
	surn := NewGedcomLine(2, "SURN", "Doe", "")
	name.AddChild(givn)
	name.AddChild(surn)
	parent2.AddChild(name)

	lines5 := parent2.ToGED()
	if len(lines5) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(lines5))
	}

	// Verify order (should include all lines)
	expectedLines := map[string]bool{
		"0 @I2@ INDI": true,
		"1 NAME Jane /Doe/": true,
		"2 GIVN Jane": true,
		"2 SURN Doe": true,
	}
	for _, line := range lines5 {
		if !expectedLines[line] {
			t.Errorf("Unexpected line in output: %q", line)
		}
	}
}

func TestGedcomLine_SetValue(t *testing.T) {
	line := NewGedcomLine(0, "HEAD", "", "")

	// Set direct value
	line.SetValue("", "test")
	if line.Value != "test" {
		t.Errorf("Expected value 'test', got %q", line.Value)
	}

	// Set simple child value
	line.SetValue("CHAR", "UTF-8")
	charLines := line.GetLines("CHAR")
	if len(charLines) != 1 {
		t.Errorf("Expected 1 CHAR line, got %d", len(charLines))
	}
	if charLines[0].Value != "UTF-8" {
		t.Errorf("Expected CHAR value 'UTF-8', got %q", charLines[0].Value)
	}

	// Set nested value (creates path)
	line.SetValue("GEDC.VERS", "5.5.5")
	versValue := line.GetValue("GEDC.VERS")
	if versValue != "5.5.5" {
		t.Errorf("Expected GEDC.VERS '5.5.5', got %q", versValue)
	}

	// Update existing value
	line.SetValue("GEDC.VERS", "5.5.1")
	versValue2 := line.GetValue("GEDC.VERS")
	if versValue2 != "5.5.1" {
		t.Errorf("Expected GEDC.VERS '5.5.1', got %q", versValue2)
	}

	// Set deeply nested value
	line.SetValue("SOUR.NAME", "MyApp")
	sourName := line.GetValue("SOUR.NAME")
	if sourName != "MyApp" {
		t.Errorf("Expected SOUR.NAME 'MyApp', got %q", sourName)
	}
}

func TestGedcomLine_GetValue_EdgeCases(t *testing.T) {
	line := NewGedcomLine(1, "NAME", "John /Doe/", "")

	// Empty selector returns direct value
	if line.GetValue("") != "John /Doe/" {
		t.Errorf("Expected 'John /Doe/', got %q", line.GetValue(""))
	}

	// Non-existent selector returns empty
	if line.GetValue("NONEXISTENT") != "" {
		t.Errorf("Expected empty string, got %q", line.GetValue("NONEXISTENT"))
	}

	// Nested non-existent
	if line.GetValue("BIRT.DATE") != "" {
		t.Errorf("Expected empty string, got %q", line.GetValue("BIRT.DATE"))
	}
}

func TestGedcomLine_GetLines_EdgeCases(t *testing.T) {
	line := NewGedcomLine(0, "INDI", "", "@I1@")

	// Empty selector returns self
	lines := line.GetLines("")
	if len(lines) != 1 || lines[0] != line {
		t.Errorf("Expected self, got %v", lines)
	}

	// Non-existent selector returns empty
	lines2 := line.GetLines("NONEXISTENT")
	if len(lines2) != 0 {
		t.Errorf("Expected empty, got %v", lines2)
	}

	// Multiple children with same tag
	child1 := NewGedcomLine(1, "NOTE", "Note 1", "")
	child2 := NewGedcomLine(1, "NOTE", "Note 2", "")
	line.AddChild(child1)
	line.AddChild(child2)

	notes := line.GetLines("NOTE")
	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}
}

func TestGedcomLine_AddChild_NilChildren(t *testing.T) {
	line := &GedcomLine{
		Level:  0,
		Tag:    "HEAD",
		Value:  "",
		XrefID: "",
		Children: nil, // nil children map
	}

	child := NewGedcomLine(1, "CHAR", "UTF-8", "")
	line.AddChild(child)

	if line.Children == nil {
		t.Error("Children map should be initialized")
	}
	if len(line.Children["CHAR"]) != 1 {
		t.Errorf("Expected 1 child, got %d", len(line.Children["CHAR"]))
	}
	if child.Parent != line {
		t.Error("Child's parent should be set")
	}
}

