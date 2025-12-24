package types

import (
	"fmt"
	"testing"
	"time"
)

// TestParseDate_BetweenEdgeCases tests edge cases for parseBetweenDate
func TestParseDate_BetweenEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		description string
	}{
		{
			name:        "valid between with different formats",
			input:       "BET 1 JAN 1900 AND 31 DEC 1900",
			shouldError: false,
			description: "Standard between format",
		},
		{
			name:        "between with lowercase",
			input:       "bet 1 jan 1900 and 31 dec 1900",
			shouldError: false,
			description: "Lowercase between format",
		},
		{
			name:        "between with abbreviated bet",
			input:       "BET. 1 JAN 1900 AND 31 DEC 1900",
			shouldError: false,
			description: "Abbreviated BET. format",
		},
		{
			name:        "between with 'to' connector",
			input:       "BET 1 JAN 1900 TO 31 DEC 1900",
			shouldError: false,
			description: "Between with TO connector",
		},
		{
			name:        "between with dash connector",
			input:       "BET 1 JAN 1900 - 31 DEC 1900",
			shouldError: false,
			description: "Between with dash connector",
		},
		{
			name:        "between with 'from' keyword",
			input:       "FROM 1 JAN 1900 TO 31 DEC 1900",
			shouldError: false,
			description: "From...to format",
		},
		{
			name:        "between with year only dates",
			input:       "BET 1900 AND 1905",
			shouldError: false,
			description: "Between with year-only dates",
		},
		{
			name:        "between with month-year dates",
			input:       "BET JAN 1900 AND DEC 1905",
			shouldError: false,
			description: "Between with month-year dates",
		},
		{
			name:        "invalid between - missing connector",
			input:       "BET 1 JAN 1900 31 DEC 1900",
			shouldError: true,
			description: "Missing connector word",
		},
		{
			name:        "invalid between - missing end date",
			input:       "BET 1 JAN 1900 AND",
			shouldError: true,
			description: "Missing end date",
		},
		{
			name:        "invalid between - missing start date",
			input:       "BET AND 31 DEC 1900",
			shouldError: true,
			description: "Missing start date",
		},
		{
			name:        "invalid between - malformed start date",
			input:       "BET INVALID AND 31 DEC 1900",
			shouldError: true,
			description: "Malformed start date",
		},
		{
			name:        "invalid between - malformed end date",
			input:       "BET 1 JAN 1900 AND INVALID",
			shouldError: true,
			description: "Malformed end date",
		},
		{
			name:        "between with reversed dates",
			input:       "BET 31 DEC 1900 AND 1 JAN 1900",
			shouldError: false,
			description: "Reversed dates (should still parse, validation is separate)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := ParseDate(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("ParseDate(%q) expected error but got none", tt.input)
				}
				if date != nil && date.IsValid() {
					t.Errorf("ParseDate(%q) should return invalid date on error", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseDate(%q) unexpected error: %v", tt.input, err)
				}
				if date == nil {
					t.Errorf("ParseDate(%q) returned nil date", tt.input)
				} else if !date.IsValid() {
					t.Errorf("ParseDate(%q) returned invalid date: %v", tt.input, date.ParseError)
				}
				// Verify it's a range date
				if !date.IsRange() {
					t.Errorf("ParseDate(%q) should return range date, got Type=%s", tt.input, date.Type)
				}
			}
		})
	}
}

// TestGedcomDate_Earliest tests the Earliest method with various date types
func TestGedcomDate_Earliest(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "exact date",
			input:    "15 JAN 1900",
			expected: time.Date(1900, time.January, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "exact date with zero month",
			input:    "1900",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "exact date with zero day",
			input:    "JAN 1900",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "before date",
			input:    "BEF 1900",
			expected: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "after date",
			input:    "AFT 1900",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "about date",
			input:    "ABT 1900",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "range date - between",
			input:    "BET 1 JAN 1900 AND 31 DEC 1900",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "range date - from to",
			input:    "FROM 1 JAN 1900 TO 31 DEC 1900",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "range date with zero month",
			input:    "BET 1900 AND 1905",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "range date with zero day",
			input:    "BET JAN 1900 AND DEC 1905",
			expected: time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "before date with full date",
			input:    "BEF 15 JAN 1900",
			expected: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "after date with full date",
			input:    "AFT 15 JAN 1900",
			expected: time.Date(1900, time.January, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "about date with full date",
			input:    "ABT 15 JAN 1900",
			expected: time.Date(1900, time.January, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) failed: %v", tt.input, err)
			}

			if !date.IsValid() {
				t.Fatalf("ParseDate(%q) returned invalid date", tt.input)
			}

			earliest := date.Earliest()
			if !earliest.Equal(tt.expected) {
				t.Errorf("Earliest() = %v, want %v", earliest, tt.expected)
			}
		})
	}

	// Test with invalid date
	t.Run("invalid date", func(t *testing.T) {
		date := &GedcomDate{}
		earliest := date.Earliest()
		if !earliest.IsZero() {
			t.Errorf("Earliest() for invalid date should return zero time, got %v", earliest)
		}
	})
}

