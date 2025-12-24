package types

import (
	"strings"
	"testing"
	"time"
)

func TestParseDate_ExactDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected struct {
			Year  int
			Month int
			Day   int
		}
	}{
		{"Full date", "15 JAN 1800", struct{ Year, Month, Day int }{1800, 1, 15}},
		{"Full date 2", "1 DEC 1900", struct{ Year, Month, Day int }{1900, 12, 1}},
		{"Full date 3", "31 MAR 2000", struct{ Year, Month, Day int }{2000, 3, 31}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate failed: %v", err)
			}

			if !date.IsValid() {
				t.Errorf("Date should be valid")
			}

			if date.Year != tt.expected.Year {
				t.Errorf("Year = %d, want %d", date.Year, tt.expected.Year)
			}
			if date.Month != tt.expected.Month {
				t.Errorf("Month = %d, want %d", date.Month, tt.expected.Month)
			}
			if date.Day != tt.expected.Day {
				t.Errorf("Day = %d, want %d", date.Day, tt.expected.Day)
			}

			if date.Type != DateTypeExact {
				t.Errorf("Type = %s, want %s", date.Type, DateTypeExact)
			}
		})
	}
}

func TestParseDate_MonthYear(t *testing.T) {
	date, err := ParseDate("JAN 1800")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	if !date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if date.Year != 1800 {
		t.Errorf("Year = %d, want 1800", date.Year)
	}
	if date.Month != 1 {
		t.Errorf("Month = %d, want 1", date.Month)
	}
	if date.Day != 0 {
		t.Errorf("Day = %d, want 0", date.Day)
	}
}

func TestParseDate_YearOnly(t *testing.T) {
	date, err := ParseDate("1800")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	if !date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if date.Year != 1800 {
		t.Errorf("Year = %d, want 1800", date.Year)
	}
	if date.Month != 0 {
		t.Errorf("Month = %d, want 0", date.Month)
	}
	if date.Day != 0 {
		t.Errorf("Day = %d, want 0", date.Day)
	}
}

func TestParseDate_About(t *testing.T) {
	date, err := ParseDate("ABT 1850")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	if !date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if date.Type != DateTypeAbout {
		t.Errorf("Type = %s, want %s", date.Type, DateTypeAbout)
	}
	if date.Year != 1850 {
		t.Errorf("Year = %d, want 1850", date.Year)
	}
}

func TestParseDate_Before(t *testing.T) {
	date, err := ParseDate("BEF 1900")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	if date.Type != DateTypeBefore {
		t.Errorf("Type = %s, want %s", date.Type, DateTypeBefore)
	}
	if date.Year != 1900 {
		t.Errorf("Year = %d, want 1900", date.Year)
	}
}

func TestParseDate_After(t *testing.T) {
	date, err := ParseDate("AFT 1900")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	if date.Type != DateTypeAfter {
		t.Errorf("Type = %s, want %s", date.Type, DateTypeAfter)
	}
	if date.Year != 1900 {
		t.Errorf("Year = %d, want 1900", date.Year)
	}
}

func TestParseDate_Between(t *testing.T) {
	date, err := ParseDate("BET 1800 AND 1850")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	if !date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if date.Type != DateTypeBetween {
		t.Errorf("Type = %s, want %s", date.Type, DateTypeBetween)
	}

	if !date.IsRange() {
		t.Errorf("IsRange() = false, want true")
	}

	if date.StartYear != 1800 {
		t.Errorf("StartYear = %d, want 1800", date.StartYear)
	}
	if date.EndYear != 1850 {
		t.Errorf("EndYear = %d, want 1850", date.EndYear)
	}
}

func TestParseDate_FromTo(t *testing.T) {
	date, err := ParseDate("FROM 1800 TO 1850")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	if !date.IsValid() {
		t.Errorf("Date should be valid")
	}

	if date.Type != DateTypeFromTo {
		t.Errorf("Type = %s, want %s", date.Type, DateTypeFromTo)
	}

	if !date.IsRange() {
		t.Errorf("IsRange() = false, want true")
	}

	if date.StartYear != 1800 {
		t.Errorf("StartYear = %d, want 1800", date.StartYear)
	}
	if date.EndYear != 1850 {
		t.Errorf("EndYear = %d, want 1850", date.EndYear)
	}
}

func TestGedcomDate_ToISO8601(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Full date", "15 JAN 1800", "1800-01-15"},
		{"Month-year", "JAN 1800", "1800-01"},
		{"Year only", "1800", "1800"},
		{"Between", "BET 1800 AND 1850", "1800"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate failed: %v", err)
			}

			iso := date.ToISO8601()
			if iso != tt.expected {
				t.Errorf("ToISO8601() = %q, want %q", iso, tt.expected)
			}
		})
	}
}

func TestGedcomDate_ToTime(t *testing.T) {
	date, err := ParseDate("15 JAN 1800")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}

	tm, err := date.ToTime()
	if err != nil {
		t.Fatalf("ToTime() failed: %v", err)
	}

	expected := time.Date(1800, time.January, 15, 0, 0, 0, 0, time.UTC)
	if !tm.Equal(expected) {
		t.Errorf("ToTime() = %v, want %v", tm, expected)
	}
}

func TestGedcomDate_Compare(t *testing.T) {
	date1, _ := ParseDate("15 JAN 1800")
	date2, _ := ParseDate("20 JAN 1800")
	date3, _ := ParseDate("15 JAN 1800")

	if date1.Compare(date2) != -1 {
		t.Errorf("date1 should be before date2")
	}

	if date2.Compare(date1) != 1 {
		t.Errorf("date2 should be after date1")
	}

	if date1.Compare(date3) != 0 {
		t.Errorf("date1 should equal date3")
	}
}

func TestGedcomDate_String(t *testing.T) {
	tests := []struct {
		input    string
		contains []string // Check that string contains these substrings
	}{
		{"15 JAN 1800", []string{"15", "JAN", "1800"}},
		{"ABT 1850", []string{"ABOUT", "1850"}},
		{"BET 1800 AND 1850", []string{"BET", "1800", "AND", "1850"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			date, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate failed: %v", err)
			}

			str := date.String()
			if str == "" {
				t.Errorf("String() returned empty string")
			}

			// Check that string contains expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(str, substr) {
					t.Errorf("String() = %q, should contain %q", str, substr)
				}
			}
		})
	}
}
