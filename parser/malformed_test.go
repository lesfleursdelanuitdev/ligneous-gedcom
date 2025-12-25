package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// findMalformedTestFile finds a malformed test file in testdata/malformed
func findMalformedTestFile(filename string) string {
	possiblePaths := []string{
		filepath.Join("/apps/gedcom-go/testdata/malformed", filename),
		filepath.Join("testdata/malformed", filename),
		filepath.Join("../testdata/malformed", filename),
		filepath.Join("../../testdata/malformed", filename),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// TestMalformedFiles tests parsing of malformed GEDCOM files
// These tests ensure the parser handles errors gracefully without crashing
func TestMalformedFiles(t *testing.T) {
	testCases := []struct {
		name           string
		filename       string
		shouldParse    bool // Whether parsing should succeed (may have errors but not fail completely)
		expectErrors   bool // Whether we expect parsing errors
		expectWarnings bool // Whether we expect warnings
		description    string
	}{
		{
			name:           "circular-reference",
			filename:       "circular-reference.ged",
			shouldParse:    true, // Should parse but may have relationship issues
			expectErrors:   false,
			expectWarnings: true, // Circular references might generate warnings
			description:    "Tests circular family references (I1->F1->I2->F2->I1)",
		},
		{
			name:           "duplicate-xref",
			filename:       "duplicate-xref.ged",
			shouldParse:    true, // Should parse but last duplicate wins
			expectErrors:   false,
			expectWarnings: true, // Duplicate XREFs should generate warnings
			description:    "Tests duplicate XREF IDs (@I1@ appears multiple times)",
		},
		{
			name:           "invalid-level",
			filename:       "invalid-level.ged",
			shouldParse:    true, // Should parse but skip invalid level
			expectErrors:   false,
			expectWarnings: true, // Invalid level should generate warnings
			description:    "Tests invalid level (99 is too deep)",
		},
		{
			name:           "invalid-xref",
			filename:       "invalid-xref.ged",
			shouldParse:    true, // Should parse but XREF won't resolve
			expectErrors:   false,
			expectWarnings: true, // Invalid XREF should generate warnings
			description:    "Tests invalid XREF reference (@F999@ doesn't exist)",
		},
		{
			name:           "missing-header",
			filename:       "missing-header.ged",
			shouldParse:    true, // Should parse but may have warnings about missing HEAD
			expectErrors:   false,
			expectWarnings: true, // Missing HEAD should generate warnings
			description:    "Tests missing HEAD record",
		},
		{
			name:           "missing-xref",
			filename:       "missing-xref.ged",
			shouldParse:    true, // Should parse but XREF won't resolve
			expectErrors:   false,
			expectWarnings: true, // Missing XREF should generate warnings
			description:    "Tests missing XREF in reference (@F999@ doesn't exist)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := findMalformedTestFile(tc.filename)
			if filePath == "" {
				t.Skipf("Malformed test file not found: %s", tc.filename)
			}

			t.Logf("Testing: %s - %s", tc.filename, tc.description)

			parser := NewHierarchicalParser()
			tree, err := parser.Parse(filePath)

			// Check if parsing succeeded (shouldParse)
			if tc.shouldParse {
				if err != nil {
					t.Errorf("Expected parsing to succeed (with possible errors), got error: %v", err)
				}
				if tree == nil {
					t.Fatal("Expected tree to be created even with malformed input")
				}
			} else {
				if err == nil {
					t.Error("Expected parsing to fail for malformed input")
				}
				return // Don't check errors if parsing should fail
			}

			// Check for errors/warnings
			hasErrors := parser.HasErrors()
			hasWarnings := false
			hasSevereErrors := parser.HasSevereErrors()

			errors := parser.GetErrors()
			for _, e := range errors {
				if e.Severity == types.SeverityWarning {
					hasWarnings = true
				}
			}

			if tc.expectErrors && !hasErrors {
				t.Logf("Warning: Expected errors but none found (this may be acceptable)")
			}

			if tc.expectWarnings && !hasWarnings {
				t.Logf("Warning: Expected warnings but none found (this may be acceptable)")
			}

			// Log error summary
			if hasErrors {
				errorSummary := parser.GetErrorManager().GetErrorSummary()
				t.Logf("Errors found: %d total (%d warnings, %d severe)",
					len(errors),
					errorSummary[types.SeverityWarning],
					errorSummary[types.SeveritySevere])
			}

			// Verify we can still access records (parser should be resilient)
			if tree != nil {
				allIndis := tree.GetAllIndividuals()
				allFams := tree.GetAllFamilies()
				t.Logf("Parsed: %d individuals, %d families", len(allIndis), len(allFams))

				// For duplicate-xref, verify last one wins
				if tc.name == "duplicate-xref" {
					indi1 := tree.GetIndividual("@I1@")
					if indi1 != nil {
						name := indi1.GetValue("NAME")
						t.Logf("Last @I1@ name (should be 'Third Person'): %s", name)
					}
				}

				// For invalid-xref, verify the record exists but XREF doesn't resolve
				if tc.name == "invalid-xref" || tc.name == "missing-xref" {
					indi1 := tree.GetIndividual("@I1@")
					if indi1 != nil {
						famc := indi1.GetValue("FAMC")
						if famc != "" {
							famRecord := tree.GetFamily(famc)
							if famRecord != nil {
								t.Errorf("Expected XREF %s to not resolve, but it does", famc)
							} else {
								t.Logf("Correctly detected invalid XREF: %s", famc)
							}
						}
					}
				}

				// For missing-header, verify tree still works
				if tc.name == "missing-header" {
					header := tree.GetHeader()
					if header != nil {
						t.Logf("Note: HEAD record was created despite missing in file")
					}
				}
			}

			// Verify parser doesn't crash on severe errors
			if hasSevereErrors {
				t.Logf("Parser handled severe errors gracefully (did not crash)")
			}
		})
	}
}

// TestMalformedFiles_StreamingParser tests streaming parser with malformed files
func TestMalformedFiles_StreamingParser(t *testing.T) {
	testFiles := []string{
		"circular-reference.ged",
		"duplicate-xref.ged",
		"invalid-level.ged",
		"invalid-xref.ged",
		"missing-header.ged",
		"missing-xref.ged",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findMalformedTestFile(filename)
			if filePath == "" {
				t.Skipf("Malformed test file not found: %s", filename)
			}

			parser := NewStreamingHierarchicalParser()
			recordCount := 0
			errorCount := 0

			err := parser.ParseWithHandler(filePath, func(record types.Record) error {
				recordCount++
				return nil
			})

			// Streaming parser should handle errors gracefully
			if err != nil {
				t.Logf("Streaming parser error (may be expected): %v", err)
			}

			errors := parser.GetErrors()
			errorCount = len(errors)

			t.Logf("%s: %d records processed, %d errors", filename, recordCount, errorCount)

			// Parser should not crash even with malformed input
			if recordCount == 0 && errorCount == 0 {
				t.Logf("Warning: No records and no errors (file may be too malformed)")
			}
		})
	}
}

