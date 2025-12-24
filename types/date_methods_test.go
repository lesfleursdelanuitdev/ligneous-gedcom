package types

import (
	"testing"
	"time"
)

// TestGedcomDate_Latest tests the Latest method.
func TestGedcomDate_Latest(t *testing.T) {
	// Test with exact date
	date, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	latest := date.Latest()
	if latest.IsZero() {
		t.Error("Latest() should return non-zero time for valid date")
	}
	if latest.Year() != 1900 {
		t.Errorf("Expected year 1900, got %d", latest.Year())
	}
	
	// Test with range date
	date2, err := ParseDate("BET 1 Jan 1900 AND 31 Dec 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	latest2 := date2.Latest()
	if latest2.IsZero() {
		t.Error("Latest() should return non-zero time for range date")
	}
	if latest2.Year() != 1900 {
		t.Errorf("Expected year 1900, got %d", latest2.Year())
	}
	
	// Test with "before" date
	date3, err := ParseDate("BEF 1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	latest3 := date3.Latest()
	if latest3.IsZero() {
		t.Error("Latest() should return non-zero time for 'before' date")
	}
	
	// Test with "after" date
	date4, err := ParseDate("AFT 1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	latest4 := date4.Latest()
	if latest4.IsZero() {
		t.Error("Latest() should return non-zero time for 'after' date")
	}
	if latest4.Year() != 9999 {
		t.Errorf("Expected year 9999 for 'after' date, got %d", latest4.Year())
	}
	
	// Test with invalid date
	date5 := &GedcomDate{}
	latest5 := date5.Latest()
	if !latest5.IsZero() {
		t.Error("Latest() should return zero time for invalid date")
	}
	
	// Test with year only
	date6, err := ParseDate("1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	latest6 := date6.Latest()
	if latest6.IsZero() {
		t.Error("Latest() should return non-zero time for year-only date")
	}
	if latest6.Month() != 12 {
		t.Errorf("Expected month 12 for year-only date, got %d", latest6.Month())
	}
	if latest6.Day() != 31 {
		t.Errorf("Expected day 31 for year-only date, got %d", latest6.Day())
	}
}

// TestGedcomDate_Sub tests the Sub method.
func TestGedcomDate_Sub(t *testing.T) {
	// Test with two valid dates
	date1, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	date2, err := ParseDate("1 Jan 1901")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	duration := date2.Sub(date1)
	if !duration.IsKnown {
		t.Error("Sub() should return known duration for valid dates")
	}
	// Duration uses time.Duration, check that it's approximately 1 year
	expectedDuration := 365 * 24 * time.Hour
	if duration.Duration < expectedDuration-time.Hour || duration.Duration > expectedDuration+time.Hour {
		t.Errorf("Expected duration around 1 year, got %v", duration.Duration)
	}
	
	// Test with same dates
	date3, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	duration2 := date3.Sub(date3)
	if !duration2.IsKnown {
		t.Error("Sub() should return known duration for same dates")
	}
	if duration2.Duration != 0 {
		t.Errorf("Expected 0 duration for same dates, got %v", duration2.Duration)
	}
	
	// Test with invalid dates
	date4 := &GedcomDate{}
	date5 := &GedcomDate{}
	
	duration3 := date4.Sub(date5)
	if duration3.IsKnown {
		t.Error("Sub() should return unknown duration for invalid dates")
	}
	
	// Test with one invalid date
	date6, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	duration4 := date6.Sub(date4)
	if duration4.IsKnown {
		t.Error("Sub() should return unknown duration when one date is invalid")
	}
	
	// Test with reversed dates (later - earlier)
	date7, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	date8, err := ParseDate("1 Jan 1905")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	duration5 := date7.Sub(date8) // Should be negative
	if !duration5.IsKnown {
		t.Error("Sub() should return known duration for reversed dates")
	}
	if duration5.Duration >= 0 {
		t.Errorf("Expected negative duration, got %v", duration5.Duration)
	}
}

// TestGedcomDate_equalsC tests the equalsC helper function.
func TestGedcomDate_equalsC(t *testing.T) {
	// Test with d1.Years() < d2.Years()
	date1, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	date2, err := ParseDate("1 Jan 1901")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	if !equalsC(date1, date2) {
		t.Error("equalsC() should return true when d1.Years() < d2.Years()")
	}
	
	// Test with d1.Years() >= d2.Years()
	date3, err := ParseDate("1 Jan 1901")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	if equalsC(date3, date2) {
		t.Error("equalsC() should return false when d1.Years() >= d2.Years()")
	}
}

// TestGedcomDate_equalsD tests the equalsD helper function.
func TestGedcomDate_equalsD(t *testing.T) {
	// equalsD should always return false
	date1, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	date2, err := ParseDate("1 Jan 1900")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	if equalsD(date1, date2) {
		t.Error("equalsD() should always return false")
	}
	
	// Test with different dates
	date3, err := ParseDate("1 Jan 1901")
	if err != nil {
		t.Fatalf("ParseDate() returned error: %v", err)
	}
	
	if equalsD(date1, date3) {
		t.Error("equalsD() should always return false")
	}
}

