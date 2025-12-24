package types

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DateType represents the type of a GEDCOM date.
type DateType string

const (
	DateTypeExact   DateType = "EXACT"
	DateTypeAbout   DateType = "ABOUT"
	DateTypeBefore  DateType = "BEFORE"
	DateTypeAfter   DateType = "AFTER"
	DateTypeBetween DateType = "BETWEEN"
	DateTypeFrom    DateType = "FROM"
	DateTypeTo      DateType = "TO"
	DateTypeFromTo  DateType = "FROM_TO"
	DateTypeUnknown DateType = "UNKNOWN"
)

// DateConstraint describes if a date is constrained by a particular range.
// This is similar to DateType but with different semantics for comparison operations.
type DateConstraint int

const (
	// DateConstraintExact - There is no constraint. The date is at the value specified.
	DateConstraintExact DateConstraint = iota

	// DateConstraintAbout - The date is approximate. There is no defined error margin
	// but it is usually proportional to how precise the date is.
	DateConstraintAbout

	// DateConstraintBefore - The real date is before the specified date value.
	DateConstraintBefore

	// DateConstraintAfter - The real date is after the specified date value.
	DateConstraintAfter
)

// DefaultMaxYearsForSimilarity is a sensible default for the Similarity function.
// Dates that are further apart than this value will return 0.0 similarity.
const DefaultMaxYearsForSimilarity = float64(3)

// Date words constants for parsing (case-insensitive)
const (
	DateWordsBetween = "Bet.|bet|between|from"
	DateWordsAnd     = "and|to|-"
	DateWordsAbout   = "Abt.|abt|about|c.|ca|ca.|cca|cca.|circa"
	DateWordsAfter   = "Aft.|aft|after"
	DateWordsBefore  = "Bef.|bef|before"
)

// Calendar represents the calendar system used.
type Calendar string

const (
	CalendarGregorian Calendar = "GREGORIAN"
	CalendarJulian    Calendar = "JULIAN"
	CalendarHebrew    Calendar = "HEBREW"
	CalendarFrench    Calendar = "FRENCH"
	CalendarUnknown   Calendar = "UNKNOWN"
)

// GedcomDate represents a parsed GEDCOM date with structured components.
type GedcomDate struct {
	Original string   // Original GEDCOM date string
	Type     DateType // EXACT, ABOUT, BEFORE, AFTER, BETWEEN, etc.
	Calendar Calendar // GREGORIAN, JULIAN, HEBREW, etc.

	// Exact date components
	Year  int
	Month int
	Day   int

	// Range date components
	StartYear  int
	StartMonth int
	StartDay   int
	EndYear    int
	EndMonth   int
	EndDay     int

	// Parsed status
	IsParsed   bool
	ParseError error
}

var (
	// Month abbreviations (case-insensitive)
	monthMap = map[string]int{
		"jan": 1, "january": 1, "feb": 2, "february": 2,
		"mar": 3, "march": 3, "apr": 4, "april": 4,
		"may": 5, "jun": 6, "june": 6,
		"jul": 7, "july": 7, "aug": 8, "august": 8,
		"sep": 9, "september": 9, "oct": 10, "october": 10,
		"nov": 11, "november": 11, "dec": 12, "december": 12,
	}

	// Date type prefixes (case-insensitive, with variations)
	dateTypePrefixes = map[string]DateType{
		"abt": DateTypeAbout, "abt.": DateTypeAbout, "about": DateTypeAbout,
		"c.": DateTypeAbout, "ca": DateTypeAbout, "ca.": DateTypeAbout,
		"cca": DateTypeAbout, "cca.": DateTypeAbout, "circa": DateTypeAbout,
		"bef": DateTypeBefore, "bef.": DateTypeBefore, "before": DateTypeBefore,
		"aft": DateTypeAfter, "aft.": DateTypeAfter, "after": DateTypeAfter,
		"bet": DateTypeBetween, "bet.": DateTypeBetween, "between": DateTypeBetween,
		"from": DateTypeFrom,
		"to": DateTypeTo,
	}

	// Patterns for date parsing (case-insensitive)
	exactDatePattern = regexp.MustCompile(`(?i)^(\d{1,2})\s+(\w+)\s+(\d{1,4})$`)
	monthYearPattern = regexp.MustCompile(`(?i)^(\w+)\s+(\d{1,4})$`)
	yearOnlyPattern  = regexp.MustCompile(`^\d{1,4}$`)
	betweenPattern   = regexp.MustCompile(`(?i)^(bet|bet\.|between|from)\s+(.+?)\s+(and|to|-)\s+(.+)$`)
	fromToPattern    = regexp.MustCompile(`(?i)^from\s+(.+?)\s+to\s+(.+)$`)
	
	// Enhanced between pattern that handles "BET X AND Y" format
	betweenPatternEnhanced = regexp.MustCompile(`(?i)^(?:bet|bet\.|between|from)\s+(.+?)\s+(?:and|to|-)\s+(.+)$`)
)

