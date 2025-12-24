package types

import (
	"fmt"
	"strings"
)

// NameType represents the type of a GEDCOM name.
type NameType string

const (
	NameTypeBirth     NameType = "birth"
	NameTypeMarried   NameType = "married"
	NameTypeAka       NameType = "aka"
	NameTypeReligious NameType = "religious"
	NameTypeOther     NameType = "other"
	NameTypeUnknown   NameType = "unknown"
)

// GedcomName represents a parsed GEDCOM name with structured components.
// Follows GEDCOM 5.5.1 specification for NAME records.
type GedcomName struct {
	Original string   // Original GEDCOM name string
	Type     NameType // birth, married, aka, religious, etc.

	// Name components (GEDCOM 5.5.1 sub-tags)
	Prefix        string // NPFX: Dr., Mr., Mrs., etc.
	Given         string // GIVN: First/middle names
	Nickname      string // NICK: Nickname
	SurnamePrefix string // SPFX: van, de, la, etc.
	Surname       string // SURN: Last name
	Suffix        string // NSFX: Jr., Sr., III, etc.

	// Parsed status
	IsParsed   bool
	ParseError error
}

// ParseName parses a GEDCOM NAME line and returns a GedcomName.
// Takes a GedcomLine (NAME record) as input to access sub-tags.
// Supports full GEDCOM 5.5.1 specification:
//   - Sub-tags: NPFX, GIVN, NICK, SPFX, SURN, NSFX, TYPE
//   - NAME value parsing when sub-tags are missing
//   - Multiple name formats: "Given /Surname/", "Given Surname", etc.
func ParseName(nameLine *GedcomLine) (*GedcomName, error) {
	if nameLine == nil {
		return nil, fmt.Errorf("name line is nil")
	}

	if nameLine.Tag != "NAME" {
		return nil, fmt.Errorf("expected NAME tag, got %s", nameLine.Tag)
	}

	name := &GedcomName{
		Original:   strings.TrimSpace(nameLine.Value),
		Type:       NameTypeUnknown,
		IsParsed:   false,
		ParseError: nil,
	}

	// Extract sub-tags (GEDCOM 5.5.1 specification)
	if prefixLines := nameLine.GetLines("NPFX"); len(prefixLines) > 0 {
		name.Prefix = strings.TrimSpace(prefixLines[0].Value)
	}

	if givenLines := nameLine.GetLines("GIVN"); len(givenLines) > 0 {
		name.Given = strings.TrimSpace(givenLines[0].Value)
	}

	if nickLines := nameLine.GetLines("NICK"); len(nickLines) > 0 {
		name.Nickname = strings.TrimSpace(nickLines[0].Value)
	}

	if spfxLines := nameLine.GetLines("SPFX"); len(spfxLines) > 0 {
		name.SurnamePrefix = strings.TrimSpace(spfxLines[0].Value)
	}

	if surnLines := nameLine.GetLines("SURN"); len(surnLines) > 0 {
		name.Surname = strings.TrimSpace(surnLines[0].Value)
	}

	if suffixLines := nameLine.GetLines("NSFX"); len(suffixLines) > 0 {
		name.Suffix = strings.TrimSpace(suffixLines[0].Value)
	}

	// Extract name type
	if typeLines := nameLine.GetLines("TYPE"); len(typeLines) > 0 {
		typeStr := strings.ToLower(strings.TrimSpace(typeLines[0].Value))
		switch typeStr {
		case "birth":
			name.Type = NameTypeBirth
		case "married":
			name.Type = NameTypeMarried
		case "aka":
			name.Type = NameTypeAka
		case "religious":
			name.Type = NameTypeReligious
		case "unknown":
			name.Type = NameTypeUnknown
		case "":
			// Empty type, keep NameTypeUnknown (default)
		default:
			// Unknown type value, treat as "other"
			name.Type = NameTypeOther
		}
	}

	// If sub-tags are missing, try to parse from NAME value
	if name.Given == "" && name.Surname == "" && name.Original != "" {
		parseNameValue(name, name.Original)
	}

	// Validate
	if !name.IsValid() {
		name.ParseError = fmt.Errorf("name has no given name or surname")
		return name, name.ParseError
	}

	name.IsParsed = true
	return name, nil
}

