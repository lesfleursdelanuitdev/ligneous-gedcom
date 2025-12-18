# Duplicate Detection Documentation

## Overview

The Duplicate Detection system identifies potential duplicate individuals across GEDCOM files, providing similarity scores and configurable thresholds to identify records that might represent the same person. The system uses weighted similarity scoring based on names, dates, places, sex, and relationships.

---

## Features

### Core Capabilities

- **Single-File Detection**: Find duplicates within one GEDCOM file
- **Cross-File Detection**: Compare individuals across two different files
- **Single Individual Matching**: Find matches for a specific individual
- **Weighted Similarity Scoring**: Configurable weights for different metrics
- **Confidence Levels**: Categorize matches (exact, high, medium, low)
- **Phonetic Matching**: Soundex algorithm for name variations
- **Relationship Matching**: Use family relationships for better accuracy
- **Parallel Processing**: Fast comparison for large datasets
- **Performance Metrics**: Track processing time and throughput

---

## Installation

The duplicate detection package is part of the main gedcom-go module:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/duplicate"
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/duplicate"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
    // Parse GEDCOM file
    p := parser.NewHierarchicalParser()
    tree, _ := p.Parse("family.ged")

    // Create detector with default configuration
    detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())

    // Find duplicates
    result, err := detector.FindDuplicates(tree)
    if err != nil {
        panic(err)
    }

    // Process matches
    for _, match := range result.Matches {
        fmt.Printf("Potential duplicate: %s and %s\n",
            match.Individual1.XrefID(),
            match.Individual2.XrefID())
        fmt.Printf("  Similarity: %.2f%%\n", match.SimilarityScore*100)
        fmt.Printf("  Confidence: %s\n", match.Confidence)
    }
}
```

### Example Output

```
Potential duplicate: @I1@ and @I5@
  Similarity: 92.50%
  Confidence: high
  Matching fields: name, birth_date, birth_place, sex
  Differences: death_date
```

---

## Configuration

### Default Configuration

```go
config := duplicate.DefaultConfig()
// MinThreshold: 0.60
// HighConfidenceThreshold: 0.85
// ExactMatchThreshold: 0.95
// NameWeight: 0.40
// DateWeight: 0.30
// PlaceWeight: 0.15
// SexWeight: 0.05
// RelationshipWeight: 0.10
// UsePhoneticMatching: true
// UseRelationshipData: true
// UseParallelProcessing: true
// DateTolerance: 2
```

### Custom Configuration

```go
config := &duplicate.DuplicateConfig{
    // Thresholds
    MinThreshold:           0.70,  // Minimum similarity to report
    HighConfidenceThreshold: 0.90,  // High confidence threshold
    ExactMatchThreshold:    0.98,  // Exact match threshold
    
    // Weights
    NameWeight:        0.50,  // Higher weight on names
    DateWeight:        0.30,
    PlaceWeight:       0.15,
    SexWeight:         0.05,
    RelationshipWeight: 0.00,  // Disable relationship matching
    
    // Options
    UsePhoneticMatching:    true,  // Enable Soundex matching
    UseRelationshipData:    false, // Disable relationship matching
    UseParallelProcessing:  true,  // Enable parallel processing
    DateTolerance:         2,     // Years tolerance
    NumWorkers:            8,     // Manual worker count (0 = auto)
}

