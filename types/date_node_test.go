package types

import (
	"testing"
)

// TestDateNode_StartDate tests the StartDate method.
func TestDateNode_StartDate(t *testing.T) {
	// Test with valid date
	dn := NewDateNode("1 Jan 1900")
	if !dn.IsValid() {
		t.Fatal("DateNode should be valid")
	}
	
	startDate := dn.StartDate()
	if startDate == nil {
		t.Fatal("StartDate() should return non-nil for valid date")
	}
	if startDate.Year != 1900 {
		t.Errorf("Expected year 1900, got %d", startDate.Year)
	}
	
	// Test with nil DateNode
	var nilDN *DateNode
	startDate2 := nilDN.StartDate()
	if startDate2 != nil {
		t.Error("StartDate() should return nil for nil DateNode")
	}
	
	// Test with invalid date
	dn3 := NewDateNode("invalid date")
	startDate3 := dn3.StartDate()
	if startDate3 != nil {
		t.Error("StartDate() should return nil for invalid date")
	}
}

// TestDateNode_EndDate tests the EndDate method.
func TestDateNode_EndDate(t *testing.T) {
	// Test with valid date
	dn := NewDateNode("1 Jan 1900")
	if !dn.IsValid() {
		t.Fatal("DateNode should be valid")
	}
	
	endDate := dn.EndDate()
	if endDate == nil {
		t.Fatal("EndDate() should return non-nil for valid date")
	}
	
	// Test with nil DateNode
	var nilDN *DateNode
	endDate2 := nilDN.EndDate()
	if endDate2 != nil {
		t.Error("EndDate() should return nil for nil DateNode")
	}
	
	// Test with invalid date
	dn3 := NewDateNode("invalid date")
	endDate3 := dn3.EndDate()
	if endDate3 != nil {
		t.Error("EndDate() should return nil for invalid date")
	}
}

// TestDateNode_StartAndEndDates tests the StartAndEndDates method.
func TestDateNode_StartAndEndDates(t *testing.T) {
	// Test with valid date
	dn := NewDateNode("1 Jan 1900")
	if !dn.IsValid() {
		t.Fatal("DateNode should be valid")
	}
	
	start, end := dn.StartAndEndDates()
	if start == nil || end == nil {
		t.Fatal("StartAndEndDates() should return non-nil for valid date")
	}
	if start.Year != 1900 {
		t.Errorf("Expected start year 1900, got %d", start.Year)
	}
	
	// Test with nil DateNode
	var nilDN *DateNode
	start2, end2 := nilDN.StartAndEndDates()
	if start2 != nil || end2 != nil {
		t.Error("StartAndEndDates() should return nil, nil for nil DateNode")
	}
	
	// Test with invalid date
	dn3 := NewDateNode("invalid date")
	start3, end3 := dn3.StartAndEndDates()
	if start3 != nil || end3 != nil {
		t.Error("StartAndEndDates() should return nil, nil for invalid date")
	}
}

// TestDateNode_Years tests the Years method.
func TestDateNode_Years(t *testing.T) {
	// Test with valid date
	dn := NewDateNode("1 Jan 1900")
	if !dn.IsValid() {
		t.Fatal("DateNode should be valid")
	}
	
	years := dn.Years()
	// Use tolerance for floating point comparison (allow small precision errors)
	expected := 1900.0
	tolerance := 0.01
	if years < expected-tolerance || years > expected+tolerance {
		t.Errorf("Expected approximately 1900.0, got %f", years)
	}
	
	// Test with nil DateNode
	var nilDN *DateNode
	years2 := nilDN.Years()
	if years2 != 0.0 {
		t.Errorf("Expected 0.0 for nil DateNode, got %f", years2)
	}
	
	// Test with invalid date
	dn3 := NewDateNode("invalid date")
	years3 := dn3.Years()
	if years3 != 0.0 {
		t.Errorf("Expected 0.0 for invalid date, got %f", years3)
	}
}

