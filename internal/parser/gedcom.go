package parser

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

// BasicParser is a parser that only handles level 0 records (no hierarchy yet).
// This is Step 1.5 of the incremental development plan.
type BasicParser struct {
	tree *gedcom.GedcomTree
}

// NewBasicParser creates a new BasicParser.
func NewBasicParser() *BasicParser {
	return &BasicParser{
		tree: gedcom.NewGedcomTree(),
	}
}

// Parse parses a GEDCOM file and extracts only level 0 records.
// All level > 0 lines are skipped.
func (bp *BasicParser) Parse(filePath string) (*gedcom.GedcomTree, error) {
	// Step 1: Validate file
	if err := ValidateFile(filePath); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Step 2: Detect encoding
	encoding, err := DetectEncoding(filePath)
	if err != nil {
		return nil, fmt.Errorf("encoding detection failed: %w", err)
	}
	bp.tree.SetEncoding(string(encoding))

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

	// Step 5: Parse file line by line
	scanner := bufio.NewScanner(reader)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Parse the line
		level, tag, value, xrefID, err := ParseLine(line)
		if err != nil {
			// Log warning but continue parsing
			// For Step 1.5, we'll just skip malformed lines
			continue
		}

		// Only process level 0 records
		if level == 0 {
			// Create GedcomLine
			gedcomLine := gedcom.NewGedcomLine(level, tag, value, xrefID)
			gedcomLine.LineNumber = lineNumber

			// Create Record from line
			record := gedcom.NewBaseRecord(gedcomLine)

			// Add to tree
			bp.tree.AddRecord(record)
		}
		// Skip all level > 0 lines for now
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return bp.tree, nil
}

// GetTree returns the parsed tree.
func (bp *BasicParser) GetTree() *gedcom.GedcomTree {
	return bp.tree
}

