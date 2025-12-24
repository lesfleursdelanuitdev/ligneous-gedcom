package duplicate

import (
	"strings"
	"unicode"
)

// Soundex implements the Soundex algorithm for phonetic name matching.
// Soundex is a phonetic algorithm for indexing names by sound.
// Returns a 4-character code: first letter + 3 digits.
func Soundex(s string) string {
	if s == "" {
		return ""
	}

	// Convert to uppercase and remove non-letters
	s = strings.ToUpper(s)
	var letters []rune
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters = append(letters, r)
		}
	}

	if len(letters) == 0 {
		return ""
	}

	// First letter
	first := letters[0]
	result := []rune{first}

	// Soundex mapping
	soundexMap := map[rune]int{
		'B': 1, 'F': 1, 'P': 1, 'V': 1,
		'C': 2, 'G': 2, 'J': 2, 'K': 2, 'Q': 2, 'S': 2, 'X': 2, 'Z': 2,
		'D': 3, 'T': 3,
		'L': 4,
		'M': 5, 'N': 5,
		'R': 6,
	}

	// Process remaining letters
	prevCode := 0
	for i := 1; i < len(letters); i++ {
		code, ok := soundexMap[letters[i]]
		if !ok {
			// H, W, and vowels are ignored
			continue
		}

		// Skip if same code as previous
		if code == prevCode {
			continue
		}

		result = append(result, rune('0'+code))
		prevCode = code

		// Stop when we have 3 digits
		if len(result) == 4 {
			break
		}
	}

	// Pad with zeros if needed
	for len(result) < 4 {
		result = append(result, '0')
	}

	return string(result)
}

// phoneticSimilarity calculates similarity based on Soundex codes.
func phoneticSimilarity(s1, s2 string) float64 {
	code1 := Soundex(s1)
	code2 := Soundex(s2)

	if code1 == "" || code2 == "" {
		return 0.0
	}

	if code1 == code2 {
		return 0.9 // High similarity for phonetic match
	}

	// Check if first letter matches
	if code1[0] == code2[0] {
		// Count matching digits
		matches := 0
		for i := 1; i < 4; i++ {
			if code1[i] == code2[i] {
				matches++
			}
		}
		// Partial phonetic match
		return 0.5 + float64(matches)*0.1 // 0.6-0.8 for partial matches
	}

	return 0.0
}
