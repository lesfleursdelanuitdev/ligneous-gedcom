package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestFamilyCollectionQuery_AllUniquenessCriteria tests all uniqueness criteria for families
func TestFamilyCollectionQuery_AllUniquenessCriteria(t *testing.T) {
	tree := CreateTestTree()

	// Create families with various characteristics
	// Family 1: Husband @I1@, Wife @I2@, 2 children
	fam1 := CreateTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@", "@I4@"})
	tree.AddRecord(fam1)

	// Family 2: Same husband, different wife, 1 child
	fam2 := CreateTestFamily("@F2@", "@I1@", "@I5@", []string{"@I6@"})
	tree.AddRecord(fam2)

	// Family 3: Different husband, same wife, 2 children (same count as F1)
	fam3 := CreateTestFamily("@F3@", "@I7@", "@I2@", []string{"@I8@", "@I9@"})
	tree.AddRecord(fam3)

	// Family 4: Same husband+wife combination as F1 (for husband_wife test)
	fam4 := CreateTestFamily("@F4@", "@I1@", "@I2@", []string{"@I10@"})
	tree.AddRecord(fam4)

	// Family 5: Marriage date
	fam5Line := types.NewGedcomLine(0, "FAM", "", "@F5@")
	marrLine := types.NewGedcomLine(1, "MARR", "", "")
	marrLine.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	fam5Line.AddChild(marrLine)
	fam5 := types.NewFamilyRecord(fam5Line)
	tree.AddRecord(fam5)

	// Family 6: Same marriage date
	fam6Line := types.NewGedcomLine(0, "FAM", "", "@F6@")
	marrLine2 := types.NewGedcomLine(1, "MARR", "", "")
	marrLine2.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1900", ""))
	fam6Line.AddChild(marrLine2)
	fam6 := types.NewFamilyRecord(fam6Line)
	tree.AddRecord(fam6)

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	tests := []struct {
		name     string
		uniqueBy FamilyUniqueBy
		expected int // Expected number of unique families
	}{
		{"ByXref", FamilyUniqueByXref, 6}, // All families are unique
		{"ByHusband", FamilyUniqueByHusband, 2}, // @I1@ (F1,F2,F4), @I7@ (F3) - F5,F6 have no husband (skipped)
		{"ByWife", FamilyUniqueByWife, 2}, // @I2@ (F1,F3,F4), @I5@ (F2) - F5,F6 have no wife (skipped)
		{"ByChildren", FamilyUniqueByChildren, 3}, // 2 children (F1,F3), 1 child (F2,F4), 0 children (F5,F6) - all included
		{"ByMarriageDate", FamilyUniqueByMarriageDate, 1}, // F5, F6 share date "1 JAN 1900" - others have no date (skipped)
		{"ByHusbandWife", FamilyUniqueByHusbandWife, 4}, // @I1@+@I2@ (F1,F4), @I1@+@I5@ (F2), @I7@+@I2@ (F3), empty (F5,F6 skipped)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := q.Families().Unique().By(tt.uniqueBy).Execute()
			if err != nil {
				t.Fatalf("Execute() failed: %v", err)
			}
			if len(results) != tt.expected {
				t.Errorf("Expected %d unique families, got %d", tt.expected, len(results))
			}
		})
	}
}

// TestFamilyCollectionQuery_Filter tests family filtering
func TestFamilyCollectionQuery_Filter(t *testing.T) {
	tree := CreateTestTree()

	// Create families with different numbers of children
	fam1 := CreateTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@", "@I4@", "@I5@"}) // 3 children
	tree.AddRecord(fam1)

	fam2 := CreateTestFamily("@F2@", "@I6@", "@I7@", []string{"@I8@"}) // 1 child
	tree.AddRecord(fam2)

	fam3 := CreateTestFamily("@F3@", "@I9@", "@I10@", []string{}) // 0 children
	tree.AddRecord(fam3)

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Filter families with 2+ children
	results, err := q.Families().
		Filter(func(fam *types.FamilyRecord) bool {
			return len(fam.GetChildren()) >= 2
		}).
		All()

	if err != nil {
		t.Fatalf("All() failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 family with 2+ children, got %d", len(results))
	}

	if results[0].XrefID() != "@F1@" {
		t.Errorf("Expected @F1@, got %s", results[0].XrefID())
	}
}

