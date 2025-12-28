# Query Package Testing Analysis

## Current Status

**Test Coverage: 69.6%** (Target: 80%+)
- **271 functions** with 0% coverage (excluding hybrid storage)
- **33 test files** currently exist
- Most tests use **synthetic data** (programmatically created)
- **Limited use** of real GEDCOM files from testdata folder

## Available Test Data Files

| File | Size | Lines | Description |
|------|------|-------|-------------|
| `royal92.ged` | 488KB | 30,683 | Large real-world dataset (Royal genealogy) |
| `pres2020.ged` | 1.1MB | ~large | Very large dataset (US Presidents) |
| `tree1.ged` | 212KB | 12,714 | Medium dataset |
| `gracis.ged` | 163KB | 10,324 | Medium dataset |
| `xavier.ged` | 101KB | 5,822 | Smaller dataset |

## Critical Areas Needing Testing

### 1. Functions with 0% Coverage (High Priority)

#### Filter Execution
- **`filterByBool`** (filter_execution.go:380) - 0%
  - Used by hybrid storage for boolean filters (HasChildren, HasSpouse, IsLiving)
  - **Test Strategy**: Use real GEDCOM files to test individuals with/without children, spouses, living status

#### Notes Query
- **`GetAllNotes`** (notes_query.go:20) - 0%
  - **Test Strategy**: Use royal92.ged or pres2020.ged which likely contain NOTE records

#### Graph Metrics
- **`Metrics`** (graph_metrics.go:31) - 0%
  - **Test Strategy**: Test metrics collection on real datasets

#### Graph Hybrid Helpers
- **`debugBuildLog`** (graph_hybrid_helpers.go:20) - 0%
  - Debug function, lower priority

### 2. Functions with Low Coverage (<80%)

#### Relationship Calculations
- **`getCollateralRelationshipType`** (relationships.go:186) - 61.5%
  - Tests cousins, uncles, aunts, nephews, nieces
  - **Test Strategy**: Use real datasets to find actual cousin relationships
  
- **`getAncestralRelationshipType`** (relationships.go:178) - 66.7%
  - Tests grandparent, great-grandparent relationships
  - **Test Strategy**: Use royal92.ged which has deep family trees

- **`min`** (relationships.go:220) - 66.7%
- **`abs`** (relationships.go:227) - 66.7%
  - Helper functions, need edge cases

#### Ancestor/Descendant Queries
- **`findAncestorsWithDepth`** (ancestor_query.go:200) - 72.2%
  - Tests depth-limited ancestor queries
  - **Test Strategy**: Use pres2020.ged or royal92.ged with deep lineages

- **`Count`** (ancestor_query.go:133) - 75.0%
- **`Exists`** (ancestor_query.go:142) - 75.0%
  - Need tests for empty results, non-existent individuals

- **`Count`** (descendant_query.go:120) - 75.0%
  - Similar to ancestor Count

#### Graph Relationship Helpers
- **`GetChildren`** (graph_relationships.go:21) - 75.0%
- **`GetParents`** (graph_relationships.go:32) - 75.0%
- **`GetSiblings`** (graph_relationships.go:43) - 75.0%
- **`GetFamilyHusband`** (graph_relationships.go:54) - 75.0%
- **`GetFamilyWife`** (graph_relationships.go:65) - 75.0%
- **`GetFamilyChildren`** (graph_relationships.go:76) - 75.0%
  - **Test Strategy**: Use real GEDCOM files to test with actual family structures

#### Path Finding
- **`reconstructBidirectionalPath`** (path_finding.go:156) - 80.0%
  - Tests bidirectional BFS path reconstruction
  - **Test Strategy**: Use datasets with complex relationship paths

- **`AllPaths`** (path_finding.go:232) - 83.3%
  - Tests finding all paths between individuals
  - **Test Strategy**: Use royal92.ged to find multiple paths between related individuals

- **`allPathsDFS`** (path_finding.go:263) - 85.0%
  - DFS implementation for all paths
  - **Test Strategy**: Test with datasets containing multiple relationship paths

#### Filter Execution
- **`executeHybrid`** (filter_execution.go:208) - 58.5%
  - Hybrid storage execution (excluded from current focus)
  
- **`buildCacheKey`** (filter_execution.go:333) - 75.0%
  - Cache key generation, needs edge cases

## Recommended Testing Strategy

### Phase 1: Integration Tests with Real Data

#### 1.1 Create Test Helper for testdata Files
- Create `query/testdata_helper.go` similar to `parser/testdata_helper.go`
- Standardize path resolution for testdata files
- Support both relative and absolute paths

#### 1.2 Test Query Operations on Real Files
**Priority: HIGH**

For each testdata file (royal92.ged, pres2020.ged, tree1.ged, gracis.ged, xavier.ged):

1. **Graph Building**
   - Test BuildGraph with each file
   - Verify node counts match expected
   - Verify edge counts are reasonable

2. **Basic Queries**
   - Test Individual queries on known XREFs
   - Test Family queries
   - Test Filter queries (by name, sex, dates)

