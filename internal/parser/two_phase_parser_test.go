package parser

import (
	"bufio"
	"os"
	"testing"
)

func TestTwoPhaseParser_Parse(t *testing.T) {
	parser := NewTwoPhaseParser()

	// Create a simple test file content
	testContent := `0 HEAD
1 GEDC
2 VERS 5.5.5
0 @I1@ INDI
1 NAME John /Doe/
1 SEX M
1 BIRT
2 DATE 1 Jan 1900
2 PLAC New York
0 @F1@ FAM
1 HUSB @I1@
1 WIFE @I2@
0 TRLR
`

	// Write to temp file
	tmpFile := "/tmp/test_two_phase.ged"
	err := writeTestFileTwoPhase(tmpFile, testContent)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	// Parse
	tree, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify structure
	individuals := tree.GetAllIndividuals()
	if len(individuals) != 1 {
		t.Errorf("Expected 1 individual, got %d", len(individuals))
	}

	families := tree.GetAllFamilies()
	if len(families) != 1 {
		t.Errorf("Expected 1 family, got %d", len(families))
	}

	// Verify individual
	indi := individuals["@I1@"]
	if indi == nil {
		t.Fatal("Individual @I1@ not found")
	}
	if indi.GetValue("NAME") != "John /Doe/" {
		t.Errorf("Expected name 'John /Doe/', got '%s'", indi.GetValue("NAME"))
	}
	if indi.GetValue("SEX") != "M" {
		t.Errorf("Expected sex 'M', got '%s'", indi.GetValue("SEX"))
	}

	// Verify birth event
	birthDate := indi.GetValue("BIRT.DATE")
	if birthDate != "1 Jan 1900" {
		t.Errorf("Expected birth date '1 Jan 1900', got '%s'", birthDate)
	}
	birthPlace := indi.GetValue("BIRT.PLAC")
	if birthPlace != "New York" {
		t.Errorf("Expected birth place 'New York', got '%s'", birthPlace)
	}

	// Verify family
	fam := families["@F1@"]
	if fam == nil {
		t.Fatal("Family @F1@ not found")
	}
	if fam.GetValue("HUSB") != "@I1@" {
		t.Errorf("Expected husband @I1@, got '%s'", fam.GetValue("HUSB"))
	}
}

func TestTwoPhaseParser_CollectRecords(t *testing.T) {
	parser := NewTwoPhaseParser()

	testContent := `0 HEAD
1 GEDC
0 @I1@ INDI
1 NAME John /Doe/
0 @I2@ INDI
1 NAME Jane /Doe/
0 TRLR
`

	tmpFile := "/tmp/test_collect.ged"
	err := writeTestFileTwoPhase(tmpFile, testContent)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Open and collect records
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	encoding, _ := DetectEncoding(tmpFile)
	reader, _ := GetReader(file, encoding)
	scanner := bufio.NewScanner(reader)

	err = parser.collectRecords(scanner)
	if err != nil {
		t.Fatalf("collectRecords failed: %v", err)
	}

	// Should have collected 4 records (HEAD, INDI, INDI, TRLR)
	// Note: TRLR is also a level 0 record
	if len(parser.records) < 3 {
		t.Errorf("Expected at least 3 records, got %d", len(parser.records))
	}

	// Check HEAD record
	if parser.records[0].Tag != "HEAD" {
		t.Errorf("Expected first record to be HEAD, got %s", parser.records[0].Tag)
	}

	// Check first INDI record
	if parser.records[1].Tag != "INDI" || parser.records[1].XrefID != "@I1@" {
		t.Errorf("Expected second record to be INDI @I1@")
	}
	if len(parser.records[1].RawLines) != 1 {
		t.Errorf("Expected 1 raw line for @I1@, got %d", len(parser.records[1].RawLines))
	}

	// Check second INDI record (skip TRLR if present)
	indi2Found := false
	for i := 2; i < len(parser.records); i++ {
		if parser.records[i].Tag == "INDI" && parser.records[i].XrefID == "@I2@" {
			indi2Found = true
			break
		}
	}
	if !indi2Found {
		t.Errorf("Expected to find INDI @I2@ record")
	}
}

func BenchmarkTwoPhaseParser_Parse(b *testing.B) {
	file := "/apps/family-tree/gedcom/gracis.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		b.Skipf("File not found: %s", file)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewTwoPhaseParser()
		_, err := parser.Parse(file)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

func BenchmarkTwoPhaseParser_vs_Sequential(b *testing.B) {
	file := "/apps/family-tree/gedcom/gracis.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		b.Skipf("File not found: %s", file)
	}

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parser := NewHierarchicalParser()
			_, err := parser.Parse(file)
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
		}
	})

	b.Run("TwoPhase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parser := NewTwoPhaseParser()
			_, err := parser.Parse(file)
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
		}
	})
}

// Helper function for two-phase parser tests
func writeTestFileTwoPhase(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

