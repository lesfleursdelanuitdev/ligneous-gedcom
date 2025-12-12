# Incremental Development Plan - GEDCOM Parser

## Overview

This document outlines an incremental approach to building the GEDCOM parser, starting with the core parsing functionality. We'll build in small, testable increments.

## Phase 1: Parser Implementation

### Goal
Build a working GEDCOM parser that can read a file and construct an in-memory tree structure.

### Algorithm: Stack-Based Hierarchical Parser

#### Core Algorithm

The parser uses a **stack-based algorithm** to build the hierarchical tree structure:

```
Algorithm: ParseGEDCOM(file)
Input: file path
Output: GedcomTree structure

1. Initialize:
   - parentsStack = []  // Stack of parent lines
   - currentValue = ""  // Accumulated CONC/CONT value
   - lastTag = nil       // Last processed tag info
   - lineNumber = 0
   - tree = NewGedcomTree()

2. Open file and detect encoding

3. For each line in file:
   a. lineNumber++
   b. line = trim(line)
   
   c. If line is empty, skip
   
   d. Parse line → (level, tag, value, xrefID, error)
      If error: log warning, continue
   
   e. Handle CONC/CONT:
      If tag == "CONC" or tag == "CONT":
         - Validate CONC/CONT rules
         - Accumulate value in currentValue
         - Continue to next line
   
   f. Apply accumulated value:
      If currentValue != "":
         - Set parentsStack[-1].Value = currentValue
         - Reset currentValue
   
   g. Handle level 0 (top-level record):
      If level == 0:
         - Create GedcomLine(level, tag, value, xrefID)
         - Create Record from line
         - Add record to tree
         - Reset parentsStack = [line]
         - Continue to next line
   
   h. Find parent level:
      While parentsStack not empty AND parentsStack[-1].Level >= level:
         Pop from parentsStack
      
      If parentsStack is empty:
         - Log warning (orphaned line)
         - Continue to next line
   
   i. Add as child:
      - Create GedcomLine(level, tag, value, "")
      - parent = parentsStack[-1]
      - parent.AddChild(line)
      - Push line to parentsStack
   
   j. Update lastTag

4. Handle remaining CONC/CONT value

5. Build xref index

6. Return tree
```

#### Why Stack-Based?

**Advantages:**
1. **Natural fit**: GEDCOM hierarchy matches stack LIFO behavior
2. **Efficient**: O(1) parent lookup
3. **Simple**: Easy to understand and implement
4. **Handles nesting**: Automatically handles arbitrary depth
5. **Error recovery**: Can handle orphaned lines gracefully

**How it works:**
- When level decreases, pop stack until parent level < current level
- Stack always contains current parent chain
- Top of stack = immediate parent for next child

**Example:**
```
0 HEAD          → Stack: [HEAD]
1 GEDC          → Stack: [HEAD, GEDC]
2 VERS 5.5.5    → Stack: [HEAD, GEDC, VERS]
1 CHAR UTF-8    → Stack: [HEAD, CHAR]  (popped GEDC, VERS)
0 @I1@ INDI     → Stack: [INDI]  (popped HEAD, CHAR)
1 NAME John     → Stack: [INDI, NAME]
2 GIVN John     → Stack: [INDI, NAME, GIVN]
2 SURN Doe      → Stack: [INDI, NAME, SURN]
```

### Incremental Steps

#### Step 1.1: Line Parser (Foundation)
**Goal**: Parse a single GEDCOM line into components

**What to build:**
- Function: `ParseLine(line string) (level int, tag string, value string, xrefID string, err error)`
- Handle all line formats:
  - `0 HEAD`
  - `0 @I1@ INDI`
  - `1 NAME John /Doe/`
  - `2 DATE 1 Jan 1900`

**Tests:**
- Valid lines (all formats)
- Invalid level (non-numeric, negative)
- Missing parts
- XREF detection
- Edge cases (empty value, only level+tag)

**Deliverable:**
- `internal/parser/line.go` with `ParseLine()` function
- Comprehensive test suite

---

#### Step 1.2: Encoding Detection
**Goal**: Detect file encoding (UTF-8, UTF-16, ANSEL)

