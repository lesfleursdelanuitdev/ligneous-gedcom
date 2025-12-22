package query

import (
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestHybridCache_Basic(t *testing.T) {
	// Create cache
	cache, err := NewHybridCache(100, 50, 10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Create a test node
	indiLine := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	indi := gedcom.NewIndividualRecord(indiLine)
	node := NewIndividualNode("@I1@", indi)

	// Test node cache
	nodeID := uint32(1)
	cache.SetNode(nodeID, node)
	retrieved, found := cache.GetNode(nodeID)
	if !found {
		t.Error("Node not found in cache")
	}
	if retrieved != node {
		t.Error("Retrieved node doesn't match")
	}

	// Test XREF cache
	xref := "@I1@"
	cache.SetXrefToID(xref, nodeID)
	retrievedID, found := cache.GetXrefToID(xref)
	if !found {
		t.Error("XREF mapping not found in cache")
	}
	if retrievedID != nodeID {
		t.Errorf("Expected nodeID %d, got %d", nodeID, retrievedID)
	}

	// Test ID to XREF cache
	retrievedXref, found := cache.GetIDToXref(nodeID)
	if !found {
		t.Error("ID to XREF mapping not found in cache")
	}
	if retrievedXref != xref {
		t.Errorf("Expected XREF %s, got %s", xref, retrievedXref)
	}

	// Test query cache
	queryKey := "test:query:key"
	nodeIDs := []uint32{1, 2, 3}
	cache.SetQuery(queryKey, nodeIDs)
	retrievedIDs, found := cache.GetQuery(queryKey)
	if !found {
		t.Error("Query result not found in cache")
	}
	if len(retrievedIDs) != len(nodeIDs) {
		t.Errorf("Expected %d IDs, got %d", len(nodeIDs), len(retrievedIDs))
	}

	// Test stats
	stats := cache.Stats()
	if stats.NodeCacheSize == 0 {
		t.Error("Node cache should have entries")
	}
	if stats.XrefCacheSize == 0 {
		t.Error("XREF cache should have entries")
	}
	if stats.QueryCacheSize == 0 {
		t.Error("Query cache should have entries")
	}

	// Test clear
	cache.Clear()
	stats = cache.Stats()
	if stats.NodeCacheSize != 0 {
		t.Error("Node cache should be empty after clear")
	}
}

func TestHybridCache_WithGraph(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create a simple GEDCOM tree
	tree := gedcom.NewGedcomTree()
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	// Build hybrid graph
	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath)
	if err != nil {
		t.Fatalf("Failed to build hybrid graph: %v", err)
	}
	defer graph.Close()

	// Verify cache exists
	if graph.hybridCache == nil {
		t.Fatal("Hybrid cache should be initialized")
	}

	// Get individual (should load from BadgerDB and cache)
	node1 := graph.GetIndividual("@I1@")
	if node1 == nil {
		t.Fatal("Failed to get individual")
	}

	// Get again (should use cache)
	node2 := graph.GetIndividual("@I1@")
	if node2 != node1 {
		t.Error("Second call should return cached node")
	}

	// Check cache stats
	stats := graph.hybridCache.Stats()
	if stats.NodeCacheSize == 0 {
		t.Error("Node cache should have entries after loading")
	}
	if stats.XrefCacheSize == 0 {
		t.Error("XREF cache should have entries after loading")
	}
}

