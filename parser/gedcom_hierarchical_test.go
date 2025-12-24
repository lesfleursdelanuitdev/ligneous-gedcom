package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHierarchicalParser_SimpleHierarchy(t *testing.T) {
	// Test simple 2-level hierarchy
	testContent := `0 HEAD
1 GEDC
0 @I1@ INDI
1 NAME John /Doe/
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify HEAD has GEDC child
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Expected HEAD record")
	}
	gedcLines := header.FirstLine().GetLines("GEDC")
	if len(gedcLines) != 1 {
		t.Errorf("Expected HEAD to have 1 GEDC child, got %d", len(gedcLines))
	}

	// Verify INDI has NAME child
	indi := tree.GetIndividual("@I1@")
	if indi == nil {
		t.Fatal("Expected INDI record")
	}
	nameLines := indi.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Errorf("Expected INDI to have 1 NAME child, got %d", len(nameLines))
	}
}

func TestHierarchicalParser_DeepHierarchy(t *testing.T) {
	// Test deep hierarchy (5+ levels)
	testContent := `0 @I1@ INDI
1 NAME John /Doe/
2 GIVN John
3 NICK Johnny
2 SURN Doe
1 BIRT
2 DATE 1 Jan 1900
3 TIME 12:00
4 SOUR @S1@
5 PAGE 123
2 PLAC New York
3 MAP
4 LATI N40.7128
4 LONG W74.0060
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	indi := tree.GetIndividual("@I1@")
	if indi == nil {
		t.Fatal("Expected INDI record")
	}

	// Verify deep nesting: INDI -> BIRT -> DATE -> TIME
	birtLines := indi.FirstLine().GetLines("BIRT")
	if len(birtLines) != 1 {
		t.Fatalf("Expected 1 BIRT child, got %d", len(birtLines))
	}
	dateLines := birtLines[0].GetLines("DATE")
	if len(dateLines) != 1 {
		t.Fatalf("Expected 1 DATE child, got %d", len(dateLines))
	}
	if dateLines[0].Value != "1 Jan 1900" {
		t.Errorf("Expected DATE value '1 Jan 1900', got %q", dateLines[0].Value)
	}
	timeLines := dateLines[0].GetLines("TIME")
	if len(timeLines) != 1 {
		t.Fatalf("Expected 1 TIME child, got %d", len(timeLines))
	}
	if timeLines[0].Value != "12:00" {
		t.Errorf("Expected TIME value '12:00', got %q", timeLines[0].Value)
	}

	// Verify even deeper: INDI -> BIRT -> DATE -> TIME -> SOUR -> PAGE
	// SOUR is at level 4, so it's a child of TIME (level 3), not DATE
	sourLines := timeLines[0].GetLines("SOUR")
	if len(sourLines) != 1 {
		t.Fatalf("Expected 1 SOUR child under TIME, got %d", len(sourLines))
	}
	if sourLines[0].Value != "@S1@" {
		t.Errorf("Expected SOUR value '@S1@', got %q", sourLines[0].Value)
	}
	pageLines := sourLines[0].GetLines("PAGE")
	if len(pageLines) != 1 {
		t.Fatalf("Expected 1 PAGE child, got %d", len(pageLines))
	}
	if pageLines[0].Value != "123" {
		t.Errorf("Expected PAGE value '123', got %q", pageLines[0].Value)
	}

	// Verify PLAC -> MAP -> LATI/LONG
	placLines := birtLines[0].GetLines("PLAC")
	if len(placLines) != 1 {
		t.Fatalf("Expected 1 PLAC child, got %d", len(placLines))
	}
	if placLines[0].Value != "New York" {
		t.Errorf("Expected PLAC value 'New York', got %q", placLines[0].Value)
	}
	mapLines := placLines[0].GetLines("MAP")
	if len(mapLines) != 1 {
		t.Fatalf("Expected 1 MAP child, got %d", len(mapLines))
	}
	latiLines := mapLines[0].GetLines("LATI")
	if len(latiLines) != 1 {
		t.Fatalf("Expected 1 LATI child, got %d", len(latiLines))
	}
	if latiLines[0].Value != "N40.7128" {
		t.Errorf("Expected LATI value 'N40.7128', got %q", latiLines[0].Value)
	}
}

