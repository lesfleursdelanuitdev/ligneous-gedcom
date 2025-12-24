package exporter

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func TestXMLExporter_ExportToString(t *testing.T) {
	errorManager := types.NewErrorManager()
	exporter := NewXMLExporter(errorManager)

	tree := types.NewGedcomTree()

	// Create header
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Create individual
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indiLine.AddChild(nameLine)
	indiRecord := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indiRecord)

	// Export
	result, err := exporter.ExportToString(tree)
	if err != nil {
		t.Fatalf("ExportToString failed: %v", err)
	}

	// Verify it's valid XML
	var xmlData XMLGedcom
	if err := xml.Unmarshal([]byte(result), &xmlData); err != nil {
		t.Fatalf("Exported string is not valid XML: %v", err)
	}

	// Verify structure
	if xmlData.Version != "5.5.5" {
		t.Errorf("Expected version 5.5.5, got %s", xmlData.Version)
	}
	if len(xmlData.Individuals) != 1 {
		t.Errorf("Expected 1 individual, got %d", len(xmlData.Individuals))
	}
}

func TestXMLExporter_ExportToFile(t *testing.T) {
	errorManager := types.NewErrorManager()
	exporter := NewXMLExporter(errorManager)

	tree := types.NewGedcomTree()

	// Create minimal tree
	headerLine := types.NewGedcomLine(0, "HEAD", "", "")
	headerRecord := types.NewHeaderRecord(headerLine)
	tree.AddRecord(headerRecord)

	// Export to file
	tmpFile := "/tmp/test_export.xml"
	defer os.Remove(tmpFile)

	err := exporter.ExportToFile(tree, tmpFile)
	if err != nil {
		t.Fatalf("ExportToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Read and verify XML
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	var xmlData XMLGedcom
	if err := xml.Unmarshal(content, &xmlData); err != nil {
		t.Fatalf("Exported file is not valid XML: %v", err)
	}
}