// ParseDate parses a GEDCOM date string and returns a GedcomDate.
// Supports various GEDCOM date formats:
//   - "15 JAN 1800" (exact date)
//   - "JAN 1800" (month-year)
//   - "1800" (year only)
//   - "ABT 1850" (about)
//   - "BEF 1900" (before)
//   - "AFT 1900" (after)
//   - "BET 1800 AND 1850" (between)
//   - "FROM 1800 TO 1850" (range)
func ParseDate(dateStr string) (*GedcomDate, error) {
	if dateStr == "" {
		return nil, fmt.Errorf("empty date string")
	}

	date := &GedcomDate{
		Original:   strings.TrimSpace(dateStr),
		Calendar:   CalendarGregorian, // Default to Gregorian
		IsParsed:   false,
		ParseError: nil,
	}

	// Normalize to lowercase for case-insensitive matching
	normalizedDate := strings.ToLower(strings.TrimSpace(dateStr))

	// Check for date type prefixes (case-insensitive)
	parts := strings.Fields(normalizedDate)
	if len(parts) > 0 {
		// Try exact match first
		if dateType, ok := dateTypePrefixes[parts[0]]; ok {
			date.Type = dateType
			// For BETWEEN and FROM, keep the full string for parsing
			if dateType == DateTypeBetween || dateType == DateTypeFrom {
				// Keep original for pattern matching
			} else {
				normalizedDate = strings.Join(parts[1:], " ")
			}
		}
	}

	// Parse based on type
	var err error
	switch date.Type {
	case DateTypeBetween:
		err = parseBetweenDate(date, normalizedDate)
	case DateTypeFrom:
		err = parseFromToDate(date, normalizedDate)
	default:
		err = parseSingleDate(date, normalizedDate)
	}

	if err != nil {
		date.ParseError = err
		return date, err
	}

	// If no type was set, default to EXACT
	if date.Type == "" {
		date.Type = DateTypeExact
	}

	date.IsParsed = true
	return date, nil
}