func TestHierarchicalParser_LevelDecreases(t *testing.T) {
	// Test level decreases (sibling handling)
	testContent := `0 HEAD
1 GEDC
2 VERS 5.5.5
1 CHAR UTF-8
0 @I1@ INDI
1 NAME John /Doe/
2 GIVN John
2 SURN Doe
1 BIRT
2 DATE 1 Jan 1900
1 DEAT
2 DATE 1 Jan 2000
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify HEAD has both GEDC and CHAR as children (siblings)
	header := tree.GetHeader()
	gedcLines := header.FirstLine().GetLines("GEDC")
	if len(gedcLines) != 1 {
		t.Errorf("Expected 1 GEDC child, got %d", len(gedcLines))
	}
	charLines := header.FirstLine().GetLines("CHAR")
	if len(charLines) != 1 {
		t.Errorf("Expected 1 CHAR child, got %d", len(charLines))
	}
	if charLines[0].Value != "UTF-8" {
		t.Errorf("Expected CHAR value 'UTF-8', got %q", charLines[0].Value)
	}

	// Verify INDI has NAME, BIRT, and DEAT as children (siblings)
	indi := tree.GetIndividual("@I1@")
	nameLines := indi.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Errorf("Expected 1 NAME child, got %d", len(nameLines))
	}
	birtLines := indi.FirstLine().GetLines("BIRT")
	if len(birtLines) != 1 {
		t.Errorf("Expected 1 BIRT child, got %d", len(birtLines))
	}
	deatLines := indi.FirstLine().GetLines("DEAT")
	if len(deatLines) != 1 {
		t.Errorf("Expected 1 DEAT child, got %d", len(deatLines))
	}

	// Verify BIRT and DEAT both have DATE children
	birtDateLines := birtLines[0].GetLines("DATE")
	if len(birtDateLines) != 1 {
		t.Errorf("Expected 1 DATE under BIRT, got %d", len(birtDateLines))
	}
	deatDateLines := deatLines[0].GetLines("DATE")
	if len(deatDateLines) != 1 {
		t.Errorf("Expected 1 DATE under DEAT, got %d", len(deatDateLines))
	}
}

func TestHierarchicalParser_LevelIncreases(t *testing.T) {
	// Test level increases (nested children)
	testContent := `0 @I1@ INDI
1 NAME John /Doe/
2 GIVN John
3 NICK Johnny
2 SURN Doe
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	indi := tree.GetIndividual("@I1@")
	nameLines := indi.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Fatalf("Expected 1 NAME child, got %d", len(nameLines))
	}

	// NAME should have GIVN and SURN as children
	givnLines := nameLines[0].GetLines("GIVN")
	if len(givnLines) != 1 {
		t.Fatalf("Expected 1 GIVN child, got %d", len(givnLines))
	}
	if givnLines[0].Value != "John" {
		t.Errorf("Expected GIVN value 'John', got %q", givnLines[0].Value)
	}

	// GIVN should have NICK as child (level increase)
	nickLines := givnLines[0].GetLines("NICK")
	if len(nickLines) != 1 {
		t.Fatalf("Expected 1 NICK child under GIVN, got %d", len(nickLines))
	}
	if nickLines[0].Value != "Johnny" {
		t.Errorf("Expected NICK value 'Johnny', got %q", nickLines[0].Value)
	}
}

func TestHierarchicalParser_OrphanedLines(t *testing.T) {
	// Test orphaned lines (lines without valid parent)
	// Note: GEDCOM allows non-consecutive levels, but in practice,
	// level 2 without level 1 parent is unusual. Our parser will
	// still add it as a child of level 0, which is technically valid.
	// For strict validation, we'd need additional checks (Step 1.8).
	testContent := `0 HEAD
2 VERS 5.5.5
0 @I1@ INDI
2 GIVN John
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// HEAD will have VERS as child (level 0 < level 2, so it's valid parent)
	// This is technically valid GEDCOM, though unusual
	header := tree.GetHeader()
	versLines := header.FirstLine().GetLines("VERS")
	if len(versLines) != 1 {
		t.Errorf("Expected 1 VERS child (non-consecutive levels are allowed), got %d", len(versLines))
	}

	// INDI will have GIVN as child (level 0 < level 2, so it's valid parent)
	indi := tree.GetIndividual("@I1@")
	givnLines := indi.FirstLine().GetLines("GIVN")
	if len(givnLines) != 1 {
		t.Errorf("Expected 1 GIVN child (non-consecutive levels are allowed), got %d", len(givnLines))
	}
}

func TestHierarchicalParser_TrulyOrphanedLines(t *testing.T) {
	// Test truly orphaned lines (no parent at all - empty stack)
	// This happens when we have a level > 0 line before any level 0 record
	testContent := `1 NAME John /Doe/
0 @I1@ INDI
1 NAME Jane /Doe/
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// First NAME should be orphaned (no level 0 parent before it)
	// It should not appear in the tree
	indi := tree.GetIndividual("@I1@")
	if indi == nil {
		t.Fatal("Expected INDI record")
	}

	// INDI should only have the second NAME (the one after level 0)
	nameLines := indi.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Errorf("Expected 1 NAME child (first one was orphaned), got %d", len(nameLines))
	}
	if nameLines[0].Value != "Jane /Doe/" {
		t.Errorf("Expected NAME value 'Jane /Doe/', got %q", nameLines[0].Value)
	}
}

