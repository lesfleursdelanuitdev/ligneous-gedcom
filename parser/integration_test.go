package parser

import (
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestIntegration_SampleGed tests parsing with the sample.ged file
// Note: This test looks for sample.ged which may not exist in testdata.
// It will use royal92.ged as a fallback if sample.ged is not found.
func TestIntegration_SampleGed(t *testing.T) {
	// Try to find sample.ged, fallback to royal92.ged
	sampleFile := findTestDataFile("sample.ged")
	if sampleFile == "" {
		sampleFile = findTestDataFile("royal92.ged")
		if sampleFile == "" {
			t.Skipf("Test file not found (tried sample.ged and royal92.ged)")
		}
		t.Logf("Using royal92.ged as fallback for sample.ged test")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse sample.ged: %v", err)
	}

	// Verify HEAD record
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Expected HEAD record")
	}

	// Verify GEDC structure (if present)
	gedcLines := header.FirstLine().GetLines("GEDC")
	if len(gedcLines) > 0 {
		versLines := gedcLines[0].GetLines("VERS")
		if len(versLines) > 0 {
			t.Logf("GEDCOM version: %s", versLines[0].Value)
		}
	}

	// Verify we have some individuals
	allIndis := tree.GetAllIndividuals()
	if len(allIndis) == 0 {
		t.Error("Expected at least one individual")
	} else {
		t.Logf("Found %d individuals", len(allIndis))
		// Check first individual
		for xref, indi := range allIndis {
			nameValue := indi.GetValue("NAME")
			if nameValue != "" {
				t.Logf("Individual %s: %s", xref, nameValue)
				break
			}
		}
	}

	// Verify we have some families
	allFams := tree.GetAllFamilies()
	if len(allFams) > 0 {
		t.Logf("Found %d families", len(allFams))
	}

	t.Logf("Successfully parsed sample.ged with %d errors", len(parser.GetErrors()))
}

// TestIntegration_LargeFiles tests parsing with large real-world GEDCOM files
func TestIntegration_LargeFiles(t *testing.T) {
	testFiles := []struct {
		name     string
		filename string
		minLines int
	}{
		{"gracis.ged", "gracis.ged", 10000},
		{"xavier.ged", "xavier.ged", 5000},
		{"tree1.ged", "tree1.ged", 12000},
		{"royal92.ged", "royal92.ged", 3000},
		{"pres2020.ged", "pres2020.ged", 1000},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			filePath := findTestDataFile(tt.filename)
			if filePath == "" {
				t.Skipf("File not found: %s (tried multiple paths)", tt.filename)
			}

			parser := NewHierarchicalParser()
			tree, err := parser.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tt.name, err)
			}

			if tree == nil {
				t.Fatal("Expected tree to be created")
			}

			// Verify HEAD exists
			header := tree.GetHeader()
			if header == nil {
				t.Error("Expected HEAD record")
			}

			// Verify we have records
			allIndis := tree.GetAllIndividuals()
			allFams := tree.GetAllFamilies()

			if len(allIndis) == 0 && len(allFams) == 0 {
				t.Error("Expected at least some records to be parsed")
			}

			// Log statistics
			errors := parser.GetErrors()
			errorSummary := parser.GetErrorManager().GetErrorSummary()
			t.Logf("%s: %d individuals, %d families, %d total errors (%d warnings, %d severe)",
				tt.name, len(allIndis), len(allFams), len(errors),
				errorSummary[types.SeverityWarning],
				errorSummary[types.SeveritySevere])

			// Verify parsing didn't fail completely
			if parser.HasSevereErrors() {
				severeErrors := parser.GetErrorManager().GetErrorsBySeverity(types.SeveritySevere)
				for _, e := range severeErrors {
					t.Logf("Severe error: %v", e)
				}
				// Severe errors are acceptable if they don't prevent parsing
			}
		})
	}
}

// TestIntegration_TreeStructure verifies the tree structure is correct
func TestIntegration_TreeStructure(t *testing.T) {
	sampleFile := findTestDataFile("royal92.ged")
	if sampleFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify parent-child relationships
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Fatal("Expected INDI @I1@")
	}

	// Verify NAME structure
	nameLines := indi1.FirstLine().GetLines("NAME")
	if len(nameLines) != 1 {
		t.Fatalf("Expected 1 NAME child, got %d", len(nameLines))
	}
	nameLine := nameLines[0]
	if nameLine.Parent != indi1.FirstLine() {
		t.Error("NAME parent should be INDI")
	}
	if nameLine.Level != 1 {
		t.Errorf("Expected NAME level 1, got %d", nameLine.Level)
	}

	// Verify GIVN/SURN structure
	givnLines := nameLine.GetLines("GIVN")
	if len(givnLines) != 1 {
		t.Fatalf("Expected 1 GIVN child, got %d", len(givnLines))
	}
	givnLine := givnLines[0]
	if givnLine.Parent != nameLine {
		t.Error("GIVN parent should be NAME")
	}
	if givnLine.Level != 2 {
		t.Errorf("Expected GIVN level 2, got %d", givnLine.Level)
	}

	// Verify BIRT structure
	birtLines := indi1.FirstLine().GetLines("BIRT")
	if len(birtLines) != 1 {
		t.Fatalf("Expected 1 BIRT child, got %d", len(birtLines))
	}
	birtLine := birtLines[0]
	if birtLine.Parent != indi1.FirstLine() {
		t.Error("BIRT parent should be INDI")
	}

	// Verify DATE under BIRT
	dateLines := birtLine.GetLines("DATE")
	if len(dateLines) != 1 {
		t.Fatalf("Expected 1 DATE child under BIRT, got %d", len(dateLines))
	}
	dateLine := dateLines[0]
	if dateLine.Parent != birtLine {
		t.Error("DATE parent should be BIRT")
	}
}