// TestEventCollectionQuery_AllUniquenessCriteria tests all uniqueness criteria for events
func TestEventCollectionQuery_AllUniquenessCriteria(t *testing.T) {
	tree := CreateTestTree()

	// Create individuals with events
	indi1 := CreateTestIndividualWithBirth("@I1@", "John /Doe/", "1 JAN 1900", "New York")
	tree.AddRecord(indi1)

	indi2 := CreateTestIndividualWithBirth("@I2@", "Jane /Doe/", "1 JAN 1900", "New York") // Same date and place
	tree.AddRecord(indi2)

	indi3 := CreateTestIndividualWithBirth("@I3@", "Bob /Smith/", "2 JAN 1900", "Boston") // Different
	tree.AddRecord(indi3)

	// Add death events
	deathLine1 := types.NewGedcomLine(1, "DEAT", "", "")
	deathLine1.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1950", ""))
	indi1Line := indi1.FirstLine()
	indi1Line.AddChild(deathLine1)

	deathLine2 := types.NewGedcomLine(1, "DEAT", "", "")
	deathLine2.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1950", "")) // Same date
	indi2Line := indi2.FirstLine()
	indi2Line.AddChild(deathLine2)

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	tests := []struct {
		name     string
		uniqueBy EventUniqueBy
		minCount int // Minimum expected unique events
	}{
		{"ByID", EventUniqueByID, 5}, // All events are unique by ID
		{"ByType", EventUniqueByType, 2}, // BIRT and DEAT
		{"ByDate", EventUniqueByDate, 3}, // 1 JAN 1900, 2 JAN 1900, 1 JAN 1950
		{"ByPlace", EventUniqueByPlace, 2}, // New York, Boston
		{"ByTypeDate", EventUniqueByTypeDate, 3}, // BIRT|1 JAN 1900 (I1,I2), BIRT|2 JAN 1900 (I3), DEAT|1 JAN 1950 (I1,I2)
		{"ByTypePlace", EventUniqueByTypePlace, 3}, // BIRT in New York, BIRT in Boston, DEAT (no place)
		{"ByOwner", EventUniqueByOwner, 3}, // @I1@, @I2@, @I3@
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := q.Events().Unique().By(tt.uniqueBy).Execute()
			if err != nil {
				t.Fatalf("Execute() failed: %v", err)
			}
			if len(results) < tt.minCount {
				t.Errorf("Expected at least %d unique events, got %d", tt.minCount, len(results))
			}
		})
	}
}

// TestEventCollectionQuery_FilterByType tests event filtering by type
func TestEventCollectionQuery_FilterByType(t *testing.T) {
	tree := CreateTestTree()

	indi1 := CreateTestIndividualWithBirth("@I1@", "John /Doe/", "1 JAN 1900", "")
	tree.AddRecord(indi1)

	// Add death event
	deathLine := types.NewGedcomLine(1, "DEAT", "", "")
	deathLine.AddChild(types.NewGedcomLine(2, "DATE", "1 JAN 1950", ""))
	indi1Line := indi1.FirstLine()
	indi1Line.AddChild(deathLine)

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Filter by birth events only
	birthEvents, err := q.Events().OfType("BIRT").All()
	if err != nil {
		t.Fatalf("All() failed: %v", err)
	}

	if len(birthEvents) != 1 {
		t.Errorf("Expected 1 birth event, got %d", len(birthEvents))
	}

	if birthEvents[0].EventType != "BIRT" {
		t.Errorf("Expected event type BIRT, got %s", birthEvents[0].EventType)
	}
}

// TestNameCollectionQuery_AllUniquenessCriteria tests all uniqueness criteria for names
func TestNameCollectionQuery_AllUniquenessCriteria(t *testing.T) {
	tree := CreateTestTree()

	// Create individuals with various names
	indi1 := CreateTestIndividual("@I1@", "John /Doe/")
	tree.AddRecord(indi1)

	indi2 := CreateTestIndividual("@I2@", "John /Smith/") // Same given, different surname
	tree.AddRecord(indi2)

	indi3 := CreateTestIndividual("@I3@", "Jane /Doe/") // Different given, same surname
	tree.AddRecord(indi3)

	indi4 := CreateTestIndividual("@I4@", "John /Doe/") // Same as I1
	tree.AddRecord(indi4)

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	tests := []struct {
		name     string
		uniqueBy NameUniqueBy
		minCount int
	}{
		{"ByFullName", NameUniqueByFullName, 3}, // John /Doe/, John /Smith/, Jane /Doe/
		{"ByGiven", NameUniqueByGiven, 2},       // John, Jane
		{"BySurname", NameUniqueBySurname, 2},   // Doe, Smith
		{"ByGivenSurname", NameUniqueByGivenSurname, 3}, // John|Doe, John|Smith, Jane|Doe
		{"BySurnameGiven", NameUniqueBySurnameGiven, 3}, // Doe|John, Smith|John, Doe|Jane
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := q.Names().Unique().By(tt.uniqueBy).Execute()
			if err != nil {
				t.Fatalf("Execute() failed: %v", err)
			}
			if len(results) < tt.minCount {
				t.Errorf("Expected at least %d unique names, got %d", tt.minCount, len(results))
			}
		})
	}
}

