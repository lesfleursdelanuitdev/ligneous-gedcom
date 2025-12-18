# Duplicate Detection Design

**Date:** 2025-01-27  
**Status:** 🎨 Design Phase

---

## Overview

Design a system to detect potential duplicate individuals across GEDCOM files, providing similarity scores and configurable thresholds to identify records that might represent the same person.

---

## Use Cases

### 1. Single File Detection (Intra-File)

**Scenario:** Detect duplicates within a single GEDCOM file.

**Use Cases:**
- Data cleanup: Find accidentally duplicated entries
- Merge detection: Identify records that should be merged
- Quality assurance: Flag potential data entry errors

**Example:**
```
File: family.ged
- @I1@: John /Doe/ (b. 1800, d. 1870)
- @I5@: John /Doe/ (b. ABT 1800, d. 1870)
→ Potential duplicate (same name, similar dates)
```

### 2. Cross-File Detection (Inter-File)

**Scenario:** Detect duplicates across two different GEDCOM files.

**Use Cases:**
- File merging: Before merging two family trees
- Data integration: Combining data from multiple sources
- Research: Finding common ancestors across different trees
- Validation: Checking if person already exists in another file

**Example:**
```
File 1: smith_family.ged
- @I10@: John /Smith/ (b. 1850, New York)

File 2: jones_family.ged
- @I25@: John /Smith/ (b. 1850, New York, NY)
→ Potential duplicate (same name, same date, same place)
```

### 3. Batch Detection

**Scenario:** Compare one individual against many (or all individuals in a file).

**Use Cases:**
- Import validation: Check if imported person already exists
- Bulk comparison: Find all potential duplicates for a person
- Database deduplication: Find all duplicates in a large dataset

---

## Design Decision: Single vs Cross-File

### Recommendation: **Support Both**

**Rationale:**
1. **Single File** is simpler and more common
2. **Cross-File** is essential for merging and integration
3. Same algorithm can work for both (just different data sources)
4. API can be unified with optional second file parameter

**Proposed API:**
```go
// Single file
duplicates := detector.FindDuplicates(tree)

// Cross-file
duplicates := detector.FindDuplicatesBetween(tree1, tree2)

// Single individual vs all
matches := detector.FindMatches(individual, tree, threshold)
```

---

## Similarity Metrics

### 1. Name Similarity

**Components:**
- Full name comparison
- Given name comparison
- Surname comparison
- Nickname/alternate name matching
- Phonetic matching (Soundex, Metaphone)
- Common name variations (John/Jon, Mary/Marie)

**Scoring:**
- Exact match: 1.0
- Phonetic match: 0.8-0.9
- Partial match (contains): 0.6-0.8
- Fuzzy match (Levenshtein distance): 0.4-0.7

**Weight:** 40% (highest - names are most distinctive)

### 2. Date Similarity

**Components:**
- Birth date comparison
- Death date comparison
- Date range overlap
- Tolerance for imprecise dates (ABT, BEF, AFT)

**Scoring:**
- Exact match: 1.0
- Same year: 0.9
- Within 1 year: 0.8
- Within 2 years: 0.7
- Within 5 years: 0.5
- Within 10 years: 0.3
- Date ranges overlap: 0.6-0.9 (depending on overlap)

**Weight:** 30% (very important - dates are strong indicators)

**Special Cases:**
- ABT dates: ±2 year tolerance
- BEF/AFT dates: Check if ranges overlap
- Missing dates: Don't penalize (0.0 contribution, but don't reduce score)

### 3. Place Similarity

**Components:**
- Birth place comparison
- Death place comparison
- Place hierarchy matching (city, state, country)
- Abbreviation handling (NY vs New York)
- Partial matches (same city, different state)

**Scoring:**
- Exact match: 1.0
- Same city + state: 0.9
- Same city: 0.7
- Same state/country: 0.5
- Partial match (contains): 0.4-0.6

**Weight:** 15% (moderate - places can change)

### 4. Sex Match

**Components:**
- Sex value comparison (M, F, U)

