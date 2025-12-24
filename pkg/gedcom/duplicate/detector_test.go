package duplicate

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

func TestFindDuplicates_SingleFile(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())
	tree := gedcom.NewGedcomTree()

	// Create duplicate individuals
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree.AddRecord(indi1)

	indi2 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi2.FirstLine().XrefID = "@I2@"
	tree.AddRecord(indi2)

	// Create non-duplicate
	indi3 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi3.FirstLine().XrefID = "@I3@"
	tree.AddRecord(indi3)

	result, err := detector.FindDuplicates(tree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Matches) == 0 {
		t.Error("expected to find at least one duplicate match")
	}

	// Check that I1 and I2 are matched
	found := false
	for _, match := range result.Matches {
		xref1 := match.Individual1.XrefID()
		xref2 := match.Individual2.XrefID()
		if (xref1 == "@I1@" && xref2 == "@I2@") || (xref1 == "@I2@" && xref2 == "@I1@") {
			found = true
			if match.SimilarityScore < 0.8 {
				t.Errorf("expected high similarity score, got %f", match.SimilarityScore)
			}
			if match.Confidence != "high" && match.Confidence != "exact" {
				t.Errorf("expected high or exact confidence, got %s", match.Confidence)
			}
		}
	}

	if !found {
		t.Error("expected to find match between @I1@ and @I2@")
	}
}

func TestFindDuplicates_NoDuplicates(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())
	tree := gedcom.NewGedcomTree()

	// Create distinct individuals
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree.AddRecord(indi1)

	indi2 := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	indi2.FirstLine().XrefID = "@I2@"
	tree.AddRecord(indi2)

	result, err := detector.FindDuplicates(tree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With default threshold (0.60), these should not match
	// But if they do match slightly, that's okay - we just check they're not high confidence
	for _, match := range result.Matches {
		if match.SimilarityScore >= 0.85 {
			t.Errorf("unexpected high similarity match: %f", match.SimilarityScore)
		}
	}
}

func TestFindDuplicatesBetween_CrossFile(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())
	tree1 := gedcom.NewGedcomTree()
	tree2 := gedcom.NewGedcomTree()

	// Create same individual in both trees
	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi1.FirstLine().XrefID = "@I1@"
	tree1.AddRecord(indi1)

	indi2 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi2.FirstLine().XrefID = "@I2@"
	tree2.AddRecord(indi2)

	result, err := detector.FindDuplicatesBetween(tree1, tree2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Matches) == 0 {
		t.Error("expected to find at least one duplicate match")
	}

	found := false
	for _, match := range result.Matches {
		xref1 := match.Individual1.XrefID()
		xref2 := match.Individual2.XrefID()
		if (xref1 == "@I1@" && xref2 == "@I2@") || (xref1 == "@I2@" && xref2 == "@I1@") {
			found = true
			if match.SimilarityScore < 0.8 {
				t.Errorf("expected high similarity score, got %f", match.SimilarityScore)
			}
		}
	}

	if !found {
		t.Error("expected to find match between @I1@ and @I2@")
	}
}

func TestFindMatches_SingleIndividual(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())
	tree := gedcom.NewGedcomTree()

	// Create target individual
	target := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	target.FirstLine().XrefID = "@I1@"
	tree.AddRecord(target)

	// Create potential match
	match1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	match1.FirstLine().XrefID = "@I2@"
	tree.AddRecord(match1)

	// Create non-match
	nonMatch := createTestIndividual("Jane /Smith/", "Jane", "Smith", "1850", "Boston")
	nonMatch.FirstLine().XrefID = "@I3@"
	tree.AddRecord(nonMatch)

	matches, err := detector.FindMatches(target, tree)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) == 0 {
		t.Error("expected to find at least one match")
	}

	// Check that @I2@ is in matches
	found := false
	for _, match := range matches {
		xref := match.Individual2.XrefID()
		if xref == "@I2@" {
			found = true
			if match.SimilarityScore < 0.8 {
				t.Errorf("expected high similarity score, got %f", match.SimilarityScore)
			}
		}
		// Should not match self
		if match.Individual2.XrefID() == "@I1@" {
			t.Error("should not match self")
		}
	}

	if !found {
		t.Error("expected to find match with @I2@")
	}
}

func TestCompare_TwoIndividuals(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())

	indi1 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")
	indi2 := createTestIndividual("John /Doe/", "John", "Doe", "1800", "New York")

	score, err := detector.Compare(indi1, indi2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if score < 0.8 {
		t.Errorf("expected high similarity score, got %f", score)
	}
}

func TestDetermineConfidence(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())

	tests := []struct {
		score    float64
		expected string
	}{
		{0.96, "exact"},
		{0.90, "high"},
		{0.75, "medium"},
		{0.65, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			confidence := detector.determineConfidence(tt.score)
			if confidence != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, confidence)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MinThreshold != 0.60 {
		t.Errorf("expected MinThreshold 0.60, got %f", config.MinThreshold)
	}
	if config.HighConfidenceThreshold != 0.85 {
		t.Errorf("expected HighConfidenceThreshold 0.85, got %f", config.HighConfidenceThreshold)
	}
	if config.ExactMatchThreshold != 0.95 {
		t.Errorf("expected ExactMatchThreshold 0.95, got %f", config.ExactMatchThreshold)
	}
	if config.NameWeight != 0.40 {
		t.Errorf("expected NameWeight 0.40, got %f", config.NameWeight)
	}
	if config.DateWeight != 0.30 {
		t.Errorf("expected DateWeight 0.30, got %f", config.DateWeight)
	}
}