// TestPlaceCollectionQuery_AllUniquenessCriteria tests all uniqueness criteria for places
func TestPlaceCollectionQuery_AllUniquenessCriteria(t *testing.T) {
	tree := CreateTestTree()

	// Create individuals with places
	indi1 := CreateTestIndividualWithBirth("@I1@", "John /Doe/", "1 JAN 1900", "New York, NY, USA")
	tree.AddRecord(indi1)

	indi2 := CreateTestIndividualWithBirth("@I2@", "Jane /Doe/", "1 JAN 1900", "New York, NY, USA") // Same place
	tree.AddRecord(indi2)

	indi3 := CreateTestIndividualWithBirth("@I3@", "Bob /Smith/", "1 JAN 1900", "Boston, MA, USA") // Different city
	tree.AddRecord(indi3)

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	tests := []struct {
		name     string
		uniqueBy PlaceUniqueBy
		minCount int
	}{
		{"ByFullString", PlaceUniqueByFullString, 2}, // New York, NY, USA and Boston, MA, USA
		{"ByCity", PlaceUniqueByCity, 2},             // New York, Boston
		{"ByState", PlaceUniqueByState, 2},           // NY, MA
		{"ByCountry", PlaceUniqueByCountry, 1},       // USA (all same)
		{"ByCityState", PlaceUniqueByCityState, 2},   // New York|NY, Boston|MA
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := q.Places().Unique().By(tt.uniqueBy).Execute()
			if err != nil {
				t.Fatalf("Execute() failed: %v", err)
			}
			if len(results) < tt.minCount {
				t.Errorf("Expected at least %d unique places, got %d", tt.minCount, len(results))
			}
		})
	}
}

// TestCollectionQuery_ErrorHandling tests error handling in collection queries
func TestCollectionQuery_ErrorHandling(t *testing.T) {
	// Test with empty tree
	tree := CreateTestTree()
	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// All should return empty slice, not error
	families, err := q.Families().All()
	if err != nil {
		t.Errorf("All() should not error on empty tree: %v", err)
	}
	if len(families) != 0 {
		t.Errorf("Expected empty slice, got %d families", len(families))
	}

	events, err := q.Events().All()
	if err != nil {
		t.Errorf("All() should not error on empty tree: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected empty slice, got %d events", len(events))
	}

	names, err := q.Names().All()
	if err != nil {
		t.Errorf("All() should not error on empty tree: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("Expected empty slice, got %d names", len(names))
	}

	places, err := q.Places().All()
	if err != nil {
		t.Errorf("All() should not error on empty tree: %v", err)
	}
	if len(places) != 0 {
		t.Errorf("Expected empty slice, got %d places", len(places))
	}
}

// TestCollectionQuery_CountEdgeCases tests Count() method with edge cases
func TestCollectionQuery_CountEdgeCases(t *testing.T) {
	tree := CreateTestTree()

	// Add some families
	AddTestFamily(tree, "@F1@", "@I1@", "@I2@", []string{"@I3@"})
	AddTestFamily(tree, "@F2@", "@I4@", "@I5@", []string{"@I6@", "@I7@"})

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Test Count()
	count, err := q.Families().Count()
	if err != nil {
		t.Fatalf("Count() failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test Count with Unique
	uniqueCount, err := q.Families().Unique().By(FamilyUniqueByChildren).Count()
	if err != nil {
		t.Fatalf("Count() failed: %v", err)
	}
	if uniqueCount != 2 { // One with 1 child, one with 2 children
		t.Errorf("Expected unique count 2, got %d", uniqueCount)
	}
}

// TestCollectionQuery_FilterCombinations tests combining filters with uniqueness
func TestCollectionQuery_FilterCombinations(t *testing.T) {
	tree := CreateTestTree()

	// Create families with different characteristics
	fam1 := CreateTestFamily("@F1@", "@I1@", "@I2@", []string{"@I3@", "@I4@", "@I5@"}) // 3 children
	tree.AddRecord(fam1)

	fam2 := CreateTestFamily("@F2@", "@I6@", "@I7@", []string{"@I8@"}) // 1 child
	tree.AddRecord(fam2)

	fam3 := CreateTestFamily("@F3@", "@I9@", "@I10@", []string{"@I11@", "@I12@"}) // 2 children
	tree.AddRecord(fam3)

	q, err := CreateTestQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}

	// Filter by 2+ children, then get unique by children count
	results, err := q.Families().
		Filter(func(fam *types.FamilyRecord) bool {
			return len(fam.GetChildren()) >= 2
		}).
		Unique().
		By(FamilyUniqueByChildren).
		Execute()

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Should have 2 unique families (one with 2 children, one with 3)
	if len(results) != 2 {
		t.Errorf("Expected 2 unique families, got %d", len(results))
	}
}

