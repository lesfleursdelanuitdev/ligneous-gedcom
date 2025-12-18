package duplicate

import (
	"testing"
)

func TestSoundex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Smith", "S530"},
		{"Smyth", "S530"},
		{"Smythe", "S530"},
		{"John", "J500"},
		{"Jon", "J500"},
		{"Mary", "M600"},
		{"Marie", "M600"},
		{"Robert", "R163"},
		{"Rupert", "R163"},
		{"", ""},
		{"A", "A000"},
		{"123", ""}, // No letters
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Soundex(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPhoneticSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected float64
	}{
		{
			name:     "exact phonetic match",
			s1:       "Smith",
			s2:       "Smyth",
			expected: 0.9,
		},
		{
			name:     "different phonetic codes",
			s1:       "Smith",
			s2:       "Jones",
			expected: 0.0,
		},
		{
			name:     "same first letter, partial match",
			s1:       "Smith",
			s2:       "Smythe",
			expected: 0.9, // Same code
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := phoneticSimilarity(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}