**Scoring:**
- Match: 1.0
- Mismatch: 0.0 (strong negative indicator)
- Unknown (U): 0.5 (neutral)

**Weight:** 5% (low - but mismatch is strong negative)

### 5. Relationship Similarity

**Components:**
- Common parents
- Common spouses
- Common children
- Family overlap

**Scoring:**
- Same parents: +0.2 bonus
- Same spouse: +0.2 bonus
- Common children: +0.1 per child (max +0.3)
- Family overlap: +0.1-0.2

**Weight:** 10% (bonus - strong indicator if present)

**Note:** This requires graph/tree structure, so may not be available for all comparisons.

### 6. Additional Attributes

**Components:**
- Occupation
- Residence
- Nationality
- Religion
- Notes (text similarity)

**Scoring:**
- Exact match: 0.3-0.5 per attribute
- Partial match: 0.1-0.3 per attribute

**Weight:** Optional (low priority, can be enabled)

---

## Similarity Score Calculation

### Weighted Sum Formula

```
Total Score = (Name × 0.40) + (Date × 0.30) + (Place × 0.15) + 
              (Sex × 0.05) + (Relationships × 0.10) + (Attributes × optional)
```

### Normalization

- Scores range from 0.0 to 1.0
- Missing data doesn't reduce score (contributes 0.0 to that component)
- Components are weighted, so missing data reduces total score proportionally

### Example Calculation

```
Individual 1: John /Doe/, b. 1800, New York, M
Individual 2: John /Doe/, b. 1800, New York, M

Name: 1.0 × 0.40 = 0.40
Date: 1.0 × 0.30 = 0.30
Place: 1.0 × 0.15 = 0.15
Sex: 1.0 × 0.05 = 0.05
Relationships: 0.0 × 0.10 = 0.00
────────────────────────────
Total: 0.90 (90% similarity)
```

### Example with Missing Data

```
Individual 1: John /Doe/, b. 1800, New York, M
Individual 2: John /Doe/, b. 1800, [no place], M

Name: 1.0 × 0.40 = 0.40
Date: 1.0 × 0.30 = 0.30
Place: 0.0 × 0.15 = 0.00 (missing)
Sex: 1.0 × 0.05 = 0.05
Relationships: 0.0 × 0.10 = 0.00
────────────────────────────
Total: 0.75 (75% similarity)
```

---

## Thresholds

### Default Thresholds

| Threshold Level | Score Range | Meaning | Action |
|----------------|-------------|---------|--------|
| **Exact Match** | 0.95 - 1.0 | Almost certainly the same | Auto-merge candidate |
| **High Confidence** | 0.85 - 0.94 | Very likely the same | Manual review recommended |
| **Medium Confidence** | 0.70 - 0.84 | Possibly the same | Manual review required |
| **Low Confidence** | 0.60 - 0.69 | Unlikely but possible | Review if other indicators |
| **No Match** | 0.0 - 0.59 | Different people | Ignore |

### Configurable Thresholds

Users should be able to configure:
- **Minimum threshold**: Below this, don't report (default: 0.60)
- **High confidence threshold**: Above this, strong candidate (default: 0.85)
- **Exact match threshold**: Above this, very likely duplicate (default: 0.95)

### Context-Dependent Thresholds

**Strict Mode (Merging):**
- Minimum: 0.85 (only high confidence)
- Use when: Merging files, want to avoid false positives

**Loose Mode (Research):**
- Minimum: 0.60 (include all possibilities)
- Use when: Research, finding potential connections

**Balanced Mode (Default):**
- Minimum: 0.70 (medium confidence)
- Use when: General duplicate detection

---

## Algorithm Design

### Phase 1: Pre-filtering (Performance Optimization)

**Goal:** Reduce comparison space before expensive similarity calculations.

**Strategies:**
1. **Name Index:** Group by surname (or first letter of surname)
2. **Date Range Index:** Group by birth year ranges (±10 years)
3. **Place Index:** Group by country/state
4. **Sex Filter:** Only compare same sex (unless one is Unknown)

**Result:** Compare only candidates that could potentially match.

### Phase 2: Similarity Calculation

