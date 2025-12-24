package query

import (
	"path/filepath"
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

// TestHybridStorage_EndToEnd tests a complete workflow with hybrid storage
func TestHybridStorage_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	// Create test data
	tree := gedcom.NewGedcomTree()

	// Add individuals
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	sex1Line := gedcom.NewGedcomLine(1, "SEX", "M", "")
	birt1Line := gedcom.NewGedcomLine(1, "BIRT", "", "")
	date1Line := gedcom.NewGedcomLine(2, "DATE", "15 JAN 1800", "")
	birt1Line.AddChild(date1Line)
	indi1Line.AddChild(name1Line)
	indi1Line.AddChild(sex1Line)
	indi1Line.AddChild(birt1Line)
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	indi2Line := gedcom.NewGedcomLine(0, "INDI", "", "@I2@")
	name2Line := gedcom.NewGedcomLine(1, "NAME", "Jane /Smith/", "")
	sex2Line := gedcom.NewGedcomLine(1, "SEX", "F", "")
	indi2Line.AddChild(name2Line)
	indi2Line.AddChild(sex2Line)
	indi2 := gedcom.NewIndividualRecord(indi2Line)
	tree.AddRecord(indi2)

	// Add family
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

	// Test 1: Query by name
	fq := NewFilterQuery(graph)
	results, err := fq.ByName("John").Execute()
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test 2: Get individual
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Failed to get individual")
	}
	if node.Individual.GetName() != "John /Doe/" {
		t.Errorf("Expected 'John /Doe/', got '%s'", node.Individual.GetName())
	}

	// Test 3: Get family
	famNode := graph.GetFamily("@F1@")
	if famNode == nil {
		t.Fatal("Failed to get family")
	}

	// Test 4: Query by sex
	results, err = fq.BySex("M").Execute()
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test 5: Cache statistics
	if graph.hybridCache != nil {
		stats := graph.hybridCache.Stats()
		if stats.NodeCacheSize == 0 && stats.XrefCacheSize == 0 {
			t.Error("Cache should have entries after queries")
		}
	}
}

// TestHybridStorage_Serialization tests serialization/deserialization
func TestHybridStorage_Serialization(t *testing.T) {
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	tree := gedcom.NewGedcomTree()
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := gedcom.NewGedcomLine(1, "NAME", "Test /Person/", "")
	indi1Line.AddChild(name1Line)
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}
	defer graph.Close()

	// Serialize node
	node := graph.GetIndividual("@I1@")
	if node == nil {
		t.Fatal("Failed to get node")
	}

	data, err := SerializeNode(node, graph)
	if err != nil {
		t.Fatalf("Failed to serialize node: %v", err)
	}

	// Deserialize node
	deserialized, err := DeserializeNode(data, graph)
	if err != nil {
		t.Fatalf("Failed to deserialize node: %v", err)
	}

	if deserialized.ID() != node.ID() {
		t.Errorf("Expected ID %s, got %s", node.ID(), deserialized.ID())
	}
	if deserialized.NodeType() != node.NodeType() {
		t.Errorf("Expected type %s, got %s", node.NodeType(), deserialized.NodeType())
	}
}

// TestHybridStorage_QueryHelpers tests SQLite query helpers
func TestHybridStorage_QueryHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	sqlitePath := filepath.Join(tmpDir, "test_indexes.db")
	badgerPath := filepath.Join(tmpDir, "test_graph")

	tree := gedcom.NewGedcomTree()
	indi1Line := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	name1Line := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	indi1Line.AddChild(name1Line)
	indi1 := gedcom.NewIndividualRecord(indi1Line)
	tree.AddRecord(indi1)

	graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}
	defer graph.Close()

	helpers := graph.queryHelpers
	if helpers == nil {
		t.Fatal("Query helpers should be initialized")
	}

	// Test FindByXref
	nodeID, err := helpers.FindByXref("@I1@")
	if err != nil {
		t.Fatalf("FindByXref failed: %v", err)
	}
	if nodeID == 0 {
		t.Error("Expected non-zero node ID")
	}

	// Test FindXrefByID
	xref, err := helpers.FindXrefByID(nodeID)
	if err != nil {
		t.Fatalf("FindXrefByID failed: %v", err)
	}
	if xref != "@I1@" {
		t.Errorf("Expected '@I1@', got '%s'", xref)
	}

	// Test FindByName
	nodeIDs, err := helpers.FindByName("john")
	if err != nil {
		t.Fatalf("FindByName failed: %v", err)
	}
	if len(nodeIDs) == 0 {
		t.Error("Expected at least one result")
	}

	// Test GetAllIndividualIDs
	allIDs, err := helpers.GetAllIndividualIDs()
	if err != nil {
		t.Fatalf("GetAllIndividualIDs failed: %v", err)
	}
	if len(allIDs) != 1 {
		t.Errorf("Expected 1 individual, got %d", len(allIDs))
	}
}

