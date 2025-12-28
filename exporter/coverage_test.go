package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestCSVExporter_NewCSVExporter tests NewCSVExporter
func TestCSVExporter_NewCSVExporter(t *testing.T) {
	errorManager := types.NewErrorManager()
	csvExporter := NewCSVExporter(errorManager)

	if csvExporter == nil {
		t.Fatal("NewCSVExporter returned nil")
	}

	if csvExporter.BaseExporter == nil {
		t.Error("BaseExporter should not be nil")
	}
}

// TestCSVExporter_AllTestDataFiles tests CSV export with all testdata files
func TestCSVExporter_AllTestDataFiles(t *testing.T) {
	testFiles := []string{"royal92.ged", "gracis.ged", "xavier.ged", "tree1.ged"}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			testFile := findTestDataFile(filename)
			if testFile == "" {
				t.Skipf("Test file not found: %s", filename)
			}

			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(testFile)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			errorManager := types.NewErrorManager()
			csvExporter := NewCSVExporter(errorManager)

			// Test ExportToString
			csvString, err := csvExporter.ExportToString(tree)
			if err != nil {
				t.Fatalf("ExportToString failed: %v", err)
			}

			if csvString == "" {
				t.Error("CSV string should not be empty")
			}

			lines := strings.Split(csvString, "\n")
			if len(lines) < 2 {
				t.Error("CSV should have at least header and one data row")
			}

			// Test ExportToFile
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.csv")
			err = csvExporter.ExportToFile(tree, tmpFile)
			if err != nil {
				t.Fatalf("ExportToFile failed: %v", err)
			}

			fileInfo, err := os.Stat(tmpFile)
			if err != nil {
				t.Fatalf("CSV file was not created: %v", err)
			}

			if fileInfo.Size() == 0 {
				t.Error("CSV file should not be empty")
			}

			t.Logf("%s: Exported %d bytes", filename, fileInfo.Size())
		})
	}
}

// TestXMLExporter_FamilyToXML tests familyToXML function
func TestXMLExporter_FamilyToXML(t *testing.T) {
	testFile := findTestDataFile("royal92.ged")
	if testFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(testFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	errorManager := types.NewErrorManager()
	xmlExporter := NewXMLExporter(errorManager)

	// Get a family record
	allFams := tree.GetAllFamilies()
	if len(allFams) == 0 {
		t.Skip("No families found in test file")
	}

	var testFamily types.Record
	for _, fam := range allFams {
		testFamily = fam
		break
	}

	xmlFamily := xmlExporter.familyToXML(testFamily)
	if xmlFamily == nil {
		t.Fatal("familyToXML returned nil")
	}

	if xmlFamily.ID == "" {
		t.Error("Family ID should not be empty")
	}

	t.Logf("Family XML: ID=%s, Husband=%s, Wife=%s, Children=%d",
		xmlFamily.ID, xmlFamily.Husband, xmlFamily.Wife, len(xmlFamily.Children))
}

// TestXMLExporter_SourceToXML tests sourceToXML function
func TestXMLExporter_SourceToXML(t *testing.T) {
	// Create a synthetic source record for testing
	errorManager := types.NewErrorManager()
	xmlExporter := NewXMLExporter(errorManager)

	// Create a minimal source record
	sourceLine := types.NewGedcomLine(0, "SOUR", "", "@S1@")
	sourceLine.AddChild(types.NewGedcomLine(1, "TITL", "Test Source", ""))
	sourceLine.AddChild(types.NewGedcomLine(1, "AUTH", "Test Author", ""))
	sourceLine.AddChild(types.NewGedcomLine(1, "PUBL", "Test Publication", ""))

	sourceRecord := types.NewSourceRecord(sourceLine)

	xmlSource := xmlExporter.sourceToXML(sourceRecord)
	if xmlSource == nil {
		t.Fatal("sourceToXML returned nil")
	}

	if xmlSource.ID != "@S1@" {
		t.Errorf("Expected ID @S1@, got %s", xmlSource.ID)
	}

	if xmlSource.Title != "Test Source" {
		t.Errorf("Expected Title 'Test Source', got %s", xmlSource.Title)
	}

	if xmlSource.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got %s", xmlSource.Author)
	}

	t.Logf("Source XML: ID=%s, Title=%s, Author=%s", xmlSource.ID, xmlSource.Title, xmlSource.Author)
}

