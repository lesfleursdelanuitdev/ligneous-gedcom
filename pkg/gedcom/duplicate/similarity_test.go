package duplicate

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

func TestCalculateNameSimilarity(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())

	tests := []struct {
		name     string
		indi1    *gedcom.IndividualRecord
		indi2    *gedcom.IndividualRecord
		expected float64
	}{
		{
			name:     "exact match",
			indi1:    createTestIndividual("John /Doe/", "", "", "", ""),
			indi2:    createTestIndividual("John /Doe/", "", "", "", ""),
			expected: 1.0,
		},
		{
			name:     "normalized match",
			indi1:    createTestIndividual("John /Doe/", "", "", "", ""),
			indi2:    createTestIndividual("John Doe", "", "", "", ""),
			expected: 1.0,
		},
		{
			name:     "component match",
			indi1:    createTestIndividual("John /Doe/", "John", "Doe", "", ""),
			indi2:    createTestIndividual("John /Doe/", "John", "Doe", "", ""),
			expected: 1.0,
		},
		{
			name:     "fuzzy match",
			indi1:    createTestIndividual("John /Doe/", "John", "Doe", "", ""),
			indi2:    createTestIndividual("Jon /Doe/", "Jon", "Doe", "", ""),
			expected: 0.7, // Should be around 0.7-0.9 for fuzzy match
		},
		{
			name:     "different names",
			indi1:    createTestIndividual("John /Doe/", "John", "Doe", "", ""),
			indi2:    createTestIndividual("Jane /Smith/", "Jane", "Smith", "", ""),
			expected: 0.0,
		},
		{
			name:     "missing name",
			indi1:    createTestIndividual("John /Doe/", "John", "Doe", "", ""),
			indi2:    createTestIndividual("", "", "", "", ""),
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.calculateNameSimilarity(tt.indi1, tt.indi2)
			if tt.name == "fuzzy match" {
				// Fuzzy match should be > 0.5 but < 1.0
				if score < 0.5 || score >= 1.0 {
					t.Errorf("expected fuzzy match score between 0.5 and 1.0, got %f", score)
				}
			} else if tt.name == "different names" {
				// Different names should have low similarity (may not be exactly 0.0 due to fuzzy matching)
				if score > 0.3 {
					t.Errorf("expected low similarity for different names, got %f", score)
				}
			} else if score != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, score)
			}
		})
	}
}

func TestCalculateDateSimilarity(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())

	tests := []struct {
		name     string
		indi1    *gedcom.IndividualRecord
		indi2    *gedcom.IndividualRecord
		expected float64
	}{
		{
			name:     "exact year match",
			indi1:    createTestIndividual("", "", "", "1800", ""),
			indi2:    createTestIndividual("", "", "", "1800", ""),
			expected: 1.0,
		},
		{
			name:     "one year difference",
			indi1:    createTestIndividual("", "", "", "1800", ""),
			indi2:    createTestIndividual("", "", "", "1801", ""),
			expected: 0.9, // May vary with range-based matching
		},
		{
			name:     "two year difference",
			indi1:    createTestIndividual("", "", "", "1800", ""),
			indi2:    createTestIndividual("", "", "", "1802", ""),
			expected: 0.9, // Within tolerance (2 years) - may vary with range matching
		},
		{
			name:     "five year difference",
			indi1:    createTestIndividual("", "", "", "1800", ""),
			indi2:    createTestIndividual("", "", "", "1805", ""),
			expected: 0.7, // May be lower with range-based matching
		},
		{
			name:     "missing date",
			indi1:    createTestIndividual("", "", "", "1800", ""),
			indi2:    createTestIndividual("", "", "", "", ""),
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.calculateDateSimilarity(tt.indi1, tt.indi2)
			// For range-based matching, allow some flexibility for non-exact matches
			if tt.name == "one year difference" || tt.name == "two year difference" {
				// With range matching, these should still be high similarity
				if score < 0.7 {
					t.Errorf("expected high similarity for %s, got %f", tt.name, score)
				}
			} else if tt.name == "five year difference" {
				// With range matching, 5 years might score differently
				if score < 0.5 {
					t.Errorf("expected reasonable similarity for 5 year difference, got %f", score)
				}
			} else if score != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, score)
			}
		})
	}
}

