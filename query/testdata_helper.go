package query

import (
	"os"
	"path/filepath"
	"runtime"
)

// findTestDataFile attempts to locate a test data file given its name.
// It checks multiple common paths relative to the current working directory
// and the Go module root.
func findTestDataFile(filename string) string {
	// Possible paths to check
	possiblePaths := []string{
		filepath.Join("testdata", filename),
		filepath.Join("../testdata", filename),
		filepath.Join("../../testdata", filename),
		filepath.Join("../../../testdata", filename),
		filepath.Join("/apps/gedcom-go/testdata", filename), // Absolute path for specific environments
	}

	// Get current file's directory (for relative paths)
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		currentDir := filepath.Dir(currentFile)
		possiblePaths = append(possiblePaths,
			filepath.Join(currentDir, "testdata", filename),
			filepath.Join(currentDir, "..", "testdata", filename),
			filepath.Join(currentDir, "..", "..", "testdata", filename),
			filepath.Join(currentDir, "..", "..", "..", "testdata", filename),
		)
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

