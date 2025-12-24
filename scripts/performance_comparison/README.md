# Parser Performance Comparison

This directory contains performance comparison tests between:
- **cacack/gedcom-go**: Pure Go library for parsing GEDCOM files
- **lesfleursdelanuitdev/gedcom-go**: Research-grade genealogy toolkit

## Test Files

The comparison uses GEDCOM 5.5.1 compatible test files from:
- `gedcom-go-cacack/testdata/gedcom-5.5.1/` - GEDCOM 5.5.1 specific files
- `gedcom-go-cacack/testdata/gedcom-5.5/` - GEDCOM 5.5 files (compatible with 5.5.1)
- `gedcom-go/testdata/` - User's project test files
- `family-tree/gedcom/` - Additional test files

## Running the Tests

### Detailed Comparison Test

```bash
cd /apps/gedcom-go
go test -v ./scripts/performance_comparison -run TestParserComparison
```

This will:
- Test each parser on all available GEDCOM 5.5.1 compatible files
- Report parsing time, throughput, and memory usage
- Compare results side-by-side

### Benchmark Tests

Run benchmarks for the cacack parser:
```bash
go test -bench=BenchmarkCacackParser -benchmem ./scripts/performance_comparison
```

Run benchmarks for the user parser:
```bash
go test -bench=BenchmarkUserParser -benchmem ./scripts/performance_comparison
```

Run both and compare:
```bash
go test -bench=. -benchmem ./scripts/performance_comparison | tee benchmark_results.txt
```

## What's Tested

**Parser Performance Only** - This comparison focuses solely on parsing performance:
- File reading and parsing speed
- Memory allocations
- Throughput (KB/s)

**Not Tested** (these are separate concerns):
- Graph construction
- Query performance
- Validation
- Error handling

## Expected Results

Based on initial tests:
- **cacack/gedcom-go**: Optimized for pure parsing, typically faster for small-medium files
- **lesfleursdelanuitdev/gedcom-go**: More comprehensive toolkit with additional features, may be slightly slower for pure parsing but provides more functionality

## Notes

- Both parsers are tested with the same input files for fair comparison
- Files are read into memory once to eliminate I/O variance
- Each test runs 10 iterations and reports average time
- Allocation tracking requires using `-benchmem` flag with benchmarks

