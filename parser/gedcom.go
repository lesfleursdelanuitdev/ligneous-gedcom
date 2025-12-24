package parser

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// HierarchicalParser is a full hierarchical parser that builds complete GEDCOM tree structure.
// It automatically enables parallel processing for files >= 32KB to improve performance.
// This parser merges the benefits of sequential and parallel parsing approaches.
type HierarchicalParser struct {
	tree                *types.GedcomTree
	parentsStack        *LineStack
	continuationHandler *ContinuationHandler
	errorManager        *types.ErrorManager
	factory             *types.RecordFactory // Reused factory to avoid allocations

	// Parallel processing fields (auto-enabled for files >= 32KB)
	enableParallel bool
	recordChan     chan *types.GedcomLine
	wg             sync.WaitGroup
	numWorkers     int
}

// NewHierarchicalParser creates a new HierarchicalParser.
// Parallel processing will be automatically enabled for files >= 32KB.
func NewHierarchicalParser() *HierarchicalParser {
	return &HierarchicalParser{
		tree:                types.NewGedcomTree(),
		parentsStack:        NewLineStack(),
		continuationHandler: NewContinuationHandler(),
		errorManager:        types.NewErrorManager(),
		factory:             types.NewRecordFactory(), // Create once, reuse for all records
		enableParallel:      false,                    // Will be auto-enabled based on file size
		numWorkers:          runtime.NumCPU(),
	}
}