// TestXMLExporter_RepositoryToXML tests repositoryToXML function
func TestXMLExporter_RepositoryToXML(t *testing.T) {
	// Create a synthetic repository record for testing
	errorManager := types.NewErrorManager()
	xmlExporter := NewXMLExporter(errorManager)

	// Create a minimal repository record with address
	repoLine := types.NewGedcomLine(0, "REPO", "", "@R1@")
	repoLine.AddChild(types.NewGedcomLine(1, "NAME", "Test Repository", ""))
	addrLine := types.NewGedcomLine(1, "ADDR", "", "")
	addrLine.AddChild(types.NewGedcomLine(2, "ADR1", "123 Test St", ""))
	addrLine.AddChild(types.NewGedcomLine(2, "CITY", "Test City", ""))
	addrLine.AddChild(types.NewGedcomLine(2, "CTRY", "Test Country", ""))
	repoLine.AddChild(addrLine)

	repoRecord := types.NewRepositoryRecord(repoLine)

	xmlRepo := xmlExporter.repositoryToXML(repoRecord)
	if xmlRepo == nil {
		t.Fatal("repositoryToXML returned nil")
	}

	if xmlRepo.ID != "@R1@" {
		t.Errorf("Expected ID @R1@, got %s", xmlRepo.ID)
	}

	if xmlRepo.Name != "Test Repository" {
		t.Errorf("Expected Name 'Test Repository', got %s", xmlRepo.Name)
	}

	if xmlRepo.Address == nil {
		t.Error("Address should not be nil")
	} else if xmlRepo.Address.City != "Test City" {
		t.Errorf("Expected City 'Test City', got %s", xmlRepo.Address.City)
	}

	t.Logf("Repository XML: ID=%s, Name=%s, City=%s", xmlRepo.ID, xmlRepo.Name, xmlRepo.Address.City)
}

// TestXMLExporter_MultimediaToXML tests multimediaToXML function
func TestXMLExporter_MultimediaToXML(t *testing.T) {
	// Create a synthetic multimedia record for testing
	errorManager := types.NewErrorManager()
	xmlExporter := NewXMLExporter(errorManager)

	// Create a minimal multimedia record
	mmLine := types.NewGedcomLine(0, "OBJE", "", "@M1@")
	mmLine.AddChild(types.NewGedcomLine(1, "FILE", "test.jpg", ""))
	fileFormLine := types.NewGedcomLine(2, "FORM", "jpeg", "")
	mmLine.GetLines("FILE")[0].AddChild(fileFormLine)
	mmLine.AddChild(types.NewGedcomLine(1, "TITL", "Test Image", ""))

	mmRecord := types.NewMultimediaRecord(mmLine)

	xmlMM := xmlExporter.multimediaToXML(mmRecord)
	if xmlMM == nil {
		t.Fatal("multimediaToXML returned nil")
	}

	if xmlMM.ID != "@M1@" {
		t.Errorf("Expected ID @M1@, got %s", xmlMM.ID)
	}

	if xmlMM.File != "test.jpg" {
		t.Errorf("Expected File 'test.jpg', got %s", xmlMM.File)
	}

	t.Logf("Multimedia XML: ID=%s, File=%s, Format=%s", xmlMM.ID, xmlMM.File, xmlMM.Format)
}

// TestXMLExporter_FormatAddressXML tests formatAddressXML function
func TestXMLExporter_FormatAddressXML(t *testing.T) {
	testFile := findTestDataFile("royal92.ged")
	if testFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(testFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	errorManager := types.NewErrorManager()
	xmlExporter := NewXMLExporter(errorManager)

	// Try to find a record with address (repository or submitter)
	allRepos := tree.GetAllRepositories()
	var testRecord types.Record
	for _, repo := range allRepos {
		addrLines := repo.GetLines("ADDR")
		if len(addrLines) > 0 {
			testRecord = repo
			break
		}
	}

	if testRecord == nil {
		// Try submitters
		allSubmitters := tree.GetAllSubmitters()
		for _, subm := range allSubmitters {
			addrLines := subm.GetLines("ADDR")
			if len(addrLines) > 0 {
				testRecord = subm
				break
			}
		}
	}

	if testRecord == nil {
		t.Skip("No records with addresses found in test file")
	}

	xmlAddr := xmlExporter.formatAddressXML(testRecord)
	if xmlAddr == nil {
		t.Log("formatAddressXML returned nil (no address found)")
	} else {
		t.Logf("Address XML: Lines=%d, City=%s, Country=%s",
			len(xmlAddr.Lines), xmlAddr.City, xmlAddr.Country)
	}
}

// TestXMLExporter_GetStringSlice tests getStringSlice helper function
func TestXMLExporter_GetStringSlice(t *testing.T) {
	// Test with valid slice
	m := map[string]interface{}{
		"children": []string{"@I1@", "@I2@", "@I3@"},
		"empty":    []string{},
		"notslice": "not a slice",
		"missing":  nil,
	}

	// Test valid slice
	children := getStringSlice(m, "children")
	if len(children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(children))
	}

	// Test empty slice
	empty := getStringSlice(m, "empty")
	if empty == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(empty) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(empty))
	}

	// Test non-slice value (function returns empty slice, not nil)
	notSlice := getStringSlice(m, "notslice")
	if notSlice == nil {
		t.Error("Expected empty slice for non-slice value, got nil")
	}
	if len(notSlice) != 0 {
		t.Errorf("Expected empty slice for non-slice value, got %d items", len(notSlice))
	}

	// Test missing key (function returns empty slice, not nil)
	missing := getStringSlice(m, "missing")
	if missing == nil {
		t.Error("Expected empty slice for missing key, got nil")
	}
	if len(missing) != 0 {
		t.Errorf("Expected empty slice for missing key, got %d items", len(missing))
	}
}