detector := duplicate.NewDuplicateDetector(config)
```

### Configuration Presets

#### Strict Configuration (Merging)

```go
config := &duplicate.DuplicateConfig{
    MinThreshold:          0.90,
    HighConfidenceThreshold: 0.95,
    ExactMatchThreshold:   0.98,
    NameWeight:            0.50,
    DateWeight:            0.30,
    UseRelationshipData:   true,
    DateTolerance:         1,
}
```

#### Loose Configuration (Research)

```go
config := &duplicate.DuplicateConfig{
    MinThreshold:          0.60,
    HighConfidenceThreshold: 0.75,
    ExactMatchThreshold:   0.90,
    NameWeight:            0.35,
    PlaceWeight:           0.20,
    UseRelationshipData:   false,
    DateTolerance:         5,
}
```

---

## Similarity Metrics

### 1. Name Similarity (40% weight)

**Algorithms:**
- Exact match
- Normalized match (removes slashes, case-insensitive)
- Component match (given name + surname separately)
- Phonetic match (Soundex algorithm)
- Fuzzy match (Levenshtein distance)

**Scoring:**
- Exact match: 1.0
- Phonetic match: 0.8-0.9
- Fuzzy match: 0.4-0.7
- Partial match: 0.6-0.8

**Example:**
```go
// "Smith" and "Smyth" will match phonetically
// "John" and "Jon" will match with fuzzy matching
```

### 2. Date Similarity (30% weight)

**Algorithms:**
- Exact year match
- Year difference calculation
- Date range overlap (for ABT, BEF, AFT dates)
- Tolerance-based matching

**Scoring:**
- Exact match: 1.0
- Same year: 0.9
- Within 1 year: 0.8
- Within 2 years: 0.7
- Within 5 years: 0.5
- Within 10 years: 0.3

**Example:**
```go
// "1800" and "ABT 1800" will match (within tolerance)
// "1800" and "1801" will have high similarity (0.9)
```

### 3. Place Similarity (15% weight)

**Algorithms:**
- Exact match
- Component matching (city, state, country)
- Hierarchy matching
- Abbreviation handling

**Scoring:**
- Exact match: 1.0
- Same city + state: 0.9
- Same city: 0.7
- Same state/country: 0.5

**Example:**
```go
// "New York" and "New York, NY" will match
// "NY" and "New York" will match (abbreviation)
```

### 4. Sex Match (5% weight)

**Scoring:**
- Match: 1.0
- Mismatch: 0.0 (strong negative indicator)
- Unknown (U): 0.5 (neutral)

### 5. Relationship Similarity (10% weight)

**Components:**
- Common parents (FAMC)
- Common spouses (FAMS)
- Common children

**Scoring:**
- Same parents: +0.2 bonus
- Same spouse: +0.2 bonus
- Common children: +0.1 per child (max +0.3)

**Note:** Requires graph structure (tree must be set).

---

## Similarity Score Calculation

### Weighted Sum Formula

```
Total Score = (Name × 0.40) + (Date × 0.30) + (Place × 0.15) + 
              (Sex × 0.05) + (Relationships × 0.10)
```

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

### Missing Data Handling

Missing fields don't reduce the score - they simply contribute 0.0 to that component:

```
Individual 1: John /Doe/, b. 1800, New York, M
Individual 2: John /Doe/, b. 1800, [no place], M

