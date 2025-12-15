package gedcom

import (
	"testing"
)

func TestParsePlace_Simple(t *testing.T) {
	place, err := ParsePlace("Rapid City")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	if !place.IsValid() {
		t.Errorf("Place should be valid")
	}

	if place.City != "Rapid City" {
		t.Errorf("City = %q, want %q", place.City, "Rapid City")
	}

	if len(place.Components) != 1 {
		t.Errorf("Components length = %d, want 1", len(place.Components))
	}
}

func TestParsePlace_CityState(t *testing.T) {
	place, err := ParsePlace("Rapid City, South Dakota")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	if !place.IsValid() {
		t.Errorf("Place should be valid")
	}

	if place.City != "Rapid City" {
		t.Errorf("City = %q, want %q", place.City, "Rapid City")
	}

	if place.State != "South Dakota" {
		t.Errorf("State = %q, want %q", place.State, "South Dakota")
	}

	if len(place.Components) != 2 {
		t.Errorf("Components length = %d, want 2", len(place.Components))
	}
}

func TestParsePlace_FullHierarchy(t *testing.T) {
	place, err := ParsePlace("Rapid City, Pennington, South Dakota, USA")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	if !place.IsValid() {
		t.Errorf("Place should be valid")
	}

	if place.City != "Rapid City" {
		t.Errorf("City = %q, want %q", place.City, "Rapid City")
	}

	if place.County != "Pennington" {
		t.Errorf("County = %q, want %q", place.County, "Pennington")
	}

	if place.State != "South Dakota" {
		t.Errorf("State = %q, want %q", place.State, "South Dakota")
	}

	if place.Country != "USA" {
		t.Errorf("Country = %q, want %q", place.Country, "USA")
	}

	if len(place.Components) != 4 {
		t.Errorf("Components length = %d, want 4", len(place.Components))
	}
}

func TestParsePlace_WithAbbreviations(t *testing.T) {
	place, err := ParsePlace("New York, NY, USA")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	if !place.IsValid() {
		t.Errorf("Place should be valid")
	}

	if place.City != "New York" {
		t.Errorf("City = %q, want %q", place.City, "New York")
	}

	if place.State != "NY" {
		t.Errorf("State = %q, want %q", place.State, "NY")
	}

	if place.Country != "USA" {
		t.Errorf("Country = %q, want %q", place.Country, "USA")
	}
}

func TestGedcomPlace_ToFormatted(t *testing.T) {
	place, err := ParsePlace("Rapid City, South Dakota")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	formatted := place.ToFormatted(", ")
	if formatted != "Rapid City, South Dakota" {
		t.Errorf("ToFormatted() = %q, want %q", formatted, "Rapid City, South Dakota")
	}

	formatted2 := place.ToFormatted(" | ")
	if formatted2 != "Rapid City | South Dakota" {
		t.Errorf("ToFormatted() = %q, want %q", formatted2, "Rapid City | South Dakota")
	}
}

func TestGedcomPlace_GetComponent(t *testing.T) {
	place, err := ParsePlace("Rapid City, Pennington, South Dakota, USA")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	if place.GetComponent(0) != "Rapid City" {
		t.Errorf("GetComponent(0) = %q, want %q", place.GetComponent(0), "Rapid City")
	}

	if place.GetComponent(1) != "Pennington" {
		t.Errorf("GetComponent(1) = %q, want %q", place.GetComponent(1), "Pennington")
	}

	if place.GetComponent(2) != "South Dakota" {
		t.Errorf("GetComponent(2) = %q, want %q", place.GetComponent(2), "South Dakota")
	}

	if place.GetComponent(3) != "USA" {
		t.Errorf("GetComponent(3) = %q, want %q", place.GetComponent(3), "USA")
	}

	if place.GetComponent(4) != "" {
		t.Errorf("GetComponent(4) = %q, want empty string", place.GetComponent(4))
	}
}

func TestGedcomPlace_Normalize(t *testing.T) {
	place, err := ParsePlace("  Rapid City  ,  South Dakota  ")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	normalized := place.Normalize()

	if normalized.City != "Rapid City" {
		t.Errorf("Normalized City = %q, want %q", normalized.City, "Rapid City")
	}

	if normalized.State != "South Dakota" {
		t.Errorf("Normalized State = %q, want %q", normalized.State, "South Dakota")
	}
}

func TestGedcomPlace_String(t *testing.T) {
	place, err := ParsePlace("Rapid City, South Dakota")
	if err != nil {
		t.Fatalf("ParsePlace failed: %v", err)
	}

	str := place.String()
	if str != "Rapid City, South Dakota" {
		t.Errorf("String() = %q, want %q", str, "Rapid City, South Dakota")
	}
}

func TestParsePlace_EmptyString(t *testing.T) {
	_, err := ParsePlace("")
	if err == nil {
		t.Error("ParsePlace should return error for empty string")
	}
}
