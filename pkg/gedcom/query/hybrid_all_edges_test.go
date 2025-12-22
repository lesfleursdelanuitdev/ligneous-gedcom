package query

import (
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// TestHybridStorage_AllEdgeTypes tests that all edge types are properly stored and loaded
func TestHybridStorage_AllEdgeTypes(t *testing.T) {
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create test data with all edge types
	tree := gedcom.NewGedcomTree()

	// Individual
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "NOTE", "@N1@", ""))
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "SOUR", "@S1@", ""))
	indi1Line.AddChild(gedcom.NewGedcomLine(1, "BIRT", "", ""))
	indi1Line.AddChild(gedcom.NewGedcomLine(2, "DATE", "15 JAN 1800", ""))
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Note
	note1Line := gedcom.NewGedcomLine(0, "NOTE", "", "@N1@")
	note1Line.AddChild(gedcom.NewGedcomLine(1, "CONT", "This is a note", ""))
	note1Line.AddChild(gedcom.NewGedcomLine(1, "SOUR", "@S1@", ""))
	note1 := gedcom.NewNoteRecord(note1Line)
	tree.AddRecord(note1)

	// Source
	source1Line := gedcom.NewGedcomLine(0, "SOUR", "", "@S1@")
	source1Line.AddChild(gedcom.NewGedcomLine(1, "TITL", "Source Title", ""))
	source1Line.AddChild(gedcom.NewGedcomLine(1, "REPO", "@R1@", ""))
	source1 := gedcom.NewSourceRecord(source1Line)
	tree.AddRecord(source1)

	// Repository
	repo1Line := gedcom.NewGedcomLine(0, "REPO", "", "@R1@")
	repo1Line.AddChild(gedcom.NewGedcomLine(1, "NAME", "Repository Name", ""))
	repo1 := gedcom.NewRepositoryRecord(repo1Line)
	tree.AddRecord(repo1)

	// Family
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "HUSB", "@I1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "NOTE", "@N1@", ""))
	fam1Line.AddChild(gedcom.NewGedcomLine(1, "SOUR", "@S1@", ""))
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Build hybrid graph
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	if err != nil {
		t.Fatalf("Failed to build hybrid graph: %v", err)
	}
	defer graph.Close()

	// Test NOTE edge from individual
	indiNode := graph.GetIndividual("@I1@")
	if indiNode == nil {
		t.Fatal("Failed to get individual")
	}

	noteEdges := 0
	for _, edge := range indiNode.OutEdges() {
		if edge.EdgeType == EdgeTypeNOTE {
			noteEdges++
			if edge.To.ID() != "@N1@" {
				t.Errorf("Expected NOTE edge to point to @N1@, got %s", edge.To.ID())
			}
		}
	}
	if noteEdges == 0 {
		t.Error("Expected NOTE edge from individual")
	}

	// Test SOUR edge from individual
	sourEdges := 0
	for _, edge := range indiNode.OutEdges() {
		if edge.EdgeType == EdgeTypeSOUR {
			sourEdges++
			if edge.To.ID() != "@S1@" {
				t.Errorf("Expected SOUR edge to point to @S1@, got %s", edge.To.ID())
			}
		}
	}
	if sourEdges == 0 {
		t.Error("Expected SOUR edge from individual")
	}

	// Test SOUR edge from note
	noteNode := graph.GetNote("@N1@")
	if noteNode == nil {
		t.Fatal("Failed to get note")
	}

	noteSourEdges := 0
	for _, edge := range noteNode.OutEdges() {
		if edge.EdgeType == EdgeTypeSOUR {
			noteSourEdges++
			if edge.To.ID() != "@S1@" {
				t.Errorf("Expected SOUR edge from note to point to @S1@, got %s", edge.To.ID())
			}
		}
	}
	if noteSourEdges == 0 {
		t.Error("Expected SOUR edge from note")
	}

	// Test REPO edge from source
	sourceNode := graph.GetSource("@S1@")
	if sourceNode == nil {
		t.Fatal("Failed to get source")
	}

	repoEdges := 0
	for _, edge := range sourceNode.OutEdges() {
		if edge.EdgeType == EdgeTypeREPO {
			repoEdges++
			if edge.To.ID() != "@R1@" {
				t.Errorf("Expected REPO edge to point to @R1@, got %s", edge.To.ID())
			}
		}
	}
	if repoEdges == 0 {
		t.Error("Expected REPO edge from source")
	}

	// Test NOTE and SOUR edges from family
	famNode := graph.GetFamily("@F1@")
	if famNode == nil {
		t.Fatal("Failed to get family")
	}

	famNoteEdges := 0
	famSourEdges := 0
	for _, edge := range famNode.OutEdges() {
		if edge.EdgeType == EdgeTypeNOTE {
			famNoteEdges++
		}
		if edge.EdgeType == EdgeTypeSOUR {
			famSourEdges++
		}
	}
	if famNoteEdges == 0 {
		t.Error("Expected NOTE edge from family")
	}
	if famSourEdges == 0 {
		t.Error("Expected SOUR edge from family")
	}

	// Test has_event edge (from birth event)
	hasEventEdges := 0
	for _, edge := range indiNode.OutEdges() {
		if edge.EdgeType == EdgeTypeHasEvent {
			hasEventEdges++
		}
	}
	if hasEventEdges == 0 {
		t.Error("Expected has_event edge from individual")
	}

	t.Logf("✓ All edge types verified: NOTE=%d, SOUR=%d, REPO=%d, has_event=%d",
		noteEdges+noteSourEdges+famNoteEdges, sourEdges+famSourEdges+noteSourEdges, repoEdges, hasEventEdges)
}