func TestHierarchicalParser_WithCONC_CONT(t *testing.T) {
	// Test hierarchical parsing with CONC/CONT continuation
	testContent := `0 @N1@ NOTE This is a note
1 CONC that continues
1 CONT on a new line
0 @I1@ INDI
1 NAME John /Doe/
2 NOTE This is a long note
3 CONC that continues
3 CONT on multiple lines
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify NOTE with continuation
	note := tree.GetRecordByXref("@N1@")
	if note == nil {
		t.Fatal("Expected NOTE record")
	}
	noteValue := note.GetValue("")
	expected := "This is a notethat continues\non a new line"
	if noteValue != expected {
		t.Errorf("Expected NOTE value %q, got %q", expected, noteValue)
	}

	// Verify INDI -> NAME -> NOTE with continuation
	indi := tree.GetIndividual("@I1@")
	nameLines := indi.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Fatalf("Expected 1 NAME child, got %d", len(nameLines))
	}
	noteLines := nameLines[0].GetLines("NOTE")
	if len(noteLines) != 1 {
		t.Fatalf("Expected 1 NOTE child under NAME, got %d", len(noteLines))
	}
	noteValue2 := noteLines[0].GetValue("")
	expected2 := "This is a long notethat continues\non multiple lines"
	if noteValue2 != expected2 {
		t.Errorf("Expected NOTE value %q, got %q", expected2, noteValue2)
	}
}

func TestHierarchicalParser_RealWorldExample(t *testing.T) {
	// Test with a real-world GEDCOM structure
	testContent := `0 HEAD
1 GEDC
2 VERS 5.5.5
1 CHAR UTF-8
1 SOUR MyApp
2 VERS 1.0
0 @I1@ INDI
1 NAME Robert Eugene /Williams/
2 GIVN Robert Eugene
2 SURN Williams
1 SEX M
1 BIRT
2 DATE 2 Oct 1822
2 PLAC Weston, Madison, Connecticut
1 DEAT
2 DATE 14 Apr 1905
2 PLAC New York, New York
1 FAMS @F1@
0 @F1@ FAM
1 HUSB @I1@
1 WIFE @I2@
1 CHIL @I3@
1 MARR
2 DATE Dec 1859
0 TRLR
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify HEAD structure
	header := tree.GetHeader()
	gedcLines := header.FirstLine().GetLines("GEDC")
	if len(gedcLines) != 1 {
		t.Fatalf("Expected 1 GEDC child, got %d", len(gedcLines))
	}
	versLines := gedcLines[0].GetLines("VERS")
	if len(versLines) != 1 {
		t.Fatalf("Expected 1 VERS child, got %d", len(versLines))
	}
	if versLines[0].Value != "5.5.5" {
		t.Errorf("Expected VERS value '5.5.5', got %q", versLines[0].Value)
	}

	// Verify INDI structure
	indi := tree.GetIndividual("@I1@")
	if indi == nil {
		t.Fatal("Expected INDI record")
	}

	// Verify NAME -> GIVN, SURN
	nameLines := indi.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Fatalf("Expected 1 NAME child, got %d", len(nameLines))
	}
	if nameLines[0].Value != "Robert Eugene /Williams/" {
		t.Errorf("Expected NAME value 'Robert Eugene /Williams/', got %q", nameLines[0].Value)
	}
	givnLines := nameLines[0].GetLines("GIVN")
	if len(givnLines) != 1 || givnLines[0].Value != "Robert Eugene" {
		t.Errorf("Expected GIVN value 'Robert Eugene'")
	}

	// Verify BIRT -> DATE, PLAC
	birtLines := indi.FirstLine().GetLines("BIRT")
	if len(birtLines) != 1 {
		t.Fatalf("Expected 1 BIRT child, got %d", len(birtLines))
	}
	birtDateLines := birtLines[0].GetLines("DATE")
	if len(birtDateLines) != 1 || birtDateLines[0].Value != "2 Oct 1822" {
		t.Errorf("Expected BIRT DATE value '2 Oct 1822'")
	}

	// Verify FAM structure
	fam := tree.GetFamily("@F1@")
	if fam == nil {
		t.Fatal("Expected FAM record")
	}
	husbLines := fam.FirstLine().GetLines("HUSB")
	if len(husbLines) != 1 || husbLines[0].Value != "@I1@" {
		t.Errorf("Expected HUSB value '@I1@'")
	}
	marrLines := fam.FirstLine().GetLines("MARR")
	if len(marrLines) != 1 {
		t.Fatalf("Expected 1 MARR child, got %d", len(marrLines))
	}
	marrDateLines := marrLines[0].GetLines("DATE")
	if len(marrDateLines) != 1 || marrDateLines[0].Value != "Dec 1859" {
		t.Errorf("Expected MARR DATE value 'Dec 1859'")
	}
}

