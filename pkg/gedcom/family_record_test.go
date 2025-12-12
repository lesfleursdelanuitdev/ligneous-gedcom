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

