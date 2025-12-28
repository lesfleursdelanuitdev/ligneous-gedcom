package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
)

// TestNotesQuery_GetRecordsForNote tests GetRecordsForNote
func TestNotesQuery_GetRecordsForNote(t *testing.T) {
	filePath := findTestDataFile("royal92.ged")
	if filePath == "" {
		t.Skip("Test data file not found: royal92.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Find individuals with notes
	notesFound := 0
	for i := 0; i < len(allIndividuals) && notesFound < 10; i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		notes, err := iq.GetAllNotes()
		if err != nil {
			continue
		}

		if len(notes) > 0 {
			notesFound++
			// Test GetRecordsForNote for each note (only for referenced notes, not inline)
			for _, note := range notes {
				if !note.IsInline && note.XrefID != "" {
					graph, err := BuildGraph(tree)
					if err != nil {
						t.Errorf("Failed to build graph: %v", err)
						continue
					}
					records, err := graph.GetRecordsForNote(note.XrefID)
					if err != nil {
						t.Errorf("Failed to get records for note %s: %v", note.XrefID, err)
					}
					_ = records // Just verify it doesn't panic
				}
			}
		}
	}

	if notesFound == 0 {
		t.Skip("No individuals with notes found in test data")
	}
}

// TestNotesQuery_GetAllNotes_Family_Comprehensive tests GetAllNotes for families comprehensively
func TestNotesQuery_GetAllNotes_Family_Comprehensive(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder: %v", err)
			}

			// Get all families
			allFamilies, err := qb.Families().All()
			if err != nil {
				t.Fatalf("Failed to get families: %v", err)
			}

			if len(allFamilies) == 0 {
				t.Skip("No families found")
			}

			// Test GetAllNotes on all families
			for i := 0; i < len(allFamilies); i++ {
				xref := allFamilies[i].XrefID()
				fq := qb.Family(xref)

				notes, err := fq.GetAllNotes()
				if err != nil {
					t.Errorf("Failed to get notes for family %s: %v", xref, err)
				}
				_ = notes // Just verify it doesn't panic
			}
		})
	}
}

// TestNotesQuery_GetAllNotes_Individual_WithContinuation tests GetAllNotes with continuation lines
func TestNotesQuery_GetAllNotes_Individual_WithContinuation(t *testing.T) {
	filePath := findTestDataFile("royal92.ged")
	if filePath == "" {
		t.Skip("Test data file not found: royal92.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Test GetAllNotes on all individuals to exercise getFullNoteText
	for i := 0; i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		notes, err := iq.GetAllNotes()
		if err != nil {
			t.Errorf("Failed to get notes for %s: %v", xref, err)
		}
		_ = notes // Just verify it doesn't panic
	}
}

// TestAncestorQuery_Count_WithError tests Count with error cases
func TestAncestorQuery_Count_WithError(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Count with invalid XREF
	// Note: Execute() returns nil, nil (not an error) for invalid XREFs
	// So Count() will return 0, nil (no error, just empty result)
	iq := qb.Individual("@INVALID@")
	count, err := iq.Ancestors().Count()
	if err != nil {
		t.Errorf("Unexpected error for invalid XREF: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for invalid XREF, got %d", count)
	}
}

// TestAncestorQuery_Exists_WithError tests Exists with error cases
func TestAncestorQuery_Exists_WithError(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Exists with invalid XREF
	// Note: Execute() returns nil, nil (not an error) for invalid XREFs
	// So Exists() will return false, nil (no error, just no ancestors)
	iq := qb.Individual("@INVALID@")
	exists, err := iq.Ancestors().Exists()
	if err != nil {
		t.Errorf("Unexpected error for invalid XREF: %v", err)
	}
	if exists {
		t.Error("Expected exists false for invalid XREF, got true")
	}
}

// TestDescendantQuery_Count_WithError tests Count with error cases
func TestDescendantQuery_Count_WithError(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Count with invalid XREF
	// Note: Execute() returns nil, nil (not an error) for invalid XREFs
	// So Count() will return 0, nil (no error, just empty result)
	iq := qb.Individual("@INVALID@")
	count, err := iq.Descendants().Count()
	if err != nil {
		t.Errorf("Unexpected error for invalid XREF: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for invalid XREF, got %d", count)
	}
}

// TestDescendantQuery_Exists_WithError tests Exists with error cases
func TestDescendantQuery_Exists_WithError(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test Exists with invalid XREF
	// Note: Execute() returns nil, nil (not an error) for invalid XREFs
	// So Exists() will return false, nil (no error, just no descendants)
	iq := qb.Individual("@INVALID@")
	exists, err := iq.Descendants().Exists()
	if err != nil {
		t.Errorf("Unexpected error for invalid XREF: %v", err)
	}
	if exists {
		t.Error("Expected exists false for invalid XREF, got true")
	}
}

// TestAncestorQuery_Count_EmptyResults tests Count when no ancestors exist
func TestAncestorQuery_Count_EmptyResults(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Find individuals with no ancestors (root individuals)
	for i := 0; i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		count, err := iq.Ancestors().Count()
		if err != nil {
			t.Errorf("Failed to count ancestors for %s: %v", xref, err)
		}
		// Count might be 0 for root individuals, that's valid
		_ = count
	}
}

// TestDescendantQuery_Count_EmptyResults tests Count when no descendants exist
func TestDescendantQuery_Count_EmptyResults(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Find individuals with no descendants (leaf individuals)
	for i := 0; i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		count, err := iq.Descendants().Count()
		if err != nil {
			t.Errorf("Failed to count descendants for %s: %v", xref, err)
		}
		// Count might be 0 for leaf individuals, that's valid
		_ = count
	}
}

