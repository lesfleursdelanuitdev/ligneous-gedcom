# GEDCOM Diff Documentation

## Overview

The GEDCOM Diff system provides semantic comparison of GEDCOM files, identifying differences at a meaningful level rather than simple line-by-line text comparison. It understands GEDCOM structure and reports changes in terms of added/removed records, modified fields, relationship changes, and structural differences.

The system also tracks **change history**, recording who, when, and what changed for each modification.

---

## Features

### Core Capabilities

- **Record-Level Comparison**: Detect added, removed, and modified records
- **Field-Level Comparison**: Identify specific field changes (names, dates, places, etc.)
- **Semantic Equivalence**: Understand that "1800" ≈ "ABT 1800" (within tolerance)
- **Relationship Tracking**: Detect family structure changes
- **Change History**: Track when and what changed with timestamps
- **Multiple Matching Strategies**: XREF-based, content-based, or hybrid
- **Multiple Output Formats**: Text reports (JSON/HTML coming soon)

---

## Installation

The diff package is part of the main gedcom-go module:

```go
import "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/diff"
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/diff"
    "github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
)

func main() {
    // Parse two GEDCOM files
    p := parser.NewHierarchicalParser()
    tree1, _ := p.Parse("file1.ged")
    tree2, _ := p.Parse("file2.ged")

    // Create differ with default configuration
    differ := diff.NewGedcomDiffer(diff.DefaultConfig())

    // Compare the files
    result, err := differ.Compare(tree1, tree2)
    if err != nil {
        panic(err)
    }

    // Generate text report
    report, err := differ.GenerateReport(result)
    if err != nil {
        panic(err)
    }

    fmt.Println(report)
}
```

### Example Output

```
GEDCOM Diff Report
==================================================

Summary:
  File 1: 100 individuals, 50 families, 20 notes, 15 sources
  File 2: 105 individuals, 52 families, 22 notes, 16 sources

  Changes:
    Added:     5 individuals, 2 families
    Removed:   2 individuals, 0 families
    Modified:  3 individuals, 0 families
    Unchanged: 95 individuals, 50 families

Added Records:
--------------------------------------------------
  @I105@: INDI
    Name: John /Smith/
    Birth: 1850, New York

Modified Records:
--------------------------------------------------
  @I1@: INDI
    NAME: John /Doe/ → John /Doe Jr/
    BIRT.DATE: 1800 → ABT 1800 (semantically equivalent)
    DEAT.DATE: Added (1870)
```

---

## Configuration

### Default Configuration

```go
config := diff.DefaultConfig()
// MatchingStrategy: "xref"
// SimilarityThreshold: 0.85
// DateTolerance: 2
// IncludeUnchanged: false
// DetailLevel: "field"
// OutputFormat: "text"
// TrackHistory: true
```

### Custom Configuration

```go
config := &diff.DiffConfig{
    // Matching strategy
    MatchingStrategy: "hybrid", // "xref", "content", or "hybrid"
    
    // Thresholds
    SimilarityThreshold: 0.90,  // Higher threshold for stricter matching
    DateTolerance:       1,     // Stricter date tolerance
    
    // Options
    IncludeUnchanged: false,    // Don't include unchanged records
    DetailLevel:      "full",   // "summary", "field", or "full"
    OutputFormat:     "text",   // "text", "json", "html", "unified"
    
    // Change history
    TrackHistory: true,         // Enable change history tracking
}

differ := diff.NewGedcomDiffer(config)
```

---

## Matching Strategies

### 1. XREF-Based Matching (Default)

**Strategy:** Match records by their XREF IDs (e.g., `@I1@`).

**Use Case:** Comparing two versions of the same file.

**Pros:**
- Very fast (O(n) complexity)
- Accurate for same-file versions
- Preserves record identity

**Cons:**
- Fails when XREFs differ between files
- Doesn't work for cross-file comparison

**Example:**
```go
config := &diff.DiffConfig{
    MatchingStrategy: "xref",
}
```

### 2. Content-Based Matching

**Strategy:** Use duplicate detection to match records by content similarity.

**Use Case:** Comparing different files with potentially different XREFs.

