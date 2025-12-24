package types

import (
	"fmt"
	"strings"
)

// GedcomPlace represents a parsed GEDCOM place with hierarchical components.
type GedcomPlace struct {
	Original   string   // Original GEDCOM place string
	Components []string // Parsed components (from most specific to least specific)

	// Hierarchical components (extracted from Components)
	City       string
	County     string
	State      string
	Country    string
	PostalCode string

	// Geographic data (optional, for future geocoding support)
	Latitude  float64
	Longitude float64

	// Parsed status
	IsParsed   bool
	ParseError error
}

// ParsePlace parses a GEDCOM place string and extracts hierarchical components.
// Supports various place formats:
//   - "Rapid City" (simple)
//   - "Rapid City, South Dakota" (city, state)
//   - "Rapid City, Pennington, South Dakota, USA" (full hierarchy)
//   - "New York, NY, USA" (with abbreviations)
//
// GEDCOM places are typically comma-separated, with the most specific
// location first, followed by progressively broader locations.
func ParsePlace(placeStr string) (*GedcomPlace, error) {
	if placeStr == "" {
		return nil, fmt.Errorf("empty place string")
	}

	place := &GedcomPlace{
		Original:   strings.TrimSpace(placeStr),
		IsParsed:   false,
		ParseError: nil,
	}

	// Split by comma and trim whitespace
	parts := strings.Split(placeStr, ",")
	components := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			components = append(components, trimmed)
		}
	}

	if len(components) == 0 {
		place.ParseError = fmt.Errorf("no components found in place string")
		return place, place.ParseError
	}

	place.Components = components

	// Extract hierarchical components
	// GEDCOM format: most specific to least specific (e.g., City, County, State, Country)
	switch len(components) {
	case 1:
		// Single component - assume it's a city
		place.City = components[0]
	case 2:
		// Two components - typically City, State or City, Country
		place.City = components[0]
		// Try to determine if second is state or country
		if isLikelyCountry(components[1]) {
			place.Country = components[1]
		} else {
			place.State = components[1]
		}
	case 3:
		// Three components - typically City, State, Country
		place.City = components[0]
		place.State = components[1]
		place.Country = components[2]
	case 4:
		// Four components - typically City, County, State, Country
		place.City = components[0]
		place.County = components[1]
		place.State = components[2]
		place.Country = components[3]
	default:
		// Five or more components
		place.City = components[0]
		if len(components) >= 2 {
			place.County = components[1]
		}
		if len(components) >= 3 {
			place.State = components[2]
		}
		if len(components) >= 4 {
			place.Country = components[3]
		}
		// Additional components beyond 4 are stored in Components but not mapped
	}

	place.IsParsed = true
	return place, nil
}

// isLikelyCountry attempts to determine if a component is likely a country.
// This is a simple heuristic - could be enhanced with a country list.
func isLikelyCountry(component string) bool {
	upper := strings.ToUpper(component)
	// Common country indicators
	countryIndicators := []string{"USA", "US", "UNITED STATES", "UK", "UNITED KINGDOM", "CANADA", "AUSTRALIA", "FRANCE", "GERMANY", "ITALY", "SPAIN"}
	for _, indicator := range countryIndicators {
		if upper == indicator || strings.Contains(upper, indicator) {
			return true
		}
	}
	return false
}

// ToFormatted formats the place with the given separator.
// Returns the original string if parsing failed.
func (gp *GedcomPlace) ToFormatted(separator string) string {
	if !gp.IsParsed || len(gp.Components) == 0 {
		return gp.Original
	}
	return strings.Join(gp.Components, separator)
}

// GetComponent returns the component at the specified level.
// Level 0 is the most specific (city), higher levels are broader.
// Returns empty string if level is out of range.
func (gp *GedcomPlace) GetComponent(level int) string {
	if !gp.IsParsed || level < 0 || level >= len(gp.Components) {
		return ""
	}
	return gp.Components[level]
}

// IsValid returns true if the place was successfully parsed.
func (gp *GedcomPlace) IsValid() bool {
	return gp.IsParsed && gp.ParseError == nil
}

// Normalize returns a normalized version of the place.
// Currently just trims and standardizes capitalization.
// Could be enhanced with place name standardization.
func (gp *GedcomPlace) Normalize() *GedcomPlace {
	if !gp.IsValid() {
		return gp
	}

	normalized := &GedcomPlace{
		Original:   gp.Original,
		Components: make([]string, len(gp.Components)),
		City:       strings.TrimSpace(gp.City),
		County:     strings.TrimSpace(gp.County),
		State:      strings.TrimSpace(gp.State),
		Country:    strings.TrimSpace(gp.Country),
		PostalCode: strings.TrimSpace(gp.PostalCode),
		IsParsed:   true,
	}

	// Normalize components
	for i, comp := range gp.Components {
		normalized.Components[i] = strings.TrimSpace(comp)
	}

	return normalized
}

// String returns a string representation of the place.
func (gp *GedcomPlace) String() string {
	if !gp.IsValid() {
		return gp.Original
	}
	return gp.ToFormatted(", ")
}

// Geocode is a placeholder for future geocoding functionality.
// Would look up latitude/longitude for the place.
func (gp *GedcomPlace) Geocode() error {
	// TODO: Implement geocoding using a geocoding service
	// For now, just return nil (no-op)
	return nil
}
