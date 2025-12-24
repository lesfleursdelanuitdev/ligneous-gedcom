# Code Comparison: gedcom-go vs elliotchance/gedcom

## Overview

This document compares the parsing implementations between `gedcom-go` (your project) and `elliotchance/gedcom`.

## Key Architectural Differences

### 1. Line Parsing Strategy

#### gedcom-go: Manual Byte Parsing (`ParseLineFast`)
```go
// Manual byte-by-byte parsing for performance
func ParseLineFast(line string) (level int, tag string, value string, xrefID string, err error) {
    // Finds spaces manually, parses level, tag, xref, value
    // No regex overhead
    // ~2-3x faster than string splitting
}
```

**Advantages:**
- ✅ **Performance**: Manual byte parsing avoids regex overhead
- ✅ **Memory**: No intermediate string slices created
- ✅ **Control**: Fine-grained control over parsing logic

**Trade-offs:**
- ⚠️ More code to maintain
- ⚠️ Manual parsing logic

#### elliotchance/gedcom: Regex Parsing
```go
var lineRegexp = regexp.MustCompile(`^(\d) +(@[^@]+@ )?(\w+) ?(.*)?$`)

func parseLine(line string, ...) (Node, int, error) {
    parts := lineRegexp.FindStringSubmatch(line)
    // Extract level, pointer, tag, value from regex groups
}
```

**Advantages:**
- ✅ **Simplicity**: Regex is concise and readable
- ✅ **Maintainability**: Easy to understand pattern

**Trade-offs:**
- ⚠️ **Performance**: Regex compilation and matching overhead
- ⚠️ **Memory**: Creates string slices for capture groups

### 2. Hierarchy Management

#### gedcom-go: Custom Stack (`LineStack`)
```go
type HierarchicalParser struct {
    parentsStack *LineStack  // Custom stack implementation
    // ...
}

// Stack tracks parent-child relationships
parent, err := hp.parentsStack.FindParent(level)
```

**Approach:**
- Custom stack data structure
- Tracks `GedcomLine` objects (not full records)
- `FindParent(level)` walks up stack to find appropriate parent

**Advantages:**
- ✅ **Explicit**: Clear parent-child relationship tracking
- ✅ **Flexible**: Can handle complex nesting scenarios
- ✅ **Type-safe**: Works with `GedcomLine` objects

#### elliotchance/gedcom: Indent Slice (`Nodes{}`)
```go
indents := Nodes{}  // Slice of nodes at each indent level

// For each line:
if indent == 0 {
    indents = Nodes{node}  // Reset for root
} else {
    i := indents[indent-1]  // Get parent
    i.AddNode(node)         // Add as child
    // Update indents slice based on indent level
}
```

**Approach:**
- Uses a slice where `indents[i]` = node at level `i`
- Direct array indexing: `parent = indents[indent-1]`
- Updates slice based on indent changes

**Advantages:**
- ✅ **Simple**: Direct array indexing is fast
- ✅ **Efficient**: O(1) parent lookup
- ✅ **Memory**: Single slice, no stack overhead

**Trade-offs:**
- ⚠️ Requires careful slice management
- ⚠️ Must handle indent level changes correctly

### 3. Line Reading

#### gedcom-go: `bufio.Scanner`
```go
scanner := bufio.NewScanner(reader)
for scanner.Scan() {
    line := scanner.Text()
    // Process line
}
```

**Advantages:**
- ✅ **Standard library**: Well-tested, optimized
- ✅ **Encoding-aware**: Works with custom readers
- ✅ **Simple**: Easy to use

#### elliotchance/gedcom: Custom `readLine()`
```go
func (dec *Decoder) readLine() (string, error) {
    buf := new(bytes.Buffer)
    for {
        b, err := dec.r.ReadByte()
        if b == '\n' || b == '\r' {
            break
        }
        buf.WriteByte(b)
    }
    return string(buf.Bytes()), nil
}
```

**Advantages:**
- ✅ **Control**: Handles `\n` and `\r` explicitly
- ✅ **Flexibility**: Can handle edge cases

**Trade-offs:**
- ⚠️ More code to maintain
- ⚠️ Manual byte reading

### 4. CONC/CONT Continuation Handling

#### gedcom-go: Dedicated `ContinuationHandler`
```go
type ContinuationHandler struct {
    accumulatedValue strings.Builder
    lastTag          string
    lastLevel        int
    // ...
}

// Accumulates CONC/CONT values
hp.continuationHandler.HandleContinuation(tag, level, value)

// Applies accumulated value to previous line
if hp.continuationHandler.HasAccumulatedValue() {
    // Append to previous line's value
}
```

**Approach:**
- Separate handler for CONC/CONT logic
- Accumulates values using `strings.Builder`
- Applies accumulated value when non-CONC/CONT line encountered

**Advantages:**
- ✅ **Explicit**: Clear separation of concerns
- ✅ **Efficient**: Uses `strings.Builder` for concatenation
- ✅ **Correct**: Handles CONC (concatenate) vs CONT (newline) correctly

#### elliotchance/gedcom: No Explicit CONC/CONT Handling
```go
// No dedicated CONC/CONT handler in decoder.go
// CONC/CONT tags are treated as regular nodes
```

**Approach:**
- CONC/CONT tags are parsed as regular nodes
- No special accumulation logic in decoder
- May handle in post-processing or node traversal

**Trade-offs:**
- ⚠️ CONC/CONT values may not be automatically concatenated
- ⚠️ Requires additional processing to reconstruct full values