// TestJsonExporter_MultimediaItemToJSON tests multimediaItemToJSON function
func TestJsonExporter_MultimediaItemToJSON(t *testing.T) {
	// Create a synthetic multimedia record for testing
	errorManager := types.NewErrorManager()
	jsonExporter := NewJsonExporter(errorManager)

	// Create a minimal multimedia record
	mmLine := types.NewGedcomLine(0, "OBJE", "", "@M1@")
	mmLine.AddChild(types.NewGedcomLine(1, "FILE", "test.jpg", ""))
	fileFormLine := types.NewGedcomLine(2, "FORM", "jpeg", "")
	mmLine.GetLines("FILE")[0].AddChild(fileFormLine)
	mmLine.AddChild(types.NewGedcomLine(1, "TITL", "Test Image", ""))
	mmLine.AddChild(types.NewGedcomLine(1, "NOTE", "Test note", ""))

	mmRecord := types.NewMultimediaRecord(mmLine)

	jsonData := jsonExporter.multimediaItemToJSON(mmRecord)
	if jsonData == nil {
		t.Fatal("multimediaItemToJSON returned nil")
	}

	if jsonData["id"] != "@M1@" {
		t.Errorf("Expected id '@M1@', got %v", jsonData["id"])
	}

	if jsonData["file"] != "test.jpg" {
		t.Errorf("Expected file 'test.jpg', got %v", jsonData["file"])
	}

	if jsonData["format"] != "jpeg" {
		t.Errorf("Expected format 'jpeg', got %v", jsonData["format"])
	}

	t.Logf("Multimedia JSON: %+v", jsonData)
}

// TestBaseExporter_AddError tests AddError method
func TestBaseExporter_AddError(t *testing.T) {
	errorManager := types.NewErrorManager()
	baseExporter := NewBaseExporter(errorManager)

	// Test adding different severity errors
	baseExporter.AddError(types.SeverityWarning, "Test warning", 10, "TestContext")
	baseExporter.AddError(types.SeveritySevere, "Test severe error", 20, "TestContext")
	baseExporter.AddError(types.SeverityInfo, "Test info", 30, "TestContext")

	errors := errorManager.Errors()
	if len(errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(errors))
	}

	// Verify error details
	if errors[0].Severity != types.SeverityWarning {
		t.Error("First error should be warning")
	}
	if errors[1].Severity != types.SeveritySevere {
		t.Error("Second error should be severe")
	}
	if errors[2].Severity != types.SeverityInfo {
		t.Error("Third error should be info")
	}

	// Verify error manager state
	if !errorManager.HasErrors() {
		t.Error("ErrorManager should have errors")
	}
	if !errorManager.HasSevereErrors() {
		t.Error("ErrorManager should have severe errors")
	}
}

// TestXMLExporter_ExportWithAllRecordTypes tests XML export with all record types
func TestXMLExporter_ExportWithAllRecordTypes(t *testing.T) {
	testFile := findTestDataFile("royal92.ged")
	if testFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(testFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	errorManager := types.NewErrorManager()
	xmlExporter := NewXMLExporter(errorManager)

	// Export to string to trigger all conversion functions
	xmlString, err := xmlExporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("ExportToString failed: %v", err)
	}

	if xmlString == "" {
		t.Error("XML string should not be empty")
	}

	// Verify XML contains expected elements
	if !strings.Contains(xmlString, "<individuals>") {
		t.Error("XML should contain <individuals>")
	}
	if !strings.Contains(xmlString, "<families>") {
		t.Error("XML should contain <families>")
	}

	// Check if sources/repositories/multimedia are included if present
	allSources := tree.GetAllSources()
	allRepos := tree.GetAllRepositories()
	allMultimedia := tree.GetAllMultimedia()

	if len(allSources) > 0 && !strings.Contains(xmlString, "<sources>") {
		t.Log("Warning: Sources exist but <sources> not found in XML")
	}
	if len(allRepos) > 0 && !strings.Contains(xmlString, "<repositories>") {
		t.Log("Warning: Repositories exist but <repositories> not found in XML")
	}
	if len(allMultimedia) > 0 && !strings.Contains(xmlString, "<multimedia>") {
		t.Log("Warning: Multimedia exists but <multimedia> not found in XML")
	}

	t.Logf("XML export: %d bytes, %d sources, %d repos, %d multimedia",
		len(xmlString), len(allSources), len(allRepos), len(allMultimedia))
}

