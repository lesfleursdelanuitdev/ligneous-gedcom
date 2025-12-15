# Two-Phase Parsing Approach

## Overview

The two-phase parsing approach splits parsing into two distinct phases to enable parallelization where possible.

## Architecture

### Phase 1: Record Collection (Sequential)
- **Purpose**: Identify record boundaries and collect raw data
- **Process**: 
  - Read file sequentially
  - Identify level 0 records (HEAD, INDI, FAM, etc.)
  - Collect each record's raw child lines (unparsed)
- **Why Sequential**: Must identify record boundaries first
- **Output**: Array of `RawRecord` structures

### Phase 2: Record Parsing (Parallel)
- **Purpose**: Parse each record's children independently
- **Process**:
  - Use worker pool (4 workers by default)
  - Each worker processes records from a channel
  - Parse children using stack-based algorithm
  - Create records and add to tree
- **Why Parallel**: Records are independent once boundaries are known
- **Output**: Fully parsed `GedcomTree`

## Implementation

**File**: `internal/parser/two_phase_parser.go`

**Key Structures**:
```go
type RawRecord struct {
    Level      int
    Tag        string
    Value      string
    XrefID     string
    LineNumber int
    RawLines   []string // Unparsed child lines
}
```

## Performance

### Benchmark Results (gracis.ged - 10K lines)

- **Sequential Parser**: 6,072,244 ns/op (~6.1ms)
- **Two-Phase Parser**: 5,892,328 ns/op (~5.9ms)
- **Improvement**: ~3% faster

### Analysis

**Why Only ~3% Improvement?**
1. **I/O Bound**: File reading is still sequential
2. **Phase 1 Overhead**: Collecting records adds small overhead
3. **Record Size**: Most records are small, so parallelization overhead can exceed benefit
4. **Channel Overhead**: Goroutine communication has cost

**When Would It Help More?**
- Files with many large records (deep hierarchies)
- Files with 1000+ records
- CPU-bound parsing operations

## Usage

```go
import "github.com/yourorg/gedcom/internal/parser"

// Create two-phase parser
parser := parser.NewTwoPhaseParser()

// Parse file
tree, err := parser.Parse("large_file.ged")
if err != nil {
    log.Fatal(err)
}

// Check errors
errors := parser.GetErrors()
```

## Comparison with Other Approaches

| Approach | Speed | Complexity | Best For |
|----------|-------|------------|----------|
| **Sequential** | Baseline | Low | All files |
| **Parallel Parser** | ~4% faster | Medium | Minimal benefit |
| **Two-Phase** | ~3% faster | Medium | Large files with many records |

## Advantages

✅ **Conceptually Clean**: Clear separation of concerns
✅ **Parallelizable**: Phase 2 can use multiple workers
✅ **Maintainable**: Easier to understand than fully parallel approach
✅ **Scalable**: Can adjust worker count based on file size

## Disadvantages

⚠️ **Modest Improvement**: Only ~3% faster for typical files
⚠️ **Memory Overhead**: Stores raw lines before parsing
⚠️ **Complexity**: More complex than sequential parser
⚠️ **Line Numbers**: Approximate line numbers in phase 2

## Recommendations

### Use Two-Phase Parser When:
- ✅ Files with 1000+ records
- ✅ Records have deep hierarchies (many child lines)
- ✅ Parsing is a bottleneck
- ✅ Multiple CPU cores available

### Use Sequential Parser When:
- ✅ Small to medium files (< 1000 records)
- ✅ Simplicity is preferred
- ✅ Single-threaded environment
- ✅ Current performance is acceptable

## Future Enhancements

1. **Adaptive Worker Pool**: Adjust workers based on record count
2. **Streaming Phase 1**: Process records as they're collected
3. **Batch Processing**: Process records in batches
4. **Memory Optimization**: Stream raw lines instead of storing all

## Conclusion

The two-phase approach is a **smart architectural improvement** that enables parallelization where it makes sense. While the performance gain is modest (~3%) for typical files, it provides a foundation for future optimizations and could show better results for very large files with many complex records.

**Recommendation**: Use sequential parser for most cases. Two-phase parser is available for experimentation and large file scenarios.

