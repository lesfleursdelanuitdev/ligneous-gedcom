package parser

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// SmartParser is the recommended entry point for parsing GEDCOM files.
// It automatically selects the optimal parser based on file size.
// For all files, it uses HierarchicalParser which automatically enables
// parallel processing for files >= 32KB.
//
// For very large files (>100MB) requiring streaming, users should
// explicitly use StreamingHierarchicalParser.
type SmartParser struct {
	parser ParserInterface
}

// NewParser is the recommended entry point for parsing GEDCOM files.
// It returns a SmartParser that automatically optimizes based on file size.
func NewParser() *SmartParser {
	return NewSmartParser()
}

// ParserInterface defines the common interface for parsers
type ParserInterface interface {
	Parse(filePath string) (*types.GedcomTree, error)
	GetErrors() []*types.GedcomError
	HasErrors() bool
	GetErrorManager() *types.ErrorManager
	GetTree() *types.GedcomTree
}

// NewSmartParser creates a parser that automatically selects the best implementation
// based on file size. This optimizes for both small and large files.
func NewSmartParser() *SmartParser {
	return &SmartParser{}
}

// Parse automatically selects and uses the best parser for the file size.
// HierarchicalParser automatically enables parallel processing for files >= 32KB,
// so we always use it for full-tree parsing.
func (sp *SmartParser) Parse(filePath string) (*types.GedcomTree, error) {
	// Always use HierarchicalParser which automatically enables parallel processing
	// for files >= 32KB. This provides optimal performance without user configuration.
	sp.parser = NewHierarchicalParser()
	return sp.parser.Parse(filePath)
}

// GetErrors returns all errors collected during parsing
func (sp *SmartParser) GetErrors() []*types.GedcomError {
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
func (sp *SmartParser) GetErrorManager() *types.ErrorManager {
	if sp.parser == nil {
		return nil
	}
	return sp.parser.GetErrorManager()
}

// GetTree returns the parsed tree
func (sp *SmartParser) GetTree() *types.GedcomTree {
	if sp.parser == nil {
		return nil
	}
	return sp.parser.GetTree()
}