// parseNameValue parses the NAME value string when sub-tags are missing.
// Handles formats like "Given /Surname/", "Dr. Given /Surname/ Jr.", etc.
func parseNameValue(name *GedcomName, nameStr string) {
	nameStr = strings.TrimSpace(nameStr)
	if nameStr == "" {
		return
	}

	// Look for /Surname/ pattern (standard GEDCOM format)
	startIdx := -1
	endIdx := -1
	for i, r := range nameStr {
		if r == '/' {
			if startIdx == -1 {
				startIdx = i + 1
			} else {
				endIdx = i
				break
			}
		}
	}

	if startIdx > 0 && endIdx > startIdx {
		// Extract surname between slashes
		name.Surname = strings.TrimSpace(nameStr[startIdx:endIdx])

		// Extract given name (everything before the first slash)
		givenPart := strings.TrimSpace(nameStr[:startIdx-1])
		if givenPart != "" {
			// Try to extract prefix and suffix from given part
			parts := strings.Fields(givenPart)
			if len(parts) > 0 {
			// Check if first part is a prefix (common prefixes)
			firstPart := strings.ToUpper(parts[0])
			commonPrefixes := []string{"DR", "DR.", "MR", "MR.", "MRS", "MRS.", "MS", "MS.", "PROF", "PROF.", "REV", "REV."}
			for _, prefix := range commonPrefixes {
				if firstPart == prefix || strings.HasPrefix(firstPart, prefix) {
					name.Prefix = parts[0]
					parts = parts[1:]
					break
				}
			}

			// Check if last part is a suffix (common suffixes)
			if len(parts) > 0 {
				lastPart := strings.ToUpper(parts[len(parts)-1])
				commonSuffixes := []string{"JR", "JR.", "SR", "SR.", "II", "III", "IV", "V"}
				for _, suffix := range commonSuffixes {
					if lastPart == suffix {
						name.Suffix = parts[len(parts)-1]
						parts = parts[:len(parts)-1]
						break
					}
				}

					// Remaining parts are the given name
					if len(parts) > 0 {
						name.Given = strings.Join(parts, " ")
					}
				}
			}
		}
	} else {
		// No slashes found, try to parse as unstructured name
		// Simple heuristic: split on spaces, assume last word is surname
		parts := strings.Fields(nameStr)
		if len(parts) > 1 {
			// Last part might be surname
			name.Surname = parts[len(parts)-1]
			// Everything else is given name
			name.Given = strings.Join(parts[:len(parts)-1], " ")
		} else if len(parts) == 1 {
			// Single word - assume it's given name
			name.Given = parts[0]
		}
	}
}

// reconstructFullName reconstructs the full name from components.
// Format: "Prefix Given SurnamePrefix Surname Suffix"
func (gn *GedcomName) reconstructFullName() string {
	parts := make([]string, 0, 6)

	if gn.Prefix != "" {
		parts = append(parts, gn.Prefix)
	}

	if gn.Given != "" {
		parts = append(parts, gn.Given)
	}

	if gn.SurnamePrefix != "" {
		parts = append(parts, gn.SurnamePrefix)
	}

	if gn.Surname != "" {
		parts = append(parts, gn.Surname)
	}

	if gn.Suffix != "" {
		parts = append(parts, gn.Suffix)
	}

	if len(parts) == 0 {
		// Fallback to original if no components
		return gn.Original
	}

	return strings.Join(parts, " ")
}

// FullName returns the reconstructed full name.
// If reconstruction failed, returns the original name string.
func (gn *GedcomName) FullName() string {
	fullName := gn.reconstructFullName()
	if fullName != "" {
		return fullName
	}
	return gn.Original
}

// IsValid returns true if the name has at least a given name or surname.
func (gn *GedcomName) IsValid() bool {
	return (gn.Given != "" || gn.Surname != "") && gn.ParseError == nil
}

// String returns a string representation of the name.
func (gn *GedcomName) String() string {
	return gn.FullName()
}

// GetGivenName returns the given name (first/middle names).
func (gn *GedcomName) GetGivenName() string {
	return gn.Given
}

// GetSurname returns the surname (last name).
func (gn *GedcomName) GetSurname() string {
	return gn.Surname
}

// GetFullSurname returns the full surname including prefix (e.g., "van der Berg").
func (gn *GedcomName) GetFullSurname() string {
	if gn.SurnamePrefix != "" && gn.Surname != "" {
		return gn.SurnamePrefix + " " + gn.Surname
	}
	if gn.SurnamePrefix != "" {
		return gn.SurnamePrefix
	}
	return gn.Surname
}

// HasPrefix returns true if the name has a prefix (NPFX).
func (gn *GedcomName) HasPrefix() bool {
	return gn.Prefix != ""
}

// HasSuffix returns true if the name has a suffix (NSFX).
func (gn *GedcomName) HasSuffix() bool {
	return gn.Suffix != ""
}

// HasNickname returns true if the name has a nickname (NICK).
func (gn *GedcomName) HasNickname() bool {
	return gn.Nickname != ""
}

// HasSurnamePrefix returns true if the name has a surname prefix (SPFX).
func (gn *GedcomName) HasSurnamePrefix() bool {
	return gn.SurnamePrefix != ""
}

