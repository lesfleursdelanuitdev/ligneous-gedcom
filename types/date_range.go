package types

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// DateRange represents a period of time.
// The minimum possible period is 1 day and ranges only have a resolution of a single day.
type DateRange struct {
	start          *GedcomDate
	end            *GedcomDate
	originalString string
}

// NewZeroDateRange creates a new zero DateRange.
func NewZeroDateRange() DateRange {
	return DateRange{}
}

// NewDateRange creates a new date range between two provided dates.
// It is expected that the start date be less than or equal to the end date.
func NewDateRange(start, end *GedcomDate) DateRange {
	return DateRange{
		start: start,
		end:   end,
	}
}

// NewDateRangeWithString creates a DateRange from a GEDCOM date string.
func NewDateRangeWithString(s string) DateRange {
	dateString := cleanSpace(s)

	// Try to match a range first
	dateRangeRegexp := regexp.MustCompile(
		fmt.Sprintf(`(?i)^(%s) (.+) (%s) (.+)$`, DateWordsBetween, DateWordsAnd))
	parts := dateRangeRegexp.FindStringSubmatch(dateString)

	if len(parts) > 0 {
		datePart1 := parseDatePartsForRange(parts[2], false)
		datePart2 := parseDatePartsForRange(parts[4], true)
		return NewDateRange(datePart1, datePart2)
	}

	// Single date - create range with same start and end
	datePart1 := parseDatePartsForRange(dateString, false)
	datePart2 := parseDatePartsForRange(dateString, true)

	return DateRange{
		start:          datePart1,
		end:            datePart2,
		originalString: s,
	}
}

// parseDatePartsForRange parses date parts for a DateRange.
// This is a simplified version that creates a GedcomDate.
func parseDatePartsForRange(dateString string, isEndOfRange bool) *GedcomDate {
	// Use the existing ParseDate function
	date, err := ParseDate(dateString)
	if err != nil {
		// Return invalid date with parse error
		return &GedcomDate{
			Original:   dateString,
			ParseError: err,
			IsParsed:   false,
		}
	}

	return date
}

// cleanSpace works similar to strings.TrimSpace except that it also replaces
// consecutive spaces anywhere in the string with a single space.
func cleanSpace(s string) string {
	// Replace twice if there is an odd number of spaces in a row
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.Replace(s, "  ", " ", -1)

	// Trim whatever spaces are left on either side
	return strings.TrimSpace(s)
}

// StartDate returns the lower boundary of the date range.
func (dr DateRange) StartDate() *GedcomDate {
	if dr.start == nil {
		return &GedcomDate{}
	}
	return dr.start
}

// EndDate returns the upper boundary of the date range.
func (dr DateRange) EndDate() *GedcomDate {
	if dr.end == nil {
		return &GedcomDate{}
	}
	return dr.end
}

// StartAndEndDates returns both start and end dates.
func (dr DateRange) StartAndEndDates() (*GedcomDate, *GedcomDate) {
	return dr.StartDate(), dr.EndDate()
}

// Years works in a similar way to GedcomDate.Years() but also takes into
// consideration the StartDate() and EndDate() values of a whole date range.
// It does this by averaging out the Years() value of the StartDate() and EndDate() values.
func (dr DateRange) Years() float64 {
	start := dr.StartDate()
	end := dr.EndDate()

	if start.IsZero() && end.IsZero() {
		return 0
	}

	return (start.Years() + end.Years()) / 2.0
}

// Similarity returns a value from 0.0 to 1.0 to identify how similar two date ranges are.
// 1.0 means the dates are exactly the same, 0.0 means they are not similar at all.
//
// Similarity is calculated on a parabola that awards higher values to dates that
// are proportionally closer to each other.
func (dr DateRange) Similarity(other DateRange, maxYears float64) float64 {
	leftYears := dr.Years()
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

// Equals compares the values of two date ranges taking into consideration the date constraint.
func (dr DateRange) Equals(other DateRange) bool {
	// Compare dates by value range
	matchStartDate := dr.StartDate().Equals(other.StartDate())
	matchEndDate := dr.EndDate().Equals(other.EndDate())

	return matchStartDate && matchEndDate
}

// IsBefore returns true if the start date is before the other start date.
func (dr DateRange) IsBefore(other DateRange) bool {
	return dr.StartDate().IsBefore(other.StartDate())
}

// IsAfter returns true if the end date is after the other end date.
func (dr DateRange) IsAfter(other DateRange) bool {
	return dr.EndDate().IsAfter(other.EndDate())
}

// IsValid returns true only when the start and end dates are non-zero.
func (dr DateRange) IsValid() bool {
	start, end := dr.StartAndEndDates()
	return !start.IsZero() && !end.IsZero()
}

// IsExact will return true if the date range represents a single day with an exact constraint.
func (dr DateRange) IsExact() bool {
	start, end := dr.StartAndEndDates()
	startIsExact := start.IsExact()
	endIsExact := end.IsExact()

	// Both dates must be exact AND they must be equal (same day)
	return startIsExact && endIsExact && start.Equals(end)
}

// IsPhrase returns true if the date value is a phrase.
// A phrase is any statement offered as a date when the year is not
// recognizable to a date parser, but which gives information about when an
// event occurred. The date phrase is enclosed in matching parentheses.
func (dr DateRange) IsPhrase() bool {
	if len(dr.originalString) == 0 {
		return false
	}

	firstLetter := dr.originalString[0]
	lastLetter := dr.originalString[len(dr.originalString)-1]

	return firstLetter == '(' && lastLetter == ')'
}

// String returns a string representation of the date range.
func (dr DateRange) String() string {
	start, end := dr.StartAndEndDates()
	if start.Equals(end) {
		return start.String()
	}

	return fmt.Sprintf("Bet. %s and %s", start.String(), end.String())
}

// Sub returns the duration between two date ranges.
func (dr DateRange) Sub(other DateRange) (min Duration, max Duration) {
	start := dr.StartDate().Sub(other.StartDate())
	end := dr.EndDate().Sub(other.EndDate())

	return start, end
}

// Duration returns the duration of the date range itself.
func (dr DateRange) Duration() Duration {
	start := dr.StartDate()
	end := dr.EndDate()

	if !start.IsValid() || !end.IsValid() {
		return NewDuration(0, false, true)
	}

	startTime := start.Earliest()
	endTime := end.Latest()

	duration := endTime.Sub(startTime)
	return NewDuration(duration, true, !dr.IsExact())
}