**What to build:**
- Function: `DetectEncoding(filePath string) (encoding string, err error)`
- BOM detection (UTF-8, UTF-16)
- Fallback to UTF-8
- Read CHAR tag from header for verification

**Tests:**
- UTF-8 with BOM
- UTF-8 without BOM
- UTF-16 BE/LE
- Invalid files

**Deliverable:**
- `internal/parser/encoding.go`
- Test files with different encodings

---

#### Step 1.3: File Validation
**Goal**: Validate file before parsing

**What to build:**
- Function: `ValidateFile(filePath string) error`
- Check: exists, is file, readable, not empty
- Return clear error messages

**Tests:**
- Non-existent file
- Directory instead of file
- Unreadable file
- Empty file
- Valid file

**Deliverable:**
- Add to `internal/parser/gedcom.go`
- Error handling

---

#### Step 1.4: CONC/CONT Handler
**Goal**: Handle continuation lines correctly

**What to build:**
- Function: `HandleContinuation(tag string, level int, value string, lastTag *TagInfo, currentValue *strings.Builder) error`
- CONC: direct concatenation
- CONT: add newline then concatenate
- Validation: CONC/CONT cannot be subordinate to another CONC/CONT

**Tests:**
- Simple CONC
- Simple CONT
- Multiple CONC
- Multiple CONT
- Mixed CONC/CONT
- Invalid nesting

**Deliverable:**
- Add to `internal/parser/gedcom.go`
- Test cases

---

#### Step 1.5: Basic Tree Building (Level 0 Only)
**Goal**: Parse level 0 records only (no hierarchy yet)

**What to build:**
- Parse only level 0 lines
- Create GedcomLine for each
- Create basic Record
- Add to tree
- Skip all level > 0 lines for now

**Tests:**
- Parse file with only level 0 records
- Verify records created
- Verify xref IDs captured

**Deliverable:**
- Basic parser skeleton
- Can parse HEAD, TRLR, INDI, FAM records (without children)

---

#### Step 1.6: Stack Implementation
**Goal**: Implement parent stack management

**What to build:**
- Stack operations: push, pop, peek
- Level comparison logic
- Parent finding algorithm

**Tests:**
- Stack push/pop operations
- Level comparison
- Finding correct parent

**Deliverable:**
- Stack utilities in parser

---

#### Step 1.7: Hierarchical Parsing (Full Tree)
**Goal**: Parse complete hierarchical structure

**What to build:**
- Complete stack-based algorithm
- Handle all levels
- Build parent-child relationships
- Handle orphaned lines (log, continue)

**Tests:**
- Simple hierarchy (2 levels)
- Deep hierarchy (5+ levels)
- Orphaned lines
- Level decreases
- Level increases

**Deliverable:**
- Complete parser implementation
- Can parse full GEDCOM structure

---

#### Step 1.8: Error Handling & Recovery
**Goal**: Graceful error handling

**What to build:**
- Error collection (ErrorManager)
- Continue parsing after errors
- Log all errors with line numbers
- Return errors at end

**Tests:**
- Malformed lines
- Encoding errors
- Missing parents
- Invalid xrefs

**Deliverable:**
- Robust error handling
- Error reporting

---

#### Step 1.9: Integration & Testing
**Goal**: End-to-end parser testing

**What to build:**
- Integration tests with real GEDCOM files
- Performance testing
- Memory profiling
- Edge case handling

**Tests:**
- Parse sample.ged
- Parse large files
- Parse malformed files
- Verify tree structure

**Deliverable:**
- Complete, tested parser
- Documentation

---

## Algorithm Details

### Line Parsing Algorithm

```
ParseLine(line string) → (level, tag, value, xrefID, error)

1. Trim whitespace
2. If empty, return error
3. Split by whitespace (max 3 parts)
4. If len(parts) < 2, return error
5. Parse level:
   - level = parseInt(parts[0])
   - If invalid or < 0, return error
6. Check for xref:
   - If len(parts) == 3 AND parts[1] starts with '@':
     → Format: level xref tag
     → Return (level, parts[2], "", parts[1], nil)
   - Else if len(parts) == 3:
     → Format: level tag value
     → Return (level, parts[1], parts[2], "", nil)
   - Else:
     → Format: level tag
     → Return (level, parts[1], "", "", nil)
```

