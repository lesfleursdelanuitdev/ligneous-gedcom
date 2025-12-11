// This is an example Go implementation showing key components
// This file is for reference only - actual implementation would be in separate packages

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// ============================================================================
// Error Handling
// ============================================================================

type ErrorSeverity string

const (
	SeverityWarning ErrorSeverity = "warning"
	SeveritySevere  ErrorSeverity = "severe"
)

type GedcomError struct {
	Severity   ErrorSeverity
	Message    string
	LineNumber int
	Context    string
}

func (e *GedcomError) Error() string {
	if e.LineNumber > 0 {
		return fmt.Sprintf("%s: %s (Line %d)", e.Severity, e.Message, e.LineNumber)
	}
	return fmt.Sprintf("%s: %s", e.Severity, e.Message)
}

type ErrorManager struct {
	mu     sync.RWMutex
	errors []*GedcomError
}

func NewErrorManager() *ErrorManager {
	return &ErrorManager{
		errors: make([]*GedcomError, 0),
	}
}

func (em *ErrorManager) AddError(severity ErrorSeverity, message string, lineNumber int, context string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.errors = append(em.errors, &GedcomError{
		Severity:   severity,
		Message:    message,
		LineNumber: lineNumber,
		Context:    context,
	})
}

func (em *ErrorManager) Errors() []*GedcomError {
	em.mu.RLock()
	defer em.mu.RUnlock()
	result := make([]*GedcomError, len(em.errors))
	copy(result, em.errors)
	return result
}

func (em *ErrorManager) HasErrors() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.errors) > 0
}

func (em *ErrorManager) HasSevereErrors() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	for _, err := range em.errors {
		if err.Severity == SeveritySevere {
			return true
		}
	}
	return false
}

// ============================================================================
// Core Data Structures
// ============================================================================

type RecordType string

const (
	RecordTypeHEAD RecordType = "HEAD"
	RecordTypeINDI RecordType = "INDI"
	RecordTypeFAM  RecordType = "FAM"
	RecordTypeNOTE RecordType = "NOTE"
	RecordTypeSOUR RecordType = "SOUR"
	RecordTypeREPO RecordType = "REPO"
	RecordTypeSUBM RecordType = "SUBM"
	RecordTypeOBJE RecordType = "OBJE"
	RecordTypeTRLR RecordType = "TRLR"
)

type GedcomLine struct {
	Level      int
	Tag        string
	Value      string
	XrefID     string
	LineNumber int
	Parent     *GedcomLine
	Children   map[string][]*GedcomLine
}

func NewGedcomLine(level int, tag, value, xrefID string) *GedcomLine {
	return &GedcomLine{
		Level:    level,
		Tag:      tag,
		Value:    value,
		XrefID:   xrefID,
		Children: make(map[string][]*GedcomLine),
	}
}

func (gl *GedcomLine) AddChild(child *GedcomLine) {
	if gl.Children == nil {
		gl.Children = make(map[string][]*GedcomLine)
	}
	gl.Children[child.Tag] = append(gl.Children[child.Tag], child)
	child.Parent = gl
}

func (gl *GedcomLine) GetValue(selector string) string {
	if selector == "" {
		return gl.Value
	}

	parts := strings.Split(selector, ".")
	if len(parts) == 0 {
		return gl.Value
	}

	currentTag := parts[0]
	remaining := strings.Join(parts[1:], ".")

	if children, ok := gl.Children[currentTag]; ok {
		for _, child := range children {
			if len(parts) == 1 {
				return child.Value
			}
			if result := child.GetValue(remaining); result != "" {
				return result
			}
		}
	}

	return ""
}

func (gl *GedcomLine) GetLines(selector string) []*GedcomLine {
	if selector == "" {
		return []*GedcomLine{gl}
	}

	parts := strings.Split(selector, ".")
	currentTag := parts[0]
	remaining := strings.Join(parts[1:], ".")

	results := make([]*GedcomLine, 0)
	if children, ok := gl.Children[currentTag]; ok {
		for _, child := range children {
			if len(parts) == 1 {
				results = append(results, child)
			} else {
				results = append(results, child.GetLines(remaining)...)
			}
		}
	}

	return results
}

// ============================================================================
// Record Interface
// ============================================================================

type Record interface {
	Type() RecordType
	XrefID() string
	FirstLine() *GedcomLine
	GetValue(selector string) string
	GetValues(selector string) []string
	GetLines(selector string) []*GedcomLine
}

type BaseRecord struct {
	firstLine  *GedcomLine
	recordType RecordType
}

func NewBaseRecord(line *GedcomLine) *BaseRecord {
	return &BaseRecord{
		firstLine:  line,
		recordType: RecordType(line.Tag),
	}
}

func (br *BaseRecord) Type() RecordType {
	return br.recordType
}

func (br *BaseRecord) XrefID() string {
	return br.firstLine.XrefID
}

func (br *BaseRecord) FirstLine() *GedcomLine {
	return br.firstLine
}

func (br *BaseRecord) GetValue(selector string) string {
	return br.firstLine.GetValue(selector)
}

func (br *BaseRecord) GetValues(selector string) []string {
	lines := br.firstLine.GetLines(selector)
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.Value != "" {
			values = append(values, line.Value)
		}
	}
	return values
}

func (br *BaseRecord) GetLines(selector string) []*GedcomLine {
	return br.firstLine.GetLines(selector)
}

// ============================================================================
// Parser with Error Handling
// ============================================================================

type GedcomParser struct {
	errorManager *ErrorManager
}