Name: 1.0 × 0.40 = 0.40
Date: 1.0 × 0.30 = 0.30
Place: 0.0 × 0.15 = 0.00 (missing)
Sex: 1.0 × 0.05 = 0.05
────────────────────────────
Total: 0.75 (75% similarity)
```

---

## Confidence Levels

### Threshold Levels

| Confidence | Score Range | Meaning | Action |
|------------|-------------|---------|--------|
| **Exact Match** | 0.95 - 1.0 | Almost certainly the same | Auto-merge candidate |
| **High Confidence** | 0.85 - 0.94 | Very likely the same | Manual review recommended |
| **Medium Confidence** | 0.70 - 0.84 | Possibly the same | Manual review required |
| **Low Confidence** | 0.60 - 0.69 | Unlikely but possible | Review if other indicators |
| **No Match** | 0.0 - 0.59 | Different people | Ignore |

### Configurable Thresholds

```go
config := &duplicate.DuplicateConfig{
    MinThreshold:          0.70,  // Minimum to report
    HighConfidenceThreshold: 0.85,  // High confidence
    ExactMatchThreshold:   0.95,  // Exact match
}
```

---

## API Reference

### Types

#### DuplicateDetector

Main detector for finding duplicate individuals.

```go
type DuplicateDetector struct {
    config *DuplicateConfig
    tree   *gedcom.GedcomTree
}
```

#### DuplicateConfig

Configuration for duplicate detection.

```go
type DuplicateConfig struct {
    MinThreshold            float64
    HighConfidenceThreshold float64
    ExactMatchThreshold     float64
    NameWeight              float64
    DateWeight              float64
    PlaceWeight             float64
    SexWeight               float64
    RelationshipWeight      float64
    UsePhoneticMatching     bool
    UseRelationshipData     bool
    UseParallelProcessing   bool
    DateTolerance           int
    NumWorkers              int
}
```

#### DuplicateMatch

A potential duplicate match between two individuals.

```go
type DuplicateMatch struct {
    Individual1       *gedcom.IndividualRecord
    Individual2      *gedcom.IndividualRecord
    SimilarityScore  float64
    Confidence       string
    MatchingFields   []string
    Differences      []string
    NameScore        float64
    DateScore        float64
    PlaceScore       float64
    SexScore         float64
    RelationshipScore float64
}
```

#### DuplicateResult

Result of duplicate detection operation.

```go
type DuplicateResult struct {
    Matches          []DuplicateMatch
    TotalComparisons int
    ProcessingTime   time.Duration
    Metrics          *PerformanceMetrics
}
```

### Functions

#### NewDuplicateDetector

Creates a new duplicate detector.

```go
func NewDuplicateDetector(config *DuplicateConfig) *DuplicateDetector
```

#### DefaultConfig

Returns default configuration.

```go
func DefaultConfig() *DuplicateConfig
```

### Methods

#### FindDuplicates

Finds duplicates within a single GEDCOM tree.

```go
func (dd *DuplicateDetector) FindDuplicates(tree *gedcom.GedcomTree) (*DuplicateResult, error)
```

**Example:**
```go
result, err := detector.FindDuplicates(tree)
```

#### FindDuplicatesBetween

Finds duplicates between two GEDCOM trees.

```go
func (dd *DuplicateDetector) FindDuplicatesBetween(tree1, tree2 *gedcom.GedcomTree) (*DuplicateResult, error)
```

**Example:**
```go
result, err := detector.FindDuplicatesBetween(tree1, tree2)
```

#### FindMatches

Finds matches for a specific individual.

```go
func (dd *DuplicateDetector) FindMatches(individual *gedcom.IndividualRecord, tree *gedcom.GedcomTree) ([]DuplicateMatch, error)
```

**Example:**
```go
matches, err := detector.FindMatches(individual, tree)
```

#### Compare

Compares two individuals directly.

```go
func (dd *DuplicateDetector) Compare(indi1, indi2 *gedcom.IndividualRecord) (float64, error)
```

**Example:**
```go
score, err := detector.Compare(indi1, indi2)
```

#### SetTree

Sets the GEDCOM tree for relationship matching.

```go
func (dd *DuplicateDetector) SetTree(tree *gedcom.GedcomTree)
```

**Note:** Required for relationship similarity calculations.

---

## Examples

### Example 1: Find Duplicates in Single File

```go
// Parse file
p := parser.NewHierarchicalParser()
tree, _ := p.Parse("family.ged")

// Create detector
detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())

// Find duplicates
result, _ := detector.FindDuplicates(tree)

// Process results
fmt.Printf("Found %d potential duplicates\n", len(result.Matches))
for _, match := range result.Matches {
    if match.Confidence == "high" || match.Confidence == "exact" {
        fmt.Printf("High confidence match: %s and %s (%.2f%%)\n",
            match.Individual1.XrefID(),
            match.Individual2.XrefID(),
            match.SimilarityScore*100)
    }
}
```

### Example 2: Cross-File Comparison

```go
// Parse both files
p := parser.NewHierarchicalParser()
tree1, _ := p.Parse("file1.ged")
tree2, _ := p.Parse("file2.ged")

// Create detector
detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())

// Find duplicates between files
result, _ := detector.FindDuplicatesBetween(tree1, tree2)

// Process matches
for _, match := range result.Matches {
    fmt.Printf("Match: %s (file1) ↔ %s (file2)\n",
        match.Individual1.XrefID(),
        match.Individual2.XrefID())
    fmt.Printf("  Similarity: %.2f%%, Confidence: %s\n",
        match.SimilarityScore*100,
        match.Confidence)
}
```

### Example 3: Find Matches for Specific Individual

```go
// Get individual
individual := tree.GetIndividual("@I1@")
indi, _ := individual.(*gedcom.IndividualRecord)

