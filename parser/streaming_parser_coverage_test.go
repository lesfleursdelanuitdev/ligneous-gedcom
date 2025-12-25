package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestStreamingParser_HasSevereErrors tests HasSevereErrors method
func TestStreamingParser_HasSevereErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	
	// Create a file with a severe error (malformed line)
	content := "0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 INDI @I1@\nINVALID LINE WITHOUT LEVEL\n0 TRLR\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewStreamingHierarchicalParser()
	err := parser.ParseWithHandler(testFile, func(record types.Record) error {
		return nil
	})
	if err != nil {
		t.Logf("Parse error (expected): %v", err)
	}

	// Test HasSevereErrors
	hasSevere := parser.HasSevereErrors()
	t.Logf("HasSevereErrors: %v", hasSevere)
	
	// Should have errors
	if !parser.HasErrors() {
		t.Error("Expected parser to have errors")
	}
}

// TestStreamingParser_GetErrorManager tests GetErrorManager method
func TestStreamingParser_GetErrorManager(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	
	content := "0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 INDI @I1@\n1 NAME John /Doe/\n0 TRLR\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewStreamingHierarchicalParser()
	err := parser.ParseWithHandler(testFile, func(record types.Record) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test GetErrorManager
	errorManager := parser.GetErrorManager()
	if errorManager == nil {
		t.Fatal("Expected error manager to be non-nil")
	}

	// Verify we can use the error manager
	errors := errorManager.Errors()
	t.Logf("Found %d errors", len(errors))
}

// TestRecordIterator_GetErrors tests RecordIterator.GetErrors method
func TestRecordIterator_GetErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.ged")
	
	// Create a file with some errors
	content := "0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 INDI @I1@\n1 NAME John /Doe/\n0 TRLR\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	iterator, err := NewRecordIterator(testFile)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iterator.Close()

	// Read all records
	recordCount := 0
	for iterator.Next() {
		record := iterator.Record()
		if record != nil {
			recordCount++
		}
	}

	if iterator.Error() != nil {
		t.Logf("Iterator error: %v", iterator.Error())
	}

	// Test GetErrors
	errors := iterator.GetErrors()
	t.Logf("Found %d errors from iterator", len(errors))
	
	if recordCount == 0 {
		t.Error("Expected at least one record")
	}
}

// TestStreamingParser_AllTestDataFiles tests streaming parser against all testdata files
func TestStreamingParser_AllTestDataFiles(t *testing.T) {
	testFiles := []string{"royal92.ged", "gracis.ged", "xavier.ged", "tree1.ged", "pres2020.ged"}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("File not found: %s", filename)
			}

			parser := NewStreamingHierarchicalParser()
			recordCount := 0
			
			err := parser.ParseWithHandler(filePath, func(record types.Record) error {
				recordCount++
				return nil
			})

			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if recordCount == 0 {
				t.Error("Expected at least one record")
			}

			t.Logf("%s: %d records, %d errors", filename, recordCount, len(parser.GetErrors()))
		})
	}
}

// TestRecordIterator_AllTestDataFiles tests RecordIterator against all testdata files
func TestRecordIterator_AllTestDataFiles(t *testing.T) {
	testFiles := []string{"royal92.ged", "gracis.ged", "xavier.ged", "tree1.ged", "pres2020.ged"}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("File not found: %s", filename)
			}

			iterator, err := NewRecordIterator(filePath)
			if err != nil {
				t.Fatalf("Failed to create iterator: %v", err)
			}
			defer iterator.Close()

			recordCount := 0
			for iterator.Next() {
				record := iterator.Record()
				if record != nil {
					recordCount++
				}
			}

			if iterator.Error() != nil {
				t.Logf("Iterator error: %v", iterator.Error())
			}

			if recordCount == 0 {
				t.Error("Expected at least one record")
			}

			errors := iterator.GetErrors()
			t.Logf("%s: %d records, %d errors", filename, recordCount, len(errors))
		})
	}
}

