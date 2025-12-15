package gedcom

import (
	"fmt"
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
	// Month abbreviations
	monthMap = map[string]int{
		"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
		"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
	}

	// Date type prefixes
	dateTypePrefixes = map[string]DateType{
		"ABT":  DateTypeAbout,
		"BEF":  DateTypeBefore,
		"AFT":  DateTypeAfter,
		"BET":  DateTypeBetween,
		"FROM": DateTypeFrom,
		"TO":   DateTypeTo,
	}

	// Patterns for date parsing
	exactDatePattern = regexp.MustCompile(`^(\d{1,2})\s+([A-Z]{3})\s+(\d{4})$`)
	monthYearPattern = regexp.MustCompile(`^([A-Z]{3})\s+(\d{4})$`)
	yearOnlyPattern  = regexp.MustCompile(`^\d{4}$`)
	betweenPattern   = regexp.MustCompile(`^BET\s+(.+?)\s+AND\s+(.+)$`)
	fromToPattern    = regexp.MustCompile(`^FROM\s+(.+?)\s+TO\s+(.+)$`)
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

	upperDate := strings.ToUpper(strings.TrimSpace(dateStr))

	// Check for date type prefixes
	parts := strings.Fields(upperDate)
	if len(parts) > 0 {
		if dateType, ok := dateTypePrefixes[parts[0]]; ok {
			date.Type = dateType
			// For BETWEEN and FROM, keep the full string for parsing
			if dateType == DateTypeBetween || dateType == DateTypeFrom {
				// Keep original for pattern matching
			} else {
				upperDate = strings.Join(parts[1:], " ")
			}
		}
	}

	// Parse based on type
	var err error
	switch date.Type {
	case DateTypeBetween:
		err = parseBetweenDate(date, upperDate)
	case DateTypeFrom:
		err = parseFromToDate(date, upperDate)
	default:
		err = parseSingleDate(date, upperDate)
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
	// Try exact date: "15 JAN 1800"
	if matches := exactDatePattern.FindStringSubmatch(dateStr); matches != nil {
		day, _ := strconv.Atoi(matches[1])
		monthStr := matches[2]
		year, _ := strconv.Atoi(matches[3])

		month, ok := monthMap[monthStr]
		if !ok {
			return fmt.Errorf("invalid month: %s", monthStr)
		}

		date.Day = day
		date.Month = month
		date.Year = year
		return nil
	}

	// Try month-year: "JAN 1800"
	if matches := monthYearPattern.FindStringSubmatch(dateStr); matches != nil {
		monthStr := matches[1]
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
		date.Year = year
		return nil
	}

	return fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseBetweenDate parses a "BET X AND Y" date.
func parseBetweenDate(date *GedcomDate, dateStr string) error {
	matches := betweenPattern.FindStringSubmatch(strings.ToUpper(dateStr))
	if matches == nil {
		return fmt.Errorf("invalid BETWEEN date format: %s", dateStr)
	}

	startStr := strings.TrimSpace(matches[1])
	endStr := strings.TrimSpace(matches[2])

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

// parseFromToDate parses a "FROM X TO Y" date.
func parseFromToDate(date *GedcomDate, dateStr string) error {
	matches := fromToPattern.FindStringSubmatch(strings.ToUpper(dateStr))
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
		return time.Date(gd.StartYear, time.Month(gd.StartMonth), gd.StartDay, 0, 0, 0, 0, time.UTC), nil
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