**For each candidate pair:**
1. Calculate name similarity
2. Calculate date similarity
3. Calculate place similarity
4. Calculate sex match
5. Calculate relationship similarity (if graph available)
6. Calculate attribute similarity (optional)
7. Compute weighted sum
8. Apply thresholds

### Phase 3: Result Ranking

**Sort by:**
1. Similarity score (descending)
2. Number of matching fields (tiebreaker)
3. Data completeness (tiebreaker)

---

## Performance Considerations

### Complexity

**Naive Approach:** O(n²) comparisons
- For 10,000 individuals: 50,000,000 comparisons
- Too slow for large files

**Optimized Approach:** O(n × m) where m << n
- Pre-filtering reduces m significantly
- For 10,000 individuals with 10% match rate: ~1,000,000 comparisons
- Still expensive but manageable

### Optimization Strategies

1. **Indexing:**
   - Build indexes on surname, birth year, place
   - Only compare within same index buckets

2. **Early Termination:**
   - If name similarity < 0.3, skip other calculations
   - If date difference > 20 years, skip

3. **Caching:**
   - Cache parsed dates/places
   - Cache similarity scores for repeated comparisons

4. **Parallelization:**
   - Compare candidates in parallel
   - Use worker pools for large datasets

5. **Sampling:**
   - For very large files, sample comparisons
   - Focus on high-probability matches first

### Expected Performance

| File Size | Comparisons | Time (estimated) |
|-----------|------------|------------------|
| 100 individuals | ~5,000 | < 1 second |
| 1,000 individuals | ~500,000 | ~5 seconds |
| 10,000 individuals | ~5,000,000 | ~1 minute |
| 100,000 individuals | ~50,000,000 | ~10 minutes |

---

## API Design

### Core Interface

```go
type DuplicateDetector struct {
    config *DuplicateConfig
}

type DuplicateConfig struct {
    // Thresholds
    MinThreshold      float64  // Minimum similarity to report (default: 0.60)
    HighConfidenceThreshold float64  // High confidence threshold (default: 0.85)
    ExactMatchThreshold float64  // Exact match threshold (default: 0.95)
    
    // Weights
    NameWeight        float64  // Name similarity weight (default: 0.40)
    DateWeight        float64  // Date similarity weight (default: 0.30)
    PlaceWeight       float64  // Place similarity weight (default: 0.15)
    SexWeight         float64  // Sex match weight (default: 0.05)
    RelationshipWeight float64  // Relationship weight (default: 0.10)
    
    // Options
    UsePhoneticMatching bool    // Use Soundex/Metaphone (default: true)
    UseRelationshipData bool    // Use family relationships (default: true)
    DateTolerance      int      // Years tolerance for dates (default: 2)
    MaxComparisons     int      // Limit comparisons for performance (0 = unlimited)
}

type DuplicateMatch struct {
    Individual1    *IndividualRecord
    Individual2   *IndividualRecord
    SimilarityScore float64
    Confidence      string  // "exact", "high", "medium", "low"
    MatchingFields  []string  // Which fields matched
    Differences     []string  // Which fields differ
}

type DuplicateResult struct {
    Matches         []DuplicateMatch
    TotalComparisons int
    ProcessingTime  time.Duration
}
```

### Methods

```go
// Single file detection
func (dd *DuplicateDetector) FindDuplicates(tree *GedcomTree) (*DuplicateResult, error)

// Cross-file detection
func (dd *DuplicateDetector) FindDuplicatesBetween(tree1, tree2 *GedcomTree) (*DuplicateResult, error)

// Find matches for single individual
func (dd *DuplicateDetector) FindMatches(individual *IndividualRecord, tree *GedcomTree) ([]DuplicateMatch, error)

// Compare two individuals directly
func (dd *DuplicateDetector) Compare(indi1, indi2 *IndividualRecord) (float64, error)
```

### Query API Integration

```go
// Add to QueryBuilder
q.Duplicates().MinThreshold(0.85).Execute()

// Find duplicates for specific individual
q.Individual("@I1@").FindDuplicates().MinThreshold(0.70).Execute()
```

---