func NewGedcomParser(errorManager *ErrorManager) *GedcomParser {
	return &GedcomParser{
		errorManager: errorManager,
	}
}

// validateFile checks file existence, readability, and non-empty
func (gp *GedcomParser) validateFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			gp.errorManager.AddError(SeveritySevere, fmt.Sprintf("File does not exist: %s", filePath), 0, "File Validation")
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		gp.errorManager.AddError(SeveritySevere, fmt.Sprintf("Cannot access file: %s", err.Error()), 0, "File Validation")
		return fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		gp.errorManager.AddError(SeveritySevere, fmt.Sprintf("Path is a directory, not a file: %s", filePath), 0, "File Validation")
		return fmt.Errorf("path is a directory: %s", filePath)
	}

	if info.Size() == 0 {
		gp.errorManager.AddError(SeveritySevere, fmt.Sprintf("File is empty: %s", filePath), 0, "File Validation")
		return fmt.Errorf("file is empty: %s", filePath)
	}

	return nil
}

// parseLine parses a single GEDCOM line with explicit error handling
func (gp *GedcomParser) parseLine(line string) (level int, tag, value, xrefID string, err error) {
	parts := strings.Fields(line)

	if len(parts) < 2 {
		return 0, "", "", "", fmt.Errorf("line has insufficient parts: %s", line)
	}

	// Parse level
	level, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", "", fmt.Errorf("invalid level '%s': %w", parts[0], err)
	}

	if level < 0 {
		return 0, "", "", "", fmt.Errorf("level cannot be negative: %d", level)
	}

	// Parse tag and value/xref
	if len(parts) == 3 && strings.HasPrefix(parts[1], "@") {
		// Has xref: level xref tag
		return level, parts[2], "", parts[1], nil
	} else if len(parts) == 3 {
		// Has value: level tag value
		return level, parts[1], parts[2], "", nil
	} else {
		// Only tag: level tag
		return level, parts[1], "", "", nil
	}
}

// Parse parses a GEDCOM file with comprehensive error handling
func (gp *GedcomParser) Parse(filePath string) ([]Record, error) {
	// Validate file first
	if err := gp.validateFile(filePath); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		gp.errorManager.AddError(SeveritySevere, fmt.Sprintf("Failed to open file: %s", err.Error()), 0, "File I/O")
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var records []Record
	var parentsStack []*GedcomLine
	var currentValue strings.Builder
	lineNumber := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse line (returns error if malformed)
		level, tag, value, xrefID, err := gp.parseLine(line)
		if err != nil {
			// Log error but continue parsing
			gp.errorManager.AddError(
				SeverityWarning,
				fmt.Sprintf("Malformed line: %s", err.Error()),
				lineNumber,
				"Line Parsing",
			)
			continue
		}

		// Handle CONC/CONT
		if tag == "CONC" || tag == "CONT" {
			if len(parentsStack) == 0 {
				gp.errorManager.AddError(
					SeverityWarning,
					"CONC/CONT without parent line",
					lineNumber,
					"Line Parsing",
				)
				continue
			}

			if tag == "CONC" {
				currentValue.WriteString(value)
			} else {
				currentValue.WriteString("\n")
				currentValue.WriteString(value)
			}
			continue
		}

		// Apply accumulated continuation value
		if currentValue.Len() > 0 && len(parentsStack) > 0 {
			parentsStack[len(parentsStack)-1].Value = currentValue.String()
			currentValue.Reset()
		}

		// Create line
		gedcomLine := NewGedcomLine(level, tag, value, xrefID)
		gedcomLine.LineNumber = lineNumber

		// Handle based on level
		if level == 0 {
			// Top-level record
			record := NewBaseRecord(gedcomLine)
			records = append(records, record)
			parentsStack = []*GedcomLine{gedcomLine}
		} else {
			// Child line - find parent
			for len(parentsStack) > 0 && parentsStack[len(parentsStack)-1].Level >= level {
				parentsStack = parentsStack[:len(parentsStack)-1]
			}

			if len(parentsStack) == 0 {
				gp.errorManager.AddError(
					SeverityWarning,
					fmt.Sprintf("Orphaned line at level %d with no parent: %s", level, tag),
					lineNumber,
					"Line Parsing",
				)
				continue
			}

			parent := parentsStack[len(parentsStack)-1]
			parent.AddChild(gedcomLine)
			parentsStack = append(parentsStack, gedcomLine)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		gp.errorManager.AddError(SeveritySevere, fmt.Sprintf("Scanner error: %s", err.Error()), 0, "File I/O")
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// Handle remaining continuation value
	if currentValue.Len() > 0 && len(parentsStack) > 0 {
		parentsStack[len(parentsStack)-1].Value = currentValue.String()
	}

	return records, nil
}

// ============================================================================
// Example Usage
// ============================================================================

func main() {
	errorManager := NewErrorManager()
	parser := NewGedcomParser(errorManager)

	records, err := parser.Parse("sample.ged")
	if err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		return
	}

	// Check for errors
	if errorManager.HasErrors() {
		fmt.Println("Validation errors found:")
		for _, err := range errorManager.Errors() {
			fmt.Printf("  %s\n", err)
		}
	}

	// Process records
	fmt.Printf("Parsed %d records\n", len(records))
	for _, record := range records {
		fmt.Printf("Record: %s (Xref: %s)\n", record.Type(), record.XrefID())
		if record.Type() == RecordTypeINDI {
			name := record.GetValue("NAME")
			if name != "" {
				fmt.Printf("  Name: %s\n", name)
			}
		}
	}
}

