package query

import (
	"fmt"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestNotesQuery_InlineNotes_Individual tests inline notes for individuals
func TestNotesQuery_InlineNotes_Individual(t *testing.T) {
	// Create a tree with inline notes
	tree := types.NewGedcomTree()

	// Add individual with inline note
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	nameLine := types.NewGedcomLine(1, "NAME", "Test /Person/", "")
	indiLine.AddChild(nameLine)

	// Add inline note with continuation
	noteLine := types.NewGedcomLine(1, "NOTE", "This is a note", "")
	noteLine.AddChild(types.NewGedcomLine(2, "CONT", "This is a continuation line", ""))
	noteLine.AddChild(types.NewGedcomLine(2, "CONC", "This is concatenated", ""))
	noteLine.AddChild(types.NewGedcomLine(2, "CONT", "Another continuation", ""))
	indiLine.AddChild(noteLine)

	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph and query
	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test GetAllNotes
	iq := qb.Individual("@I1@")
	notes, err := iq.GetAllNotes()
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected at least one note, got 0")
	}

	// Verify inline note
	foundInline := false
	for _, note := range notes {
		if note.IsInline {
			foundInline = true
			if note.Text == "" {
				t.Error("Expected note text, got empty")
			}
			// Verify continuation lines are included
			if len(note.Text) < 20 {
				t.Errorf("Expected note text with continuation, got short text: %s", note.Text)
			}
		}
	}

	if !foundInline {
		t.Error("Expected to find inline note")
	}
}

// TestNotesQuery_InlineNotes_Family tests inline notes for families
func TestNotesQuery_InlineNotes_Family(t *testing.T) {
	// Create a tree with inline notes
	tree := types.NewGedcomTree()

	// Add individuals
	husbandLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	husbandLine.AddChild(types.NewGedcomLine(1, "NAME", "Husband /Test/", ""))
	husband := types.NewIndividualRecord(husbandLine)
	tree.AddRecord(husband)

	wifeLine := types.NewGedcomLine(0, "INDI", "", "@I2@")
	wifeLine.AddChild(types.NewGedcomLine(1, "NAME", "Wife /Test/", ""))
	wife := types.NewIndividualRecord(wifeLine)
	tree.AddRecord(wife)

	// Add family with inline note
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))

	// Add inline note with continuation
	noteLine := types.NewGedcomLine(1, "NOTE", "Family note text", "")
	noteLine.AddChild(types.NewGedcomLine(2, "CONT", "Continuation line 1", ""))
	noteLine.AddChild(types.NewGedcomLine(2, "CONC", "Concatenated text", ""))
	noteLine.AddChild(types.NewGedcomLine(2, "CONT", "Continuation line 2", ""))
	famLine.AddChild(noteLine)

	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	// Build graph and query
	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test GetAllNotes for family
	fq := qb.Family("@F1@")
	notes, err := fq.GetAllNotes()
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected at least one note, got 0")
	}

	// Verify inline note
	foundInline := false
	for _, note := range notes {
		if note.IsInline {
			foundInline = true
			if note.Text == "" {
				t.Error("Expected note text, got empty")
			}
			// Verify continuation lines are included
			if len(note.Text) < 20 {
				t.Errorf("Expected note text with continuation, got short text: %s", note.Text)
			}
		}
	}

	if !foundInline {
		t.Error("Expected to find inline note")
	}
}

// TestNotesQuery_ReferencedNote tests referenced notes (via XREF)
func TestNotesQuery_ReferencedNote(t *testing.T) {
	// Create a tree with referenced note
	tree := types.NewGedcomTree()

	// Add note record
	noteLine := types.NewGedcomLine(0, "NOTE", "", "@N1@")
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", "This is a referenced note", ""))
	noteLine.AddChild(types.NewGedcomLine(1, "CONC", " with concatenation", ""))
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", " and continuation", ""))
	note := types.NewNoteRecord(noteLine)
	tree.AddRecord(note)

	// Add individual referencing the note
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "Test /Person/", ""))
	indiLine.AddChild(types.NewGedcomLine(1, "NOTE", "@N1@", ""))
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph and query
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test GetAllNotes
	iq := qb.Individual("@I1@")
	notes, err := iq.GetAllNotes()
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected at least one note, got 0")
	}

	// Verify referenced note
	foundReferenced := false
	for _, note := range notes {
		if !note.IsInline && note.XrefID == "@N1@" {
			foundReferenced = true
			if note.Text == "" {
				t.Error("Expected note text, got empty")
			}
			// Verify continuation lines are included
			if len(note.Text) < 20 {
				t.Errorf("Expected note text with continuation, got short text: %s", note.Text)
			}
			if note.NoteRecord == nil {
				t.Error("Expected NoteRecord for referenced note, got nil")
			}
		}
	}

	if !foundReferenced {
		t.Error("Expected to find referenced note")
	}

	// Test GetRecordsForNote
	records, err := graph.GetRecordsForNote("@N1@")
	if err != nil {
		t.Fatalf("Failed to get records for note: %v", err)
	}

	if len(records) == 0 {
		t.Error("Expected at least one record referencing the note, got 0")
	}
}

