# Parser Optimization Analysis

## Current Performance Issues

Based on code analysis and comparison with cacack/gedcom-go parser, here are the identified bottlenecks:

### 1. **ParseLine Function** (High Impact)
**Location:** `internal/parser/line.go`

**Issues:**
- Uses `strings.SplitN` multiple times (creates multiple slices)
- Calls `strings.TrimSpace` even though line is already trimmed
- Multiple string operations per line

**Optimization:**
- Use manual byte/string indexing instead of SplitN
- Remove redundant TrimSpace call
- Use byte-level parsing for better performance

### 2. **String Concatenation** (Medium Impact)
**Location:** `internal/parser/gedcom.go` (lines 103-107)

**Issues:**
- Uses `+=` for string concatenation in CONC/CONT handling
- Creates new strings on each concatenation

**Optimization:**
- Use `strings.Builder` for accumulating CONC/CONT values

### 3. **Factory Pattern** (Low-Medium Impact)
**Location:** `internal/parser/gedcom.go` (line 118)

**Issues:**
- Creates new `RecordFactory` for every level-0 record
- Factory could be reused

**Optimization:**
- Create factory once and reuse it

### 4. **Double TrimSpace** (Low Impact)
**Location:** `internal/parser/gedcom.go` (line 73) and `line.go` (line 27)

**Issues:**
- Line is trimmed in Parse() and again in ParseLine()

**Optimization:**
- Remove TrimSpace from ParseLine (assume input is already trimmed)

### 5. **Scanner Buffer Size** (Low Impact)
**Location:** `internal/parser/gedcom.go` (line 65)

**Issues:**
- Uses default scanner buffer (64KB)
- For large files, larger buffer could reduce syscalls

**Optimization:**
- Use `bufio.NewScanner` with custom buffer size (256KB-1MB)

## Comparison Results

### cacack/gedcom-go vs HierarchicalParser
- **Small files (0.2 KB):** cacack is 5-8x faster
- **Medium files (4.58 KB):** cacack is 1.5x faster
- **Large files:** Need to test

### Internal Parser Comparison
- **HierarchicalParser:** Baseline (sequential)
- **ParallelHierarchicalParser:** May be slower due to goroutine overhead for small files
- **TwoPhaseParser:** May be faster for large files with many records

## Recommended Optimizations (Priority Order)

### Priority 1: Optimize ParseLine
- Manual byte parsing instead of SplitN
- Remove redundant TrimSpace
- **Expected improvement:** 2-3x faster parsing

### Priority 2: Use strings.Builder for CONC/CONT
- Replace `+=` with Builder
- **Expected improvement:** 10-20% faster for files with many continuations

### Priority 3: Reuse RecordFactory
- Create factory once per parser instance
- **Expected improvement:** 5-10% faster

### Priority 4: Optimize Scanner Buffer
- Use larger buffer for large files
- **Expected improvement:** 5-10% faster for large files

### Priority 5: Profile and Optimize Stack Operations
- Check if FindParent does linear search
- Consider using slice with binary search or better data structure
- **Expected improvement:** Variable, depends on hierarchy depth

## Implementation Plan

1. Create optimized `ParseLineFast` function
2. Update all parsers to use optimized version
3. Add benchmarks to verify improvements
4. Test with various file sizes
5. Compare with cacack parser performance

