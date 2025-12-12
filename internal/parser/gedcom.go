package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// HierarchicalParser is a full hierarchical parser that builds complete GEDCOM tree structure.
// This is Step 1.7 of the incremental development plan.
type HierarchicalParser struct {
	tree                *gedcom.GedcomTree
	parentsStack        *LineStack
	continuationHandler *ContinuationHandler
}

// NewHierarchicalParser creates a new HierarchicalParser.
func NewHierarchicalParser() *HierarchicalParser {
	return &HierarchicalParser{
		tree:                gedcom.NewGedcomTree(),
		parentsStack:        NewLineStack(),
		continuationHandler: NewContinuationHandler(),
	}
}

// Parse parses a GEDCOM file and builds the complete hierarchical tree structure.
func (hp *HierarchicalParser) Parse(filePath string) (*gedcom.GedcomTree, error) {
	// Step 1: Validate file
	if err := ValidateFile(filePath); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Step 2: Detect encoding
	encoding, err := DetectEncoding(filePath)
	if err != nil {
		return nil, fmt.Errorf("encoding detection failed: %w", err)
	}
	hp.tree.SetEncoding(string(encoding))

	// Step 3: Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Step 4: Get reader with proper encoding (handles BOM skipping)
	reader, err := GetReader(file, encoding)
	if err != nil {
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

		// Parse the line
		level, tag, value, xrefID, err := ParseLine(line)
		if err != nil {
			// Log warning but continue parsing
			// TODO: Add proper error logging in Step 1.8
			continue
		}

		// Handle CONC/CONT continuation lines
		if tag == "CONC" || tag == "CONT" {
			if err := hp.continuationHandler.HandleContinuation(tag, level, value); err != nil {
				// Invalid continuation, skip it
				// TODO: Log error in Step 1.8
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
			gedcomLine := gedcom.NewGedcomLine(level, tag, value, xrefID)
			gedcomLine.LineNumber = lineNumber

			// Create Record from line
			record := gedcom.NewBaseRecord(gedcomLine)

			// Add to tree
			hp.tree.AddRecord(record)

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
			// TODO: Log warning in Step 1.8
			// Skip this line
			continue
		}

		// Create child line (no xref for level > 0)
		childLine := gedcom.NewGedcomLine(level, tag, value, "")
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
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return hp.tree, nil
}

// GetTree returns the parsed tree.
func (hp *HierarchicalParser) GetTree() *gedcom.GedcomTree {
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
func (bp *BasicParser) Parse(filePath string) (*gedcom.GedcomTree, error) {
	return bp.parser.Parse(filePath)
}

// GetTree returns the parsed tree.
func (bp *BasicParser) GetTree() *gedcom.GedcomTree {
	return bp.parser.GetTree()
}

