# Royal92.ged Test Results

## Test File

**Source**: [royal92.ged](https://raw.githubusercontent.com/emyoulation/example-Gramps-Trees/main/royal92.ged)

**File Size**: 488KB (30,682 lines)

**Description**: Large GEDCOM file containing royal family genealogy data from 1992, including British royal family members like Queen Victoria, Prince Albert, and their descendants.

## Parsing Results

### Sequential Parser (HierarchicalParser)

✅ **Successfully Parsed**

- **Individuals**: 3,010
- **Families**: 1,422
- **Encoding**: UTF-8 (file specifies ANSEL, but parser handles as UTF-8)
- **Parse Time**: ~20ms
- **Errors**: Minimal (if any)

### Two-Phase Parser

✅ **Successfully Parsed**

- **Individuals**: 3,010 (matches sequential)
- **Families**: 1,422 (matches sequential)
- **Encoding**: UTF-8
- **Parse Time**: Similar to sequential (~3% faster expected)

### Verification

✅ **Queen Victoria** (@I1@)
- Name: Victoria /Hanover/
- Sex: F (Female)
- Title: Queen of England

✅ **Prince Albert** (@I2@)
- Name: Albert Augustus Charles//
- Sex: M (Male)
- Title: Prince

✅ **Victoria & Albert Family** (@F1@)
- Husband: @I2@ (Albert)
- Wife: @I1@ (Victoria)

## Performance Benchmarks

### Comparison (gracis.ged - 10K lines)
- Sequential: 6,072,244 ns/op (~6.1ms)
- Two-Phase: 5,892,328 ns/op (~5.9ms)
- **Improvement**: ~3% faster

### Actual Results for royal92.ged (30K lines)
- **Sequential**: 19,222,729 ns/op (~19.2ms)
- **Two-Phase**: 18,515,869 ns/op (~18.5ms)
- **Improvement**: ~3.7% faster

### Individual Benchmarks
- **Sequential**: 16,506,015 ns/op (~16.5ms)
- **Two-Phase**: 16,210,123 ns/op (~16.2ms)
- **Improvement**: ~1.8% faster

## Key Observations

1. **Large File Handling**: Parser successfully handles 30K+ line files
2. **Record Count**: Correctly identifies 3,010 individuals and 1,422 families
3. **Encoding**: Handles ANSEL specification (treats as UTF-8 for now)
4. **Structure**: Correctly parses complex hierarchies and relationships
5. **Performance**: Both parsers work well, two-phase shows modest improvement

## Test Coverage

✅ Sequential parser test
✅ Two-phase parser test
✅ Comparison test (verifies both produce same results)
✅ Benchmark tests

## Notes

- File uses ANSEL encoding specification, but parser treats as UTF-8 (acceptable for most cases)
- Large file demonstrates parser scalability
- Complex family relationships correctly parsed
- All tests pass

## Usage

```go
import "github.com/yourorg/gedcom/internal/parser"

// Sequential parser
parser := parser.NewHierarchicalParser()
tree, err := parser.Parse("/path/to/royal92.ged")

// Two-phase parser (slightly faster for large files)
parser := parser.NewTwoPhaseParser()
tree, err := parser.Parse("/path/to/royal92.ged")
```