**Pros:**
- Works across different files
- Handles XREF mismatches
- Finds semantic matches

**Cons:**
- Slower (requires similarity calculations)
- May have false positives
- Requires similarity threshold configuration

**Example:**
```go
config := &diff.DiffConfig{
    MatchingStrategy:   "content",
    SimilarityThreshold: 0.85,
}
```

### 3. Hybrid Matching (Recommended)

**Strategy:** Try XREF first, fallback to content matching for unmatched records.

**Use Case:** General-purpose comparison.

**Pros:**
- Fast for same-file, accurate for cross-file
- Best of both worlds
- Configurable behavior

**Cons:**
- More complex implementation

**Example:**
```go
config := &diff.DiffConfig{
    MatchingStrategy: "hybrid",
}
```

---

## Change Types

### 1. Added Records

Records that exist in file 2 but not in file 1.

```go
for _, added := range result.Changes.Added {
    fmt.Printf("Added: %s (%s)\n", added.Xref, added.Type)
    fmt.Printf("  Record: %v\n", added.Record)
}
```

### 2. Removed Records

Records that exist in file 1 but not in file 2.

```go
for _, removed := range result.Changes.Removed {
    fmt.Printf("Removed: %s (%s)\n", removed.Xref, removed.Type)
}
```

### 3. Modified Records

Records that exist in both files but have different content.

```go
for _, modified := range result.Changes.Modified {
    fmt.Printf("Modified: %s (%s)\n", modified.Xref, modified.Type)
    for _, change := range modified.Changes {
        fmt.Printf("  %s: %v → %v (%s)\n",
            change.Path,
            change.OldValue,
            change.NewValue,
            change.Type)
    }
}
```

### 4. Field Changes

Individual field changes within modified records.

