package parser

import (
	"os"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/pkg/gedcom"
)

// SmartParser automatically selects the best parser based on file size.
// For small files (< 32KB), it uses HierarchicalParser to avoid goroutine overhead.
// For larger files (>= 32KB), it uses ParallelHierarchicalParser for better performance.
// This threshold is based on performance analysis showing parallel overhead dominates
// below 32KB, while parallel parser is 12-22% faster on files >= 400KB.
type SmartParser struct {
	parser ParserInterface
}

// ParserInterface defines the common interface for parsers
type ParserInterface interface {
	Parse(filePath string) (*gedcom.GedcomTree, error)
	GetErrors() []*gedcom.GedcomError
	HasErrors() bool
	GetErrorManager() *gedcom.ErrorManager
	GetTree() *gedcom.GedcomTree
}

// NewSmartParser creates a parser that automatically selects the best implementation
// based on file size. This optimizes for both small and large files.
func NewSmartParser() *SmartParser {
	return &SmartParser{}
}

// Parse automatically selects and uses the best parser for the file size.
func (sp *SmartParser) Parse(filePath string) (*gedcom.GedcomTree, error) {
	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()
	const smallFileThreshold = 32 * 1024 // 32KB - auto-fallback to avoid parallel overhead

	// For small files, use regular parser to avoid goroutine overhead
	// For larger files, use parallel parser for better performance
	// Threshold set to 32KB based on performance analysis showing parallel overhead
	// dominates below this size, while parallel parser is 12-22% faster on larger files
	if fileSize < smallFileThreshold {
		sp.parser = NewHierarchicalParser()
	} else {
		sp.parser = NewParallelHierarchicalParser()
	}

	return sp.parser.Parse(filePath)
}

// GetErrors returns all errors collected during parsing
func (sp *SmartParser) GetErrors() []*gedcom.GedcomError {
	if sp.parser == nil {
		return nil
	}
	return sp.parser.GetErrors()
}

// HasErrors returns true if any errors were encountered
func (sp *SmartParser) HasErrors() bool {
	if sp.parser == nil {
		return false
	}
	return sp.parser.HasErrors()
}

// GetErrorManager returns the error manager
func (sp *SmartParser) GetErrorManager() *gedcom.ErrorManager {
	if sp.parser == nil {
		return nil
	}
	return sp.parser.GetErrorManager()
}

// GetTree returns the parsed tree
func (sp *SmartParser) GetTree() *gedcom.GedcomTree {
	if sp.parser == nil {
		return nil
	}
	return sp.parser.GetTree()
}

