package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestBuildGraphHybrid_Basic(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create a simple GEDCOM tree
	tree := gedcom.NewGedcomTree()

	// Add a few individuals
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := gedcom.NewGedcomLine(1, "NAME", "Jane /Doe/", "")
	indi2Line.AddChild(name2Line)
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Add a family
	fam1Line := gedcom.NewGedcomLine(0, "FAM", "", "@F1@")
	husbLine := gedcom.NewGedcomLine(1, "HUSB", "@I1@", "")
	wifeLine := gedcom.NewGedcomLine(1, "WIFE", "@I2@", "")
	fam1Line.AddChild(husbLine)
	fam1Line.AddChild(wifeLine)
	fam1 := gedcom.NewFamilyRecord(fam1Line)
	tree.AddRecord(fam1)

	// Build hybrid graph
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	if err != nil {
		t.Fatalf("Failed to build hybrid graph: %v", err)
	}
	defer graph.Close()

	// Verify graph was created
	if graph == nil {
		t.Fatal("Graph is nil")
	}

	// Verify SQLite has data
	storage := graph.hybridStorage
	if storage == nil {
		t.Fatal("Hybrid storage is nil")
	}

	var count int
	err = storage.SQLite().QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query nodes: %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", count)
	}

	// Verify BadgerDB has data
	err = storage.BadgerDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("node:")
		it := txn.NewIterator(opts)
		defer it.Close()

		nodeCount := 0
		for it.Rewind(); it.Valid(); it.Next() {
			nodeCount++
		}

		if nodeCount < 2 {
			t.Errorf("Expected at least 2 nodes in BadgerDB, got %d", nodeCount)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Failed to query BadgerDB: %v", err)
	}
}

func TestBuildGraphHybrid_Close(t *testing.T) {
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create minimal tree
	tree := gedcom.NewGedcomTree()
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	if err != nil {
		t.Fatalf("Failed to build hybrid graph: %v", err)
	}

	// Close should work
	if err := graph.Close(); err != nil {
		t.Errorf("Failed to close graph: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(sqlitePath); err != nil {
		t.Errorf("SQLite file should exist: %v", err)
	}
}

// Close closes the graph and its hybrid storage
func (g *Graph) Close() error {
	if g.hybridStorage != nil {
		return g.hybridStorage.Close()
	}
	return nil
}

