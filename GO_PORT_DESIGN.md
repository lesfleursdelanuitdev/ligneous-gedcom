# GEDCOM Parser - Go Implementation Design

## Overview

This document outlines the design for porting the Python GEDCOM parser to Go, addressing all identified issues from the Python implementation and leveraging Go's strengths for a more robust, type-safe, and performant solution.

## Design Principles

1. **Type Safety**: Leverage Go's strong typing to prevent runtime errors
2. **Explicit Error Handling**: Use Go's error return pattern instead of exceptions
3. **Interface-Based Design**: Use interfaces for extensibility and testability
4. **Zero-Value Usability**: Make types safe to use with zero values
5. **Comprehensive Testing**: Built-in testing from the start
6. **Performance**: Take advantage of Go's efficiency for large files

## Architecture

### Package Structure

```
gedcom/
├── cmd/
│   └── gedcom-cli/          # CLI application
│       └── main.go
├── internal/                 # Internal packages (not importable)
│   ├── parser/
│   │   ├── gedcom.go        # GEDCOM parser
│   │   ├── json.go          # JSON parser
│   │   └── line.go          # Line parsing utilities
│   ├── exporter/
│   │   ├── gedcom.go        # GEDCOM exporter
│   │   └── json.go          # JSON exporter
│   ├── validator/
│   │   ├── validator.go     # Base validator interface
│   │   ├── individual.go
│   │   ├── family.go
│   │   └── ...
│   └── record/
│       ├── record.go        # Base record interface
│       ├── individual.go
│       ├── family.go
│       └── factory.go       # Record factory
├── pkg/                      # Public API
│   ├── gedcom.go            # Main Gedcom struct
│   ├── line.go              # GedcomLine struct
│   ├── error.go             # Error types
│   └── types.go             # Type definitions
├── go.mod
├── go.sum
└── README.md
```

## Core Types

### Error Handling

```go
// pkg/error.go

package gedcom

import "fmt"

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
    SeverityWarning ErrorSeverity = "warning"
    SeveritySevere  ErrorSeverity = "severe"
)

// Error represents a GEDCOM parsing/validation error
type Error struct {
    Severity   ErrorSeverity
    Message    string
    LineNumber int
    Context    string
}

func (e *Error) Error() string {
    if e.LineNumber > 0 {
        return fmt.Sprintf("%s: %s (Line %d)", e.Severity, e.Message, e.LineNumber)
    }
    return fmt.Sprintf("%s: %s", e.Severity, e.Message)
}

// ErrorManager manages collection of errors
type ErrorManager struct {
    errors []*Error
}

func NewErrorManager() *ErrorManager {
    return &ErrorManager{
        errors: make([]*Error, 0),
    }
}

func (em *ErrorManager) AddError(severity ErrorSeverity, message string, lineNumber int, context string) {
    em.errors = append(em.errors, &Error{
        Severity:   severity,
        Message:    message,
        LineNumber: lineNumber,
        Context:    context,
    })
}

func (em *ErrorManager) Errors() []*Error {
    return em.errors
}

func (em *ErrorManager) HasErrors() bool {
    return len(em.errors) > 0
}

func (em *ErrorManager) HasSevereErrors() bool {
    for _, err := range em.errors {
        if err.Severity == SeveritySevere {
            return true
        }
    }
    return false
}
```

### Core Data Structures

```go
// pkg/types.go

package gedcom

// RecordType represents the type of a GEDCOM record
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

// GedcomLine represents a single line in a GEDCOM file
type GedcomLine struct {
    Level      int
    Tag        string
    Value      string
    XrefID     string
    LineNumber int
    Parent     *GedcomLine
    Children   map[string][]*GedcomLine
}

func NewGedcomLine(level int, tag, value string, xrefID string) *GedcomLine {
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

// GetValue retrieves a value using dot notation (e.g., "BIRT.DATE")
func (gl *GedcomLine) GetValue(selector string) string {
    if selector == "" {
        return gl.Value
    }
    
    parts := splitSelector(selector)
    if len(parts) == 0 {
        return gl.Value
    }
    
    currentTag := parts[0]
    remaining := parts[1:]
    
    if children, ok := gl.Children[currentTag]; ok {
        for _, child := range children {
            if len(remaining) == 0 {
                return child.Value
            }
            if result := child.GetValue(joinSelector(remaining)); result != "" {
                return result
            }
        }
    }
    
    return ""
}

// GetLines retrieves all lines matching a selector
func (gl *GedcomLine) GetLines(selector string) []*GedcomLine {
    if selector == "" {
        return []*GedcomLine{gl}
    }
    
    parts := splitSelector(selector)
    currentTag := parts[0]
    remaining := parts[1:]
    
    results := make([]*GedcomLine, 0)
    if children, ok := gl.Children[currentTag]; ok {
        for _, child := range children {
            if len(remaining) == 0 {
                results = append(results, child)
            } else {
                results = append(results, child.GetLines(joinSelector(remaining))...)
            }
        }
    }
    
    return results
}
```

