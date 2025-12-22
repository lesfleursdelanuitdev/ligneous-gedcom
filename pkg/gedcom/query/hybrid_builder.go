package query

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// BuildGraphHybrid builds a graph using hybrid storage (SQLite + BadgerDB)
// This function coordinates the building process by delegating to:
// - buildGraphInSQLite: Builds indexes in SQLite (see hybrid_sqlite_builder.go)
// - buildGraphInBadgerDB: Stores graph structure in BadgerDB (see hybrid_badger_builder.go)
func BuildGraphHybrid(tree *gedcom.GedcomTree, sqlitePath, badgerPath string) (*Graph, error) {
	// Initialize hybrid storage
	storage, err := NewHybridStorage(sqlitePath, badgerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize hybrid storage: %w", err)
	}

	// Create graph structure (will use hybrid storage)
	graph := NewGraph(tree)
	graph.hybridStorage = storage
	graph.hybridMode = true

	// Initialize query helpers with prepared statements
	queryHelpers, err := NewHybridQueryHelpers(storage.SQLite())
	if err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to create query helpers: %w", err)
	}
	graph.queryHelpers = queryHelpers

	// Initialize hybrid cache with default sizes
	// Node cache: 50K nodes (configurable)
	// XREF cache: 25K entries (configurable)
	// Query cache: 5K queries (configurable)
	hybridCache, err := NewHybridCache(50000, 25000, 5000)
	if err != nil {
		queryHelpers.Close()
		storage.Close()
		return nil, fmt.Errorf("failed to create hybrid cache: %w", err)
	}
	graph.hybridCache = hybridCache

	// Build graph in both databases
	if err := buildGraphInSQLite(storage, tree, graph); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to build SQLite indexes: %w", err)
	}

	if err := buildGraphInBadgerDB(storage, tree, graph); err != nil {
		storage.Close()
		return nil, fmt.Errorf("failed to build BadgerDB graph: %w", err)
	}

	return graph, nil
}
