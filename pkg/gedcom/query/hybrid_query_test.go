package query

import (
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestFilterQuery_Hybrid(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create a simple GEDCOM tree
	tree := gedcom.NewGedcomTree()

	// Add individuals
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := gedcom.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	indi2Line.AddChild(name2Line)
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Build hybrid graph
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath)
	if err != nil {
		t.Fatalf("Failed to build hybrid graph: %v", err)
	}
	defer graph.Close()

	// Test by name (substring match should work)
	fq1 := NewFilterQuery(graph)
	results, err := fq1.ByName("john").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'john', got %d", len(results))
		if len(results) == 0 {
			// Debug: check what's in SQLite
			var count int
			var name, nameLower string
			err = graph.hybridStorage.SQLite().QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'individual'").Scan(&count)
			if err != nil {
				t.Logf("Error querying SQLite: %v", err)
			} else {
				t.Logf("Found %d individuals in SQLite", count)
			}
			err = graph.hybridStorage.SQLite().QueryRow("SELECT name, name_lower FROM nodes WHERE type = 'individual' LIMIT 1").Scan(&name, &nameLower)
			if err == nil {
				t.Logf("Sample name: '%s', name_lower: '%s'", name, nameLower)
			}
		}
		return
	}
	if results[0].GetName() != "John /Doe/" {
		t.Errorf("Expected 'John /Doe/', got '%s'", results[0].GetName())
	}

	// Test by name exact (case-insensitive) - create new query to avoid state issues
	fq2 := NewFilterQuery(graph)
	results, err = fq2.ByNameExact("jane /smith/").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for exact 'jane /smith/', got %d", len(results))
		if len(results) == 0 {
			// Debug: check what's stored and what the query returns
			var name, nameLower string
			err = graph.hybridStorage.SQLite().QueryRow("SELECT name, name_lower FROM nodes WHERE type = 'individual' AND xref = '@I2@'").Scan(&name, &nameLower)
			if err == nil {
				t.Logf("I2 name: '%s', name_lower: '%s'", name, nameLower)
			}
			// Check what the query helper returns
			nodeIDs, err := graph.queryHelpers.FindByNameExact("jane /smith/")
			if err == nil {
				t.Logf("FindByNameExact('jane /smith/') returned %d node IDs: %v", len(nodeIDs), nodeIDs)
			} else {
				t.Logf("FindByNameExact error: %v", err)
			}
		}
	}

	// Test by name starts (case-insensitive) - create new query to avoid state issues
	fq3 := NewFilterQuery(graph)
	results, err = fq3.ByNameStarts("jane").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for starts 'jane', got %d", len(results))
		if len(results) == 0 {
			// Debug: check what the query helper returns
			nodeIDs, err := graph.queryHelpers.FindByNameStarts("jane")
			if err == nil {
				t.Logf("FindByNameStarts('jane') returned %d node IDs: %v", len(nodeIDs), nodeIDs)
			} else {
				t.Logf("FindByNameStarts error: %v", err)
			}
		}
	}
}