### Record Interface

```go
// pkg/record.go

package gedcom

// Record represents a GEDCOM record
type Record interface {
    Type() RecordType
    XrefID() string
    FirstLine() *GedcomLine
    GetValue(selector string) string
    GetValues(selector string) []string
    GetLines(selector string) []*GedcomLine
    ToGED() []string
    ToJSON() interface{}
}

// BaseRecord provides default implementation
type BaseRecord struct {
    firstLine *GedcomLine
    recordType RecordType
}

func NewBaseRecord(line *GedcomLine) *BaseRecord {
    return &BaseRecord{
        firstLine: line,
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
```

### Main Gedcom Structure

```go
// pkg/gedcom.go

package gedcom

import (
    "fmt"
    "sync"
)

// Gedcom represents a parsed GEDCOM file
type Gedcom struct {
    mu sync.RWMutex
    
    // Records organized by type
    header      Record
    individuals map[string]Record
    families    map[string]Record
    notes       map[string]Record
    sources     map[string]Record
    repositories map[string]Record
    submitters  map[string]Record
    multimedia  map[string]Record
    
    // Cross-reference index
    xrefIndex map[string]Record
    
    // Metadata
    encoding string
    version  string
    
    // Components
    errorManager *ErrorManager
    validator    Validator
    parser       Parser
    exporter     Exporter
    
    // Record counts for ID generation
    recordCounts map[RecordType]int
}

// NewGedcom creates a new Gedcom instance
func NewGedcom() *Gedcom {
    return &Gedcom{
        individuals:  make(map[string]Record),
        families:     make(map[string]Record),
        notes:        make(map[string]Record),
        sources:      make(map[string]Record),
        repositories: make(map[string]Record),
        submitters:   make(map[string]Record),
        multimedia:   make(map[string]Record),
        xrefIndex:    make(map[string]Record),
        recordCounts: make(map[RecordType]int),
        errorManager: NewErrorManager(),
    }
}

// Parse parses a file using the specified parser
func (g *Gedcom) Parse(parserType string, filePath string) error {
    g.mu.Lock()
    defer g.mu.Unlock()
    
    parser, err := g.getParser(parserType)
    if err != nil {
        return fmt.Errorf("invalid parser type: %w", err)
    }
    
    if err := parser.Parse(filePath, g); err != nil {
        return fmt.Errorf("parsing failed: %w", err)
    }
    
    g.buildXrefIndex()
    g.validate()
    
    return nil
}

// AddRecord adds a record to the appropriate collection
func (g *Gedcom) AddRecord(record Record) error {
    g.mu.Lock()
    defer g.mu.Unlock()
    
    recordType := record.Type()
    
    switch recordType {
    case RecordTypeHEAD:
        g.header = record
        g.extractHeaderInfo(record)
    case RecordTypeTRLR:
        // Trailer is handled separately
        return nil
    default:
        xrefID := record.XrefID()
        if xrefID == "" {
            xrefID = g.generateXrefID(recordType)
        }
        
        switch recordType {
        case RecordTypeINDI:
            g.individuals[xrefID] = record
        case RecordTypeFAM:
            g.families[xrefID] = record
        case RecordTypeNOTE:
            g.notes[xrefID] = record
        case RecordTypeSOUR:
            g.sources[xrefID] = record
        case RecordTypeREPO:
            g.repositories[xrefID] = record
        case RecordTypeSUBM:
            g.submitters[xrefID] = record
        case RecordTypeOBJE:
            g.multimedia[xrefID] = record
        default:
            return fmt.Errorf("unknown record type: %s", recordType)
        }
        
        g.xrefIndex[xrefID] = record
        g.recordCounts[recordType]++
    }
    
    return nil
}

// Thread-safe getters
func (g *Gedcom) GetRecord(xrefID string) (Record, bool) {
    g.mu.RLock()
    defer g.mu.RUnlock()
    record, ok := g.xrefIndex[xrefID]
    return record, ok
}

func (g *Gedcom) Individuals() map[string]Record {
    g.mu.RLock()
    defer g.mu.RUnlock()
    result := make(map[string]Record)
    for k, v := range g.individuals {
        result[k] = v
    }
    return result
}
```

## Parser Implementation

### Parser Interface

