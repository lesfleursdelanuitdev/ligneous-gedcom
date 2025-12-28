package query

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
)

// TestAncestorQuery_Count_EdgeCases tests Count with edge cases
func TestAncestorQuery_Count_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Test Count on individuals with ancestors
	for i := 0; i < 10 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		count, err := iq.Ancestors().Count()
		if err != nil {
			t.Errorf("Failed to count ancestors for %s: %v", xref, err)
		}
		_ = count // Just verify it doesn't panic
	}

	// Test Count with MaxGenerations
	for i := 0; i < 5 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		count, err := iq.Ancestors().MaxGenerations(2).Count()
		if err != nil {
			t.Errorf("Failed to count ancestors with max generations for %s: %v", xref, err)
		}
		_ = count
	}
}

// TestAncestorQuery_Exists_EdgeCases tests Exists with edge cases
func TestAncestorQuery_Exists_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Test Exists on individuals with ancestors
	for i := 0; i < 10 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		exists, err := iq.Ancestors().Exists()
		if err != nil {
			t.Errorf("Failed to check ancestors existence for %s: %v", xref, err)
		}
		_ = exists // Just verify it doesn't panic
	}

	// Test Exists with MaxGenerations
	for i := 0; i < 5 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		exists, err := iq.Ancestors().MaxGenerations(1).Exists()
		if err != nil {
			t.Errorf("Failed to check ancestors existence with max generations for %s: %v", xref, err)
		}
		_ = exists
	}
}

// TestDescendantQuery_Count_EdgeCases tests Count with edge cases
func TestDescendantQuery_Count_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Test Count on individuals with descendants
	for i := 0; i < 10 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		count, err := iq.Descendants().Count()
		if err != nil {
			t.Errorf("Failed to count descendants for %s: %v", xref, err)
		}
		_ = count // Just verify it doesn't panic
	}

	// Test Count with MaxGenerations
	for i := 0; i < 5 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		count, err := iq.Descendants().MaxGenerations(2).Count()
		if err != nil {
			t.Errorf("Failed to count descendants with max generations for %s: %v", xref, err)
		}
		_ = count
	}
}

// TestDescendantQuery_Exists_EdgeCases tests Exists with edge cases
func TestDescendantQuery_Exists_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Test Exists on individuals with descendants
	for i := 0; i < 10 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		exists, err := iq.Descendants().Exists()
		if err != nil {
			t.Errorf("Failed to check descendants existence for %s: %v", xref, err)
		}
		_ = exists // Just verify it doesn't panic
	}

	// Test Exists with MaxGenerations
	for i := 0; i < 5 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		exists, err := iq.Descendants().MaxGenerations(1).Exists()
		if err != nil {
			t.Errorf("Failed to check descendants existence with max generations for %s: %v", xref, err)
		}
		_ = exists
	}
}

// TestAncestorQuery_ExecuteWithPaths_EdgeCases tests ExecuteWithPaths with edge cases
func TestAncestorQuery_ExecuteWithPaths_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Test ExecuteWithPaths on individuals with ancestors
	for i := 0; i < 10 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		ancestorsWithPaths, err := iq.Ancestors().ExecuteWithPaths()
		if err != nil {
			t.Errorf("Failed to get ancestors with paths for %s: %v", xref, err)
		}
		_ = ancestorsWithPaths // Just verify it doesn't panic
	}

	// Test ExecuteWithPaths with MaxGenerations
	for i := 0; i < 5 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		ancestorsWithPaths, err := iq.Ancestors().MaxGenerations(2).ExecuteWithPaths()
		if err != nil {
			t.Errorf("Failed to get ancestors with paths (max gen) for %s: %v", xref, err)
		}
		_ = ancestorsWithPaths
	}
}