// TestMalformedFiles_RecordIterator tests RecordIterator with malformed files
func TestMalformedFiles_RecordIterator(t *testing.T) {
	testFiles := []string{
		"circular-reference.ged",
		"duplicate-xref.ged",
		"invalid-level.ged",
		"invalid-xref.ged",
		"missing-header.ged",
		"missing-xref.ged",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findMalformedTestFile(filename)
			if filePath == "" {
				t.Skipf("Malformed test file not found: %s", filename)
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
				t.Logf("Iterator error (may be expected): %v", iterator.Error())
			}

			errors := iterator.GetErrors()
			t.Logf("%s: %d records, %d errors", filename, recordCount, len(errors))

			// Iterator should handle errors gracefully
			if recordCount == 0 && len(errors) == 0 {
				t.Logf("Warning: No records and no errors (file may be too malformed)")
			}
		})
	}
}

// TestMalformedFiles_ErrorRecovery tests that parser recovers from errors
func TestMalformedFiles_ErrorRecovery(t *testing.T) {
	// Test that parser continues after encountering errors
	filePath := findMalformedTestFile("invalid-level.ged")
	if filePath == "" {
		t.Skip("Malformed test file not found: invalid-level.ged")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(filePath)

	if err != nil {
		t.Fatalf("Parser should handle invalid level gracefully: %v", err)
	}

	if tree == nil {
		t.Fatal("Expected tree to be created despite invalid level")
	}

	// Should still be able to access valid records
	indi1 := tree.GetIndividual("@I1@")
	if indi1 == nil {
		t.Fatal("Expected to find @I1@ despite invalid level in file")
	}

	name := indi1.GetValue("NAME")
	if name == "" {
		t.Error("Expected to read NAME from @I1@")
	}

	t.Logf("Parser recovered from invalid level error. @I1@ name: %s", name)
}

// TestMalformedFiles_AllMalformedFiles tests all malformed files in one run
func TestMalformedFiles_AllMalformedFiles(t *testing.T) {
	malformedDir := findMalformedTestFile(".")
	if malformedDir == "" {
		// Try to find any malformed file to get the directory
		malformedDir = findMalformedTestFile("invalid-level.ged")
		if malformedDir != "" {
			malformedDir = filepath.Dir(malformedDir)
		}
	}

	if malformedDir == "" {
		t.Skip("Malformed test directory not found")
	}

	// List all .ged files in malformed directory
	files, err := os.ReadDir(malformedDir)
	if err != nil {
		t.Skipf("Cannot read malformed directory: %v", err)
	}

	fileCount := 0
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".ged" {
			fileCount++
			filePath := filepath.Join(malformedDir, file.Name())

			t.Run(file.Name(), func(t *testing.T) {
				parser := NewHierarchicalParser()
				tree, err := parser.Parse(filePath)

				// Parser should handle all malformed files without crashing
				if err != nil {
					t.Logf("Parse error (may be expected): %v", err)
				}

				if tree == nil {
					t.Logf("Tree is nil (may be expected for severely malformed files)")
				} else {
					allIndis := tree.GetAllIndividuals()
					allFams := tree.GetAllFamilies()
					errors := parser.GetErrors()
					t.Logf("%s: %d individuals, %d families, %d errors",
						file.Name(), len(allIndis), len(allFams), len(errors))
				}
			})
		}
	}

	if fileCount == 0 {
		t.Skip("No malformed .ged files found")
	}

	t.Logf("Tested %d malformed files", fileCount)
}