// TestNotesQuery_MixedNotes tests both inline and referenced notes
func TestNotesQuery_MixedNotes(t *testing.T) {
	// Create a tree with both inline and referenced notes
	tree := types.NewGedcomTree()

	// Add note record
	noteLine := types.NewGedcomLine(0, "NOTE", "", "@N1@")
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", "Referenced note text", ""))
	note := types.NewNoteRecord(noteLine)
	tree.AddRecord(note)

	// Add individual with both inline and referenced notes
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "Test /Person/", ""))

	// Inline note
	inlineNoteLine := types.NewGedcomLine(1, "NOTE", "Inline note text", "")
	inlineNoteLine.AddChild(types.NewGedcomLine(2, "CONT", "Inline continuation", ""))
	indiLine.AddChild(inlineNoteLine)

	// Referenced note
	indiLine.AddChild(types.NewGedcomLine(1, "NOTE", "@N1@", ""))

	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph and query
	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test GetAllNotes
	iq := qb.Individual("@I1@")
	notes, err := iq.GetAllNotes()
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	if len(notes) < 2 {
		t.Errorf("Expected at least 2 notes (inline and referenced), got %d", len(notes))
	}

	// Verify we have both types
	hasInline := false
	hasReferenced := false
	for _, note := range notes {
		if note.IsInline {
			hasInline = true
		} else if note.XrefID != "" {
			hasReferenced = true
		}
	}

	if !hasInline {
		t.Error("Expected to find inline note")
	}
	if !hasReferenced {
		t.Error("Expected to find referenced note")
	}
}

// TestNotesQuery_NoteWithMultipleContinuations tests notes with multiple CONT/CONC lines
func TestNotesQuery_NoteWithMultipleContinuations(t *testing.T) {
	// Create a tree with note having many continuation lines
	tree := types.NewGedcomTree()

	// Add note record with many continuations
	noteLine := types.NewGedcomLine(0, "NOTE", "", "@N1@")
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", "Line 1", ""))
	noteLine.AddChild(types.NewGedcomLine(1, "CONC", " concatenated", ""))
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", "Line 2", ""))
	noteLine.AddChild(types.NewGedcomLine(1, "CONC", " more concat", ""))
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", "Line 3", ""))
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", "Line 4", ""))
	note := types.NewNoteRecord(noteLine)
	tree.AddRecord(note)

	// Add individual referencing the note
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "Test /Person/", ""))
	indiLine.AddChild(types.NewGedcomLine(1, "NOTE", "@N1@", ""))
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph and query
	_, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test GetAllNotes
	iq := qb.Individual("@I1@")
	notes, err := iq.GetAllNotes()
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	if len(notes) == 0 {
		t.Error("Expected at least one note, got 0")
	}

	// Verify note text includes all continuations
	for _, note := range notes {
		if !note.IsInline && note.XrefID == "@N1@" {
			if len(note.Text) < 30 {
				t.Errorf("Expected note text with all continuations, got short text: %s", note.Text)
			}
			// Should contain multiple lines
			if len(note.Text) < 50 {
				t.Errorf("Expected longer note text, got: %s", note.Text)
			}
		}
	}
}

// TestNotesQuery_GetRecordsForNote_MultipleReferences tests GetRecordsForNote with multiple references
func TestNotesQuery_GetRecordsForNote_MultipleReferences(t *testing.T) {
	// Create a tree with note referenced by multiple records
	tree := types.NewGedcomTree()

	// Add note record
	noteLine := types.NewGedcomLine(0, "NOTE", "", "@N1@")
	noteLine.AddChild(types.NewGedcomLine(1, "CONT", "Shared note", ""))
	note := types.NewNoteRecord(noteLine)
	tree.AddRecord(note)

	// Add multiple individuals referencing the note
	for i := 1; i <= 3; i++ {
		indiLine := types.NewGedcomLine(0, "INDI", "", fmt.Sprintf("@I%d@", i))
		indiLine.AddChild(types.NewGedcomLine(1, "NAME", fmt.Sprintf("Person %d /Test/", i), ""))
		indiLine.AddChild(types.NewGedcomLine(1, "NOTE", "@N1@", ""))
		indi := types.NewIndividualRecord(indiLine)
		tree.AddRecord(indi)
	}

	// Add family referencing the note
	famLine := types.NewGedcomLine(0, "FAM", "", "@F1@")
	famLine.AddChild(types.NewGedcomLine(1, "HUSB", "@I1@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "WIFE", "@I2@", ""))
	famLine.AddChild(types.NewGedcomLine(1, "NOTE", "@N1@", ""))
	fam := types.NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetRecordsForNote
	records, err := graph.GetRecordsForNote("@N1@")
	if err != nil {
		t.Fatalf("Failed to get records for note: %v", err)
	}

	if len(records) < 4 {
		t.Errorf("Expected at least 4 records (3 individuals + 1 family), got %d", len(records))
	}
}

// TestNotesQuery_GetRecordsForNote_InvalidNote tests GetRecordsForNote with invalid note
func TestNotesQuery_GetRecordsForNote_InvalidNote(t *testing.T) {
	// Create a minimal tree
	tree := types.NewGedcomTree()
	indiLine := types.NewGedcomLine(0, "INDI", "", "@I1@")
	indiLine.AddChild(types.NewGedcomLine(1, "NAME", "Test /Person/", ""))
	indi := types.NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	// Build graph
	graph, err := BuildGraph(tree)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Test GetRecordsForNote with invalid note
	records, err := graph.GetRecordsForNote("@INVALID@")
	if err == nil {
		t.Error("Expected error for invalid note, got nil")
	}
	if records != nil {
		t.Errorf("Expected nil records for invalid note, got %v", records)
	}
}