## Implementation Phases

### Phase 1: Basic Similarity (MVP)

**Scope:**
- Name similarity (exact, partial, fuzzy)
- Date similarity (year-based)
- Place similarity (exact match)
- Sex match
- Single file only
- Basic threshold

**Deliverables:**
- `DuplicateDetector` struct
- Basic similarity calculations
- Single file duplicate detection
- Simple threshold filtering

### Phase 2: Advanced Similarity

**Scope:**
- Phonetic matching (Soundex, Metaphone)
- Date range handling (ABT, BEF, AFT)
- Place hierarchy matching
- Relationship similarity
- Configurable weights

**Deliverables:**
- Enhanced similarity algorithms
- Relationship-based matching
- Configurable scoring

### Phase 3: Cross-File Detection

**Scope:**
- Cross-file comparison
- Batch comparison
- Performance optimizations
- Indexing

**Deliverables:**
- Cross-file API
- Performance optimizations
- Indexing system

### Phase 4: Advanced Features

**Scope:**
- Merge suggestions
- Conflict resolution
- Export/import duplicate reports
- CLI integration

**Deliverables:**
- Merge recommendations
- Report generation
- CLI commands

---

## Edge Cases and Challenges

### 1. Name Variations

**Challenge:** Same person with different name formats
- "John /Doe/" vs "John Doe" vs "J. /Doe/"
- "Mary /Smith/" vs "Mary (Smith) /Jones/" (married name)

**Solution:**
- Normalize names (remove slashes, handle parentheses)
- Compare given name and surname separately
- Handle nickname variations

### 2. Date Imprecision

**Challenge:** Imprecise dates (ABT, BEF, AFT)
- "ABT 1800" vs "1800" vs "BET 1798 AND 1802"

**Solution:**
- Use date ranges for comparison
- Calculate overlap percentage
- Apply tolerance based on date type

### 3. Place Variations

**Challenge:** Same place with different formats
- "New York" vs "New York, NY" vs "New York, New York, USA"

**Solution:**
- Parse places into components
- Compare at different hierarchy levels
- Handle abbreviations

### 4. Missing Data

**Challenge:** One record has data, other doesn't
- Person 1: Name + Date + Place
- Person 2: Name only

**Solution:**
- Don't penalize missing data
- Use available data only
- Adjust weights dynamically based on available fields

### 5. False Positives

**Challenge:** Different people with similar data
- Common names (John Smith)
- Same birth year and place (twins, different people)

**Solution:**
- Use relationship data to distinguish
- Require multiple matching fields
- Higher thresholds for common names

### 6. Performance on Large Files

**Challenge:** O(n²) complexity is too slow

**Solution:**
- Aggressive pre-filtering
- Indexing
- Parallel processing
- Sampling for very large files

---

## Similarity Algorithms

### Name Similarity

#### 1. Exact Match
```
"John /Doe/" == "John /Doe/" → 1.0
```

#### 2. Normalized Match
```
Normalize: Remove slashes, trim, lowercase
"John /Doe/" → "john doe"
"John Doe" → "john doe"
→ 1.0
```

#### 3. Component Match
```
Given name: "John" == "John" → 1.0
Surname: "Doe" == "Doe" → 1.0
Average: 1.0
```

#### 4. Phonetic Match (Soundex)
```
"Smith" → S530
"Smyth" → S530
→ 0.9 (phonetic match)
```

#### 5. Fuzzy Match (Levenshtein)
```
"John" vs "Jon" → Distance: 1, Similarity: 0.75
"Mary" vs "Marie" → Distance: 1, Similarity: 0.80
```

#### 6. Partial Match
```
"John /Doe/" contains "John" → 0.6
"John /Doe/" starts with "John" → 0.7
```

### Date Similarity

#### 1. Exact Year Match
```
1800 == 1800 → 1.0
```

#### 2. Year Difference
```
1800 vs 1801 → Difference: 1 year → 0.9
1800 vs 1805 → Difference: 5 years → 0.7
1800 vs 1810 → Difference: 10 years → 0.5
```

