package duplicate

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// calculateNameSimilarity calculates the similarity between two individuals' names.
func (dd *DuplicateDetector) calculateNameSimilarity(indi1, indi2 *gedcom.IndividualRecord) float64 {
	name1 := indi1.GetName()
	name2 := indi2.GetName()

	if name1 == "" && name2 == "" {
		return 0.0
	}
	if name1 == "" || name2 == "" {
		return 0.0 // Can't compare if one is missing
	}

	// Try exact match first
	if normalizeName(name1) == normalizeName(name2) {
		return 1.0
	}

	// Try component match (given name + surname)
	given1 := normalizeString(indi1.GetGivenName())
	given2 := normalizeString(indi2.GetGivenName())
	surname1 := normalizeString(indi1.GetSurname())
	surname2 := normalizeString(indi2.GetSurname())

	givenScore := 0.0
	surnameScore := 0.0

	if given1 != "" && given2 != "" {
		givenScore = stringSimilarity(given1, given2)
	}
	if surname1 != "" && surname2 != "" {
		surnameScore = stringSimilarity(surname1, surname2)
	}

	// If both components match, return high score
	if givenScore >= 0.8 && surnameScore >= 0.8 {
		return (givenScore + surnameScore) / 2.0
	}

	// Try fuzzy match on full name
	fullNameScore := stringSimilarity(normalizeName(name1), normalizeName(name2))
	if fullNameScore > (givenScore+surnameScore)/2.0 {
		return fullNameScore
	}

	// Return average of component scores
	if givenScore > 0.0 || surnameScore > 0.0 {
		if givenScore > 0.0 && surnameScore > 0.0 {
			return (givenScore + surnameScore) / 2.0
		}
		if givenScore > 0.0 {
			return givenScore * 0.7 // Penalize missing surname
		}
		return surnameScore * 0.7 // Penalize missing given name
	}

	// Try partial match
	if strings.Contains(normalizeName(name1), normalizeName(name2)) ||
		strings.Contains(normalizeName(name2), normalizeName(name1)) {
		return 0.6
	}

	return 0.0
}

// calculateDateSimilarity calculates the similarity between two individuals' dates.
func (dd *DuplicateDetector) calculateDateSimilarity(indi1, indi2 *gedcom.IndividualRecord) float64 {
	birthDate1 := indi1.GetBirthDate()
	birthDate2 := indi2.GetBirthDate()

	if birthDate1 == "" && birthDate2 == "" {
		return 0.0
	}
	if birthDate1 == "" || birthDate2 == "" {
		return 0.0 // Can't compare if one is missing
	}

	// Try to parse dates
	date1, err1 := indi1.GetBirthDateParsed()
	date2, err2 := indi2.GetBirthDateParsed()

	// If both parse successfully, use range-based comparison
	if err1 == nil && err2 == nil && date1 != nil && date2 != nil {
		return dateSimilarity(date1, date2, dd.config.DateTolerance)
	}

	// Fallback to string-based comparison for unparseable dates
	return dateStringSimilarity(birthDate1, birthDate2)
}

// calculatePlaceSimilarity calculates the similarity between two individuals' places.
func (dd *DuplicateDetector) calculatePlaceSimilarity(indi1, indi2 *gedcom.IndividualRecord) float64 {
	birthPlace1 := indi1.GetBirthPlace()
	birthPlace2 := indi2.GetBirthPlace()

	if birthPlace1 == "" && birthPlace2 == "" {
		return 0.0
	}
	if birthPlace1 == "" || birthPlace2 == "" {
		return 0.0 // Can't compare if one is missing
	}

	// Try exact match
	normalized1 := normalizePlace(birthPlace1)
	normalized2 := normalizePlace(birthPlace2)
	if normalized1 == normalized2 {
		return 1.0
	}

	// Try parsed place comparison
	place1, err1 := indi1.GetBirthPlaceParsed()
	place2, err2 := indi2.GetBirthPlaceParsed()

	if err1 == nil && err2 == nil {
		return placeComponentSimilarity(place1, place2)
	}

	// Fallback to string similarity
	return stringSimilarity(normalized1, normalized2)
}

// calculateSexSimilarity calculates the similarity between two individuals' sex values.
func (dd *DuplicateDetector) calculateSexSimilarity(indi1, indi2 *gedcom.IndividualRecord) float64 {
	sex1 := strings.ToUpper(strings.TrimSpace(indi1.GetSex()))
	sex2 := strings.ToUpper(strings.TrimSpace(indi2.GetSex()))

	if sex1 == "" && sex2 == "" {
		return 0.5 // Both unknown - neutral
	}
	if sex1 == "" || sex2 == "" {
		return 0.5 // One unknown - neutral
	}

	// Unknown (U) matches anything with 0.5
	if sex1 == "U" && sex2 == "U" {
		return 0.5 // Both unknown - neutral
	}
	if sex1 == "U" || sex2 == "U" {
		return 0.5 // One unknown - neutral
	}

	if sex1 == sex2 {
		return 1.0
	}

	// Mismatch is strong negative indicator
	if (sex1 == "M" && sex2 == "F") || (sex1 == "F" && sex2 == "M") {
		return 0.0
	}

	return 0.0
}

