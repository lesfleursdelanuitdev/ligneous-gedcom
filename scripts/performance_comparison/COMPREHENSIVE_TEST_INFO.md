# Comprehensive Performance Comparison Test

## Test Configuration

- **Iterations per file:** 2,000
- **Total test files:** 9 files
- **Total iterations:** 18,000 parser runs
- **Estimated duration:** 10-30 minutes (depending on file sizes)

## Test Files

### From User's Project
1. **user-royal92** - `testdata/royal92.ged` (488 KB, 30,683 lines)

### From family-tree/gedcom
2. **family-tree-gracis** - `family-tree/gedcom/gracis.ged` (163 KB, 10,324 lines)
3. **family-tree-xavier** - `family-tree/gedcom/xavier.ged` (101 KB, 5,822 lines)
4. **family-tree-tree1** - `family-tree/gedcom/tree1.ged` (211 KB, 12,714 lines)

### From gedcom-go-cacack/testdata/gedcom-5.5
5. **cacack-5.5-royal92** - `gedcom-go-cacack/testdata/gedcom-5.5/royal92.ged` (458 KB)
6. **cacack-5.5-pres2020** - `gedcom-go-cacack/testdata/gedcom-5.5/pres2020.ged` (1.1 MB)
7. **cacack-5.5-minimal** - `gedcom-go-cacack/testdata/gedcom-5.5/minimal.ged` (170 B)

### From gedcom-go-cacack/testdata/gedcom-5.5.1
8. **cacack-5.5.1-comprehensive** - `gedcom-go-cacack/testdata/gedcom-5.5.1/comprehensive.ged` (4.6 KB)
9. **cacack-5.5.1-minimal** - `gedcom-go-cacack/testdata/gedcom-5.5.1/minimal.ged` (204 B)

## Metrics Collected

For each file and parser, the test collects:

1. **Average time** - Mean of 2,000 iterations
2. **Min/Max time** - Fastest and slowest runs
3. **Percentiles:**
   - P50 (median)
   - P95 (95th percentile)
   - P99 (99th percentile)
4. **Throughput** - KB/s and MB/s
5. **Comparison ratio** - Speed difference between parsers

## Running the Test

```bash
cd /apps/gedcom-go
go test -v ./scripts/performance_comparison -run TestComprehensiveComparison -timeout 30m
```

The test will:
- Run 2,000 iterations for each file
- Compare ParallelHierarchicalParser vs cacack parser
- Generate detailed statistics
- Print a summary table at the end

## Output

The test outputs:
1. Per-file detailed statistics
2. Summary table comparing all files
3. Overall statistics (how many files each parser wins, average ratio)

## Expected Results

Based on previous tests:
- **Small files (< 1 KB):** cacack parser likely faster (3-6x)
- **Medium files (100-500 KB):** ParallelHierarchicalParser likely faster (1.05-1.1x)
- **Large files (> 1 MB):** Need to test (pres2020.ged)

The 2,000 iterations provide:
- High statistical confidence
- Better understanding of variance
- More accurate percentile calculations

