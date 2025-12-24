package query

import (
	"strings"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// toLower converts a string to lowercase
// Simple lowercase conversion - in production, might want to use proper Unicode handling
func toLower(s string) string {
	return strings.ToLower(s)
}

// parseBirthDate parses birth date from individual record
// This is simplified - in production, use proper date parsing
// For now, return nil (will be improved)
func parseBirthDate(indi *types.IndividualRecord) *int64 {
	birthDateStr := indi.GetBirthDate()
	if birthDateStr == "" {
		return nil
	}

	// Try to parse the date
	// This is simplified - in production, use proper date parsing
	// For now, return nil (will be improved)
	return nil
}

// boolToInt converts a boolean to an integer (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

