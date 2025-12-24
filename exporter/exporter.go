package exporter

import (
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// Exporter is the interface that all exporters must implement.
type Exporter interface {
	// ExportToFile exports the tree to a file.
	ExportToFile(tree *types.GedcomTree, filePath string) error

	// ExportToString exports the tree to a string.
	ExportToString(tree *types.GedcomTree) (string, error)
}

// BaseExporter provides common functionality for all exporters.
type BaseExporter struct {
	errorManager *types.ErrorManager
}

// NewBaseExporter creates a new BaseExporter.
func NewBaseExporter(errorManager *types.ErrorManager) *BaseExporter {
	return &BaseExporter{
		errorManager: errorManager,
	}
}

// AddError is a helper method to add errors.
func (be *BaseExporter) AddError(severity types.ErrorSeverity, message string, lineNumber int, context string) {
	be.errorManager.AddError(severity, message, lineNumber, context)
}