// Create detector
detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())
detector.SetTree(tree) // Required for relationship matching

// Find matches
matches, _ := detector.FindMatches(indi, tree)

// Process matches
for _, match := range matches {
    fmt.Printf("Potential match: %s (similarity: %.2f%%)\n",
        match.Individual2.XrefID(),
        match.SimilarityScore*100)
}
```

### Example 4: Compare Two Individuals

```go
indi1 := tree.GetIndividual("@I1@").(*gedcom.IndividualRecord)
indi2 := tree.GetIndividual("@I5@").(*gedcom.IndividualRecord)

detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())
score, _ := detector.Compare(indi1, indi2)

if score >= 0.85 {
    fmt.Printf("High similarity: %.2f%%\n", score*100)
}
```

### Example 5: Performance Metrics

```go
result, _ := detector.FindDuplicates(tree)

if result.Metrics != nil {
    fmt.Printf("Processing time: %v\n", result.Metrics.ProcessingTime)
    fmt.Printf("Total comparisons: %d\n", result.Metrics.TotalComparisons)
    fmt.Printf("Throughput: %.2f comparisons/sec\n", result.Metrics.Throughput)
    fmt.Printf("Parallel workers: %d\n", result.Metrics.ParallelWorkers)
}
```

### Example 6: Custom Configuration

```go
// Strict configuration for merging
config := &duplicate.DuplicateConfig{
    MinThreshold:          0.90,
    HighConfidenceThreshold: 0.95,
    NameWeight:            0.50,
    DateWeight:            0.30,
    UseRelationshipData:   true,
    DateTolerance:         1,
}

detector := duplicate.NewDuplicateDetector(config)
result, _ := detector.FindDuplicates(tree)

// Only high-confidence matches
for _, match := range result.Matches {
    if match.Confidence == "high" || match.Confidence == "exact" {
        // Process match
    }
}
```

---

## Phonetic Matching

### Soundex Algorithm

The system uses the Soundex algorithm for phonetic name matching.

**How it works:**
- Converts names to 4-character codes
- First letter + 3 digits based on consonants
- Similar-sounding names get the same code

**Examples:**
- "Smith" → S530
- "Smyth" → S530 (matches)
- "Smythe" → S530 (matches)

**Enable/Disable:**
```go
config := &duplicate.DuplicateConfig{
    UsePhoneticMatching: true, // Default: true
}
```

---

## Relationship Matching

### Using Family Relationships

Relationship matching uses family connections to improve accuracy:

- **Common Parents**: Same parents (FAMC) → strong indicator
- **Common Spouse**: Same spouse (FAMS) → strong indicator
- **Common Children**: Shared children → additional indicator

**Requirements:**
- `UseRelationshipData: true`
- Tree must be set via `SetTree()`

**Example:**
```go
detector := duplicate.NewDuplicateDetector(config)
detector.SetTree(tree) // Required for relationship matching

result, _ := detector.FindDuplicates(tree)

