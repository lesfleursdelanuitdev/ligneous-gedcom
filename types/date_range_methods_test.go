package types

import (
	"testing"
	"time"
)

// TestNewZeroDateRange tests the NewZeroDateRange function.
func TestNewZeroDateRange(t *testing.T) {
	dr := NewZeroDateRange()
	
	if dr.start != nil || dr.end != nil {
		t.Error("NewZeroDateRange() should return DateRange with nil start and end")
	}
	
	// Test that it's a valid zero value
	if dr.IsValid() {
		t.Error("Zero DateRange should not be valid")
	}
}

// TestDateRange_StartAndEndDates tests the StartAndEndDates method.
func TestDateRange_StartAndEndDates(t *testing.T) {
	// Test with valid date range
	start, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	end, err := ParseDate("31 Dec 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	dr := NewDateRange(start, end)
	startDate, endDate := dr.StartAndEndDates()
	
	if startDate == nil || endDate == nil {
		t.Fatal("StartAndEndDates() should return non-nil dates for valid range")
	}
	
	if startDate.Year != 1900 {
		t.Errorf("Expected start year 1900, got %d", startDate.Year)
	}
	if endDate.Year != 1900 {
		t.Errorf("Expected end year 1900, got %d", endDate.Year)
	}
	
	// Test with zero DateRange (returns zero GedcomDate structs, not nil)
	dr2 := NewZeroDateRange()
	start2, end2 := dr2.StartAndEndDates()
	if start2 == nil || end2 == nil {
		t.Error("StartAndEndDates() should return non-nil (but zero) dates for zero DateRange")
	}
	if !start2.IsZero() || !end2.IsZero() {
		t.Error("StartAndEndDates() should return zero dates for zero DateRange")
	}
}

// TestDateRange_IsBefore tests the IsBefore method.
func TestDateRange_IsBefore(t *testing.T) {
	// Test with date range before another
	start1, _ := ParseDate("1 Jan 1900")
	end1, _ := ParseDate("31 Dec 1900")
	dr1 := NewDateRange(start1, end1)
	
	start2, _ := ParseDate("1 Jan 1901")
	end2, _ := ParseDate("31 Dec 1901")
	dr2 := NewDateRange(start2, end2)
	
	if !dr1.IsBefore(dr2) {
		t.Error("IsBefore() should return true when range is before another")
	}
	
	// Test with date range not before another
	if dr2.IsBefore(dr1) {
		t.Error("IsBefore() should return false when range is not before another")
	}
	
	// Test with overlapping ranges
	start3, _ := ParseDate("1 Jan 1900")
	end3, _ := ParseDate("31 Dec 1901")
	dr3 := NewDateRange(start3, end3)
	
	if dr1.IsBefore(dr3) {
		t.Error("IsBefore() should return false for overlapping ranges")
	}
}

// TestDateRange_IsAfter tests the IsAfter method.
func TestDateRange_IsAfter(t *testing.T) {
	// Test with date range after another
	start1, _ := ParseDate("1 Jan 1900")
	end1, _ := ParseDate("31 Dec 1900")
	dr1 := NewDateRange(start1, end1)
	
	start2, _ := ParseDate("1 Jan 1901")
	end2, _ := ParseDate("31 Dec 1901")
	dr2 := NewDateRange(start2, end2)
	
	if !dr2.IsAfter(dr1) {
		t.Error("IsAfter() should return true when range is after another")
	}
	
	// Test with date range not after another
	if dr1.IsAfter(dr2) {
		t.Error("IsAfter() should return false when range is not after another")
	}
	
	// Test with overlapping ranges
	start3, _ := ParseDate("1 Jan 1900")
	end3, _ := ParseDate("31 Dec 1901")
	dr3 := NewDateRange(start3, end3)
	
	if dr2.IsAfter(dr3) {
		t.Error("IsAfter() should return false for overlapping ranges")
	}
}

// TestDateRange_IsValid tests the IsValid method.
func TestDateRange_IsValid(t *testing.T) {
	// Test with valid date range
	start, _ := ParseDate("1 Jan 1900")
	end, _ := ParseDate("31 Dec 1900")
	dr := NewDateRange(start, end)
	
	if !dr.IsValid() {
		t.Error("IsValid() should return true for valid date range")
	}
	
	// Test with zero DateRange
	dr2 := NewZeroDateRange()
	if dr2.IsValid() {
		t.Error("IsValid() should return false for zero DateRange")
	}
	
	// Test with nil dates
	dr3 := NewDateRange(nil, nil)
	if dr3.IsValid() {
		t.Error("IsValid() should return false for DateRange with nil dates")
	}
	
	// Test with one nil date
	dr4 := NewDateRange(start, nil)
	if dr4.IsValid() {
		t.Error("IsValid() should return false for DateRange with one nil date")
	}
}

// TestDateRange_IsExact tests the IsExact method.
func TestDateRange_IsExact(t *testing.T) {
	// Test with exact date (same start and end)
	start, _ := ParseDate("1 Jan 1900")
	dr := NewDateRange(start, start)
	
	if !dr.IsExact() {
		t.Error("IsExact() should return true for exact date")
	}
	
	// Test with date range (both dates are exact, so range is exact)
	start2, _ := ParseDate("1 Jan 1900")
	end2, _ := ParseDate("1 Jan 1900") // Same date = exact
	dr2 := NewDateRange(start2, end2)
	
	if !dr2.IsExact() {
		t.Error("IsExact() should return true for exact date range (same start and end)")
	}
	
	// Test with different start and end dates
	start3, _ := ParseDate("1 Jan 1900")
	end3, _ := ParseDate("31 Dec 1900")
	dr3 := NewDateRange(start3, end3)
	
	if dr3.IsExact() {
		t.Error("IsExact() should return false for date range with different start and end")
	}
	
	// Test with zero DateRange
	dr3b := NewZeroDateRange()
	if dr3b.IsExact() {
		t.Error("IsExact() should return false for zero DateRange")
	}
}

// TestDateRange_IsPhrase tests the IsPhrase method.
func TestDateRange_IsPhrase(t *testing.T) {
	// Test with phrase date (enclosed in parentheses)
	dr := NewDateRangeWithString("(sometime in the 1900s)")
	if !dr.IsPhrase() {
		t.Error("IsPhrase() should return true for phrase date (enclosed in parentheses)")
	}
	
	// Test with non-phrase date
	dr2b := NewDateRangeWithString("BETWEEN 1900 AND 1901")
	if dr2b.IsPhrase() {
		t.Error("IsPhrase() should return false for non-phrase date")
	}
	
	// Test with exact date
	start, _ := ParseDate("1 Jan 1900")
	dr3 := NewDateRange(start, start)
	if dr3.IsPhrase() {
		t.Error("IsPhrase() should return false for exact date")
	}
	
	// Test with zero DateRange
	dr4 := NewZeroDateRange()
	if dr4.IsPhrase() {
		t.Error("IsPhrase() should return false for zero DateRange")
	}
}

// TestDateRange_String tests the String method.
func TestDateRange_String(t *testing.T) {
	// Test with valid date range
	start, _ := ParseDate("1 Jan 1900")
	end, _ := ParseDate("31 Dec 1900")
	dr := NewDateRange(start, end)
	
	str := dr.String()
	if str == "" {
		t.Error("String() should return non-empty string for valid date range")
	}
	
	// Test with exact date
	dr2 := NewDateRange(start, start)
	str2 := dr2.String()
	if str2 == "" {
		t.Error("String() should return non-empty string for exact date")
	}
	
	// Test with zero DateRange
	dr3 := NewZeroDateRange()
	str3 := dr3.String()
	if str3 == "" {
		t.Error("String() should return non-empty string for zero DateRange")
	}
	
	// Test with phrase date
	dr4 := NewDateRangeWithString("BETWEEN 1900 AND 1901")
	str4 := dr4.String()
	if str4 == "" {
		t.Error("String() should return non-empty string for phrase date")
	}
}

// TestDateRange_Sub tests the Sub method.
func TestDateRange_Sub(t *testing.T) {
	// Test with two valid date ranges
	start1, _ := ParseDate("1 Jan 1900")
	end1, _ := ParseDate("31 Dec 1900")
	dr1 := NewDateRange(start1, end1)
	
	start2, _ := ParseDate("1 Jan 1901")
	end2, _ := ParseDate("31 Dec 1901")
	dr2 := NewDateRange(start2, end2)
	
	minDuration, maxDuration := dr2.Sub(dr1)
	if !minDuration.IsKnown || !maxDuration.IsKnown {
		t.Error("Sub() should return known durations for valid date ranges")
	}
	// Check that durations are approximately 1 year
	expectedDuration := 365 * 24 * time.Hour
	if minDuration.Duration < expectedDuration-time.Hour || minDuration.Duration > expectedDuration+time.Hour {
		t.Errorf("Expected min duration around 1 year, got %v", minDuration.Duration)
	}
	
	// Test with same date ranges
	dr3 := NewDateRange(start1, end1)
	minDuration2, maxDuration2 := dr3.Sub(dr1)
	if !minDuration2.IsKnown || !maxDuration2.IsKnown {
		t.Error("Sub() should return known durations for same date ranges")
	}
	if minDuration2.Duration != 0 || maxDuration2.Duration != 0 {
		t.Errorf("Expected 0 duration for same date ranges, got min=%v max=%v", minDuration2.Duration, maxDuration2.Duration)
	}
	
	// Test with zero DateRange
	dr4 := NewZeroDateRange()
	minDuration3, maxDuration3 := dr4.Sub(dr1)
	if minDuration3.IsKnown || maxDuration3.IsKnown {
		t.Error("Sub() should return unknown durations when one range is zero")
	}
}

// TestDateRange_Duration tests the Duration method.
func TestDateRange_Duration(t *testing.T) {
	// Test with valid date range
	start, _ := ParseDate("1 Jan 1900")
	end, _ := ParseDate("31 Dec 1900")
	dr := NewDateRange(start, end)
	
	duration := dr.Duration()
	if !duration.IsKnown {
		t.Error("Duration() should return known duration for valid date range")
	}
	if duration.Duration <= 0 {
		t.Errorf("Expected positive duration, got %v", duration.Duration)
	}
	
	// Test with exact date (may have small duration due to time precision)
	dr2 := NewDateRange(start, start)
	duration2 := dr2.Duration()
	if !duration2.IsKnown {
		t.Error("Duration() should return known duration for exact date")
	}
	// Exact date may have a small duration (less than 1 day) due to time calculations
	if duration2.Duration > 24*time.Hour {
		t.Errorf("Expected duration < 1 day for exact date, got %v", duration2.Duration)
	}
	
	// Test with zero DateRange
	dr3 := NewZeroDateRange()
	duration3 := dr3.Duration()
	if duration3.IsKnown {
		t.Error("Duration() should return unknown duration for zero DateRange")
	}
	
	// Test with multi-year range
	start4, _ := ParseDate("1 Jan 1900")
	end4, _ := ParseDate("31 Dec 1905")
	dr4 := NewDateRange(start4, end4)
	
	duration4 := dr4.Duration()
	if !duration4.IsKnown {
		t.Error("Duration() should return known duration for multi-year range")
	}
	// 1900-1905 is approximately 6 years (including both endpoints)
	expectedMinDuration := 5 * 365 * 24 * time.Hour
	expectedMaxDuration := 6 * 366 * 24 * time.Hour // Account for leap years
	if duration4.Duration < expectedMinDuration || duration4.Duration > expectedMaxDuration {
		t.Errorf("Expected duration around 5-6 years, got %v (which is %.2f years)", 
			duration4.Duration, float64(duration4.Duration.Hours())/24/365)
	}
}

