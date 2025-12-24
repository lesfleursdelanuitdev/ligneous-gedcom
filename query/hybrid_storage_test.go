package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

func TestHybridStorage_Initialization(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create hybrid storage
	hs, err := NewHybridStorage(sqlitePath, badgerPath, nil)
	if err != nil {
		t.Fatalf("Failed to create hybrid storage: %v", err)
	}
	defer hs.Close()

	// Verify SQLite is initialized
	if hs.SQLite() == nil {
		t.Error("SQLite database is nil")
	}

	// Verify BadgerDB is initialized
	if hs.BadgerDB() == nil {
		t.Error("BadgerDB database is nil")
	}

	// Test SQLite schema - try a simple query
	var count int
	err = hs.SQLite().QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query SQLite: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 nodes, got %d", count)
	}

	// Test BadgerDB - try a simple read
	err = hs.BadgerDB().View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("test_key"))
		if err != badger.ErrKeyNotFound {
			// Key not found is expected for empty database
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to access BadgerDB: %v", err)
	}
}

func TestHybridStorage_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	hs, err := NewHybridStorage(sqlitePath, badgerPath, nil)
	if err != nil {
		t.Fatalf("Failed to create hybrid storage: %v", err)
	}

	// Close should not error
	if err := hs.Close(); err != nil {
		t.Errorf("Failed to close hybrid storage: %v", err)
	}

	// Verify files exist (they should, even after close)
	if _, err := os.Stat(sqlitePath); err != nil {
		t.Errorf("SQLite file should exist: %v", err)
	}

	// BadgerDB creates a directory, check that
	if _, err := os.Stat(badgerPath); err != nil {
		t.Errorf("BadgerDB directory should exist: %v", err)
	}
}