// parseSingleDate parses a single date (exact, about, before, after, or year-only).
func parseSingleDate(date *GedcomDate, dateStr string) error {
	// Normalize to lowercase for case-insensitive matching
	dateStr = strings.ToLower(strings.TrimSpace(dateStr))

	// Try exact date: "15 jan 1800" or "15 JAN 1800"
	if matches := exactDatePattern.FindStringSubmatch(dateStr); matches != nil {
		day, _ := strconv.Atoi(matches[1])
		monthStr := strings.ToLower(matches[2])
		year, _ := strconv.Atoi(matches[3])

		month, ok := monthMap[monthStr]
		if !ok {
			return fmt.Errorf("invalid month: %s", monthStr)
		}

		// Validate the date
		_, err := time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", year, month, day))
		if err != nil {
			return fmt.Errorf("invalid date: %d %s %d", day, monthStr, year)
		}

		date.Day = day
		date.Month = month
		date.Year = year
		return nil
	}

	// Try month-year: "JAN 1800" or "january 1800"
	if matches := monthYearPattern.FindStringSubmatch(dateStr); matches != nil {
		monthStr := strings.ToLower(matches[1])
		year, _ := strconv.Atoi(matches[2])

		month, ok := monthMap[monthStr]
		if !ok {
			return fmt.Errorf("invalid month: %s", monthStr)
		}

		date.Month = month
		date.Year = year
		return nil
	}

	// Try year only: "1800"
	if yearOnlyPattern.MatchString(dateStr) {
		year, _ := strconv.Atoi(dateStr)
		if year < 0 || year > 9999 {
			return fmt.Errorf("year out of range: %d", year)
		}
		date.Year = year
		return nil
	}

	return fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseBetweenDate parses a "BET X AND Y" date (case-insensitive).
func parseBetweenDate(date *GedcomDate, dateStr string) error {
	// Use enhanced pattern that captures only the dates, not the keywords
	matches := betweenPatternEnhanced.FindStringSubmatch(dateStr)
	if matches == nil {
		// Fallback to original pattern
		matches = betweenPattern.FindStringSubmatch(dateStr)
		if matches == nil {
			return fmt.Errorf("invalid BETWEEN date format: %s", dateStr)
		}
		// Original pattern: [0]=full, [1]=keyword, [2]=start, [3]=connector, [4]=end
		if len(matches) >= 5 {
			startStr := strings.TrimSpace(matches[2])
			endStr := strings.TrimSpace(matches[4])
			return parseBetweenDates(date, startStr, endStr)
		}
	}

	// Enhanced pattern: [0]=full, [1]=start, [2]=end
	startStr := strings.TrimSpace(matches[1])
	endStr := strings.TrimSpace(matches[2])
	return parseBetweenDates(date, startStr, endStr)
}

// parseBetweenDates parses the start and end dates for a BETWEEN range.
func parseBetweenDates(date *GedcomDate, startStr, endStr string) error {

	// Parse start date
	startDate := &GedcomDate{}
	if err := parseSingleDate(startDate, startStr); err != nil {
		return fmt.Errorf("invalid start date in BETWEEN: %w", err)
	}
	date.StartYear = startDate.Year
	date.StartMonth = startDate.Month
	date.StartDay = startDate.Day

	// Parse end date
	endDate := &GedcomDate{}
	if err := parseSingleDate(endDate, endStr); err != nil {
		return fmt.Errorf("invalid end date in BETWEEN: %w", err)
	}
	date.EndYear = endDate.Year
	date.EndMonth = endDate.Month
	date.EndDay = endDate.Day

	return nil
}

// parseFromToDate parses a "FROM X TO Y" date (case-insensitive).
func parseFromToDate(date *GedcomDate, dateStr string) error {
	matches := fromToPattern.FindStringSubmatch(dateStr)
	if matches == nil {
		return fmt.Errorf("invalid FROM-TO date format: %s", dateStr)
	}

	date.Type = DateTypeFromTo

	startStr := strings.TrimSpace(matches[1])
	endStr := strings.TrimSpace(matches[2])

	// Parse start date
	startDate := &GedcomDate{}
	if err := parseSingleDate(startDate, startStr); err != nil {
		return fmt.Errorf("invalid start date in FROM-TO: %w", err)
	}
	date.StartYear = startDate.Year
	date.StartMonth = startDate.Month
	date.StartDay = startDate.Day

	// Parse end date
	endDate := &GedcomDate{}
	if err := parseSingleDate(endDate, endStr); err != nil {
		return fmt.Errorf("invalid end date in FROM-TO: %w", err)
	}
	date.EndYear = endDate.Year
	date.EndMonth = endDate.Month
	date.EndDay = endDate.Day

	return nil
}

// IsValid returns true if the date was successfully parsed.
func (gd *GedcomDate) IsValid() bool {
	return gd.IsParsed && gd.ParseError == nil
}

// IsRange returns true if this is a range date (BETWEEN or FROM-TO).
func (gd *GedcomDate) IsRange() bool {
	return gd.Type == DateTypeBetween || gd.Type == DateTypeFromTo
}

// ToTime converts the date to a time.Time.
// For range dates, returns the start date.
// Returns error if date is invalid or cannot be converted.
func (gd *GedcomDate) ToTime() (time.Time, error) {
	if !gd.IsValid() {
		return time.Time{}, fmt.Errorf("invalid date: %v", gd.ParseError)
	}

	if gd.IsRange() {
		// Default to January 1st if month/day not specified
		month := gd.StartMonth
		if month == 0 {
			month = 1
		}
		day := gd.StartDay
		if day == 0 {
			day = 1
		}
		return time.Date(gd.StartYear, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
	}

	// Default to January 1st if month/day not specified
	month := gd.Month
	if month == 0 {
		month = 1
	}
	day := gd.Day
	if day == 0 {
		day = 1
	}

	return time.Date(gd.Year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// ToISO8601 converts the date to ISO 8601 format (YYYY-MM-DD).
// For range dates, returns the start date.
// Returns empty string if date is invalid.
func (gd *GedcomDate) ToISO8601() string {
	if !gd.IsValid() {
		return ""
	}

	if gd.IsRange() {
		return formatISO8601(gd.StartYear, gd.StartMonth, gd.StartDay)
	}

	return formatISO8601(gd.Year, gd.Month, gd.Day)
}

// formatISO8601 formats year, month, day as ISO 8601.
func formatISO8601(year, month, day int) string {
	if year == 0 {
		return ""
	}

	monthStr := fmt.Sprintf("%02d", month)
	dayStr := fmt.Sprintf("%02d", day)

	if month == 0 {
		return fmt.Sprintf("%04d", year)
	}
	if day == 0 {
		return fmt.Sprintf("%04d-%s", year, monthStr)
	}

	return fmt.Sprintf("%04d-%s-%s", year, monthStr, dayStr)
}

// Earliest returns the earliest possible time for this date.
func (gd *GedcomDate) Earliest() time.Time {
	if !gd.IsValid() {
		return time.Time{}
	}

	if gd.IsRange() {
		month := gd.StartMonth
		if month == 0 {
			month = 1
		}
		day := gd.StartDay
		if day == 0 {
			day = 1
		}
		return time.Date(gd.StartYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}

	month := gd.Month
	if month == 0 {
		month = 1
	}
	day := gd.Day
	if day == 0 {
		day = 1
	}

	// Adjust for date types
	switch gd.Type {
	case DateTypeBefore:
		// Before: use earliest possible (year 1)
		return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	case DateTypeAfter:
		// After: use the date itself
		return time.Date(gd.Year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	case DateTypeAbout:
		// About: use the date itself (could subtract years for range, but keeping simple)
		return time.Date(gd.Year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	default:
		return time.Date(gd.Year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}
}

// Latest returns the latest possible time for this date.
func (gd *GedcomDate) Latest() time.Time {
	if !gd.IsValid() {
		return time.Time{}
	}

	if gd.IsRange() {
		month := gd.EndMonth
		if month == 0 {
			month = 12
		}
		day := gd.EndDay
		if day == 0 {
			day = 31 // Last day of month (approximate)
		}
		return time.Date(gd.EndYear, time.Month(month), day, 23, 59, 59, 0, time.UTC)
	}

	month := gd.Month
	if month == 0 {
		month = 12
	}
	day := gd.Day
	if day == 0 {
		day = 31 // Last day of month (approximate)
	}

	// Adjust for date types
	switch gd.Type {
	case DateTypeBefore:
		// Before: use the date itself
		return time.Date(gd.Year, time.Month(month), day, 23, 59, 59, 0, time.UTC)
	case DateTypeAfter:
		// After: use latest possible (year 9999)
		return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	case DateTypeAbout:
		// About: use the date itself (could add years for range, but keeping simple)
		return time.Date(gd.Year, time.Month(month), day, 23, 59, 59, 0, time.UTC)
	default:
		return time.Date(gd.Year, time.Month(month), day, 23, 59, 59, 0, time.UTC)
	}
}

// Compare compares two dates. Returns:
//   - -1 if this date is before other
//   - 0 if dates are equal
//   - 1 if this date is after other
//
// Uses earliest time for comparison.
func (gd *GedcomDate) Compare(other *GedcomDate) int {
	if !gd.IsValid() || !other.IsValid() {
		return 0
	}

	thisEarliest := gd.Earliest()
	otherEarliest := other.Earliest()

	if thisEarliest.Before(otherEarliest) {
		return -1
	}
	if thisEarliest.After(otherEarliest) {
		return 1
	}
	return 0
}

// String returns a string representation of the date.
func (gd *GedcomDate) String() string {
	if !gd.IsValid() {
		return gd.Original
	}

	if gd.IsRange() {
		start := formatDateComponents(gd.StartYear, gd.StartMonth, gd.StartDay)
		end := formatDateComponents(gd.EndYear, gd.EndMonth, gd.EndDay)
		if gd.Type == DateTypeBetween {
			return fmt.Sprintf("BET %s AND %s", start, end)
		}
		return fmt.Sprintf("FROM %s TO %s", start, end)
	}

	dateStr := formatDateComponents(gd.Year, gd.Month, gd.Day)
	if gd.Type != DateTypeExact {
		return fmt.Sprintf("%s %s", gd.Type, dateStr)
	}

	return dateStr
}

// formatDateComponents formats year, month, day as GEDCOM date string.
func formatDateComponents(year, month, day int) string {
	if year == 0 {
		return ""
	}

	if month == 0 {
		return fmt.Sprintf("%d", year)
	}

	monthNames := []string{"", "JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	monthStr := monthNames[month]

	if day == 0 {
		return fmt.Sprintf("%s %d", monthStr, year)
	}

	return fmt.Sprintf("%d %s %d", day, monthStr, year)
}

// DateConstraintFromString returns the constraint for the provided keyword.
// If the word is not understood, DateConstraintExact will be returned.
// This function is not case sensitive.
func DateConstraintFromString(word string) DateConstraint {
	lowerWord := strings.ToLower(strings.TrimSpace(word))
	if lowerWord == "" {
		return DateConstraintExact
	}

	// Check if word matches any of the date word constants
	if wordInWords(lowerWord, DateWordsAbout) {
		return DateConstraintAbout
	}
	if wordInWords(lowerWord, DateWordsAfter) {
		return DateConstraintAfter
	}
	if wordInWords(lowerWord, DateWordsBefore) {
		return DateConstraintBefore
	}

	return DateConstraintExact
}

// wordInWords checks if a word is in a pipe-separated list of words.
func wordInWords(word, words string) bool {
	for _, w := range strings.Split(strings.ToLower(words), "|") {
		if w == word {
			return true
		}
	}
	return false
}

// String returns the constraint abbreviation for non-exact dates.
// Exact dates will return an empty string.
func (dc DateConstraint) String() string {
	switch dc {
	case DateConstraintAbout:
		return strings.Split(DateWordsAbout, "|")[0]
	case DateConstraintAfter:
		return strings.Split(DateWordsAfter, "|")[0]
	case DateConstraintBefore:
		return strings.Split(DateWordsBefore, "|")[0]
	default:
		return ""
	}
}

// Constraint returns the DateConstraint for this date based on its Type.
func (gd *GedcomDate) Constraint() DateConstraint {
	switch gd.Type {
	case DateTypeAbout:
		return DateConstraintAbout
	case DateTypeBefore:
		return DateConstraintBefore
	case DateTypeAfter:
		return DateConstraintAfter
	default:
		return DateConstraintExact
	}
}

// Years returns the number of years of a date as a floating-point.
// It can be used as an approximation to get a general idea of how far apart dates are.
//
// For specific dates, it's calculated as the number of days that have passed,
// divided by the number of days in that year (to correct for leap years).
//
// Since some date components can be missing (like the day or month), Years
// compensates by returning the midpoint (average) of the maximum and minimum value.
//
// When only a year is provided, 0.5 will be added to the year.
func (gd *GedcomDate) Years() float64 {
	if !gd.IsValid() {
		return 0
	}

	if gd.IsRange() {
		start := gd.getStartDateForYears()
		end := gd.getEndDateForYears()
		return (start.Years() + end.Years()) / 2.0
	}

	return gd.getSingleDateYears()
}

// getSingleDateYears calculates Years for a single (non-range) date.
func (gd *GedcomDate) getSingleDateYears() float64 {
	hasDay := gd.Day != 0
	hasMonth := gd.Month != 0
	hasYear := gd.Year != 0

	if hasDay && hasMonth && hasYear {
		// Calculate the total number of days in this year to account for leap years
		t := time.Date(gd.Year, time.Month(gd.Month), gd.Day, 0, 0, 0, 0, time.UTC)
		daysInYear := time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC).
			AddDate(0, 0, -1).YearDay() + 1

		fractional := float64(t.YearDay()) / float64(daysInYear)
		return float64(t.Year()) + fractional
	}

	if hasMonth && hasYear {
		// Average of first and last day of month
		start := time.Date(gd.Year, time.Month(gd.Month), 1, 0, 0, 0, 0, time.UTC)
		lastDay := time.Date(gd.Year, time.Month(gd.Month)+1, 1, 0, 0, 0, 0, time.UTC).
			AddDate(0, 0, -1).Day()

		startDaysInYear := time.Date(start.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC).
			AddDate(0, 0, -1).YearDay() + 1
		endDaysInYear := startDaysInYear // Same year

		startFractional := float64(start.YearDay()) / float64(startDaysInYear)
		endFractional := float64(lastDay) / float64(endDaysInYear)

		startYears := float64(start.Year()) + startFractional
		endYears := float64(start.Year()) + endFractional

		return (startYears + endYears) / 2
	}

	if hasYear {
		return float64(gd.Year) + 0.5
	}

	return 0
}

// getStartDateForYears returns a GedcomDate representing the start of the range.
func (gd *GedcomDate) getStartDateForYears() *GedcomDate {
	return &GedcomDate{
		Year:  gd.StartYear,
		Month: gd.StartMonth,
		Day:   gd.StartDay,
		Type:  DateTypeExact,
		IsParsed: true,
	}
}

// getEndDateForYears returns a GedcomDate representing the end of the range.
func (gd *GedcomDate) getEndDateForYears() *GedcomDate {
	return &GedcomDate{
		Year:  gd.EndYear,
		Month: gd.EndMonth,
		Day:   gd.EndDay,
		Type:  DateTypeExact,
		IsParsed: true,
	}
}

// Similarity returns a value from 0.0 to 1.0 to identify how similar two dates are.
// 1.0 means the dates are exactly the same, 0.0 means they are not similar at all.
//
// Similarity is calculated on a parabola that awards higher values to dates that
// are proportionally closer to each other.
//
// The maxYears allows the error margin to be adjusted. Dates that are further
// apart (in any direction) than maxYears will always return 0.0.
func (gd *GedcomDate) Similarity(other *GedcomDate, maxYears float64) float64 {
	if !gd.IsValid() || !other.IsValid() {
		return 0.5 // Return neutral value when dates are invalid
	}

	leftYears := gd.Years()
	rightYears := other.Years()
	yearsApart := leftYears - rightYears
	if yearsApart < 0 {
		yearsApart = -yearsApart
	}

	similarity := math.Pow(yearsApart/maxYears, 2)

	// When dates are too far apart, similarity goes asymptotic
	if similarity > 1 {
		return 0
	}

	return 1 - similarity
}

// IsExact returns true if all parts of the date are complete and the date
// constraint is exact. This means it points to a specific day.
func (gd *GedcomDate) IsExact() bool {
	if !gd.IsValid() {
		return false
	}

	if gd.IsRange() {
		return gd.StartDay != 0 && gd.StartMonth != 0 && gd.StartYear != 0 &&
			gd.EndDay != 0 && gd.EndMonth != 0 && gd.EndYear != 0 &&
			gd.Type == DateTypeExact
	}

	return gd.Day != 0 && gd.Month != 0 && gd.Year != 0 && gd.Type == DateTypeExact
}

// IsBefore returns true if this date is before the other date.
func (gd *GedcomDate) IsBefore(other *GedcomDate) bool {
	if !gd.IsValid() || !other.IsValid() {
		return false
	}

	leftYears := gd.Years()
	rightYears := other.Years()

	return leftYears < rightYears
}

// IsAfter returns true if this date is after the other date.
func (gd *GedcomDate) IsAfter(other *GedcomDate) bool {
	if !gd.IsValid() || !other.IsValid() {
		return false
	}

	leftYears := gd.Years()
	rightYears := other.Years()

	return leftYears > rightYears
}

// Equals compares two dates taking into consideration the date constraint.
//
// Unlike Compare(), Equals() takes into account what the date and its constraint
// represents, rather than just its raw value.
//
// For example, "3 Sep 1943" == "Bef. Oct 1943" returns true because 3 Sep 1943
// is before Oct 1943.
//
// If either date is invalid, false is always returned.
func (gd *GedcomDate) Equals(other *GedcomDate) bool {
	if !gd.IsValid() || !other.IsValid() {
		return false
	}

	// If both dates are exactly the same (same day, month, year, constraint), return true
	if gd.Is(other) {
		return true
	}

	// Use constraint-aware comparison matrix
	matchers := [][]func(d1, d2 *GedcomDate) bool{
		{equalsA, equalsA, equalsB, equalsC}, // Exact row
		{equalsA, equalsA, equalsD, equalsD}, // About row
		{equalsC, equalsD, equalsC, equalsD}, // Before row
		{equalsB, equalsD, equalsD, equalsB}, // After row
	}

	c1 := gd.Constraint()
	c2 := other.Constraint()

	return matchers[c2][c1](gd, other)
}

// Is compares two dates. Dates are only considered to be the same if the day,
// month, year and constraint are all the same.
func (gd *GedcomDate) Is(other *GedcomDate) bool {
	if !gd.IsValid() || !other.IsValid() {
		return false
	}

	if gd.IsRange() != other.IsRange() {
		return false
	}

	if gd.IsRange() {
		return gd.StartDay == other.StartDay &&
			gd.StartMonth == other.StartMonth &&
			gd.StartYear == other.StartYear &&
			gd.EndDay == other.EndDay &&
			gd.EndMonth == other.EndMonth &&
			gd.EndYear == other.EndYear &&
			gd.Constraint() == other.Constraint()
	}

	return gd.Day == other.Day &&
		gd.Month == other.Month &&
		gd.Year == other.Year &&
		gd.Constraint() == other.Constraint()
}

// equalsA: Match if the day, month and year are all equal.
func equalsA(d1, d2 *GedcomDate) bool {
	if d1.IsRange() || d2.IsRange() {
		return false // Ranges need special handling
	}

	if d1.Day != d2.Day {
		return false
	}
	if d1.Month != d2.Month {
		return false
	}
	return d1.Year == d2.Year
}

// equalsB: Match if left.Years() > right.Years().
func equalsB(d1, d2 *GedcomDate) bool {
	return d1.Years() > d2.Years()
}

// equalsC: Match if left.Years() < right.Years().
func equalsC(d1, d2 *GedcomDate) bool {
	return d1.Years() < d2.Years()
}

// equalsD: Never a match.
func equalsD(d1, d2 *GedcomDate) bool {
	return false
}

// IsZero returns true if the day, month and year are not provided.
func (gd *GedcomDate) IsZero() bool {
	if gd.IsRange() {
		return (gd.StartDay == 0 && gd.StartMonth == 0 && gd.StartYear == 0) &&
			(gd.EndDay == 0 && gd.EndMonth == 0 && gd.EndYear == 0)
	}

	return gd.Day == 0 && gd.Month == 0 && gd.Year == 0
}

// Sub returns the duration between two dates.
func (gd *GedcomDate) Sub(other *GedcomDate) Duration {
	if !gd.IsValid() || !other.IsValid() {
		return NewDuration(0, false, true)
	}

	a := gd.Earliest()
	b := other.Earliest()

	isKnown := gd.ParseError == nil && other.ParseError == nil
	isEstimate := !gd.IsExact() || !other.IsExact()

	// Create Duration directly to preserve negative values (NewDuration forces positive)
	duration := a.Sub(b)
	return Duration{
		Duration:   duration,
		IsKnown:    isKnown,
		IsEstimate: isEstimate,
	}
}