#### 3. Date Range Overlap
```
"ABT 1800" (range: 1798-1802) vs "1800" (range: 1800-1800)
Overlap: 1 year out of 5 year range → 0.6
```

#### 4. Imprecise Date Handling
```
"ABT 1800" vs "1800" → Use range, calculate overlap
"BEF 1850" vs "AFT 1840" → Check if ranges overlap
```

### Place Similarity

#### 1. Exact Match
```
"New York" == "New York" → 1.0
```

#### 2. Component Match
```
Place 1: City="New York", State="NY"
Place 2: City="New York", State="New York"
→ City match: 1.0, State match: 0.8 (abbreviation)
Average: 0.9
```

#### 3. Hierarchy Match
```
"New York, NY, USA" vs "New York, NY"
→ Both have "New York" and "NY" → 0.9
```

#### 4. Partial Match
```
"New York" contains "York" → 0.5
"New York, NY" contains "New York" → 0.7
```

---

## Configuration Examples

### Strict Configuration (Merging)

```go
config := &DuplicateConfig{
    MinThreshold: 0.90,
    HighConfidenceThreshold: 0.95,
    ExactMatchThreshold: 0.98,
    NameWeight: 0.50,        // Higher weight on names
    DateWeight: 0.30,
    PlaceWeight: 0.15,
    SexWeight: 0.05,
    UseRelationshipData: true,  // Require relationship match
    DateTolerance: 1,        // Stricter date tolerance
}
```

### Loose Configuration (Research)

```go
config := &DuplicateConfig{
    MinThreshold: 0.60,
    HighConfidenceThreshold: 0.75,
    ExactMatchThreshold: 0.90,
    NameWeight: 0.35,        // Lower weight on names
    DateWeight: 0.30,
    PlaceWeight: 0.20,       // Higher weight on places
    SexWeight: 0.05,
    UseRelationshipData: false,  // Don't require relationships
    DateTolerance: 5,        // More lenient date tolerance
}
```

### Balanced Configuration (Default)

```go
config := &DuplicateConfig{
    MinThreshold: 0.70,
    HighConfidenceThreshold: 0.85,
    ExactMatchThreshold: 0.95,
    NameWeight: 0.40,
    DateWeight: 0.30,
    PlaceWeight: 0.15,
    SexWeight: 0.05,
    RelationshipWeight: 0.10,
    UseRelationshipData: true,
    DateTolerance: 2,
}
```

---

## Output Format

### DuplicateMatch Structure

```go
type DuplicateMatch struct {
    // Individuals
    Individual1 *IndividualRecord
    Individual2 *IndividualRecord
    
    // Similarity
    SimilarityScore float64  // 0.0 - 1.0
    Confidence      string   // "exact", "high", "medium", "low"
    
    // Details
    MatchingFields  []string  // ["name", "birth_date", "birth_place"]
    Differences     []string  // ["death_date", "death_place"]
    
    // Breakdown
    NameScore       float64
    DateScore       float64
    PlaceScore      float64
    SexScore        float64
    RelationshipScore float64
    
    // Metadata
    ComparisonTime  time.Duration
}
```

### Example Output

```json
{
  "matches": [
    {
      "individual1": "@I1@",
      "individual2": "@I5@",
      "similarity_score": 0.92,
      "confidence": "high",
      "matching_fields": ["name", "birth_date", "birth_place", "sex"],
      "differences": ["death_date"],
      "breakdown": {
        "name": 1.0,
        "date": 0.9,
        "place": 1.0,
        "sex": 1.0,
        "relationships": 0.0
      }
    }
  ],
  "total_comparisons": 1250,
  "processing_time": "2.3s"
}
```

---

## Integration Points

### 1. Query API Integration

```go
// Add duplicate detection to query API
q.Duplicates().MinThreshold(0.85).Execute()

// Find duplicates for specific individual
q.Individual("@I1@").FindDuplicates().Execute()
```

### 2. CLI Integration

```bash
# Find duplicates in file
gedcom duplicates family.ged --threshold 0.85

# Compare two files
gedcom duplicates file1.ged file2.ged --threshold 0.80

# Find matches for specific individual
gedcom duplicates family.ged --individual @I1@ --threshold 0.70

# Export results
gedcom duplicates family.ged -o duplicates.json --format json
```

