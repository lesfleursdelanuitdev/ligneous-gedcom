# Test Data Update

## Summary

Updated comprehensive comparison tests to use the new test data files that are now available in the `testdata/` directory.

## Files Updated

1. **`comprehensive_comparison_test.go`** - Updated to use testdata directory files
2. **`internal_parser_comparison_test.go`** - Updated to use testdata directory files

## Test Files Now Used

### From `/apps/gedcom-go/testdata/` (Primary Source)
- ✅ **user-royal92** - `testdata/royal92.ged` (488 KB)
- ✅ **user-pres2020** - `testdata/pres2020.ged` (1.1 MB) - NEW
- ✅ **user-gracis** - `testdata/gracis.ged` (163 KB) - NEW
- ✅ **user-xavier** - `testdata/xavier.ged` (101 KB) - NEW
- ✅ **user-tree1** - `testdata/tree1.ged` (212 KB) - NEW

### From `/apps/gedcom-go-cacack/testdata/` (Reference)
- **cacack-5.5-royal92** - `gedcom-go-cacack/testdata/gedcom-5.5/royal92.ged` (458 KB)
- **cacack-5.5-pres2020** - `gedcom-go-cacack/testdata/gedcom-5.5/pres2020.ged` (1.1 MB)
- **cacack-5.5-minimal** - `gedcom-go-cacack/testdata/gedcom-5.5/minimal.ged` (170 B)
- **cacack-5.5.1-comprehensive** - `gedcom-go-cacack/testdata/gedcom-5.5.1/comprehensive.ged` (4.6 KB)
- **cacack-5.5.1-minimal** - `gedcom-go-cacack/testdata/gedcom-5.5.1/minimal.ged` (204 B)

## Total Test Files

**10 test files** will be tested (5 from user's testdata + 5 from cacack testdata)

## Benefits

1. **All files in one place** - Easier to manage test data
2. **More comprehensive testing** - Now includes pres2020, gracis, xavier, tree1
3. **Consistent paths** - All user files use `../../testdata/` path
4. **Better coverage** - Tests files from 101 KB to 1.1 MB

## File Sizes

| File | Size | Lines | Type |
|------|------|-------|------|
| royal92.ged | 488 KB | 30,683 | Large |
| pres2020.ged | 1.1 MB | 49,432 | Very Large |
| tree1.ged | 212 KB | 12,714 | Medium-Large |
| gracis.ged | 163 KB | 10,324 | Medium |
| xavier.ged | 101 KB | 5,822 | Medium |

## Next Run

When you run the comprehensive comparison test, it will now test all 10 files:
- 5 files from your testdata directory
- 5 files from cacack testdata directory

This provides a more comprehensive performance comparison across different file sizes.

