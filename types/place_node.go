package types

import "strings"

// PlaceNode represents a structured place with jurisdictional components.
// Similar to elliotchance's PlaceNode but adapted for gedcom-go's structure.
type PlaceNode struct {
	// Original is the original place string from GEDCOM
	Original string

	// Components are the parsed jurisdictional components
	// Format: Name, County, State, Country
	Components []string

	// Name is the first component (city/town name)
	Name string

	// County is the second component
	County string

	// State is the third component (state/province)
	State string

	// Country is the fourth component
	Country string

	// Latitude and Longitude if available
	Latitude  string
	Longitude string

	// OriginalLine is the original GedcomLine for accessing sub-tags
	OriginalLine *GedcomLine
}

// NewPlaceNode creates a new PlaceNode from a place string.
func NewPlaceNode(placeStr string) *PlaceNode {
	if placeStr == "" {
		return &PlaceNode{}
	}

	pn := &PlaceNode{
		Original: strings.TrimSpace(placeStr),
	}

	// Parse jurisdictional components (comma-separated)
	components := strings.Split(placeStr, ",")
	for i, comp := range components {
		components[i] = strings.TrimSpace(comp)
	}
	pn.Components = components

	// Extract components
	if len(components) > 0 {
		pn.Name = components[0]
	}
	if len(components) > 1 {
		pn.County = components[1]
	}
	if len(components) > 2 {
		pn.State = components[2]
	}
	if len(components) > 3 {
		pn.Country = components[3]
	}

	return pn
}

// NewPlaceNodeFromLine creates a PlaceNode from a GedcomLine (PLAC tag).
func NewPlaceNodeFromLine(line *GedcomLine) *PlaceNode {
	if line == nil || line.Tag != "PLAC" {
		return nil
	}

	pn := NewPlaceNode(line.Value)

	// Extract latitude/longitude if present
	if latLines := line.GetLines("LATI"); len(latLines) > 0 {
		pn.Latitude = latLines[0].Value
	}
	if lonLines := line.GetLines("LONG"); len(lonLines) > 0 {
		pn.Longitude = lonLines[0].Value
	}

	// Extract FORM sub-tag if present (format specification)
	// Store original line for additional sub-tags
	pn.OriginalLine = line

	return pn
}

// IsValid returns true if the place has at least a name.
func (pn *PlaceNode) IsValid() bool {
	return pn != nil && pn.Name != ""
}

// JurisdictionalName returns the full jurisdictional name.
// Returns the original string if components aren't properly formatted.
func (pn *PlaceNode) JurisdictionalName() string {
	if pn == nil {
		return ""
	}

	if len(pn.Components) >= 4 {
		// Properly formatted: Name,County,State,Country
		return strings.Join(pn.Components, ",")
	}

	// Return original if not in standard format
	return pn.Original
}

// String returns the place as a string.
func (pn *PlaceNode) String() string {
	if pn == nil {
		return ""
	}
	return pn.Original
}

// Format returns the place formatted with the specified separator.
func (pn *PlaceNode) Format(separator string) string {
	if pn == nil {
		return ""
	}
	return strings.Join(pn.Components, separator)
}