### 5. Node/Record Creation

#### gedcom-go: Factory Pattern with Reuse
```go
type HierarchicalParser struct {
    factory *gedcom.RecordFactory  // Created once, reused
}

// Create record using factory
record := hp.factory.CreateRecord(gedcomLine)
```

**Approach:**
- Factory instance created once per parser
- Reused for all records
- Factory creates typed records based on tag

**Advantages:**
- ✅ **Performance**: Avoids repeated factory creation
- ✅ **Memory**: Single factory instance
- ✅ **Type-safe**: Creates appropriate record types

#### elliotchance/gedcom: Function-Based Creation
```go
func newNode(document *Document, family *FamilyNode, tag Tag, value, pointer string) Node {
    // Large switch statement
    switch tag {
    case TagIndividual:
        node = newIndividualNode(document, pointer, children...)
    case TagFamily:
        node = newFamilyNode(document, pointer, children...)
    // ... many cases
    }
    return node
}
```

**Approach:**
- Function-based node creation
- Large switch statement for tag-to-type mapping
- Creates typed nodes (IndividualNode, FamilyNode, etc.)

**Advantages:**
- ✅ **Explicit**: Clear tag-to-type mapping
- ✅ **Comprehensive**: Handles many tag types

**Trade-offs:**
- ⚠️ Large switch statement (100+ cases)
- ⚠️ Function call overhead per node

### 6. Error Handling

#### gedcom-go: Error Manager (Collects, Continues)
```go
type ErrorManager struct {
    errors []*GedcomError
}

// Collects errors but continues parsing
hp.errorManager.AddError(gedcom.SeverityWarning, msg, lineNumber, category)
// Parsing continues even with errors
```

**Approach:**
- Collects errors/warnings in `ErrorManager`
- Continues parsing despite errors
- Returns tree + errors at end

**Advantages:**
- ✅ **Resilient**: Can parse partial/invalid files
- ✅ **Informative**: Collects all errors for reporting
- ✅ **Non-fatal**: Warnings don't stop parsing

#### elliotchance/gedcom: Immediate Error Return (Stops)
```go
node, indent, err := parseLine(line, document, family)
if err != nil {
    if dec.AllowMultiLine && previousNode != nil {
        // Try to recover
        previousNode.RawSimpleNode().value += "\n" + line
        continue
    }
    return nil, fmt.Errorf("line %d: %s", lineNumber, err)
}
```

**Approach:**
- Returns error immediately on parse failure
- Stops parsing (unless `AllowMultiLine` recovery)
- Optional recovery for multi-line values

**Advantages:**
- ✅ **Strict**: Ensures valid GEDCOM
- ✅ **Simple**: Clear error propagation

**Trade-offs:**
- ⚠️ **Less resilient**: Stops on first error
- ⚠️ **Less informative**: Only reports first error

### 7. Encoding Handling

#### gedcom-go: Explicit Encoding Detection
```go
// Step 2: Detect encoding
encoding, err := DetectEncoding(filePath)

// Step 4: Get reader with proper encoding
reader, err := GetReader(file, encoding)
```

**Approach:**
- Explicit encoding detection from file
- Creates encoding-aware reader
- Handles BOM automatically

#### elliotchance/gedcom: BOM Handling Only
```go
document.HasBOM = dec.consumeOptionalBOM()
// Uses bufio.Reader directly (assumes UTF-8)
```

**Approach:**
- Only handles BOM detection
- Assumes UTF-8 encoding
- No explicit encoding detection

## Performance Implications

### gedcom-go Advantages:
1. **Manual byte parsing** → ~2-3x faster line parsing
2. **Factory reuse** → Less allocation overhead
3. **strings.Builder for CONC/CONT** → Efficient string concatenation
4. **Optimized scanner buffer** → Better I/O performance

### elliotchance/gedcom Advantages:
1. **Direct array indexing** → O(1) parent lookup (vs stack walk)
2. **Simpler code** → Less overhead from abstractions

## Code Complexity

### gedcom-go:
- **More files**: Separate files for stack, continuation handler, line parsing
- **More abstractions**: Factory, ErrorManager, ContinuationHandler
- **More code**: ~500+ lines for core parser

### elliotchance/gedcom:
- **Fewer files**: Most logic in `decoder.go`
- **Fewer abstractions**: Direct function calls
- **Less code**: ~400 lines for decoder

## Recommendations

### From elliotchance/gedcom to gedcom-go:
1. ✅ **Keep manual byte parsing** - Already faster
2. ✅ **Keep factory reuse** - Already optimized
3. ✅ **Keep ErrorManager** - More resilient
4. ✅ **Keep ContinuationHandler** - Correct CONC/CONT handling

### Potential Improvements from elliotchance/gedcom:
1. **Consider indent slice** - O(1) parent lookup vs stack walk
   - Could improve performance for deeply nested structures
   - Trade-off: Less explicit, requires careful slice management

2. **Simplify node creation** - If factory overhead becomes issue
   - But current factory reuse is already optimized

## Conclusion

**gedcom-go** has:
- ✅ Better performance optimizations (manual parsing, factory reuse)
- ✅ More resilient error handling
- ✅ Better CONC/CONT handling
- ✅ More explicit architecture

**elliotchance/gedcom** has:
- ✅ Simpler code structure
- ✅ O(1) parent lookup (indent slice)
- ✅ Fewer abstractions

**Overall**: Your parser is already well-optimized. The main potential improvement would be switching from stack-based to indent-slice-based parent lookup, but this is a trade-off between explicitness and performance.