// TestAncestorQuery_findAncestorsWithDepth tests findAncestorsWithDepth
func TestAncestorQuery_findAncestorsWithDepth(t *testing.T) {
	filePath := findTestDataFile("gracis.ged")
	if filePath == "" {
		t.Skip("Test data file not found: gracis.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) == 0 {
		t.Skip("No individuals found")
	}

	// Test with various MaxGenerations to exercise findAncestorsWithDepth
	for i := 0; i < 10 && i < len(allIndividuals); i++ {
		xref := allIndividuals[i].XrefID()
		iq := qb.Individual(xref)

		// Test with depth 1
		ancestors1, err := iq.Ancestors().MaxGenerations(1).Execute()
		if err != nil {
			t.Errorf("Failed to get ancestors (depth 1) for %s: %v", xref, err)
		}
		_ = ancestors1

		// Test with depth 2
		ancestors2, err := iq.Ancestors().MaxGenerations(2).Execute()
		if err != nil {
			t.Errorf("Failed to get ancestors (depth 2) for %s: %v", xref, err)
		}
		_ = ancestors2

		// Test with depth 3
		ancestors3, err := iq.Ancestors().MaxGenerations(3).Execute()
		if err != nil {
			t.Errorf("Failed to get ancestors (depth 3) for %s: %v", xref, err)
		}
		_ = ancestors3
	}
}

// TestPathQuery_AllPaths_EdgeCases tests AllPaths with edge cases
func TestPathQuery_AllPaths_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("gracis.ged")
	if filePath == "" {
		t.Skip("Test data file not found: gracis.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) < 2 {
		t.Skip("Need at least 2 individuals for path finding")
	}

	// Test AllPaths between various pairs
	testCount := 10
	if len(allIndividuals) < testCount {
		testCount = len(allIndividuals)
	}

	for i := 0; i < testCount-1; i++ {
		fromXref := allIndividuals[i].XrefID()
		toXref := allIndividuals[i+1].XrefID()

		pathQuery := qb.Individual(fromXref).PathTo(toXref)
		allPaths, err := pathQuery.All()
		if err != nil {
			// Path might not exist, that's okay
			_ = err
		}
		_ = allPaths // Just verify it doesn't panic
	}
}

// TestPathQuery_MaxLength_EdgeCases tests MaxLength filter with edge cases
func TestPathQuery_MaxLength_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("gracis.ged")
	if filePath == "" {
		t.Skip("Test data file not found: gracis.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) < 2 {
		t.Skip("Need at least 2 individuals for path finding")
	}

	// Test MaxLength with various values
	for i := 0; i < 5 && i+1 < len(allIndividuals); i++ {
		fromXref := allIndividuals[i].XrefID()
		toXref := allIndividuals[i+1].XrefID()

		pathQuery := qb.Individual(fromXref).PathTo(toXref)

		// Test with MaxLength 1
		paths1, err := pathQuery.MaxLength(1).All()
		if err != nil {
			_ = err // Path might not exist
		}
		_ = paths1

		// Test with MaxLength 2
		paths2, err := pathQuery.MaxLength(2).All()
		if err != nil {
			_ = err
		}
		_ = paths2

		// Test with MaxLength 3
		paths3, err := pathQuery.MaxLength(3).All()
		if err != nil {
			_ = err
		}
		_ = paths3
	}
}

// TestPathQuery_IncludeMarital_IncludeBlood_EdgeCases tests path query filters with edge cases
func TestPathQuery_IncludeMarital_IncludeBlood_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("gracis.ged")
	if filePath == "" {
		t.Skip("Test data file not found: gracis.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) < 2 {
		t.Skip("Need at least 2 individuals for path finding")
	}

	// Test IncludeMarital and IncludeBlood
	for i := 0; i < 5 && i+1 < len(allIndividuals); i++ {
		fromXref := allIndividuals[i].XrefID()
		toXref := allIndividuals[i+1].XrefID()

		pathQuery := qb.Individual(fromXref).PathTo(toXref)

		// Test IncludeMarital(false)
		paths1, err := pathQuery.IncludeMarital(false).All()
		if err != nil {
			_ = err
		}
		_ = paths1

		// Test IncludeBlood(false)
		paths2, err := pathQuery.IncludeBlood(false).All()
		if err != nil {
			_ = err
		}
		_ = paths2

		// Test both false
		paths3, err := pathQuery.IncludeMarital(false).IncludeBlood(false).All()
		if err != nil {
			_ = err
		}
		_ = paths3
	}
}

// TestRelationshipQuery_EdgeCases tests relationship queries with edge cases
func TestRelationshipQuery_EdgeCases(t *testing.T) {
	filePath := findTestDataFile("gracis.ged")
	if filePath == "" {
		t.Skip("Test data file not found: gracis.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all individuals
	allIndividuals, err := qb.AllIndividuals().Execute()
	if err != nil {
		t.Fatalf("Failed to get all individuals: %v", err)
	}

	if len(allIndividuals) < 2 {
		t.Skip("Need at least 2 individuals")
	}

	// Test relationship queries between various pairs
	testCount := 15
	if len(allIndividuals) < testCount {
		testCount = len(allIndividuals)
	}

	for i := 0; i < testCount-1; i++ {
		fromXref := allIndividuals[i].XrefID()
		toXref := allIndividuals[i+1].XrefID()

		relationshipQuery := qb.Individual(fromXref).RelationshipTo(toXref)
		result, err := relationshipQuery.Execute()
		if err != nil {
			// Relationship might not exist, that's okay
			_ = err
		}
		if result != nil {
			// Test GetRelationshipType
			relType := result.GetRelationshipType()
			_ = relType

			// Test IsBloodRelation
			isBlood := result.IsBloodRelation()
			_ = isBlood

			// Test IsMaritalRelation
			isMarital := result.IsMaritalRelation()
			_ = isMarital
		}
	}
}

// TestFilterQuery_BuildCacheKey tests buildCacheKey indirectly through filters
func TestFilterQuery_BuildCacheKey(t *testing.T) {
	filePath := findTestDataFile("xavier.ged")
	if filePath == "" {
		t.Skip("Test data file not found: xavier.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Test various filter combinations to exercise buildCacheKey
	// Note: buildCacheKey is called internally, so we test it indirectly

	// Test with name filter
	results1, err := qb.Filter().ByName("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}
	_ = results1

	// Test with name exact filter
	results2, err := qb.Filter().ByNameExact("John Doe").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}
	_ = results2

	// Test with name starts filter
	results3, err := qb.Filter().ByNameStarts("John").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}
	_ = results3

	// Test with sex filter
	results4, err := qb.Filter().BySex("M").Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}
	_ = results4

	// Test with hasChildren filter
	results5, err := qb.Filter().HasChildren().Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}
	_ = results5

	// Test with hasSpouse filter
	results6, err := qb.Filter().HasSpouse().Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}
	_ = results6

	// Test with multiple filters combined
	results7, err := qb.Filter().BySex("M").HasChildren().Execute()
	if err != nil {
		t.Fatalf("Failed to execute filter: %v", err)
	}
	_ = results7
}

// TestGetAllNotes_Family tests GetAllNotes for families
func TestGetAllNotes_Family(t *testing.T) {
	filePath := findTestDataFile("royal92.ged")
	if filePath == "" {
		t.Skip("Test data file not found: royal92.ged")
	}

	// Parse and build query
	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(filePath)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	qb, err := NewQuery(tree)
	if err != nil {
		t.Fatalf("Failed to create query builder: %v", err)
	}

	// Get all families
	allFamilies, err := qb.Families().All()
	if err != nil {
		t.Fatalf("Failed to get families: %v", err)
	}

	if len(allFamilies) == 0 {
		t.Skip("No families found")
	}

	// Test GetAllNotes on first few families
	testCount := 20
	if len(allFamilies) < testCount {
		testCount = len(allFamilies)
	}

	for i := 0; i < testCount; i++ {
		xref := allFamilies[i].XrefID()
		fq := qb.Family(xref)

		notes, err := fq.GetAllNotes()
		if err != nil {
			t.Errorf("Failed to get notes for family %s: %v", xref, err)
		}
		_ = notes // Just verify it doesn't panic
	}
}

// TestGetAllNotes_Individual_Comprehensive tests GetAllNotes comprehensively
func TestGetAllNotes_Individual_Comprehensive(t *testing.T) {
	testFiles := []string{
		"xavier.ged",
		"gracis.ged",
	}

	if !testing.Short() {
		testFiles = append(testFiles, "tree1.ged", "royal92.ged")
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := findTestDataFile(filename)
			if filePath == "" {
				t.Skipf("Test data file not found: %s", filename)
			}

			// Parse and build query
			p := parser.NewHierarchicalParser()
			tree, err := p.Parse(filePath)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			qb, err := NewQuery(tree)
			if err != nil {
				t.Fatalf("Failed to create query builder: %v", err)
			}

			// Get all individuals
			allIndividuals, err := qb.AllIndividuals().Execute()
			if err != nil {
				t.Fatalf("Failed to get all individuals: %v", err)
			}

			if len(allIndividuals) == 0 {
				t.Skip("No individuals found")
			}

			// Test GetAllNotes on first 20 individuals
			testCount := 20
			if len(allIndividuals) < testCount {
				testCount = len(allIndividuals)
			}

			for i := 0; i < testCount; i++ {
				xref := allIndividuals[i].XrefID()
				iq := qb.Individual(xref)

				notes, err := iq.GetAllNotes()
				if err != nil {
					t.Errorf("Failed to get notes for %s: %v", xref, err)
				}
				_ = notes // Just verify it doesn't panic
			}
		})
	}
}