// TestDateNode_IsExact tests the IsExact method.
func TestDateNode_IsExact(t *testing.T) {
	// Test with exact date
	dn := NewDateNode("1 Jan 1900")
	if !dn.IsValid() {
		t.Fatal("DateNode should be valid")
	}
	
	if !dn.IsExact() {
		t.Error("IsExact() should return true for exact date")
	}
	
	// Test with approximate date
	dn2 := NewDateNode("ABT 1900")
	if !dn2.IsValid() {
		t.Fatal("DateNode should be valid")
	}
	
	if dn2.IsExact() {
		t.Error("IsExact() should return false for approximate date")
	}
	
	// Test with nil DateNode
	var nilDN *DateNode
	if nilDN.IsExact() {
		t.Error("IsExact() should return false for nil DateNode")
	}
	
	// Test with invalid date
	dn3 := NewDateNode("invalid date")
	if dn3.IsExact() {
		t.Error("IsExact() should return false for invalid date")
	}
}

// TestDateNode_Equals tests the Equals method.
func TestDateNode_Equals(t *testing.T) {
	// Test with equal dates
	dn1 := NewDateNode("1 Jan 1900")
	dn2 := NewDateNode("1 Jan 1900")
	
	if !dn1.Equals(dn2) {
		t.Error("Equals() should return true for equal dates")
	}
	
	// Test with different dates
	dn3 := NewDateNode("1 Jan 1900")
	dn4 := NewDateNode("1 Jan 1901")
	
	if dn3.Equals(dn4) {
		t.Error("Equals() should return false for different dates")
	}
	
	// Test with nil DateNodes
	var nilDN1, nilDN2 *DateNode
	if !nilDN1.Equals(nilDN2) {
		t.Error("Equals() should return true for both nil DateNodes")
	}
	
	if nilDN1.Equals(dn1) {
		t.Error("Equals() should return false when one is nil and other is not")
	}
	
	if dn1.Equals(nilDN1) {
		t.Error("Equals() should return false when one is nil and other is not")
	}
	
	// Test with invalid dates
	dn5 := NewDateNode("invalid date")
	dn6 := NewDateNode("invalid date")
	
	if dn5.Equals(dn6) {
		t.Error("Equals() should return false for invalid dates")
	}
}

// TestDateNode_Similarity tests the Similarity method.
func TestDateNode_Similarity(t *testing.T) {
	// Test with similar dates
	dn1 := NewDateNode("1 Jan 1900")
	dn2 := NewDateNode("1 Jan 1901")
	
	similarity := dn1.Similarity(dn2, 10.0)
	if similarity <= 0.0 {
		t.Errorf("Similarity() should return > 0 for similar dates, got %f", similarity)
	}
	
	// Test with very different dates
	dn3 := NewDateNode("1 Jan 1900")
	dn4 := NewDateNode("1 Jan 2000")
	
	similarity2 := dn3.Similarity(dn4, 10.0)
	if similarity2 >= 0.5 {
		t.Errorf("Similarity() should return < 0.5 for very different dates, got %f", similarity2)
	}
	
	// Test with nil DateNodes
	var nilDN1, nilDN2 *DateNode
	similarity3 := nilDN1.Similarity(nilDN2, 10.0)
	if similarity3 != 0.5 {
		t.Errorf("Similarity() should return 0.5 for both nil DateNodes, got %f", similarity3)
	}
	
	// Test with one nil DateNode
	similarity4 := nilDN1.Similarity(dn1, 10.0)
	if similarity4 != 0.5 {
		t.Errorf("Similarity() should return 0.5 when one is nil, got %f", similarity4)
	}
	
	// Test with invalid dates
	dn5 := NewDateNode("invalid date")
	dn6 := NewDateNode("invalid date")
	
	similarity5 := dn5.Similarity(dn6, 10.0)
	if similarity5 != 0.5 {
		t.Errorf("Similarity() should return 0.5 for invalid dates, got %f", similarity5)
	}
}

// TestDateNode_String tests the String method.
func TestDateNode_String(t *testing.T) {
	// Test with valid date
	dn := NewDateNode("1 Jan 1900")
	str := dn.String()
	if str == "" {
		t.Error("String() should return non-empty string for valid date")
	}
	
	// Test with nil DateNode
	var nilDN *DateNode
	str2 := nilDN.String()
	if str2 != "" {
		t.Errorf("String() should return empty string for nil DateNode, got %q", str2)
	}
	
	// Test with invalid date (should return original)
	dn3 := NewDateNode("invalid date")
	str3 := dn3.String()
	if str3 != "invalid date" {
		t.Errorf("String() should return original string for invalid date, got %q", str3)
	}
}