// Check relationship scores
for _, match := range result.Matches {
    if match.RelationshipScore > 0 {
        fmt.Printf("Relationship match: %.2f\n", match.RelationshipScore)
    }
}
```

---

## Performance

### Parallel Processing

The system automatically uses parallel processing for large datasets (>10 individuals).

**Configuration:**
```go
config := &duplicate.DuplicateConfig{
    UseParallelProcessing: true, // Default: true
    NumWorkers:            8,     // Manual count (0 = auto-detect)
}
```

**Performance:**
- Sequential: O(n²) comparisons
- Parallel: 4-8x faster on multi-core systems
- Auto-detects optimal worker count (1.5x CPU cores)

### Expected Performance

| File Size | Comparisons | Time (Sequential) | Time (Parallel) |
|-----------|-------------|-------------------|------------------|
| 100 individuals | ~5,000 | < 1 second | < 1 second |
| 1,000 individuals | ~500,000 | ~5 seconds | ~1 second |
| 10,000 individuals | ~5,000,000 | ~1 minute | ~10 seconds |
| 100,000 individuals | ~50,000,000 | ~10 minutes | ~2 minutes |

### Optimization Features

- **Pre-filtering**: Indexes by surname, birth year, place
- **Early Termination**: Skips low-probability matches
- **Memory Pooling**: Reduces allocations
- **Caching**: Caches parsed dates/places

---

## Best Practices

### 1. Choose Appropriate Thresholds

- **Merging**: Use high threshold (0.85-0.90)
- **Research**: Use lower threshold (0.60-0.70)
- **Quality Check**: Use medium threshold (0.70-0.80)

### 2. Enable Relationship Matching

Always enable when comparing within the same file:

```go
config.UseRelationshipData = true
detector.SetTree(tree)
```

### 3. Use Phonetic Matching

Enable for better name matching:

```go
config.UsePhoneticMatching = true
```

### 4. Configure Date Tolerance

Adjust based on data quality:

- High quality: `DateTolerance: 1`
- Normal: `DateTolerance: 2` (default)
- Low quality: `DateTolerance: 5`

### 5. Review High-Confidence Matches First

Process matches by confidence level:

```go
// Exact matches first
for _, match := range result.Matches {
    if match.Confidence == "exact" {
        // Auto-merge candidate
    }
}

// High confidence next
for _, match := range result.Matches {
    if match.Confidence == "high" {
        // Manual review
    }
}
```

---

## Troubleshooting

### Issue: Too Many False Positives

**Problem:** Many matches that aren't actually duplicates.

**Solution:**
- Increase `MinThreshold`
- Increase `HighConfidenceThreshold`
- Enable `UseRelationshipData` for better accuracy
- Review similarity scores and adjust weights

### Issue: Missing Obvious Duplicates

**Problem:** Duplicates not being detected.

**Solution:**
- Lower `MinThreshold`
- Enable `UsePhoneticMatching`
- Increase `DateTolerance`
- Check if data is missing (names, dates)

### Issue: Slow Performance

**Problem:** Detection takes too long.

**Solution:**
- Ensure `UseParallelProcessing: true`
- Use `MaxComparisons` to limit comparisons
- Check if pre-filtering is working
- Consider sampling for very large files

### Issue: Relationship Matching Not Working

**Problem:** Relationship scores are always 0.

**Solution:**
- Ensure `UseRelationshipData: true`
- Call `SetTree()` before comparison
- Verify family relationships exist in tree
- Check that individuals have FAMC/FAMS links

---

## Advanced Usage

### Custom Similarity Weights

Adjust weights based on your data characteristics:

```go
config := &duplicate.DuplicateConfig{
    // If names are very reliable
    NameWeight: 0.50,
    DateWeight: 0.30,
    PlaceWeight: 0.15,
    
    // If dates are more reliable
    // NameWeight: 0.30,
    // DateWeight: 0.50,
    // PlaceWeight: 0.15,
}
```

### Limiting Comparisons

For very large files, limit comparisons:

```go
config := &duplicate.DuplicateConfig{
    MaxComparisons: 100000, // Limit to 100k comparisons
}
```

### Processing Specific Subsets

Compare only specific individuals:

```go
// Get subset
individuals := []*gedcom.IndividualRecord{indi1, indi2, indi3}

// Compare manually
detector := duplicate.NewDuplicateDetector(config)
for i := 0; i < len(individuals); i++ {
    for j := i + 1; j < len(individuals); j++ {
        score, _ := detector.Compare(individuals[i], individuals[j])
        if score >= config.MinThreshold {
            // Process match
        }
    }
}
```

---

## Related Documentation

- [Diff Documentation](diff.md) - Semantic comparison of GEDCOM files
- [Query API Documentation](query-api.md) - Graph-based query API
- [Types Documentation](types.md) - Core GEDCOM data types

---

## Examples Repository

For more examples, see the [examples directory](../examples/) in the repository.

---

**Last Updated:** 2025-01-27