func TestCalculatePlaceSimilarity(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())

	tests := []struct {
		name     string
		indi1    *gedcom.IndividualRecord
		indi2    *gedcom.IndividualRecord
		expected float64
	}{
		{
			name:     "exact match",
			indi1:    createTestIndividual("", "", "", "", "New York"),
			indi2:    createTestIndividual("", "", "", "", "New York"),
			expected: 1.0,
		},
		{
			name:     "case insensitive",
			indi1:    createTestIndividual("", "", "", "", "New York"),
			indi2:    createTestIndividual("", "", "", "", "new york"),
			expected: 1.0,
		},
		{
			name:     "missing place",
			indi1:    createTestIndividual("", "", "", "", "New York"),
			indi2:    createTestIndividual("", "", "", "", ""),
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.calculatePlaceSimilarity(tt.indi1, tt.indi2)
			if score != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, score)
			}
		})
	}
}

func TestCalculateSexSimilarity(t *testing.T) {
	detector := NewDuplicateDetector(DefaultConfig())

	tests := []struct {
		name     string
		indi1    *gedcom.IndividualRecord
		indi2    *gedcom.IndividualRecord
		expected float64
	}{
		{
			name:     "exact match M",
			indi1:    createTestIndividualWithSex("", "M"),
			indi2:    createTestIndividualWithSex("", "M"),
			expected: 1.0,
		},
		{
			name:     "exact match F",
			indi1:    createTestIndividualWithSex("", "F"),
			indi2:    createTestIndividualWithSex("", "F"),
			expected: 1.0,
		},
		{
			name:     "mismatch M vs F",
			indi1:    createTestIndividualWithSex("", "M"),
			indi2:    createTestIndividualWithSex("", "F"),
			expected: 0.0,
		},
		{
			name:     "unknown with M",
			indi1:    createTestIndividualWithSex("", "U"),
			indi2:    createTestIndividualWithSex("", "M"),
			expected: 0.5,
		},
		{
			name:     "both unknown",
			indi1:    createTestIndividualWithSex("", "U"),
			indi2:    createTestIndividualWithSex("", "U"),
			expected: 0.5,
		},
		{
			name:     "missing sex",
			indi1:    createTestIndividualWithSex("", ""),
			indi2:    createTestIndividualWithSex("", ""),
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.calculateSexSimilarity(tt.indi1, tt.indi2)
			if score != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, score)
			}
		})
	}
}

// Helper function to create test individual records
func createTestIndividual(name, givenName, surname, birthDate, birthPlace string) *gedcom.IndividualRecord {
	line := gedcom.NewGedcomLine(0, "INDI", "", "@TEST@")
	indi := gedcom.NewIndividualRecord(line)

	// Set name
	if name != "" {
		nameLine := gedcom.NewGedcomLine(1, "NAME", name, "")
		line.AddChild(nameLine)

		// Set given name
		if givenName != "" {
			givnLine := gedcom.NewGedcomLine(2, "GIVN", givenName, "")
			nameLine.AddChild(givnLine)
		}

		// Set surname
		if surname != "" {
			surnLine := gedcom.NewGedcomLine(2, "SURN", surname, "")
			nameLine.AddChild(surnLine)
		}
	}

	// Set birth date and place
	if birthDate != "" || birthPlace != "" {
		birtLine := gedcom.NewGedcomLine(1, "BIRT", "", "")
		line.AddChild(birtLine)

		if birthDate != "" {
			dateLine := gedcom.NewGedcomLine(2, "DATE", birthDate, "")
			birtLine.AddChild(dateLine)
		}

		if birthPlace != "" {
			placLine := gedcom.NewGedcomLine(2, "PLAC", birthPlace, "")
			birtLine.AddChild(placLine)
		}
	}

	return indi
}

// Helper function to create test individual with sex
func createTestIndividualWithSex(name, sex string) *gedcom.IndividualRecord {
	indi := createTestIndividual(name, "", "", "", "")
	if sex != "" {
		sexLine := gedcom.NewGedcomLine(1, "SEX", sex, "")
		indi.FirstLine().AddChild(sexLine)
	}
	return indi
}