// Parse parses a GEDCOM file and builds the complete hierarchical tree structure.
// Returns the tree and any parsing errors (warnings don't stop parsing).
// Parallel processing is automatically enabled for files >= 32KB.
func (hp *HierarchicalParser) Parse(filePath string) (*types.GedcomTree, error) {
	// Step 1: Validate file
	if err := ValidateFile(filePath); err != nil {
		hp.errorManager.AddError(types.SeveritySevere, fmt.Sprintf("File validation failed: %v", err), 0, "File Validation")
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Step 2: Check file size and auto-enable parallel processing if beneficial
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		hp.errorManager.AddError(types.SeveritySevere, fmt.Sprintf("Failed to stat file: %v", err), 0, "File I/O")
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	const parallelThreshold = 32 * 1024 // 32KB - threshold for parallel processing
	if fileInfo.Size() >= parallelThreshold {
		hp.enableParallel = true
		hp.recordChan = make(chan *types.GedcomLine, 100) // Buffered channel
		// Start record processor goroutine
		hp.wg.Add(1)
		go hp.processRecords()
	}

	// Step 3: Detect encoding
	encoding, err := DetectEncoding(filePath)
	if err != nil {
		hp.errorManager.AddError(types.SeveritySevere, fmt.Sprintf("Encoding detection failed: %v", err), 0, "Encoding Detection")
		return nil, fmt.Errorf("encoding detection failed: %w", err)
	}
	hp.tree.SetEncoding(string(encoding))

	// Step 4: Open file
	file, err := os.Open(filePath)
	if err != nil {
		hp.errorManager.AddError(types.SeveritySevere, fmt.Sprintf("Failed to open file: %v", err), 0, "File I/O")
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Step 5: Get reader with proper encoding (handles BOM skipping)
	reader, err := GetReader(file, encoding)
	if err != nil {
		hp.errorManager.AddError(types.SeveritySevere, fmt.Sprintf("Failed to create reader: %v", err), 0, "Encoding")
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	// Step 5: Parse file line by line using stack-based algorithm
	scanner := bufio.NewScanner(reader)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Skip empty lines
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parse the line using optimized parser (line is already trimmed)
		level, tag, value, xrefID, err := ParseLineFast(line)
		if err != nil {
			// Log warning but continue parsing
			hp.errorManager.AddError(types.SeverityWarning, fmt.Sprintf("Malformed line: %v", err), lineNumber, "Line Parsing")
			continue
		}

		// Handle CONC/CONT continuation lines
		if tag == "CONC" || tag == "CONT" {
			if err := hp.continuationHandler.HandleContinuation(tag, level, value); err != nil {
				// Invalid continuation, skip it
				hp.errorManager.AddError(types.SeverityWarning, fmt.Sprintf("Invalid continuation: %v", err), lineNumber, "CONC/CONT Handling")
				continue
			}
			// Continue to next line (value is accumulated)
			continue
		}

		// Apply accumulated CONC/CONT value to previous line
		if hp.continuationHandler.HasAccumulatedValue() {
			if !hp.parentsStack.IsEmpty() {
				topLine := hp.parentsStack.Peek()
				accumulatedValue := hp.continuationHandler.GetAccumulatedValue()
				// Append to existing value if any
				if topLine.Value != "" {
					topLine.Value += accumulatedValue
				} else {
					topLine.Value = accumulatedValue
				}
			}
		}

		// Handle level 0 (top-level record)
		if level == 0 {
			// Create GedcomLine
			gedcomLine := types.NewGedcomLine(level, tag, value, xrefID)
			gedcomLine.LineNumber = lineNumber

			// Process record (parallel if enabled, sequential otherwise)
			if hp.enableParallel {
				// Send to parallel processor (non-blocking if channel has space)
				select {
				case hp.recordChan <- gedcomLine:
					// Record sent for parallel processing
				default:
					// Channel full, process synchronously (fallback)
					record := hp.factory.CreateRecord(gedcomLine)
					hp.tree.AddRecord(record)
				}
			} else {
				// Sequential processing
				record := hp.factory.CreateRecord(gedcomLine)
				hp.tree.AddRecord(record)
			}

			// Reset stack (new top-level record)
			hp.parentsStack.Clear()
			hp.parentsStack.Push(gedcomLine)

			// Update last tag
			hp.continuationHandler.SetLastTag(tag, level)

			continue
		}

		// Handle level > 0 (child lines)
		// Find parent using stack
		parent, err := hp.parentsStack.FindParent(level)
		if err != nil {
			// Orphaned line - no parent found
			hp.errorManager.AddError(types.SeverityWarning, fmt.Sprintf("Orphaned line: %s", err.Error()), lineNumber, "Hierarchy")
			// Skip this line
			continue
		}

		// Create child line (no xref for level > 0)
		childLine := types.NewGedcomLine(level, tag, value, "")
		childLine.LineNumber = lineNumber

		// Add as child to parent
		parent.AddChild(childLine)

		// Push to stack
		hp.parentsStack.Push(childLine)

		// Update last tag
		hp.continuationHandler.SetLastTag(tag, level)
	}

	// Handle remaining CONC/CONT value (if file ends with continuation)
	if hp.continuationHandler.HasAccumulatedValue() {
		if !hp.parentsStack.IsEmpty() {
			topLine := hp.parentsStack.Peek()
			accumulatedValue := hp.continuationHandler.GetAccumulatedValue()
			if topLine.Value != "" {
				topLine.Value += accumulatedValue
			} else {
				topLine.Value = accumulatedValue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		hp.errorManager.AddError(types.SeveritySevere, fmt.Sprintf("Error reading file: %v", err), lineNumber, "File I/O")
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// If parallel processing was enabled, close channel and wait for workers to finish
	if hp.enableParallel {
		close(hp.recordChan)
		hp.wg.Wait()
	}

	// Return tree (errors are available via GetErrors())
	return hp.tree, nil
}

// processRecords processes level 0 records in a separate goroutine.
// This allows the main parsing loop to continue while records are being created.
// Only used when parallel processing is enabled.
func (hp *HierarchicalParser) processRecords() {
	defer hp.wg.Done()

	// Use the reused factory from the parser instance
	for line := range hp.recordChan {
		record := hp.factory.CreateRecord(line)
		hp.tree.AddRecord(record)
	}
}

// GetErrors returns all errors collected during parsing
func (hp *HierarchicalParser) GetErrors() []*types.GedcomError {
	return hp.errorManager.Errors()
}

// HasErrors returns true if any errors were encountered
func (hp *HierarchicalParser) HasErrors() bool {
	return hp.errorManager.HasErrors()
}

// HasSevereErrors returns true if any severe errors were encountered
func (hp *HierarchicalParser) HasSevereErrors() bool {
	return hp.errorManager.HasSevereErrors()
}

// GetErrorManager returns the error manager (for advanced usage)
func (hp *HierarchicalParser) GetErrorManager() *types.ErrorManager {
	return hp.errorManager
}

// GetTree returns the parsed tree.
func (hp *HierarchicalParser) GetTree() *types.GedcomTree {
	return hp.tree
}

// BasicParser is kept for backward compatibility but now uses HierarchicalParser
// This maintains the API from Step 1.5
type BasicParser struct {
	parser *HierarchicalParser
}

// NewBasicParser creates a new BasicParser (now uses hierarchical parsing).
func NewBasicParser() *BasicParser {
	return &BasicParser{
		parser: NewHierarchicalParser(),
	}
}

// Parse parses a GEDCOM file using hierarchical parsing.
func (bp *BasicParser) Parse(filePath string) (*types.GedcomTree, error) {
	return bp.parser.Parse(filePath)
}

// GetTree returns the parsed tree.
func (bp *BasicParser) GetTree() *types.GedcomTree {
	return bp.parser.GetTree()
}
