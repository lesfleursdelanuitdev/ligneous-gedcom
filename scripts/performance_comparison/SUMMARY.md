# Parser Performance Comparison Summary

## Test Results

### Internal Parser Comparison (royal92.ged - 488KB)

| Parser | Time | Throughput | vs Baseline |
|--------|------|------------|-------------|
| HierarchicalParser | 19.6ms | 24,864 KB/s | Baseline (1.00x) |
| ParallelHierarchicalParser | 18.5ms | 26,415 KB/s | **1.06x faster** |
| TwoPhaseParser | 19.4ms | 25,112 KB/s | **1.01x faster** |

### External Comparison (cacack/gedcom-go)

| File Size | cacack Parser | User Parser | Speedup |
|-----------|---------------|-------------|---------|
| 0.2 KB | 5-8µs | 36-43µs | **5-8x faster** |
| 4.58 KB | 96-110µs | 155-165µs | **1.5x faster** |
| 488 KB | ~15-20ms (estimated) | 19.6ms | ~1.0-1.3x faster |

## Key Findings

1. **ParallelHierarchicalParser** shows slight improvement (6%) for medium files
2. **TwoPhaseParser** shows minimal improvement (1%) - overhead may cancel benefits
3. **cacack parser** is significantly faster for small files, but gap narrows for larger files

## Optimization Opportunities

### High Priority
1. **Optimize ParseLine** - Use manual byte parsing instead of SplitN
   - Expected: 2-3x improvement
   - Implementation: `ParseLineFast` function created

2. **Remove redundant TrimSpace** - Line is trimmed twice
   - Expected: 5-10% improvement

### Medium Priority
3. **Use strings.Builder for CONC/CONT** - Replace `+=` concatenation
   - Expected: 10-20% improvement for files with many continuations

4. **Reuse RecordFactory** - Create once per parser instance
   - Expected: 5-10% improvement

### Low Priority
5. **Optimize Scanner Buffer** - Use larger buffer for large files
   - Expected: 5-10% improvement for large files

## Next Steps

1. ✅ Created optimized `ParseLineFast` function
2. ⏳ Update parsers to use `ParseLineFast`
3. ⏳ Add benchmarks to verify improvements
4. ⏳ Test with various file sizes
5. ⏳ Compare final performance with cacack parser

## How to Use Optimized Parser

To use the optimized parser, replace `ParseLine` calls with `ParseLineFast`:

```go
// Before
level, tag, value, xrefID, err := ParseLine(line)

// After (ensure line is already trimmed)
level, tag, value, xrefID, err := ParseLineFast(line)
```

Note: `ParseLineFast` assumes the input line is already trimmed. Remove the `strings.TrimSpace` call before parsing.