3. **Relationship Queries**
   - Test Ancestors() on individuals with known ancestors
   - Test Descendants() on individuals with known descendants
   - Test Siblings(), Parents(), Children(), Spouses()
   - Test Cousins(), Uncles(), Nephews(), Grandparents(), Grandchildren()

4. **Path Finding**
   - Test PathTo() between known related individuals
   - Test AllPaths() for complex relationships
   - Test ShortestPath() for direct relationships

5. **Relationship Calculations**
   - Test CalculateRelationship() for various relationship types
   - Test collateral relationships (cousins, etc.)
   - Test ancestral relationships (grandparents, etc.)

6. **Collection Queries**
   - Test Names(), Places(), Events(), Families() collections
   - Test Unique(), By(), All() methods
   - Test Count() and Execute() methods

### Phase 2: Edge Cases and Error Handling

#### 2.1 Non-existent Individuals
- Test queries on XREFs that don't exist
- Test relationships between unrelated individuals
- Test path finding when no path exists

#### 2.2 Empty Results
- Test Count() when no results
- Test Exists() for false cases
- Test collections with empty datasets

#### 2.3 Boundary Conditions
- Test MaxGenerations limits
- Test MaxLength for path queries
- Test filters with extreme values

### Phase 3: Complex Scenarios

#### 3.1 Large Dataset Testing
- Use pres2020.ged (1.1MB) for performance and correctness
- Test deep ancestor/descendant queries
- Test complex path finding

#### 3.2 Multi-generational Families
- Use royal92.ged for deep family trees
- Test relationship calculations across many generations
- Test collateral relationships in large families

#### 3.3 Notes and Sources
- Test GetAllNotes() with files containing NOTE records
- Test source and repository queries
- Test multimedia queries

## Specific Test Files to Create

### 1. `integration_testdata_test.go`
- Tests using all testdata files
- Basic query operations on real data
- Verify correctness against known data

### 2. `relationship_real_data_test.go`
- Relationship queries using real GEDCOM files
- Test all relationship types (direct, ancestral, collateral)
- Test edge cases (unrelated individuals, etc.)

### 3. `path_finding_real_data_test.go`
- Path finding using real datasets
- Test shortest paths, all paths
- Test bidirectional path reconstruction

### 4. `collection_real_data_test.go`
- Collection queries on real data
- Test Names, Places, Events, Families collections
- Test uniqueness and filtering

### 5. `filter_real_data_test.go`
- Filter queries on real data
- Test all filter types (name, date, place, sex, etc.)
- Test filterByBool with real individuals

## Coverage Goals by Category

| Category | Current | Target | Priority |
|----------|---------|--------|----------|
| Filter Execution | 58-100% | 85%+ | HIGH |
| Relationship Calculations | 61-100% | 85%+ | HIGH |
| Path Finding | 80-97% | 90%+ | MEDIUM |
| Ancestor/Descendant Queries | 72-94% | 85%+ | MEDIUM |
| Graph Relationships | 75-100% | 85%+ | MEDIUM |
| Collection Queries | 90-100% | 95%+ | LOW |
| Notes Query | 0% | 80%+ | LOW |
| Graph Metrics | 0% | 80%+ | LOW |

## Implementation Recommendations

### 1. Create Test Helper
```go
// query/testdata_helper.go
func findTestDataFile(filename string) string {
    // Similar to parser/testdata_helper.go
    // Check multiple paths including /apps/gedcom-go/testdata
}
```

### 2. Test Structure
- Use table-driven tests where possible
- Create reusable test fixtures for each testdata file
- Cache parsed trees/graphs for multiple tests

### 3. Known Data Extraction
- Extract known XREFs from testdata files for testing
- Document expected counts (individuals, families, etc.)
- Create test fixtures with known relationships

### 4. Coverage Focus
- Prioritize functions with 0% coverage
- Focus on functions with <80% coverage
- Exclude hybrid storage operations (as requested)

## Expected Impact

By implementing tests using real testdata files:

1. **Coverage Increase**: Expected to reach 75-80%+ coverage
2. **Real-world Validation**: Tests will validate against actual GEDCOM data
3. **Edge Case Discovery**: Real data will reveal edge cases not found in synthetic data
4. **Performance Validation**: Large files (pres2020.ged) will validate performance
5. **Correctness Assurance**: Tests will catch regressions in real-world scenarios

## Next Steps

1. ✅ Create testdata helper (similar to parser package)
2. ✅ Create integration tests for basic operations
3. ✅ Test relationship queries with real data
4. ✅ Test path finding with real data
5. ✅ Test collection queries with real data
6. ✅ Test filterByBool with real individuals
7. ✅ Test notes query with files containing notes
8. ✅ Test graph metrics on real datasets

## Notes

- **Hybrid Storage**: Excluded from current focus (as requested)
- **Performance Tests**: Already exist but could be enhanced with real data
- **Synthetic Data**: Keep existing tests, add real data tests as complement
- **Test Data Size**: Use smaller files (xavier.ged, gracis.ged) for unit tests, larger files (royal92.ged, pres2020.ged) for integration tests