### Stack Management Algorithm

```
FindParentLevel(parentsStack, currentLevel) → []*GedcomLine

1. While len(parentsStack) > 0:
   a. If parentsStack[-1].Level < currentLevel:
      → Found parent, return stack
   b. Else:
      → Pop from stack
2. Return empty stack (orphaned line)
```

### CONC/CONT Algorithm

```
HandleContinuation(tag, level, value, lastTag, currentValue)

1. Validate:
   - If lastTag is CONC/CONT and level < lastTag.level:
     → Error: CONC/CONT cannot be subordinate
2. Accumulate:
   - If tag == "CONC":
     → currentValue.WriteString(value)
   - Else if tag == "CONT":
     → currentValue.WriteString("\n")
     → currentValue.WriteString(value)
3. Update lastTag
```

### Tree Building Algorithm

```
BuildTree(file) → GedcomTree

1. Initialize stack, currentValue, lastTag
2. For each line:
   a. Parse line
   b. Handle CONC/CONT (accumulate, continue)
   c. Apply accumulated value if needed
   d. If level 0:
      - Create record
      - Add to tree
      - Reset stack
   e. Else:
      - Find parent (pop stack)
      - Add as child
      - Push to stack
3. Handle remaining CONC/CONT
4. Build xref index
5. Return tree
```

## Data Structures Needed

### For Parser

```go
// Tag information for CONC/CONT tracking
type TagInfo struct {
    Tag   string
    Level int
}

// Parser state
type ParserState struct {
    ParentsStack []*GedcomLine
    CurrentValue strings.Builder
    LastTag      *TagInfo
    LineNumber   int
    Tree         *GedcomTree
}
```

## Error Handling Strategy

### Error Types
1. **File Errors**: Not found, not readable, empty
2. **Encoding Errors**: Cannot detect, cannot decode
3. **Line Errors**: Malformed, invalid level, missing parts
4. **Structure Errors**: Orphaned lines, invalid hierarchy
5. **XREF Errors**: Invalid format, duplicate, missing target

### Error Severity
- **Severe**: File errors, encoding errors (stop parsing)
- **Warning**: Line errors, structure errors (continue parsing)

### Error Collection
- Collect all errors during parsing
- Continue parsing when possible
- Return errors at end
- Include line numbers and context

## Testing Strategy

### Unit Tests
- Line parsing (all formats, edge cases)
- Encoding detection
- File validation
- CONC/CONT handling
- Stack operations
- Error handling

### Integration Tests
- Parse sample.ged (known good file)
- Parse files with errors (verify recovery)
- Parse large files (performance)
- Parse edge cases (deep nesting, many children)

### Test Files Needed
- `testdata/simple.ged` - Minimal valid file
- `testdata/sample.ged` - Full example
- `testdata/malformed.ged` - Various errors
- `testdata/large.ged` - Performance test
- `testdata/deep.ged` - Deep nesting test

## Implementation Order

1. **Week 1: Foundation**
   - Step 1.1: Line parser
   - Step 1.2: Encoding detection
   - Step 1.3: File validation
   - Tests for each

2. **Week 2: Core Parsing**
   - Step 1.4: CONC/CONT handler
   - Step 1.5: Basic tree building (level 0)
   - Step 1.6: Stack implementation
   - Tests for each

3. **Week 3: Complete Parser**
   - Step 1.7: Hierarchical parsing
   - Step 1.8: Error handling
   - Step 1.9: Integration testing
   - Documentation

## Success Criteria

Parser is complete when:
- ✅ Can parse valid GEDCOM files correctly
- ✅ Builds correct tree structure
- ✅ Handles errors gracefully
- ✅ All tests pass
- ✅ Performance acceptable (< 1s for 10K lines)
- ✅ Memory efficient (streaming, no full file load)

## Next Steps After Parser

Once parser is complete:
1. Record types (Individual, Family, etc.)
2. Validators
3. Exporters
4. CLI

But for now, focus on getting a solid parser foundation!