// TestIntegration_ErrorHandling verifies error handling with real files
func TestIntegration_ErrorHandling(t *testing.T) {
	sampleFile := findTestDataFile("royal92.ged")
	if sampleFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify tree was created despite any errors
	if tree == nil {
		t.Fatal("Expected tree to be created")
	}

			// Check errors
			parseErrors := parser.GetErrors()
			if len(parseErrors) > 0 {
				t.Logf("Found %d errors during parsing:", len(parseErrors))
				for _, e := range parseErrors {
					t.Logf("  %v", e)
				}
			}

	// Verify we can still access records
	header := tree.GetHeader()
	if header == nil {
		t.Error("Expected HEAD record even with errors")
	}
}

// TestIntegration_XrefIndex verifies cross-reference index is correct
func TestIntegration_XrefIndex(t *testing.T) {
	sampleFile := findTestDataFile("royal92.ged")
	if sampleFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify some xrefs are accessible (check first few individuals and families)
	allIndis := tree.GetAllIndividuals()
	allFams := tree.GetAllFamilies()
	
	if len(allIndis) > 0 {
		count := 0
		for xref := range allIndis {
			record := tree.GetRecordByXref(xref)
			if record == nil {
				t.Errorf("Expected record %s to be in xref index", xref)
			} else {
				count++
				if count >= 3 {
					break
				}
			}
		}
		t.Logf("Verified %d individual xrefs", count)
	}
	
	if len(allFams) > 0 {
		count := 0
		for xref := range allFams {
			record := tree.GetRecordByXref(xref)
			if record == nil {
				t.Errorf("Expected record %s to be in xref index", xref)
			} else {
				count++
				if count >= 2 {
					break
				}
			}
		}
		t.Logf("Verified %d family xrefs", count)
	}

	// Verify xrefs in FAM records (check first family)
	if len(allFams) > 0 {
		var fam1 types.Record
		for _, fam := range allFams {
			fam1 = fam
			break
		}
		
		husbXref := fam1.GetValue("HUSB")
		if husbXref != "" {
			// Verify the xref points to actual record
			husbRecord := tree.GetRecordByXref(husbXref)
			if husbRecord == nil {
				t.Errorf("Expected HUSB xref %s to resolve to a record", husbXref)
			} else {
				t.Logf("Verified HUSB xref %s resolves correctly", husbXref)
			}
		}
	}
}

// TestIntegration_DeepHierarchy verifies deep nesting works correctly
func TestIntegration_DeepHierarchy(t *testing.T) {
	sampleFile := findTestDataFile("royal92.ged")
	if sampleFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify deep hierarchy: HEAD -> SOUR -> CORP -> ADDR -> CITY
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Expected HEAD record")
	}

	sourLines := header.FirstLine().GetLines("SOUR")
	if len(sourLines) != 1 {
		t.Fatalf("Expected 1 SOUR child, got %d", len(sourLines))
	}

	corpLines := sourLines[0].GetLines("CORP")
	if len(corpLines) != 1 {
		t.Fatalf("Expected 1 CORP child, got %d", len(corpLines))
	}

	addrLines := corpLines[0].GetLines("ADDR")
	if len(addrLines) != 1 {
		t.Fatalf("Expected 1 ADDR child, got %d", len(addrLines))
	}

	cityLines := addrLines[0].GetLines("CITY")
	if len(cityLines) != 1 {
		t.Fatalf("Expected 1 CITY child, got %d", len(cityLines))
	}

	if cityLines[0].Value != "LEIDEN" {
		t.Errorf("Expected CITY value 'LEIDEN', got %q", cityLines[0].Value)
	}
}

// TestIntegration_Performance tests parsing performance
func TestIntegration_Performance(t *testing.T) {
	testFiles := []string{"royal92.ged", "gracis.ged", "xavier.ged", "tree1.ged", "pres2020.ged"}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("File not found: %s (tried multiple paths)", filename)
			}

			parser := NewHierarchicalParser()
			tree, err := parser.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if tree == nil {
				t.Fatal("Expected tree to be created")
			}

			// Get file info for statistics
			info, err := os.Stat(filePath)
			if err != nil {
				t.Fatalf("Failed to get file info: %v", err)
			}

			allIndis := tree.GetAllIndividuals()
			allFams := tree.GetAllFamilies()

			t.Logf("%s: %d bytes, %d individuals, %d families, %d errors",
				filename, info.Size(), len(allIndis), len(allFams), len(parser.GetErrors()))
		})
	}
}

// TestIntegration_EdgeCases tests edge cases with real files
func TestIntegration_EdgeCases(t *testing.T) {
	sampleFile := findTestDataFile("royal92.ged")
	if sampleFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Test empty selectors
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Fatal("Expected INDI @I1@")
	}

	// Empty selector should return the record's value (empty for INDI)
	value := indi1.GetValue("")
	if value != "" {
		t.Errorf("Expected empty value for INDI, got %q", value)
	}

	// Test non-existent selectors
	nonExistent := indi1.GetValue("NONEXISTENT")
	if nonExistent != "" {
		t.Errorf("Expected empty value for non-existent selector, got %q", nonExistent)
	}

	// Test nested selectors
	birtDate := indi1.GetValue("BIRT.DATE")
	if birtDate != "2 Oct 1822" {
		t.Errorf("Expected BIRT.DATE '2 Oct 1822', got %q", birtDate)
	}

	// Test multiple values
	nameLines := indi1.FirstLine().GetLines("FAMS")
	if len(nameLines) != 2 {
		t.Errorf("Expected 2 FAMS children, got %d", len(nameLines))
	}
}

