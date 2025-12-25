package parser

import (
	"os"
	"path/filepath"
)

// findTestDataFile attempts to locate a testdata file using multiple possible paths.
// Returns the first path that exists, or empty string if none found.
func findTestDataFile(filename string) string {
	possiblePaths := []string{
		// Absolute path (for CI/CD)
		filepath.Join("/apps/gedcom-go/testdata", filename),
		// Relative from parser package
		filepath.Join("testdata", filename),
		filepath.Join("../testdata", filename),
		filepath.Join("../../testdata", filename),
		// Legacy paths (for backward compatibility)
		filepath.Join("../../../family-tree/gedcom", filename),
		filepath.Join("../../../family-tree/flask-backend/gedcom", filename),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