// normalizeName normalizes a name string for comparison.
func normalizeName(name string) string {
	// Remove GEDCOM slashes
	name = strings.ReplaceAll(name, "/", "")
	// Remove extra whitespace
	name = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(name), " ")
	// Convert to lowercase
	return strings.ToLower(name)
}

// normalizeString normalizes a string for comparison.
func normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// normalizePlace normalizes a place string for comparison.
func normalizePlace(place string) string {
	return normalizeString(place)
}

// stringSimilarity calculates similarity between two strings using Levenshtein distance.
func stringSimilarity(s1, s2 string) float64 {
	s1 = normalizeString(s1)
	s2 = normalizeString(s2)

	if s1 == s2 {
		return 1.0
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	// Convert distance to similarity (0.0 to 1.0)
	similarity := 1.0 - (float64(distance) / float64(maxLen))

	// Clamp to reasonable range for fuzzy matches
	if similarity < 0.0 {
		return 0.0
	}
	if similarity > 0.9 {
		return 0.9 // Cap fuzzy matches at 0.9 (exact match is 1.0)
	}

	return similarity
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	column := make([]int, len(r1)+1)

	for y := 1; y <= len(r1); y++ {
		column[y] = y
	}

	for x := 1; x <= len(r2); x++ {
		column[0] = x
		lastDiag := x - 1
		for y := 1; y <= len(r1); y++ {
			oldDiag := column[y]
			cost := 0
			if r1[y-1] != r2[x-1] {
				cost = 1
			}
			column[y] = min(column[y]+1, column[y-1]+1, lastDiag+cost)
			lastDiag = oldDiag
		}
	}

	return column[len(r1)]
}

// min returns the minimum of three integers.
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// dateSimilarity calculates similarity between two parsed dates.
// Enhanced to handle date ranges (ABT, BEF, AFT, BETWEEN).
func dateSimilarity(date1, date2 *gedcom.GedcomDate, tolerance int) float64 {
	// Get year ranges for both dates
	range1 := getDateRange(date1, tolerance)
	range2 := getDateRange(date2, tolerance)

	if range1.start == 0 && range1.end == 0 {
		return 0.0
	}
	if range2.start == 0 && range2.end == 0 {
		return 0.0
	}

	// Calculate overlap between ranges
	overlap := calculateRangeOverlap(range1, range2)
	if overlap <= 0 {
		// No overlap - calculate distance-based similarity
		// For exact dates, use year difference
		if range1.start == range1.end && range2.start == range2.end {
			diff := abs(range1.start - range2.start)
			switch {
			case diff == 0:
				return 1.0
			case diff <= 1:
				return 0.9
			case diff <= 2:
				return 0.8
			case diff <= 5:
				return 0.7
			case diff <= 10:
				return 0.5
			default:
				return 0.0
			}
		}
		return 0.0
	}

	// Calculate similarity based on overlap percentage
	range1Size := range1.end - range1.start + 1
	range2Size := range2.end - range2.start + 1
	avgSize := float64(range1Size+range2Size) / 2.0

	if avgSize == 0 {
		return 0.0
	}

	overlapRatio := float64(overlap) / avgSize

	// Convert overlap ratio to similarity score
	if overlapRatio >= 1.0 {
		return 1.0 // Exact or near-exact match
	} else if overlapRatio >= 0.8 {
		return 0.9 // High overlap
	} else if overlapRatio >= 0.6 {
		return 0.8 // Good overlap
	} else if overlapRatio >= 0.4 {
		return 0.7 // Moderate overlap
	} else if overlapRatio >= 0.2 {
		return 0.5 // Low overlap
	} else {
		return 0.3 // Minimal overlap
	}
}

// dateRange represents a year range for date comparison.
type dateRange struct {
	start int
	end   int
}

// getDateRange extracts a year range from a GedcomDate, handling imprecise dates.
func getDateRange(date *gedcom.GedcomDate, tolerance int) dateRange {
	if date == nil {
		return dateRange{0, 0}
	}

	// Handle different date types
	switch date.Type {
	case gedcom.DateTypeExact, gedcom.DateTypeUnknown:
		year := date.Year
		if year == 0 {
			year = date.StartYear
		}
		if year == 0 {
			// Try to extract from original string
			year = extractYear(date.Original)
		}
		if year == 0 {
			return dateRange{0, 0}
		}
		return dateRange{year, year}

	case gedcom.DateTypeAbout:
		year := date.Year
		if year == 0 {
			year = date.StartYear
		}
		if year == 0 {
			return dateRange{0, 0}
		}
		// ABT dates: ±tolerance years
		return dateRange{year - tolerance, year + tolerance}

	case gedcom.DateTypeBefore:
		year := date.Year
		if year == 0 {
			year = date.StartYear
		}
		if year == 0 {
			// Try to extract from original string
			year = extractYear(date.Original)
		}
		if year == 0 {
			return dateRange{0, 0}
		}
		// BEF dates: up to the year (with some tolerance before)
		return dateRange{year - tolerance*2, year}

	case gedcom.DateTypeAfter:
		year := date.Year
		if year == 0 {
			year = date.StartYear
		}
		if year == 0 {
			// Try to extract from original string
			year = extractYear(date.Original)
		}
		if year == 0 {
			return dateRange{0, 0}
		}
		// AFT dates: from the year (with some tolerance after)
		return dateRange{year, year + tolerance*2}

	case gedcom.DateTypeBetween, gedcom.DateTypeFromTo:
		startYear := date.StartYear
		endYear := date.EndYear
		if startYear == 0 {
			startYear = date.Year
		}
		if endYear == 0 {
			endYear = date.Year
		}
		if startYear == 0 || endYear == 0 {
			return dateRange{0, 0}
		}
		return dateRange{startYear, endYear}

	case gedcom.DateTypeFrom:
		startYear := date.StartYear
		if startYear == 0 {
			startYear = date.Year
		}
		if startYear == 0 {
			return dateRange{0, 0}
		}
		// FROM dates: from start year to some reasonable future
		return dateRange{startYear, startYear + 50}

	case gedcom.DateTypeTo:
		endYear := date.EndYear
		if endYear == 0 {
			endYear = date.Year
		}
		if endYear == 0 {
			return dateRange{0, 0}
		}
		// TO dates: from some reasonable past to end year
		return dateRange{endYear - 50, endYear}

	default:
		// Fallback to year if available
		year := date.Year
		if year == 0 {
			year = date.StartYear
		}
		if year == 0 {
			return dateRange{0, 0}
		}
		return dateRange{year, year}
	}
}

// calculateRangeOverlap calculates the overlap between two date ranges.
func calculateRangeOverlap(r1, r2 dateRange) int {
	// Find the overlap
	overlapStart := r1.start
	if r2.start > overlapStart {
		overlapStart = r2.start
	}

	overlapEnd := r1.end
	if r2.end < overlapEnd {
		overlapEnd = r2.end
	}

	// Calculate overlap size
	if overlapStart > overlapEnd {
		return 0
	}

	return overlapEnd - overlapStart + 1
}

// dateStringSimilarity calculates similarity between date strings (fallback).
func dateStringSimilarity(date1, date2 string) float64 {
	normalized1 := normalizeString(date1)
	normalized2 := normalizeString(date2)

	if normalized1 == normalized2 {
		return 1.0
	}

	// Extract years
	year1 := extractYear(date1)
	year2 := extractYear(date2)

	if year1 > 0 && year2 > 0 {
		diff := abs(year1 - year2)
		switch {
		case diff == 0:
			return 1.0
		case diff <= 1:
			return 0.9
		case diff <= 2:
			return 0.8
		case diff <= 5:
			return 0.7
		case diff <= 10:
			return 0.5
		default:
			return 0.0
		}
	}

	// Fallback to string similarity
	return stringSimilarity(normalized1, normalized2) * 0.5
}

// extractYear extracts a year from a date string.
func extractYear(dateStr string) int {
	// Try to find 4-digit year
	re := regexp.MustCompile(`\b(\d{4})\b`)
	matches := re.FindStringSubmatch(dateStr)
	if len(matches) > 1 {
		year, err := strconv.Atoi(matches[1])
		if err == nil && year > 1000 && year < 3000 {
			return year
		}
	}
	return 0
}

// placeComponentSimilarity calculates similarity between parsed places.
func placeComponentSimilarity(place1, place2 *gedcom.GedcomPlace) float64 {
	if place1 == nil || place2 == nil {
		return 0.0
	}

	// Compare components
	scores := make([]float64, 0)

	// City match
	if place1.City != "" && place2.City != "" {
		if normalizeString(place1.City) == normalizeString(place2.City) {
			scores = append(scores, 1.0)
		} else {
			scores = append(scores, stringSimilarity(place1.City, place2.City)*0.7)
		}
	}

	// State match
	if place1.State != "" && place2.State != "" {
		if normalizeString(place1.State) == normalizeString(place2.State) {
			scores = append(scores, 1.0)
		} else {
			// Check for abbreviations (NY vs New York)
			if isAbbreviation(place1.State, place2.State) {
				scores = append(scores, 0.9)
			} else {
				scores = append(scores, stringSimilarity(place1.State, place2.State)*0.7)
			}
		}
	}

	// Country match
	if place1.Country != "" && place2.Country != "" {
		if normalizeString(place1.Country) == normalizeString(place2.Country) {
			scores = append(scores, 1.0)
		} else {
			scores = append(scores, stringSimilarity(place1.Country, place2.Country)*0.7)
		}
	}

	if len(scores) == 0 {
		return 0.0
	}

	// Average the scores
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

// isAbbreviation checks if two strings might be the same (one abbreviated).
func isAbbreviation(s1, s2 string) bool {
	s1 = normalizeString(s1)
	s2 = normalizeString(s2)

	// Check if one is abbreviation of the other
	if len(s1) == 2 && len(s2) > 2 {
		return strings.HasPrefix(s2, s1)
	}
	if len(s2) == 2 && len(s1) > 2 {
		return strings.HasPrefix(s1, s2)
	}

	return false
}