**Change Types:**
- `modified`: Field value changed
- `added`: Field added (didn't exist in file 1)
- `removed`: Field removed (existed in file 1, not in file 2)
- `semantically_equivalent`: Different values but same meaning

**Example:**
```go
for _, change := range modified.Changes {
    switch change.Type {
    case diff.ChangeTypeModified:
        fmt.Printf("Modified: %s\n", change.Path)
    case diff.ChangeTypeAdded:
        fmt.Printf("Added: %s\n", change.Path)
    case diff.ChangeTypeRemoved:
        fmt.Printf("Removed: %s\n", change.Path)
    case diff.ChangeTypeSemanticallyEquivalent:
        fmt.Printf("Equivalent: %s (no real change)\n", change.Path)
    }
}
```

---

## Semantic Equivalence

The diff system understands that different values can have the same meaning.

### Date Equivalence

Dates within tolerance are considered equivalent:

- `"1800"` ≈ `"ABT 1800"` (within 2 years tolerance)
- `"1800"` ≈ `"1801"` (within 1 year)
- `"BEF 1850"` vs `"AFT 1840"` (overlapping ranges)

**Configuration:**
```go
config := &diff.DiffConfig{
    DateTolerance: 2, // Years tolerance
}
```

### Place Equivalence

Places with same components are considered equivalent:

- `"New York"` ≈ `"New York, NY"` (hierarchy match)
- `"New York"` ≈ `"new york"` (case-insensitive)
- `"NY"` ≈ `"New York"` (abbreviation match)

---

## Change History

When `TrackHistory` is enabled, the system records detailed change history.

### History Structure

```go
type ChangeHistory struct {
    Timestamp  time.Time
    Author     string // Optional: who made the change
    Reason     string // Optional: why the change was made
    ChangeType ChangeType
    Field      string // Field that changed
    OldValue   string
    NewValue   string
}
```

### Accessing History

**Global History:**
```go
for _, entry := range result.History {
    fmt.Printf("[%s] %s: %s\n",
        entry.Timestamp.Format(time.RFC3339),
        entry.ChangeType,
        entry.Field)
}
```

**Record-Level History:**
```go
for _, modified := range result.Changes.Modified {
    for _, entry := range modified.History {
        fmt.Printf("  [%s] %s\n", entry.Timestamp, entry.Field)
    }
}
```

**Field-Level History:**
```go
for _, change := range modified.Changes {
    for _, entry := range change.History {
        fmt.Printf("    [%s] %s: %s → %s\n",
            entry.Timestamp,
            entry.Field,
            entry.OldValue,
            entry.NewValue)
    }
}
```

### Example History Output

```
Change History:
--------------------------------------------------
  [2025-01-27T10:30:00Z] modified: NAME
    John /Doe/ → John /Doe Jr/
    Author: admin
    Reason: Corrected name

  [2025-01-27T10:31:00Z] semantically_equivalent: BIRT.DATE
    1800 → ABT 1800
```

---

## API Reference

### Types

#### GedcomDiffer

Main differ struct for comparing GEDCOM files.

```go
type GedcomDiffer struct {
    config *DiffConfig
}
```

#### DiffConfig

Configuration for the differ.

```go
type DiffConfig struct {
    MatchingStrategy   string
    SimilarityThreshold float64
    DateTolerance      int
    IncludeUnchanged   bool
    DetailLevel        string
    OutputFormat       string
    TrackHistory       bool
}
```

#### DiffResult

Result of comparison operation.

```go
type DiffResult struct {
    Summary    DiffSummary
    Changes    DiffChanges
    Statistics DiffStatistics
    History    []ChangeHistory
}
```

#### ChangeHistory

Individual change history entry.

```go
type ChangeHistory struct {
    Timestamp  time.Time
    Author     string
    Reason     string
    ChangeType ChangeType
    Field      string
    OldValue   string
    NewValue   string
}
```

### Functions

#### NewGedcomDiffer

Creates a new GEDCOM differ.

```go
func NewGedcomDiffer(config *DiffConfig) *GedcomDiffer
```

#### DefaultConfig

Returns default configuration.

```go
func DefaultConfig() *DiffConfig
```

### Methods

#### Compare

Compares two GEDCOM trees.

```go
func (gd *GedcomDiffer) Compare(tree1, tree2 *gedcom.GedcomTree) (*DiffResult, error)
```

**Example:**
```go
result, err := differ.Compare(tree1, tree2)
```

#### CompareFiles

Compares two GEDCOM files from disk (coming soon).

```go
func (gd *GedcomDiffer) CompareFiles(file1, file2 string) (*DiffResult, error)
```

#### GenerateReport

Generates a text report from diff results.

```go
func (gd *GedcomDiffer) GenerateReport(result *DiffResult) (string, error)
```

**Example:**
```go
report, err := differ.GenerateReport(result)
fmt.Println(report)
```

---

## Examples

### Example 1: Version Comparison

Compare two versions of the same file:

```go
// Parse both versions
p := parser.NewHierarchicalParser()
v1, _ := p.Parse("family_v1.ged")
v2, _ := p.Parse("family_v2.ged")

// Compare with XREF matching (fast for same file)
config := diff.DefaultConfig()
config.MatchingStrategy = "xref"
differ := diff.NewGedcomDiffer(config)

result, _ := differ.Compare(v1, v2)

// Print summary
fmt.Printf("Added: %d, Removed: %d, Modified: %d\n",
    len(result.Changes.Added),
    len(result.Changes.Removed),
    len(result.Changes.Modified))
```

### Example 2: Cross-File Comparison

Compare two different files:

```go
// Parse both files
p := parser.NewHierarchicalParser()
file1, _ := p.Parse("smith_family.ged")
file2, _ := p.Parse("jones_family.ged")

// Compare with content matching
config := &diff.DiffConfig{
    MatchingStrategy:   "content",
    SimilarityThreshold: 0.85,
    DateTolerance:      2,
    TrackHistory:       true,
}
differ := diff.NewGedcomDiffer(config)

result, _ := differ.Compare(file1, file2)

// Generate detailed report
report, _ := differ.GenerateReport(result)
fmt.Println(report)
```

### Example 3: Track Specific Changes

Find all name changes:

```go
result, _ := differ.Compare(tree1, tree2)

for _, modified := range result.Changes.Modified {
    for _, change := range modified.Changes {
        if change.Field == "NAME" {
            fmt.Printf("Name changed in %s: %v → %v\n",
                modified.Xref,
                change.OldValue,
                change.NewValue)
        }
    }
}
```

### Example 4: Change History Analysis

Analyze when changes were made:

```go
result, _ := differ.Compare(tree1, tree2)

// Group changes by timestamp
changesByDate := make(map[string][]ChangeHistory)
for _, entry := range result.History {
    date := entry.Timestamp.Format("2006-01-02")
    changesByDate[date] = append(changesByDate[date], entry)
}

// Print changes by date
for date, changes := range changesByDate {
    fmt.Printf("%s: %d changes\n", date, len(changes))
}
```

---

## Best Practices

### 1. Choose the Right Matching Strategy

- **Same file versions**: Use `"xref"` (fastest)
- **Different files**: Use `"content"` or `"hybrid"` (more accurate)
- **General use**: Use `"hybrid"` (balanced)

### 2. Configure Date Tolerance

Adjust based on your data quality:

- **High quality data**: `DateTolerance: 1` (stricter)
- **Normal data**: `DateTolerance: 2` (default)
- **Low quality data**: `DateTolerance: 5` (more lenient)

### 3. Enable Change History

Always enable history tracking for audit trails:

```go
config.TrackHistory = true
```

### 4. Use Appropriate Detail Level

- **Quick overview**: `DetailLevel: "summary"`
- **Standard review**: `DetailLevel: "field"` (default)
- **Deep analysis**: `DetailLevel: "full"`

### 5. Handle Large Files

For very large files (>10,000 records):

- Use XREF matching when possible (faster)
- Consider limiting comparison scope
- Process in batches if needed

---

## Performance

### Expected Performance

| File Size | XREF Strategy | Content Strategy | Hybrid Strategy |
|-----------|---------------|------------------|-----------------|
| 100 records | < 1 second | ~5 seconds | ~1 second |
| 1,000 records | ~1 second | ~1 minute | ~5 seconds |
| 10,000 records | ~5 seconds | ~10 minutes | ~30 seconds |

### Optimization Tips

1. **Use XREF when possible**: 10-100x faster than content matching
2. **Disable history for large files**: Reduces memory usage
3. **Use summary level**: Faster report generation
4. **Compare incrementally**: Only compare changed sections

---

## Troubleshooting

### Issue: Too Many "Modified" Records

**Problem:** Many records marked as modified when they should be unchanged.

**Solution:**
- Increase `DateTolerance` for date comparisons
- Check if semantic equivalence is working correctly
- Verify that field comparison logic is appropriate

### Issue: Missing Matches

**Problem:** Records that should match aren't being matched.

**Solution:**
- Use `"hybrid"` or `"content"` strategy instead of `"xref"`
- Lower `SimilarityThreshold` for content matching
- Check XREF consistency between files

### Issue: Slow Performance

**Problem:** Comparison takes too long.

**Solution:**
- Use `"xref"` strategy when possible
- Disable `TrackHistory` for large files
- Use `DetailLevel: "summary"` for faster processing

### Issue: Incorrect Semantic Equivalence

**Problem:** Dates/places marked as equivalent when they shouldn't be.

**Solution:**
- Adjust `DateTolerance` to be more strict
- Review place comparison logic
- Check date parsing accuracy

---

## Future Enhancements

### Planned Features

- **JSON Output**: Structured JSON format for programmatic processing
- **HTML Output**: Visual HTML diff with color coding
- **Unified Diff**: Git-style unified diff format
- **Three-Way Merge**: Compare base + two versions
- **Parallel Processing**: Faster comparison for large files
- **Incremental Diff**: Compare only changed sections
- **Author Tracking**: Track who made each change
- **Change Reasons**: Optional reason field for changes

---

## Related Documentation

- [Duplicate Detection Documentation](duplicate-detection.md) - Find potential duplicate individuals
- [Query API Documentation](query-api.md) - Graph-based query API
- [Types Documentation](types.md) - Core GEDCOM data types

---

## Examples Repository

For more examples, see the [examples directory](../examples/) in the repository.

---

**Last Updated:** 2025-01-27
