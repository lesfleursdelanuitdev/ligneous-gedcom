# GEDCOM Query API Documentation

Complete reference guide for the graph-based Query API.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Query Builder](#query-builder)
- [Query Types](#query-types)
  - [IndividualQuery](#individualquery)
  - [AncestorQuery](#ancestorquery)
  - [DescendantQuery](#descendantquery)
  - [RelationshipQuery](#relationshipquery)
  - [PathQuery](#pathquery)
  - [FilterQuery](#filterquery)
  - [FamilyQuery](#familyquery)
  - [MultiIndividualQuery](#multiindividualquery)
  - [GraphMetricsQuery](#graphmetricsquery)
- [Graph Operations](#graph-operations)
- [Performance Optimizations](#performance-optimizations)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Query API provides a fluent, builder-style interface for querying GEDCOM genealogical data. It builds on top of a graph representation, enabling efficient traversal, path finding, relationship calculation, and complex filtering.

### Features

- **Fluent API**: Builder pattern for intuitive query construction
- **Graph-Based**: Efficient graph algorithms for relationship queries
- **Type-Safe**: Compile-time type checking
- **Performance Optimized**: Caching, indexing, memory pooling
- **Thread-Safe**: Concurrent access support
- **Comprehensive**: All relationship types and graph operations

---

## Installation

The query package is part of the GEDCOM Go library:

```go
import "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
```

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
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
}
```

---

## Query Builder

The `QueryBuilder` is the entry point for all queries.

### Creating a Query Builder

```go
// From a GEDCOM tree (builds graph automatically)
q, err := query.NewQuery(tree)

// From an existing graph
q := query.NewQueryFromGraph(graph)
```

### Query Builder Methods

```go
// Start query from individual
q.Individual("@I1@")

// Start query from multiple individuals
q.Individuals("@I1@", "@I2@", "@I3@")

// Start query from all individuals
q.AllIndividuals()

// Start filter query
q.Filter()

// Start query from family
q.Family("@F1@")

// Get all families in the tree
families, _ := q.AllFamilies()

// Get all events in the tree
events, _ := q.AllEvents()

// Get all unique places in the tree
places, _ := q.AllPlaces()

// Get all unique names in the tree
names, _ := q.UniqueNames()

// Access graph directly
graph := q.Graph()

// Access graph metrics
metrics := q.Metrics()
```

---

## Query Types

### IndividualQuery

Query operations starting from a specific individual.

#### Direct Relationships

```go
// Parents
parents, _ := q.Individual("@I1@").Parents()

// Children
children, _ := q.Individual("@I1@").Children()

// Siblings
siblings, _ := q.Individual("@I1@").Siblings()

// Spouses
spouses, _ := q.Individual("@I1@").Spouses()
```

#### Extended Relationships

```go
// Grandparents
grandparents, _ := q.Individual("@I1@").Grandparents()

// Grandchildren
grandchildren, _ := q.Individual("@I1@").Grandchildren()

// Uncles/Aunts
uncles, _ := q.Individual("@I1@").Uncles()

// Cousins (1st cousins)
cousins, _ := q.Individual("@I1@").Cousins(1)

// Nephews/Nieces
nephews, _ := q.Individual("@I1@").Nephews()
```

#### Complex Queries

```go
// Ancestors (returns AncestorQuery)
ancestors, _ := q.Individual("@I1@").Ancestors().MaxGenerations(5).Execute()

// Descendants (returns DescendantQuery)
descendants, _ := q.Individual("@I1@").Descendants().IncludeSelf().Execute()

// Relationship to another individual
result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()

// Path to another individual
path, _ := q.Individual("@I1@").PathTo("@I2@").Shortest()
```

---

### AncestorQuery

Configurable ancestor search with options.

#### Options

```go
ancestors, _ := q.Individual("@I1@").
    Ancestors().
    MaxGenerations(3).        // Limit to 3 generations
    IncludeSelf().            // Include starting individual
    Filter(func(indi *gedcom.IndividualRecord) bool {
        return indi.GetSex() == "M"  // Only males
    }).
    Execute()
```

#### Methods

- `MaxGenerations(n)`: Limit search depth
- `IncludeSelf()`: Include starting individual
- `Filter(fn)`: Apply custom filter function
- `Execute()`: Execute query and return results
- `Count()`: Return count only
- `Exists()`: Check if any ancestors exist

#### Example

```go
// Count ancestors
count, _ := q.Individual("@I1@").Ancestors().Count()

// Check if ancestors exist
exists, _ := q.Individual("@I1@").Ancestors().Exists()

// Filter ancestors
ancestors, _ := q.Individual("@I1@").
    Ancestors().
    MaxGenerations(5).
    Filter(func(indi *gedcom.IndividualRecord) bool {
        return indi.GetBirthDate() != ""
    }).
    Execute()
```

---

### DescendantQuery

Configurable descendant search (same API as AncestorQuery).

#### Example

```go
descendants, _ := q.Individual("@I1@").
    Descendants().
    MaxGenerations(2).
    IncludeSelf().
    Execute()
```

---

### SubtreeQuery

Extract a family subtree with ancestors, descendants, siblings, and spouses.

#### Options

```go
subtree, _ := q.Individual("@I1@").
    GetSubtree().
    AncestorGenerations(2).      // Limit ancestor generations
    DescendantGenerations(3).     // Limit descendant generations
    IncludeSelf().                // Include starting individual
    IncludeSiblings().            // Include siblings
    IncludeSpouses().            // Include spouses
    Filter(func(indi *gedcom.IndividualRecord) bool {
        return indi.GetSex() == "M"  // Only males
    }).
    Execute()
```

#### Methods

- `AncestorGenerations(n)`: Limit ancestor search depth
- `DescendantGenerations(n)`: Limit descendant search depth
- `IncludeSelf()`: Include starting individual
- `ExcludeSelf()`: Exclude starting individual
- `IncludeSiblings()`: Include siblings
- `IncludeSpouses()`: Include spouses
- `Filter(fn)`: Apply custom filter function
- `Execute()`: Execute query and return SubtreeResult
- `Count()`: Return count only
- `ExecuteRecords()`: Return records directly

#### SubtreeResult

```go
type SubtreeResult struct {
    Root        *gedcom.IndividualRecord
    Ancestors   []*gedcom.IndividualRecord
    Descendants []*gedcom.IndividualRecord
    Siblings    []*gedcom.IndividualRecord
    Spouses     []*gedcom.IndividualRecord
    All         []*gedcom.IndividualRecord  // All individuals (deduplicated)
}
```

#### Example

```go
subtree, _ := q.Individual("@I3@").GetSubtree().
    AncestorGenerations(1).
    DescendantGenerations(1).
    IncludeSelf().
    IncludeSiblings().
    Execute()

fmt.Printf("Root: %s\n", subtree.Root.GetName())
fmt.Printf("Ancestors: %d\n", len(subtree.Ancestors))
fmt.Printf("Descendants: %d\n", len(subtree.Descendants))
fmt.Printf("Siblings: %d\n", len(subtree.Siblings))
fmt.Printf("Total: %d\n", len(subtree.All))
```

---

### RelationshipQuery

Calculate relationship between two individuals.

#### Execute

```go
result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
```

#### RelationshipResult

```go
type RelationshipResult struct {
    RelationshipType string  // "Parent", "Child", "Sibling", "1st Cousin", etc.
    Degree           int     // For cousins: 1st, 2nd, etc.
    Removal          int     // For removed cousins
    IsDirect         bool    // Direct relationship (parent, child, sibling)
    IsCollateral     bool    // Collateral relationship (cousin, uncle, etc.)
    Path             []string // Path between individuals
}
```

#### Example

```go
result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()

fmt.Printf("Relationship: %s\n", result.RelationshipType)
fmt.Printf("Degree: %d\n", result.Degree)
fmt.Printf("Removal: %d\n", result.Removal)
fmt.Printf("Is Direct: %v\n", result.IsDirect)
fmt.Printf("Is Collateral: %v\n", result.IsCollateral)
```

---

### PathQuery

Find paths between two individuals.

#### Shortest Path

```go
path, _ := q.Individual("@I1@").PathTo("@I2@").Shortest()
fmt.Printf("Path length: %d\n", path.Length)
```

#### All Paths

```go
paths, _ := q.Individual("@I1@").
    PathTo("@I2@").
    MaxLength(10).
    IncludeBlood(true).
    IncludeMarital(false).
    All()
```

#### Path Options

- `MaxLength(n)`: Maximum path length
- `IncludeBlood(true)`: Include blood relationships
- `IncludeMarital(false)`: Exclude marital relationships
- `Shortest()`: Return shortest path only
- `All()`: Return all paths

#### Path Result

```go
type Path struct {
    Nodes  []GraphNode
    Length int
    Edges  []*Edge
}
```

---

### FilterQuery

Filter individuals by various criteria.

#### Basic Filters

```go
// Filter by name
results, _ := q.Filter().ByName("John").Execute()

// Filter by sex
results, _ := q.Filter().BySex("M").Execute()

// Filter by birth place
results, _ := q.Filter().ByBirthPlace("New York").Execute()
```

#### Date Filters

```go
start := time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC)

// Filter by birth date range
results, _ := q.Filter().ByBirthDate(start, end).Execute()
```

#### Boolean Filters

```go
// Living individuals
results, _ := q.Filter().Living().Execute()

// Deceased individuals
results, _ := q.Filter().Deceased().Execute()

// Has children
results, _ := q.Filter().HasChildren().Execute()

// Has spouse
results, _ := q.Filter().HasSpouse().Execute()
```

#### Combined Filters

```go
// Multiple filters (AND logic)
results, _ := q.Filter().
    ByName("John").
    BySex("M").
    ByBirthDate(start, end).
    HasChildren().
    Execute()
```

#### Custom Filters

```go
// Custom filter function
results, _ := q.Filter().
    Where(func(indi *gedcom.IndividualRecord) bool {
        return len(indi.GetNames()) > 1
    }).
    Execute()
```

#### Count and Exists

```go
// Count matching individuals
count, _ := q.Filter().Living().HasSpouse().Count()

// Check if any match
exists, _ := q.Filter().ByName("John").Exists()
```

---

### FamilyQuery

Query operations starting from a family.

#### Basic Queries

```go
// Get husband
husband, _ := q.Family("@F1@").Husband()

// Get wife
wife, _ := q.Family("@F1@").Wife()

// Get children
children, _ := q.Family("@F1@").Children()

// Get parents (husband and wife)
parents, _ := q.Family("@F1@").Parents()
```

#### Event Queries

```go
// Get marriage date
marriageDate, _ := q.Family("@F1@").MarriageDate()

// Get all events
events, _ := q.Family("@F1@").Events()
```

---

### MultiIndividualQuery

Query operations on multiple individuals.

#### Union Operations

```go
// Find ancestors of all individuals (union)
ancestors, _ := q.Individuals("@I1@", "@I2@", "@I3@").Ancestors()

// Union of different queries
results, _ := q.Individuals("@I1@", "@I2@").
    Union(
        func(iq *query.IndividualQuery) ([]*gedcom.IndividualRecord, error) {
            return iq.Parents()
        },
        func(iq *query.IndividualQuery) ([]*gedcom.IndividualRecord, error) {
            return iq.Siblings()
        },
    )
```

#### Common Ancestors

```go
// Find common ancestors
common, _ := q.Individuals("@I1@", "@I2@").CommonAncestors()

// Find lowest common ancestor (LCA)
lca, _ := q.Individuals("@I1@", "@I2@").LowestCommonAncestor()
```

---

### Simple Collection Queries

Simple queries to get all records of a specific type.

#### All Families

```go
// Get all families in the tree
families, _ := q.AllFamilies()
for _, family := range families {
    fmt.Printf("Family: %s\n", family.XrefID())
}
```

#### All Events

```go
// Get all events from all individuals and families
events, _ := q.AllEvents()
for _, event := range events {
    fmt.Printf("Event: %s on %s at %s\n", 
        event.EventType, event.Date, event.Place)
}
```

#### All Places

```go
// Get all unique places found in the tree
places, _ := q.AllPlaces()
for _, place := range places {
    fmt.Printf("Place: %s\n", place)
}
```

#### Unique Names

```go
// Get all unique names (given names and surnames)
names, _ := q.UniqueNames()
fmt.Printf("Given names: %v\n", names["given"])
fmt.Printf("Surnames: %v\n", names["surname"])
```

---

### GraphMetricsQuery

Graph analytics and metrics.

#### Node Metrics

```go
metrics := q.Metrics()

// Degree (total connections)
degree, _ := metrics.Degree("@I1@")

// In-degree (incoming connections)
inDegree, _ := metrics.InDegree("@I1@")

// Out-degree (outgoing connections)
outDegree, _ := metrics.OutDegree("@I1@")
```

#### Graph Metrics

```go
// Graph diameter
diameter, _ := metrics.Diameter()

// Average path length
avgPathLength, _ := metrics.AveragePathLength()

// Average degree
avgDegree, _ := metrics.AverageDegree()

// Graph density
density, _ := metrics.Density()
```

#### Centrality Measures

```go
// Degree centrality
centrality, _ := metrics.Centrality(query.CentralityDegree)

// Betweenness centrality
betweenness, _ := metrics.Centrality(query.CentralityBetweenness)

// Closeness centrality
closeness, _ := metrics.Centrality(query.CentralityCloseness)
```

#### Connectivity

```go
// Check if two nodes are connected
connected, _ := metrics.IsConnected("@I1@", "@I2@")

// Find connected components
components, _ := metrics.ConnectedComponents()

// Longest path in graph
longestPath, _ := metrics.LongestPath()
```

---

## Graph Operations

Direct access to graph algorithms.

### Graph Access

```go
graph := q.Graph()
```

### Traversal

```go
// Breadth-First Search
graph.BFS("@I1@", func(node GraphNode) bool {
    fmt.Printf("Visited: %s\n", node.ID())
    return true  // Continue
})

// Depth-First Search
graph.DFS("@I1@", func(node GraphNode) bool {
    fmt.Printf("Visited: %s\n", node.ID())
    return true  // Continue
})
```

### Path Finding

```go
// Shortest path
path, _ := graph.ShortestPath("@I1@", "@I2@")

// All paths
allPaths, _ := graph.AllPaths("@I1@", "@I2@", 10)
```

### Ancestors

```go
// Common ancestors
common, _ := graph.CommonAncestors("@I1@", "@I2@")

// Lowest common ancestor
lca, _ := graph.LowestCommonAncestor("@I1@", "@I2@")
```

### Relationships

```go
// Calculate relationship
result, _ := graph.CalculateRelationship("@I1@", "@I2@")
```

---

## Performance Optimizations

### Caching

Query results are cached for repeated queries:

```go
// First call - builds cache
parents1, _ := q.Individual("@I1@").Parents()

// Second call - uses cache (much faster)
parents2, _ := q.Individual("@I1@").Parents()
```

### Indexing

Filter queries use indexes for fast lookups:

```go
// Uses name index
results, _ := q.Filter().ByName("John").Execute()

// Uses date index
results, _ := q.Filter().ByBirthDate(start, end).Execute()

// Uses boolean indexes
results, _ := q.Filter().HasChildren().Execute()
```

### Memory Pooling

Temporary data structures are pooled to reduce allocations.

---

## API Reference

### QueryBuilder

```go
type QueryBuilder struct {
    graph *Graph
}

func NewQuery(tree *gedcom.GedcomTree) (*QueryBuilder, error)
func NewQueryFromGraph(graph *Graph) *QueryBuilder
func (qb *QueryBuilder) Individual(xrefID string) *IndividualQuery
func (qb *QueryBuilder) Individuals(xrefIDs ...string) *MultiIndividualQuery
func (qb *QueryBuilder) AllIndividuals() *MultiIndividualQuery
func (qb *QueryBuilder) Filter() *FilterQuery
func (qb *QueryBuilder) Family(xrefID string) *FamilyQuery
func (qb *QueryBuilder) AllFamilies() ([]*gedcom.FamilyRecord, error)
func (qb *QueryBuilder) AllEvents() ([]EventInfo, error)
func (qb *QueryBuilder) AllPlaces() ([]string, error)
func (qb *QueryBuilder) UniqueNames() (map[string][]string, error)
func (qb *QueryBuilder) Graph() *Graph
func (qb *QueryBuilder) Metrics() *GraphMetricsQuery
```

### IndividualQuery

```go
type IndividualQuery struct {
    xrefID string
    graph  *Graph
}

func (iq *IndividualQuery) Parents() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Children() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Siblings() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Spouses() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Grandparents() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Grandchildren() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Uncles() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Cousins(degree int) ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Nephews() ([]*gedcom.IndividualRecord, error)
func (iq *IndividualQuery) Ancestors() *AncestorQuery
func (iq *IndividualQuery) Descendants() *DescendantQuery
func (iq *IndividualQuery) GetSubtree() *SubtreeQuery
func (iq *IndividualQuery) RelationshipTo(xrefID string) *RelationshipQuery
func (iq *IndividualQuery) PathTo(xrefID string) *PathQuery
func (iq *IndividualQuery) GetEvents() ([]EventInfo, error)
```

### FilterQuery

```go
type FilterQuery struct {
    graph   *Graph
    filters []Filter
}

func (fq *FilterQuery) Where(filter Filter) *FilterQuery
func (fq *FilterQuery) ByName(pattern string) *FilterQuery
func (fq *FilterQuery) ByBirthDate(start, end time.Time) *FilterQuery
func (fq *FilterQuery) ByBirthPlace(place string) *FilterQuery
func (fq *FilterQuery) BySex(sex string) *FilterQuery
func (fq *FilterQuery) HasChildren() *FilterQuery
func (fq *FilterQuery) HasSpouse() *FilterQuery
func (fq *FilterQuery) Living() *FilterQuery
func (fq *FilterQuery) Deceased() *FilterQuery
func (fq *FilterQuery) Execute() ([]*gedcom.IndividualRecord, error)
func (fq *FilterQuery) Count() (int, error)
func (fq *FilterQuery) Exists() (bool, error)
```

---

## Examples

### Example 1: Find All Ancestors

```go
ancestors, _ := q.Individual("@I1@").
    Ancestors().
    MaxGenerations(5).
    Execute()

for _, ancestor := range ancestors {
    fmt.Printf("Ancestor: %s\n", ancestor.GetName())
}
```

### Example 2: Find Relationship

```go
result, _ := q.Individual("@I1@").RelationshipTo("@I2@").Execute()
fmt.Printf("Relationship: %s\n", result.RelationshipType)
fmt.Printf("Degree: %d\n", result.Degree)
```

### Example 3: Find All Cousins

```go
cousins, _ := q.Individual("@I1@").Cousins(1)  // 1st cousins
for _, cousin := range cousins {
    fmt.Printf("Cousin: %s\n", cousin.GetName())
}
```

### Example 4: Find Path Between Two Individuals

```go
paths, _ := q.Individual("@I1@").PathTo("@I2@").All()
for _, path := range paths {
    fmt.Printf("Path length: %d\n", path.Length)
    for _, node := range path.Nodes {
        if indi, ok := node.(*query.IndividualNode); ok {
            fmt.Printf("  -> %s\n", indi.Individual.GetName())
        }
    }
}
```

### Example 5: Complex Filtering

```go
results, _ := q.Filter().
    ByName("John").
    BySex("M").
    ByBirthDate(
        time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Date(1900, 12, 31, 23, 59, 59, 0, time.UTC),
    ).
    HasChildren().
    Execute()
```

### Example 6: Graph Metrics

```go
metrics := q.Metrics()

// Find most connected individual
centrality, _ := metrics.Centrality(query.CentralityDegree)
maxDegree := 0.0
mostConnected := ""
for id, degree := range centrality {
    if degree > maxDegree {
        maxDegree = degree
        mostConnected = id
    }
}
fmt.Printf("Most connected: %s (degree: %.0f)\n", mostConnected, maxDegree)

// Check graph connectivity
components, _ := metrics.ConnectedComponents()
fmt.Printf("Number of connected components: %d\n", len(components))
```

### Example 7: Get All Families

```go
families, _ := q.AllFamilies()
fmt.Printf("Total families: %d\n", len(families))
for _, family := range families {
    fmt.Printf("Family: %s\n", family.XrefID())
}
```

### Example 8: Get All Events

```go
events, _ := q.AllEvents()
fmt.Printf("Total events: %d\n", len(events))
for _, event := range events {
    fmt.Printf("%s: %s at %s on %s\n", 
        event.EventType, event.Description, event.Place, event.Date)
}
```

### Example 9: Get All Unique Places

```go
places, _ := q.AllPlaces()
fmt.Printf("Unique places: %d\n", len(places))
for _, place := range places {
    fmt.Printf("Place: %s\n", place)
}
```

### Example 10: Get Unique Names

```go
names, _ := q.UniqueNames()
fmt.Printf("Unique given names: %d\n", len(names["given"]))
fmt.Printf("Unique surnames: %d\n", len(names["surname"]))
for _, surname := range names["surname"] {
    fmt.Printf("Surname: %s\n", surname)
}
```

### Example 11: Subtree Extraction

```go
subtree, _ := q.Individual("@I3@").GetSubtree().
    AncestorGenerations(2).
    DescendantGenerations(2).
    IncludeSelf().
    IncludeSiblings().
    IncludeSpouses().
    Execute()

fmt.Printf("Subtree contains %d individuals\n", len(subtree.All))
fmt.Printf("  Root: %s\n", subtree.Root.GetName())
fmt.Printf("  Ancestors: %d\n", len(subtree.Ancestors))
fmt.Printf("  Descendants: %d\n", len(subtree.Descendants))
fmt.Printf("  Siblings: %d\n", len(subtree.Siblings))
fmt.Printf("  Spouses: %d\n", len(subtree.Spouses))
```

---

## Best Practices

### Reuse Query Builder

Build the graph once and reuse:

```go
q, _ := query.NewQuery(tree)

// Reuse for multiple queries
ancestors, _ := q.Individual("@I1@").Ancestors().Execute()
descendants, _ := q.Individual("@I1@").Descendants().Execute()
```

### Use Indexed Filters

Prefer indexed filters for better performance:

```go
// Good: Uses index
results, _ := q.Filter().ByName("John").Execute()

// Less efficient: Custom filter
results, _ := q.Filter().Where(func(indi *gedcom.IndividualRecord) bool {
    return strings.Contains(indi.GetName(), "John")
}).Execute()
```

### Limit Query Depth

Use `MaxGenerations` to limit search depth:

```go
// Good: Limited depth
ancestors, _ := q.Individual("@I1@").Ancestors().MaxGenerations(5).Execute()

// May be slow: Unlimited depth
ancestors, _ := q.Individual("@I1@").Ancestors().Execute()
```

---

## Troubleshooting

### Common Issues

#### 1. "Individual not found" Error

**Problem:** Xref ID doesn't exist in graph.

**Solutions:**
- Check xref ID format: `@I1@`
- Verify individual exists in GEDCOM file
- Check graph was built correctly

#### 2. Slow Queries

**Problem:** Queries are slow on large graphs.

**Solutions:**
- Use `MaxGenerations` to limit depth
- Use indexed filters when possible
- Reuse query builder (graph is cached)

#### 3. Memory Issues

**Problem:** Out of memory on large graphs.

**Solutions:**
- Limit query depth
- Use filters to reduce result set
- Process results incrementally

---

## See Also

- [Types Documentation](types.md) - Core GEDCOM types
- [CLI Documentation](cli.md) - Command-line interface
- [Parser Documentation](parser.md) - Parsing GEDCOM files

---

**Last Updated:** 2025-01-27