// TestGedcomDate_ToTime_Comprehensive tests ToTime with various date formats
func TestGedcomDate_ToTime_Comprehensive(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Time
		shouldError bool
	}{
		{
			name:        "full date",
			input:       "15 JAN 1800",
			expected:    time.Date(1800, time.January, 15, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "month-year date",
			input:       "JAN 1800",
			expected:    time.Date(1800, time.January, 1, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "year only date",
			input:       "1800",
			expected:    time.Date(1800, time.January, 1, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "range date - between",
			input:       "BET 1 JAN 1900 AND 31 DEC 1900",
			expected:    time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "range date - from to",
			input:       "FROM 1 JAN 1900 TO 31 DEC 1900",
			expected:    time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "range date with zero month",
			input:       "BET 1900 AND 1905",
			expected:    time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "range date with zero day",
			input:       "BET JAN 1900 AND DEC 1905",
			expected:    time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "about date",
			input:       "ABT 15 JAN 1800",
			expected:    time.Date(1800, time.January, 15, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "before date",
			input:       "BEF 15 JAN 1800",
			expected:    time.Date(1800, time.January, 15, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "after date",
			input:       "AFT 15 JAN 1800",
			expected:    time.Date(1800, time.January, 15, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "invalid date",
			input:       "INVALID DATE",
			expected:    time.Time{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := ParseDate(tt.input)
			if err != nil && !tt.shouldError {
				t.Fatalf("ParseDate(%q) failed: %v", tt.input, err)
			}

			if tt.shouldError {
				// For invalid dates, ToTime should return error
				if date != nil {
					tm, err := date.ToTime()
					if err == nil {
						t.Errorf("ToTime() for invalid date should return error, got %v", tm)
					}
				}
				return
			}

			if date == nil || !date.IsValid() {
				t.Fatalf("ParseDate(%q) returned invalid date", tt.input)
			}

			tm, err := date.ToTime()
			if err != nil {
				t.Errorf("ToTime() failed: %v", err)
				return
			}

			if !tm.Equal(tt.expected) {
				t.Errorf("ToTime() = %v, want %v", tm, tt.expected)
			}
		})
	}

	// Test with explicitly invalid date
	t.Run("explicitly invalid date", func(t *testing.T) {
		date := &GedcomDate{
			ParseError: fmt.Errorf("test error"),
		}
		tm, err := date.ToTime()
		if err == nil {
			t.Errorf("ToTime() should return error for invalid date, got %v", tm)
		}
	})
}

// TestGetErrorType tests the GetErrorType function
func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "StandardError with Parse type",
			err:      NewStandardError(ErrorTypeParse, SeverityWarning, "parse error"),
			expected: ErrorTypeParse,
		},
		{
			name:     "StandardError with Validation type",
			err:      NewStandardError(ErrorTypeValidation, SeverityWarning, "validation error"),
			expected: ErrorTypeValidation,
		},
		{
			name:     "StandardError with Query type",
			err:      NewStandardError(ErrorTypeQuery, SeverityWarning, "query error"),
			expected: ErrorTypeQuery,
		},
		{
			name:     "StandardError with Storage type",
			err:      NewStandardError(ErrorTypeStorage, SeverityWarning, "storage error"),
			expected: ErrorTypeStorage,
		},
		{
			name:     "StandardError with IO type",
			err:      NewStandardError(ErrorTypeIO, SeverityWarning, "io error"),
			expected: ErrorTypeIO,
		},
		{
			name:     "StandardError with Internal type",
			err:      NewStandardError(ErrorTypeInternal, SeverityWarning, "internal error"),
			expected: ErrorTypeInternal,
		},
		{
			name:     "StandardError with context",
			err:      NewStandardErrorWithContext(ErrorTypeParse, SeverityWarning, "parse error", "test context"),
			expected: ErrorTypeParse,
		},
		{
			name:     "StandardError with cause",
			err:      NewStandardErrorWithCause(ErrorTypeValidation, SeverityWarning, "validation error", fmt.Errorf("cause")),
			expected: ErrorTypeValidation,
		},
		{
			name:     "GedcomError (not StandardError)",
			err:      &GedcomError{Severity: SeverityWarning, Message: "gedcom error"},
			expected: ErrorTypeInternal, // GedcomError defaults to Internal
		},
		{
			name:     "regular error",
			err:      fmt.Errorf("regular error"),
			expected: ErrorTypeInternal, // Regular errors default to Internal
		},
		{
			name:     "nil error",
			err:      nil,
			expected: ErrorTypeInternal, // nil errors default to Internal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorType(tt.err)
			if got != tt.expected {
				t.Errorf("GetErrorType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