```go
// internal/parser/parser.go

package parser

import "github.com/yourorg/gedcom/pkg"

// Parser defines the interface for parsing GEDCOM files
type Parser interface {
    Parse(filePath string, gedcom *gedcom.Gedcom) error
}

// BaseParser provides common parsing functionality
type BaseParser struct {
    errorManager *gedcom.ErrorManager
}

func (bp *BaseParser) validateFile(filePath string) error {
    // File validation logic
    // - Check existence
    // - Check readability
    // - Check not empty
    // Returns explicit error
}
```

### GEDCOM Parser

```go
// internal/parser/gedcom.go

package parser

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "strings"
    "unicode/utf8"
    
    "github.com/yourorg/gedcom/pkg"
)

type GedcomParser struct {
    BaseParser
}

func NewGedcomParser() *GedcomParser {
    return &GedcomParser{}
}

func (gp *GedcomParser) Parse(filePath string, g *gedcom.Gedcom) error {
    // Validate file first
    if err := gp.validateFile(filePath); err != nil {
        return fmt.Errorf("file validation failed: %w", err)
    }
    
    // Detect encoding
    encoding, err := gp.detectEncoding(filePath)
    if err != nil {
        return fmt.Errorf("encoding detection failed: %w", err)
    }
    
    // Open file with detected encoding
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()
    
    // Use appropriate reader based on encoding
    reader := gp.getReader(file, encoding)
    
    // Parse line by line
    var parentsStack []*gedcom.GedcomLine
    var currentValue strings.Builder
    var lastTag *tagInfo
    lineNumber := 0
    
    scanner := bufio.NewScanner(reader)
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
            g.ErrorManager().AddError(
                gedcom.SeverityWarning,
                fmt.Sprintf("Malformed line: %s", err.Error()),
                lineNumber,
                "Line Parsing",
            )
            continue
        }
        
        // Handle CONC/CONT
        if tag == "CONC" || tag == "CONT" {
            if err := gp.handleContinuation(tag, level, value, lastTag, &currentValue); err != nil {
                g.ErrorManager().AddError(
                    gedcom.SeverityWarning,
                    err.Error(),
                    lineNumber,
                    "Line Parsing",
                )
                continue
            }
            lastTag = &tagInfo{tag: tag, level: level}
            continue
        }
        
        // Apply accumulated continuation value
        if currentValue.Len() > 0 && len(parentsStack) > 0 {
            parentsStack[len(parentsStack)-1].Value = currentValue.String()
            currentValue.Reset()
        }
        
        // Create line
        gedcomLine := gedcom.NewGedcomLine(level, tag, value, xrefID)
        gedcomLine.LineNumber = lineNumber
        
        // Handle based on level
        if level == 0 {
            // Top-level record
            record, err := gp.createRecord(gedcomLine, g)
            if err != nil {
                g.ErrorManager().AddError(
                    gedcom.SeverityWarning,
                    fmt.Sprintf("Failed to create record: %s", err.Error()),
                    lineNumber,
                    "Record Creation",
                )
                continue
            }
            
            if err := g.AddRecord(record); err != nil {
                g.ErrorManager().AddError(
                    gedcom.SeverityWarning,
                    fmt.Sprintf("Failed to add record: %s", err.Error()),
                    lineNumber,
                    "Record Management",
                )
                continue
            }
            parentsStack = []*gedcom.GedcomLine{gedcomLine}
        } else {
            // Child line - find parent
            parentsStack = gp.findParentLevel(parentsStack, level)
            
            if len(parentsStack) == 0 {
                g.ErrorManager().AddError(
                    gedcom.SeverityWarning,
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
        
        lastTag = &tagInfo{tag: tag, level: level}
    }
    
    // Check for scanner errors
    if err := scanner.Err(); err != nil {
        return fmt.Errorf("scanner error: %w", err)
    }
    
    // Handle remaining continuation value
    if currentValue.Len() > 0 && len(parentsStack) > 0 {
        parentsStack[len(parentsStack)-1].Value = currentValue.String()
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

type tagInfo struct {
    tag   string
    level int
}
```

## Key Improvements Over Python Implementation

### 1. Type Safety

**Python Issue**: No type hints, runtime type errors
**Go Solution**: 
- Strong typing enforced at compile time
- Interfaces for extensibility
- Type-safe record handling

### 2. Error Handling

**Python Issue**: Exceptions, inconsistent error handling
**Go Solution**:
- Explicit error returns (`error` interface)
- No panics for expected errors
- Error wrapping with context
- Error collection through ErrorManager

```go
// Example: Explicit error handling
func (gp *GedcomParser) Parse(filePath string, g *gedcom.Gedcom) error {
    if err := gp.validateFile(filePath); err != nil {
        return fmt.Errorf("file validation: %w", err)
    }
    // ... rest of parsing
}
```

### 3. Concurrency Safety

**Python Issue**: No thread safety
**Go Solution**:
- `sync.RWMutex` for thread-safe access
- Immutable data structures where possible
- Clear ownership of data