### 3. Validator Integration

```go
// Add duplicate detection as validation rule
validator.AddRule(NewDuplicateDetectionRule(config))
```

---

## Testing Strategy

### Unit Tests

1. **Name Similarity Tests:**
   - Exact matches
   - Phonetic matches
   - Fuzzy matches
   - Partial matches
   - Name variations

2. **Date Similarity Tests:**
   - Exact dates
   - Year differences
   - Date ranges
   - Imprecise dates (ABT, BEF, AFT)

3. **Place Similarity Tests:**
   - Exact places
   - Component matches
   - Abbreviation handling
   - Partial matches

4. **Score Calculation Tests:**
   - Weighted sum calculation
   - Missing data handling
   - Threshold filtering

### Integration Tests

1. **Single File Detection:**
   - Small file (10 individuals, 2 duplicates)
   - Medium file (100 individuals, 5 duplicates)
   - Large file (1000 individuals, 20 duplicates)

2. **Cross-File Detection:**
   - Two files with overlapping individuals
   - Two files with no overlap
   - Files with different data quality

3. **Performance Tests:**
   - Measure comparison time
   - Verify indexing effectiveness
   - Test parallel processing

---

## Future Enhancements

### Phase 5: Machine Learning (Optional)

- Train model on known duplicates
- Learn optimal weights from data
- Improve accuracy over time

### Phase 6: Merge Suggestions

- Automatically suggest merges for high-confidence matches
- Generate merge conflict reports
- Provide merge preview

### Phase 7: Family Group Detection

- Detect duplicate families
- Match families by members
- Merge family records

---

## Open Questions

1. **Should we support fuzzy matching for places?**
   - "New York" vs "New York City" - same or different?
   - Recommendation: Use component matching, not fuzzy

2. **How to handle nicknames?**
   - "John" vs "Johnny" - same person?
   - Recommendation: Use phonetic matching + nickname dictionary

3. **Should relationship matching be required?**
   - Two people with same name/date but different parents
   - Recommendation: Make it optional, but strong indicator

4. **How to handle very common names?**
   - "John Smith" appears 100 times in file
   - Recommendation: Require higher threshold or additional matching fields

5. **Should we cache similarity scores?**
   - Same comparisons repeated
   - Recommendation: Yes, cache for performance

---

## Recommendations

### 1. Start with Single File

**Rationale:**
- Simpler to implement
- More common use case
- Can validate algorithm before adding complexity

### 2. Use Configurable Thresholds

**Rationale:**
- Different use cases need different sensitivity
- User should control false positive rate

### 3. Focus on Name + Date First

**Rationale:**
- These are the strongest indicators
- Can get good results with just these two
- Add other metrics incrementally

### 4. Implement Indexing Early

**Rationale:**
- Performance is critical
- Better to design with performance in mind
- Indexing is essential for large files

### 5. Make Relationship Matching Optional

**Rationale:**
- Not always available (cross-file, incomplete data)
- Should work without graph structure
- Can be bonus when available

---

## Summary

**Recommended Approach:**
1. **Support both single-file and cross-file** detection
2. **Use weighted similarity scoring** with configurable weights
3. **Implement multiple similarity algorithms** (exact, phonetic, fuzzy)
4. **Use configurable thresholds** for different use cases
5. **Optimize with indexing** and pre-filtering
6. **Start with MVP** (name + date similarity) then expand

**Priority Metrics:**
1. Name similarity (40% weight)
2. Date similarity (30% weight)
3. Place similarity (15% weight)
4. Sex match (5% weight)
5. Relationship similarity (10% weight, bonus)

**Default Thresholds:**
- Minimum: 0.70 (medium confidence)
- High: 0.85 (high confidence)
- Exact: 0.95 (very likely duplicate)

---

**Next Steps:**
1. Review and refine design
2. Implement Phase 1 (Basic Similarity)
3. Test with real GEDCOM files
4. Iterate based on results
