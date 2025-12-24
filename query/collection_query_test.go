package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
)

func TestFamilyCollectionQuery_All(t *testing.T) {
	// Create a simple test tree
	tree := CreateTestTree()
	
	// Add a family using test helper
	AddTestFamily(tree, "@F1@", "@I1@", "@I2@", []string{})

	// Build query
	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test All()
	families, err := q.Families().All()
	if err != nil {
		t.Fatalf("Families().All() failed: %v", err)
	}

	if len(families) != 1 {
		t.Errorf("Expected 1 family, got %d", len(families))
	}
}

func TestFamilyCollectionQuery_UniqueByChildren(t *testing.T) {
	// Create a test tree with families having different numbers of children
	tree := CreateTestTree()
	
	// Family 1: 2 children
	AddTestFamily(tree, "@F1@", "", "", []string{"@I1@", "@I2@"})

	// Family 2: 2 children (duplicate count)
	AddTestFamily(tree, "@F2@", "", "", []string{"@I3@", "@I4@"})

	// Family 3: 1 child
	AddTestFamily(tree, "@F3@", "", "", []string{"@I5@"})

	// Build query
	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Unique by children
	uniqueFamilies, err := q.Families().Unique().By(FamilyUniqueByChildren).Execute()
	if err != nil {
		t.Fatalf("Unique().By(Children).Execute() failed: %v", err)
	}

	// Should have 2 unique families (one with 2 children, one with 1 child)
	if len(uniqueFamilies) != 2 {
		t.Errorf("Expected 2 unique families by children count, got %d", len(uniqueFamilies))
	}
}

func TestEventCollectionQuery_All(t *testing.T) {
	// Create a test tree with an individual having events
	tree := CreateTestTree()
	
	tree.AddRecord(CreateTestIndividualWithBirth("@I1@", "John /Doe/", "1 JAN 1900", ""))

	// Build query
	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test All()
	events, err := q.Events().All()
	if err != nil {
		t.Fatalf("Events().All() failed: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected at least 1 event, got 0")
	}
}

func TestEventCollectionQuery_UniqueByType(t *testing.T) {
	// Create a test tree with multiple events of same type
	tree := CreateTestTree()
	
	// Individual 1 with BIRT
	tree.AddRecord(CreateTestIndividualWithBirth("@I1@", "John /Doe/", "1 JAN 1900", ""))

	// Individual 2 with BIRT
	tree.AddRecord(CreateTestIndividualWithBirth("@I2@", "Jane /Doe/", "2 JAN 1900", ""))

	// Build query
	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Unique by type
	uniqueEvents, err := q.Events().Unique().By(EventUniqueByType).Execute()
	if err != nil {
		t.Fatalf("Unique().By(Type).Execute() failed: %v", err)
	}

	// Should have 1 unique event type (BIRT)
	if len(uniqueEvents) != 1 {
		t.Errorf("Expected 1 unique event type, got %d", len(uniqueEvents))
	}
}

func TestNameCollectionQuery_UniqueBySurname(t *testing.T) {
	// Create a test tree with individuals having names
	tree := CreateTestTree()
	
	// Individual 1: John Doe
	AddTestIndividual(tree, "@I1@", "John /Doe/")

	// Individual 2: Jane Doe (same surname)
	AddTestIndividual(tree, "@I2@", "Jane /Doe/")

	// Individual 3: Bob Smith (different surname)
	AddTestIndividual(tree, "@I3@", "Bob /Smith/")

	// Build query
	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Unique by surname
	uniqueSurnames, err := q.Names().Unique().By(NameUniqueBySurname).Execute()
	if err != nil {
		t.Fatalf("Unique().By(Surname).Execute() failed: %v", err)
	}

	// Should have 2 unique surnames (Doe, Smith)
	if len(uniqueSurnames) != 2 {
		t.Errorf("Expected 2 unique surnames, got %d", len(uniqueSurnames))
	}
}

func TestPlaceCollectionQuery_All(t *testing.T) {
	// Create a test tree with places
	tree := CreateTestTree()
	
	tree.AddRecord(CreateTestIndividualWithBirth("@I1@", "John /Doe/", "", "New York"))

	// Build query
	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test All()
	places, err := q.Places().All()
	if err != nil {
		t.Fatalf("Places().All() failed: %v", err)
	}

	if len(places) == 0 {
		t.Error("Expected at least 1 place, got 0")
	}
}

func TestCollectionQuery_BackwardCompatibility(t *testing.T) {
	// Test that old methods still work
	tree := CreateTestTree()
	
	AddTestFamily(tree, "@F1@", "", "", []string{})

	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test deprecated methods still work
	families, err := q.AllFamilies()
	if err != nil {
		t.Fatalf("AllFamilies() failed: %v", err)
	}
	if len(families) != 1 {
		t.Errorf("Expected 1 family, got %d", len(families))
	}

	events, err := q.AllEvents()
	if err != nil {
		t.Fatalf("AllEvents() failed: %v", err)
	}
	_ = events // Just check it doesn't error

	places, err := q.AllPlaces()
	if err != nil {
		t.Fatalf("AllPlaces() failed: %v", err)
	}
	_ = places // Just check it doesn't error

	names, err := q.UniqueNames()
	if err != nil {
		t.Fatalf("UniqueNames() failed: %v", err)
	}
	if names == nil {
		t.Error("Expected names map, got nil")
	}
}

func TestCollectionQuery_Count(t *testing.T) {
	// Test Count() methods
	tree := CreateTestTree()
	
	// Add a family
	AddTestFamily(tree, "@F1@", "", "", []string{})

	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Count()
	count, err := q.Families().Count()
	if err != nil {
		t.Fatalf("Families().Count() failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

// Test with a real GEDCOM file if available
func TestCollectionQuery_WithRealFile(t *testing.T) {
	// This test requires a test GEDCOM file
	// Skip if not available
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse("testdata/example.ged")
	if err != nil {
		t.Skip("Test file not available, skipping")
	}

	q, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test all collection queries
	families, _ := q.Families().All()
	events, _ := q.Events().All()
	names, _ := q.Names().All()
	places, _ := q.Places().All()

	t.Logf("Found %d families, %d events, %d names, %d places", 
		len(families), len(events), len(names), len(places))
}