### 4. Memory Efficiency

**Python Issue**: Entire file loaded into memory
**Go Solution**:
- Streaming parser using `bufio.Scanner`
- Efficient string handling
- Minimal allocations

### 5. Validation

**Python Issue**: Missing validation, runtime errors
**Go Solution**:
- File validation before parsing
- Line validation with explicit errors
- Type validation at compile time

### 6. Testing

**Python Issue**: No tests
**Go Solution**:
- Built-in testing framework
- Table-driven tests
- Benchmarking support

```go
// Example test
func TestParseLine(t *testing.T) {
    tests := []struct {
        name      string
        line      string
        wantLevel int
        wantTag   string
        wantValue string
        wantErr   bool
    }{
        {
            name:      "valid line with value",
            line:      "1 NAME John /Doe/",
            wantLevel: 1,
            wantTag:   "NAME",
            wantValue: "John /Doe/",
            wantErr:   false,
        },
        {
            name:    "invalid level",
            line:    "X NAME John",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            level, tag, value, _, err := parseLine(tt.line)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseLine() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr {
                if level != tt.wantLevel || tag != tt.wantTag || value != tt.wantValue {
                    t.Errorf("parseLine() = (%d, %s, %s), want (%d, %s, %s)",
                        level, tag, value, tt.wantLevel, tt.wantTag, tt.wantValue)
                }
            }
        })
    }
}
```

## Migration Strategy

### Phase 1: Core Types (Week 1)
- [ ] Define core types (GedcomLine, Record, Error)
- [ ] Implement ErrorManager
- [ ] Create base record types
- [ ] Write comprehensive tests

### Phase 2: Parser (Week 2)
- [ ] Implement GEDCOM parser
- [ ] Implement JSON parser
- [ ] Add file validation
- [ ] Add encoding detection
- [ ] Write parser tests

### Phase 3: Validators (Week 3)
- [ ] Implement validator interface
- [ ] Port individual validators
- [ ] Port family validators
- [ ] Port cross-reference validators
- [ ] Write validator tests

### Phase 4: Exporters (Week 4)
- [ ] Implement GEDCOM exporter
- [ ] Implement JSON exporter
- [ ] Add error handling
- [ ] Write exporter tests

### Phase 5: CLI & Integration (Week 5)
- [ ] Implement CLI using cobra or similar
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Documentation

## Performance Considerations

1. **Streaming**: Parse line-by-line instead of loading entire file
2. **String Pooling**: Reuse strings where possible
3. **Pre-allocated Slices**: Use `make([]T, 0, capacity)` with known sizes
4. **Concurrent Validation**: Validate records in parallel (if needed)
5. **Memory Profiling**: Use `pprof` for optimization

## Testing Strategy

1. **Unit Tests**: Every function has tests
2. **Integration Tests**: End-to-end parsing/export cycles
3. **Fuzz Testing**: Use Go's fuzzing for malformed input
4. **Benchmark Tests**: Performance regression testing
5. **Example Tests**: Documentation through examples

## Dependencies

```go
// go.mod
module github.com/yourorg/gedcom

go 1.21

require (
    github.com/spf13/cobra v1.8.0  // CLI
    github.com/stretchr/testify v1.8.4  // Testing utilities
)
```

## Example Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/yourorg/gedcom/pkg"
)

func main() {
    g := gedcom.NewGedcom()
    
    // Parse a GEDCOM file
    if err := g.Parse("ged", "sample.ged"); err != nil {
        log.Fatalf("Failed to parse: %v", err)
    }
    
    // Check for errors
    if g.ErrorManager().HasErrors() {
        fmt.Println("Validation errors found:")
        for _, err := range g.ErrorManager().Errors() {
            fmt.Printf("  %s\n", err)
        }
    }
    
    // Access records
    individuals := g.Individuals()
    for xrefID, indi := range individuals {
        fmt.Printf("Individual %s: %s\n", xrefID, indi.GetValue("NAME"))
    }
    
    // Export to JSON
    if err := g.Export("json", "output.json"); err != nil {
        log.Fatalf("Failed to export: %v", err)
    }
}
```

## Conclusion

This Go implementation addresses all issues from the Python version:

✅ **Type Safety**: Compile-time type checking
✅ **Error Handling**: Explicit error returns, no exceptions
✅ **Thread Safety**: Mutex-protected access
✅ **Memory Efficiency**: Streaming parser
✅ **Validation**: Comprehensive file and line validation
✅ **Testing**: Built-in from the start
✅ **Performance**: Leverages Go's efficiency
✅ **Maintainability**: Clear interfaces, good separation of concerns

The design follows Go idioms and best practices while maintaining the same functionality as the Python implementation.

