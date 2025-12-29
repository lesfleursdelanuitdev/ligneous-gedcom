# ligneous-gedcom

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/lesfleursdelanuitdev/ligneous-gedcom)](https://goreportcard.com/report/github.com/lesfleursdelanuitdev/ligneous-gedcom)

**ligneous-gedcom** is a genealogy toolkit designed to facilitate easily discovering information from GEDCOM data. It provides a faster and more modern parser and validator implementation, helping you find relationships, detect duplicates, understand data quality, and export meaningful subsets.

Built with Go for performance and reliability, supporting the full GEDCOM 5.5.1 specification.

## Package Overview

ligneous-gedcom provides a comprehensive set of packages for working with GEDCOM data:

- **Parser & Validator**: Parses and validates GEDCOM files against the GEDCOM 5.5.1 specification, ensuring data integrity and compliance
- **Query Package**: Perform powerful queries on your dataset, including relationship queries (ancestors, descendants, siblings, spouses), path finding, and filtered searches
- **CLI Package**: Interactive command line interface that provides easy access to the query package and other functionality through an intuitive terminal-based interface
- **Duplicate Package**: Detect potential duplicate individuals using similarity scoring, phonetic matching, and relationship analysis
- **Diff Package**: Compare two GEDCOM files and identify semantic differences with change tracking
- **Exporter Package**: Export your GEDCOM data to multiple formats including JSON, XML, YAML, CSV, and GEDCOM for integration with other systems

## Quick Start

**New to ligneous-gedcom? Start here:**

```bash
# 1. Install the tool
# Option A: Prebuilt binaries (coming soon - check GitHub Releases)
# Option B: Using Go (if you have Go installed)
go install github.com/lesfleursdelanuitdev/ligneous-gedcom/cmd/gedcom@latest

# 2. Start interactive exploration (recommended for first-time users)
gedcom interactive family.ged

# 3. Try these commands in interactive mode:
# > search John Smith
# > individual @I123@
# > ancestors @I123@ 5
# > relationship @I123@ @I456@
```

**Common tasks:**

```bash
# Find potential duplicates (top 200 most likely matches)
gedcom duplicates family.ged --top 200

# Validate your data
gedcom validate advanced family.ged --severity warning

# Generate data quality report
gedcom quality family.ged --format json -o quality-report.json

# Compare two GEDCOM files
gedcom diff file1.ged file2.ged --strategy hybrid -o diff-report.txt

# Search with filters
gedcom search family.ged --name "John" --sex M --birth-year 1900-1950

# Export a family branch for sharing
gedcom export --descendants @I123@ --depth 8 -o branch.json
```

> **📖 Next Steps:** Read the [User Workflows](#user-workflows) section to see example workflows for typical family research scenarios.

## What This Tool Does

ligneous-gedcom helps you:

- **Find relationships** — Discover how people are connected, calculate degrees of relationship, trace family lines
- **Detect potential duplicates** — Identify records that might refer to the same person, with explanations and confidence scores
- **Validate data quality** — Understand what's missing, what's inconsistent, and what needs attention ✅
- **Explore interactively** — Navigate family trees naturally, ask questions, follow connections ✅
- **Export meaningful subsets** — Extract specific branches, regions, or time periods for sharing or analysis ⚠️ (Branch export: ✅, Region/Time filters: 🔄 In progress)

## What This Tool Does Not Do

To set clear expectations:

- **It does not automatically merge people** — Duplicate detection produces suggestions, not automatic merges
- **It does not silently discard records** — All warnings are surfaced; nothing is hidden
- **It does not claim certainty where data is ambiguous** — Results are ranked by confidence, with explanations
- **It does not require you to be a programmer** — The CLI and interactive mode are designed for genealogists. Most users start in `interactive` mode and never write code.

## Design Philosophy

Genealogical data is incomplete and often contradictory. This tool is built on principles that respect that reality:

- **Transparency over convenience** — Warnings are shown instead of silent failures
- **Scoped questions over global scans** — Large datasets require focused queries, not "scan everything"
- **Suggestions over assertions** — Duplicate detection produces ranked candidates, not definitive answers
- **Safety over speed** — Better to warn and skip than to produce misleading results

This means the tool will tell you when a search is too broad, when data quality is poor, and when results are uncertain — because that honesty is what serious research requires.

### 🛠️ Available Commands

- **`interactive`** - Explore your data interactively (start here!)
- **`duplicates`** - Find potential duplicates with explanations
- **`search`** - Search with multiple filters (name, date, place, etc.)
- **`validate`** - Check data quality and find inconsistencies
- **`export`** - Export to JSON, XML, YAML, CSV, or GEDCOM formats
- **`parse`** - Parse and validate GEDCOM files
- **`quality`** - Generate comprehensive data quality reports
- **`diff`** - Compare two GEDCOM files and show semantic differences

## Installation

### Prebuilt Binaries (Recommended)

**Prebuilt binaries are planned for GitHub Releases (macOS, Windows, Linux).**

For now, install via Go (see below). Once binaries are available, you'll be able to download and run without installing Go.

### Using Go Install

```bash
go install github.com/lesfleursdelanuitdev/ligneous-gedcom/cmd/gedcom@latest
```

**Note:** If you don't use Go, wait for prebuilt binaries or use the source installation method below.

### From Source

```bash
git clone https://github.com/lesfleursdelanuitdev/ligneous-gedcom.git
cd ligneous-gedcom
go build -o gedcom ./cmd/gedcom
```

## Example Usage

### Interactive Exploration

```bash
# Start interactive mode
gedcom interactive family.ged

# Example session:
gedcom> search John
Search results for 'John':
  @I1@: John /Doe/
  @I5@: John /Smith/

gedcom> individual @I1@
Individual: @I1@
  Name: John /Doe/
  Sex: M
  Birth: 1900
  Death: 1970

gedcom> parents @I1@
Parents of @I1@:
  @I10@: James /Doe/
  @I11@: Mary /Doe/

gedcom> ancestors @I1@ 3
Ancestors of @I1@ (max 3 generations):
  @I10@: James /Doe/
  @I11@: Mary /Doe/
  ...

gedcom> relationship @I1@ @I2@
Relationship from @I1@ to @I2@:
  Type: 1st Cousin
  Degree: 1
  Removal: 0

gedcom> exit
```

### Finding Duplicates

```bash
# Find top 200 potential duplicates with explanations
gedcom duplicates family.ged --top 200 --explain

# Find matches for a specific person (most common use case)
gedcom duplicates family.ged --individual @I123@ --top 20 --explain

# Find duplicates in a specific time period and place
gedcom duplicates family.ged --place "Guyana" --year 1850-1920 --top 200

# Scope by surname and time period for more focused results
gedcom duplicates family.ged --surname "Smith" --year 1880-1920 --top 200
```

### Advanced Search

```bash
# Search by name
gedcom search family.ged --name "John"

# Multiple filters
gedcom search family.ged \
  --name "John" \
  --sex M \
  --birth-year 1900-1950 \
  --birth-place "New York" \
  --has-children \
  --format json

# Count only
gedcom search family.ged --name "John" --count-only

# Sorted results
gedcom search family.ged --name "John" --sort name --limit 10
```

### Exporting Data

```bash
# Export to JSON
gedcom export json family.ged -o family.json --pretty

# Export to XML
gedcom export xml family.ged -o family.xml

# Export to YAML
gedcom export yaml family.ged -o family.yaml

# Export a family branch (descendants)
gedcom export --descendants @I123@ --depth 8 -o branch.json

# Export by surname and place
gedcom export --surname "Bisht" --place "Uttarakhand" --year 1750-1900 -o bisht_family.json

# Export a disconnected component (family cluster)
gedcom export --component 3 -o cluster3.json
```

### Validation

```bash
# Basic validation
gedcom validate basic family.ged

# Advanced validation with severity levels
gedcom validate advanced family.ged --severity warning

# Generate validation report
gedcom validate advanced family.ged --output report.html --format html
```

### Programmatic Usage (Go API)

> **Note:** These are illustrative code snippets. For complete examples, see the [documentation](docs/).

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/duplicate"
)

func main() {
    // Parse GEDCOM file
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse("family.ged")
    if err != nil {
        panic(err)
    }

    // Create query builder
    q, err := query.NewQuery(tree)
    if err != nil {
        panic(err)
    }

    // Find ancestors
    ancestors, _ := q.Individual("@I1@").Ancestors().MaxGenerations(5).Execute()
    for _, ancestor := range ancestors {
        fmt.Printf("Ancestor: %s\n", ancestor.GetName())
    }

    // Calculate relationship
    result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
    fmt.Printf("Relationship: %s (Degree: %d)\n", result.RelationshipType, result.Degree)

    // Find duplicate individuals
    detector := duplicate.NewDuplicateDetector(duplicate.DefaultConfig())
    duplicates, _ := detector.FindDuplicates(tree)
    for _, match := range duplicates.Matches {
        fmt.Printf("Potential duplicate: %s and %s (similarity: %.2f, confidence: %s)\n",
            match.Individual1.XrefID(),
            match.Individual2.XrefID(),
            match.SimilarityScore,
            match.Confidence)
    }
}
```

## CLI Examples

### Parse and Validate

```bash
# Basic parse
gedcom parse file family.ged

# Parse with validation
gedcom parse validate family.ged --strict

# Quick syntax check
gedcom parse check family.ged
```

### Advanced Validation

```bash
# Basic validation
gedcom validate basic family.ged

# Advanced validation with report
gedcom validate advanced family.ged \
  --severity warning \
  -o report.json \
  --format json
```

### Export Formats

```bash
# Export to JSON
gedcom export json family.ged -o family.json --pretty

# Export to XML
gedcom export xml family.ged -o family.xml

# Export to YAML
gedcom export yaml family.ged -o family.yaml
```

### Advanced Search

```bash
# Search by name
gedcom search family.ged --name "John"

# Multiple filters
gedcom search family.ged \
  --name "John" \
  --sex M \
  --birth-year 1900-1950 \
  --birth-place "New York" \
  --has-children \
  --format json

# Count only
gedcom search family.ged --name "John" --count-only

# Sorted results
gedcom search family.ged --name "John" --sort name --limit 10
```

### Interactive Mode

```bash
# Start interactive mode
gedcom interactive family.ged

# Example session:
gedcom> search John
Search results for 'John':
  @I1@: John /Doe/
  @I5@: John /Smith/

gedcom> individual @I1@
Individual: @I1@
  Name: John /Doe/
  Sex: M
  Birth: 1900
  Death: 1970

gedcom> parents @I1@
Parents of @I1@:
  @I10@: James /Doe/
  @I11@: Mary /Doe/

gedcom> ancestors @I1@ 3
Ancestors of @I1@ (max 3 generations):
  @I10@: James /Doe/
  @I11@: Mary /Doe/
  ...

gedcom> relationship @I1@ @I2@
Relationship from @I1@ to @I2@:
  Type: 1st Cousin
  Degree: 1
  Removal: 0

gedcom> exit
```

## User Workflows

> **📖 Important:** This section explains how to use ligneous-gedcom effectively for your family research goals.

### For Private Family Researchers (50 to at most 50K individuals)

**Your typical scenario:** Your own family tree or a few related families, with a mix of complete and incomplete records.

**What to expect:**
- Duplicate detection runs in **minutes, not hours**
- Results are ranked with clear explanations ("same parents + birth year close + place similar")
- Interactive exploration feels natural and fast
- Most operations complete in seconds

**Example workflows:**

```bash
# Find matches for a specific person (most common use case)
gedcom duplicates family.ged --individual @I123@ --top 20 --explain

# Find top 200 potential duplicates with explanations
gedcom duplicates family.ged --top 200 --explain

# Find duplicates in a specific time period and place
gedcom duplicates family.ged --place "Guyana" --year 1850-1920 --top 200

# Explore a specific family line interactively
gedcom interactive family.ged
> ancestors @I123@ 5
> descendants @I123@ 3
> relationship @I123@ @I456@

# Export a branch for sharing
gedcom export --descendants @I123@ --depth 8 -o branch.json
```


## Technical Details

### Architecture

**Package Structure:**

```
ligneous-gedcom/
├── types/               # Core GEDCOM types and data structures
│   ├── Tree, Record, Line, Error
│   ├── Individual, Family, Note, etc.
│   ├── Date, Name, Place, Event types
├── parser/              # Parsing logic
├── validator/           # Validation logic
├── exporter/            # Export functionality
├── query/               # Graph-based Query API
├── diff/                # GEDCOM diff system
├── duplicate/           # Duplicate detection system
└── cmd/gedcom/          # CLI application
```

**Design Principles:**

- **Type Safety**: Strong typing throughout
- **Thread Safety**: Mutex-protected shared state
- **Performance**: Optimized with caching, indexing, pooling
- **Extensibility**: Interface-based design
- **Error Handling**: Explicit error returns with severity levels

### Core Capabilities

**Full GEDCOM 5.5.1 Support:**
- Complete parser implementation
- Advanced validation with severity levels
- Multiple export formats (GEDCOM, JSON, XML, YAML)

**Graph-Based Query API:**
- Relationship queries (parents, children, siblings, spouses)
- Ancestor/descendant traversal with generation limits
- Path finding (shortest path, all paths between individuals)
- Relationship calculation (degree, type, removal)
- Common ancestors and LCA (Lowest Common Ancestor)
- Graph analytics (centrality, diameter, connected components)

**Advanced Features:**
- Incremental graph updates (50-200x faster than full rebuild)
- Query result caching (100x speedup for repeated queries)
- Indexed search with multiple criteria (20-200x faster filtering)
- Parallel duplicate detection (4-8x faster on multi-core systems)
- Blocking-based duplicate detection (O(n²) → O(n) complexity)

## Performance & Benchmarks

### What This Means in Practice

**For typical family research (hundreds to tens of thousands of individuals):**
- Searches and relationship queries are **instant** (< 1 second)
- Duplicate detection usually completes in **minutes**
- Interactive exploration feels natural and responsive
- Most operations complete in seconds

**Memory requirements:**
- Small family trees (hundreds to thousands): ~50-150 MB
- Medium family trees (10K-50K individuals): ~150 MB - 3 GB

### Reproducing the Benchmarks

**Run benchmarks:**

```bash
# All benchmarks
go test -bench=. -benchmem ./...

# Run tests with your own GEDCOM files
go test ./...
```

### Performance Characteristics

**Scaling Behavior:**
- **Parsing:** ~50,000-100,000 individuals/second
- **Graph Construction:** ~10,000-20,000 individuals/second
- **Query Performance:** Sub-millisecond for most queries (constant time with caching)
- **Memory:** ~14-15 MB per 1,000 individuals
- **Duplicate Detection:** Optimized with blocking strategy for efficient processing

### Optimizations

- **Query Result Caching**: 100x speedup for repeated queries
- **Indexing**: 20-200x faster filtering (O(1) or O(log n) instead of O(V))
- **Bidirectional BFS**: ~2x faster path finding (O(V/2 + E/2) average case)
- **Memory Pooling**: Reduced allocations and GC pressure
- **Incremental Updates**: 50-200x faster than full rebuild
- **Parallel Duplicate Detection**: 4-8x faster on multi-core systems
- **Blocking Strategy**: Reduces duplicate detection from O(n²) to O(n × avg_block_size)

### Detailed Benchmarks

**Small Scale (10K individuals):**
- Graph construction: ~100ms
- Cached queries: ~45ns (cache hit)
- Indexed filtering: O(1) or O(log n) instead of O(V)
- Shortest path: O(V/2 + E/2) average case
- All operations scale efficiently with dataset size

## Documentation

### Core Documentation
- **[CLI Documentation](docs/cli.md)** - Complete CLI reference guide
- **[Parser Documentation](docs/parser.md)** - Parse GEDCOM files with multiple parser types
- **[Validator Documentation](docs/validator.md)** - Validate GEDCOM files with comprehensive rules
- **[Exporter Documentation](docs/exporter.md)** - Export GEDCOM data to JSON, XML, YAML, CSV, and GEDCOM formats
- **[Query API Documentation](docs/query-api.md)** - Graph-based query API for relationship queries
- **[Types Documentation](docs/types.md)** - Core GEDCOM data types and structures
- **[Duplicate Detection Documentation](docs/duplicate-detection.md)** - Find potential duplicate individuals with similarity scoring
- **[Diff Documentation](docs/diff.md)** - Semantic comparison of GEDCOM files with change history tracking

### Architecture & Examples
- **[Architecture Documentation](docs/ARCHITECTURE.md)** - System architecture, design patterns, and scalability
- **[API Examples](docs/API_EXAMPLES.md)** - Comprehensive code examples for all major features
- **[Error Handling Guide](docs/ERROR_HANDLING.md)** - Error handling patterns, examples, and best practices

## Testing

The codebase has been thoroughly tested with multiple real GEDCOM files of varying sizes from the `testdata` folder:

**Real GEDCOM Test Files:**
- **xavier.ged** (smallest): 317 individuals, 107 families, 5,821 lines
- **gracis.ged**: 585 individuals, 180 families, 10,323 lines
- **tree1.ged**: 1,032 individuals, 310 families, 12,713 lines
- **royal92.ged**: 3,010 individuals, 1,422 families, 30,682 lines
- **pres2020.ged** (largest): 2,322 individuals, 1,115 families, 49,431 lines

**Stress Testing:**
- Comprehensive stress testing with synthetic test data to validate performance and correctness
- Memory and performance characteristics verified across typical dataset sizes

```bash
# Run all tests
go test ./... -timeout 10m

# Run tests with coverage
go test ./... -cover -timeout 10m

# Run benchmarks
go test ./... -bench=. -timeout 10m
```

**Test Coverage:**

The core codebase maintains **83.4% test coverage** across all essential packages (excluding CLI and scripts). Test coverage applies to all core functionality:

- ✅ **Parser** (`parser/`): Comprehensive coverage with 15+ test files covering all parser types, edge cases, and malformed input handling
- ✅ **Validator** (`validator/`): Comprehensive coverage with 10+ test files covering all validation rules and severity levels
- ✅ **Exporter** (`exporter/`): Comprehensive coverage with 8+ test files covering all export formats (JSON, XML, YAML, CSV, GEDCOM)
- ✅ **Query API** (`query/`): Comprehensive coverage with 15+ test files covering graph operations, relationship queries, filtering, and hybrid storage modes
- ✅ **Core Types** (`types/`): Comprehensive coverage with 10+ test files covering all GEDCOM data structures and type conversions
- ✅ **Duplicate Detection** (`duplicate/`): Comprehensive coverage including similarity scoring, phonetic matching, relationship analysis, blocking strategies, and parallel processing
- ✅ **GEDCOM Diff** (`diff/`): Comprehensive coverage including XREF matching, field comparison, content comparison, and change history tracking

All core packages (`query`, `parser`, `types`, `duplicate`, `diff`, `exporter`, `validator`) are thoroughly tested with unit tests, integration tests, and performance benchmarks.

## Requirements

- **Go 1.23+**
- **Dependencies**: See [go.mod](go.mod)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- GEDCOM 5.5.1 specification
- Go community for excellent tooling and libraries

## Status

✅ **Production Ready**

The codebase is mature, well-tested, and ready for real-world use. All core functionality is implemented and tested. The tool has been validated on datasets ranging from small family trees (hundreds of individuals) to extended family research (tens of thousands of individuals).

---

**Made with ❤️ for genealogical research**
