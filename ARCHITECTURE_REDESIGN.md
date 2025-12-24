# Architecture Redesign: Records as Data, Graph as Query Engine

## Your Vision

You want a clear separation of concerns:

1. **Parser → Records**: Simple in-memory structure for **validation only**
2. **Records → Graph Nodes**: Records become **parts of** graph nodes (not duplicated)
3. **Graph → Queries**: Graph handles **all relationship queries**

This is a cleaner architecture than the current state. Let me explain what you're saying and what needs to change.

---

## What You're Saying

### Current Architecture (Problematic)

```
GEDCOM File
    ↓
Parser → Records (with relationship methods)
    ↓
GedcomTree (stores records)
    ↓
BuildGraph() → Graph Nodes (also with relationship methods)
    ↓
Graph (stores nodes + edges)

Problem: TWO ways to query relationships:
- record.Spouses()  ← Tree-based traversal
- node.getSpousesFromEdges()  ← Edge-based traversal
```

### Your Desired Architecture

```
GEDCOM File
    ↓
Parser → Records (DATA ONLY - no relationship methods)
    ↓
GedcomTree (stores records for validation)
    ↓
✅ VALIDATE RECORDS (validator package)
    ↓
BuildGraph() → Graph Nodes (wraps records, adds edges)
    ↓
✅ VALIDATE GRAPH (graph integrity, edge consistency)
    ↓
Graph (ONLY way to query relationships)
```

**Key Principles:**
- **Records = Data Container**: Just store parsed GEDCOM data, no relationship logic
- **Validation After Parsing**: Validate records/tree structure before building graph
- **Graph = Query Engine**: All relationship queries go through graph
- **Validation After Graph**: Validate graph integrity and edge consistency
- **No Duplication**: Records are referenced by nodes, not copied

---

## What Needs to Change

### 1. Remove Relationship Methods from Records

**Current State:**
```go
// types/individual_record.go
type IndividualRecord struct {
    *BaseRecord
}

// ❌ REMOVE: These relationship methods
func (ir *IndividualRecord) Spouses() ([]*IndividualRecord, error)
func (ir *IndividualRecord) Children() ([]*IndividualRecord, error)
func (ir *IndividualRecord) Parents() ([]*IndividualRecord, error)
func (ir *IndividualRecord) Siblings() ([]*IndividualRecord, error)
func (ir *IndividualRecord) SpouseChildren() (map[*IndividualRecord][]*IndividualRecord, error)
```

**What Records Should Have:**
```go
// types/individual_record.go
type IndividualRecord struct {
    *BaseRecord
}

// ✅ KEEP: Data access methods
func (ir *IndividualRecord) GetName() string
func (ir *IndividualRecord) GetBirthDate() string
func (ir *IndividualRecord) GetFamiliesAsSpouse() []string  // Just returns XREFs
func (ir *IndividualRecord) GetFamiliesAsChild() []string    // Just returns XREFs

// ❌ REMOVE: All relationship traversal methods
// Relationships are ONLY available through graph
```

**Files to Modify:**
- `types/individual_record.go` - Remove relationship methods (lines ~430-850)
- `types/family_record.go` - Remove relationship methods if any
- `types/individual_record_relationship_test.go` - Move tests to query package
- `types/family_record_relationship_test.go` - Move tests to query package

---

### 2. Make Graph the Only Way to Query Relationships

**Current State:**
```go
// query/relationship_helpers.go
func (node *IndividualNode) getSpousesFromEdges() []*IndividualNode  // Private
```

**What Should Happen:**
```go
// query/relationships.go (or new file)
// ✅ PUBLIC: These become the ONLY way to get relationships
func (node *IndividualNode) Spouses() []*IndividualNode
func (node *IndividualNode) Children() []*IndividualNode
func (node *IndividualNode) Parents() []*IndividualNode
func (node *IndividualNode) Siblings() []*IndividualNode

// ✅ Also add convenience methods on Graph
func (g *Graph) GetSpouses(xrefID string) ([]*IndividualNode, error)
func (g *Graph) GetChildren(xrefID string) ([]*IndividualNode, error)
func (g *Graph) GetParents(xrefID string) ([]*IndividualNode, error)
```

**Files to Modify:**
- `query/relationship_helpers.go` - Make methods public, add Graph-level methods
- `query/node.go` - Add public relationship methods to IndividualNode
- `query/graph.go` - Add convenience query methods

---

### 3. Ensure Records Are Referenced, Not Duplicated

**Current State:**
```go
// query/node.go
type IndividualNode struct {
    *BaseNode
    Individual *types.IndividualRecord  // ✅ Good: Reference, not copy
}

type BaseNode struct {
    xrefID   string
    nodeType NodeType
    record   types.Record  // ✅ Good: Also references record
    inEdges  []*Edge
    outEdges []*Edge
}
```

**Status:** ✅ **Already correct!** Nodes reference records, don't duplicate them.

**What to Verify:**
- Ensure `BuildGraph()` doesn't copy record data
- Ensure graph nodes always reference records from tree
- No serialization/deserialization that duplicates data

**Files to Check:**
- `query/builder.go` - Verify `createNodes()` only creates references
- `query/hybrid_serialization.go` - Ensure deserialization references tree records
- `query/lazy_node.go` - Ensure lazy loading references tree records

---

### 4. Simplify GedcomTree to Be Validation-Only

**Current State:**
```go
// types/tree.go
type GedcomTree struct {
    individuals  map[string]Record
    families     map[string]Record
    // ... more maps
    xrefIndex    map[string]Record
    uuidIndex    map[string]Record
}
```

**What Should Happen:**
```go
// types/tree.go
type GedcomTree struct {
    // ✅ KEEP: Storage for validation
    individuals  map[string]Record
    families     map[string]Record
    // ... more maps
    
    // ✅ KEEP: Indexes for validation
    xrefIndex    map[string]Record
    uuidIndex    map[string]Record
    
    // ❌ REMOVE or DEPRECATE: Graph reference
    // queryGraph interface{}  // Not needed if graph is separate
}

// ✅ KEEP: Simple getters for validation
func (gt *GedcomTree) GetIndividual(xrefID string) Record
func (gt *GedcomTree) GetAllIndividuals() map[string]Record

// ❌ REMOVE: Any methods that do relationship traversal
```

**Files to Modify:**
- `types/tree.go` - Remove graph reference, simplify to data container
- `types/individual_record.go` - Remove `getTree()` usage in relationship methods

---

### 5. Update BuildGraph to Be Explicit

**Current State:**
```go
// query/builder.go
func BuildGraph(tree *types.GedcomTree) (*Graph, error) {
    graph := NewGraph(tree)
    createNodes(graph, tree)  // Creates nodes referencing records
    createEdges(graph, tree)  // Creates edges
    return graph, nil
}
```

**What Should Happen:**
```go
// query/builder.go
func BuildGraph(tree *types.GedcomTree) (*Graph, error) {
    graph := NewGraph(tree)
    
    // Phase 1: Create nodes (reference records, don't copy)
    if err := createNodes(graph, tree); err != nil {
        return nil, fmt.Errorf("failed to create nodes: %w", err)
    }
    
    // Phase 2: Create edges (builds relationship graph)
    if err := createEdges(graph, tree); err != nil {
        return nil, fmt.Errorf("failed to create edges: %w", err)
    }
    
    // Phase 3: Build indexes for query performance
    graph.indexes.buildIndexes(graph)
    
    return graph, nil
}

// ✅ ENSURE: createNodes only creates references
func createNodes(graph *Graph, tree *types.GedcomTree) error {
    individuals := tree.GetAllIndividuals()
    for xrefID, record := range individuals {
        indi, ok := record.(*types.IndividualRecord)
        if !ok {
            continue
        }
        // ✅ Reference, don't copy
        node := NewIndividualNode(xrefID, indi)
        graph.AddNode(node)
    }
    // ... similar for other types
}
```

**Status:** ✅ **Already mostly correct**, but verify no data copying happens.

---

### 6. Add Validation Steps

**Current State:**
- ✅ **Validation after parsing exists**: `validator` package validates `GedcomTree`
- ❌ **Validation after graph building MISSING**: No graph integrity validation

**What Should Happen:**

#### Step 1: Validate Records (After Parsing)

```go
// Current: Already exists in validator package
// cmd/gedcom/commands/validate.go or similar

p := parser.NewHierarchicalParser()
tree, err := p.Parse("file.ged")
if err != nil {
    return err
}

// ✅ VALIDATE RECORDS (already exists)
errorManager := types.NewErrorManager()
validator := validator.NewGedcomValidator(errorManager)
if err := validator.Validate(tree); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

// Check for validation errors
if errorManager.HasSevereErrors() {
    return fmt.Errorf("severe validation errors found")
}
```

**What Record Validation Checks:**
- ✅ Header record structure
- ✅ Individual record structure (required fields, valid tags)
- ✅ Family record structure
- ✅ Cross-reference validity (XREFs resolve)
- ✅ Required records (SUBM, etc.)
- ✅ Date format validity
- ✅ Name structure validity

**Files:**
- `validator/validator.go` - Already exists ✅
- `validator/individual_validator.go` - Already exists ✅
- `validator/family_validator.go` - Already exists ✅
- `validator/cross_reference_validator.go` - Already exists ✅

#### Step 2: Validate Graph (After Building)

```go
// NEW: Add graph validation
// query/graph_validator.go (to be created)

func ValidateGraph(graph *Graph) error {
    errorManager := types.NewErrorManager()
    validator := NewGraphValidator(errorManager)
    return validator.Validate(graph)
}

// query/builder.go - Updated
func BuildGraph(tree *types.GedcomTree) (*Graph, error) {
    graph := NewGraph(tree)
    
    // Phase 1: Create nodes
    if err := createNodes(graph, tree); err != nil {
        return nil, fmt.Errorf("failed to create nodes: %w", err)
    }
    
    // Phase 2: Create edges
    if err := createEdges(graph, tree); err != nil {
        return nil, fmt.Errorf("failed to create edges: %w", err)
    }
    
    // Phase 3: Build indexes
    graph.indexes.buildIndexes(graph)
    
    // ✅ NEW: Validate graph integrity
    if err := ValidateGraph(graph); err != nil {
        return nil, fmt.Errorf("graph validation failed: %w", err)
    }
    
    return graph, nil
}
```

**What Graph Validation Should Check:**
- ✅ **Edge Consistency**: All edges have valid From/To nodes
- ✅ **Bidirectional Edges**: FAMC/FAMS edges are properly bidirectional
- ✅ **Family Structure**: Families have valid HUSB/WIFE/CHIL edges
- ✅ **Orphaned Nodes**: No nodes without any edges (if required)
- ✅ **Circular References**: No impossible relationship cycles
- ✅ **Edge Type Validity**: All edges have valid edge types
- ✅ **Node-Record Consistency**: All nodes reference valid records
- ✅ **Relationship Integrity**: Spouse/child/parent relationships are consistent

**Files to Create:**
- `query/graph_validator.go` - NEW: Graph validation logic
- `query/graph_validator_test.go` - NEW: Tests for graph validation

**Example Graph Validator:**

```go
// query/graph_validator.go
package query

import "github.com/lesfleursdelanuitdev/ligneous-gedcom/types"

type GraphValidator struct {
    errorManager *types.ErrorManager
}

func NewGraphValidator(errorManager *types.ErrorManager) *GraphValidator {
    return &GraphValidator{errorManager: errorManager}
}

func (gv *GraphValidator) Validate(graph *Graph) error {
    // Validate edge consistency
    if err := gv.validateEdges(graph); err != nil {
        return err
    }
    
    // Validate family structure
    if err := gv.validateFamilies(graph); err != nil {
        return err
    }
    
    // Validate relationship integrity
    if err := gv.validateRelationships(graph); err != nil {
        return err
    }
    
    return nil
}

func (gv *GraphValidator) validateEdges(graph *Graph) error {
    // Check all edges have valid From/To nodes
    // Check bidirectional edges are consistent
    // ...
}

func (gv *GraphValidator) validateFamilies(graph *Graph) error {
    // Check families have valid structure
    // Check HUSB/WIFE/CHIL edges are correct
    // ...
}

func (gv *GraphValidator) validateRelationships(graph *Graph) error {
    // Check relationship consistency
    // Check for impossible cycles
    // ...
}
```

**Integration Points:**

```go
// Recommended workflow
func ProcessGEDCOM(filename string) (*Graph, error) {
    // Step 1: Parse
    p := parser.NewHierarchicalParser()
    tree, err := p.Parse(filename)
    if err != nil {
        return nil, fmt.Errorf("parse failed: %w", err)
    }
    
    // Step 2: Validate Records
    errorManager := types.NewErrorManager()
    recordValidator := validator.NewGedcomValidator(errorManager)
    if err := recordValidator.Validate(tree); err != nil {
        return nil, fmt.Errorf("record validation failed: %w", err)
    }
    if errorManager.HasSevereErrors() {
        return nil, fmt.Errorf("severe validation errors: %v", errorManager.Errors())
    }
    
    // Step 3: Build Graph
    graph, err := query.BuildGraph(tree)
    if err != nil {
        return nil, fmt.Errorf("graph build failed: %w", err)
    }
    
    // Step 4: Validate Graph (NEW)
    graphErrorManager := types.NewErrorManager()
    graphValidator := query.NewGraphValidator(graphErrorManager)
    if err := graphValidator.Validate(graph); err != nil {
        return nil, fmt.Errorf("graph validation failed: %w", err)
    }
    if graphErrorManager.HasSevereErrors() {
        return nil, fmt.Errorf("severe graph errors: %v", graphErrorManager.Errors())
    }
    
    return graph, nil
}
```

---

## Migration Impact

### Breaking Changes

**1. Record Relationship Methods Removed:**
```go
// ❌ BREAKING: This will no longer work
record, _ := tree.GetIndividual("@I1@")
spouses, err := record.(*types.IndividualRecord).Spouses()  // ❌ Method removed

// ✅ NEW: Must use graph
graph, _ := query.BuildGraph(tree)
node := graph.GetIndividual("@I1@")
spouses := node.Spouses()  // ✅ Works
```

**2. Tests Need to Move:**
- `types/individual_record_relationship_test.go` → `query/individual_node_relationship_test.go`
- `types/family_record_relationship_test.go` → `query/family_node_relationship_test.go`

**3. API Changes:**
- Any code using `record.Spouses()` must use `graph.GetSpouses()` or `node.Spouses()`
- Relationship queries become graph-only

---

## Benefits of This Architecture

### 1. Clear Separation of Concerns
- **Records**: Data storage and validation
- **Graph**: Query engine and relationship traversal
- **No overlap**: Each has a single responsibility

### 2. No Duplication
- Relationship logic exists in ONE place (graph)
- Records are referenced, not copied
- Single source of truth for relationships

### 3. Better Performance
- Graph can optimize relationship queries
- Can add caching, indexes, lazy loading
- Records stay simple and fast

### 4. Easier to Maintain
- One place to fix relationship bugs
- One place to add new relationship queries
- Clearer code organization

### 5. Validation vs Query Separation
- Records can be validated independently
- Graph can be built optionally (for querying)
- Can validate without building graph

---

## Implementation Plan

### Phase 1: Preparation (No Breaking Changes)
1. ✅ Document current architecture (done)
2. ✅ Document validation requirements (done)
3. Add graph-level relationship methods alongside record methods
4. Add deprecation warnings to record relationship methods
5. **Create graph validator** (`query/graph_validator.go`)
6. **Add graph validation to BuildGraph**
7. Update documentation

### Phase 2: Migration (Breaking Changes)
1. Move relationship tests from `types` to `query`
2. Remove relationship methods from records
3. Update all internal code to use graph
4. Update examples and documentation
5. **Ensure validation runs in all entry points**

### Phase 3: Cleanup
1. Remove graph reference from GedcomTree
2. Simplify record code
3. Optimize graph relationship queries
4. Add comprehensive graph tests
5. **Enhance graph validation rules**

---

## Code Examples: Before vs After

### Before (Current - Duplicated)

```go
// Using records (tree-based)
tree := parseGEDCOM(file)
record := tree.GetIndividual("@I1@")
spouses, _ := record.(*types.IndividualRecord).Spouses()  // Tree traversal

// Using graph (edge-based)
graph, _ := query.BuildGraph(tree)
node := graph.GetIndividual("@I1@")
spouses := node.getSpousesFromEdges()  // Edge traversal (private)
```

### After (Your Vision)

```go
// Validation only (records)
tree := parseGEDCOM(file)
record := tree.GetIndividual("@I1@")
name := record.(*types.IndividualRecord).GetName()  // ✅ Data access
// record.Spouses()  // ❌ Doesn't exist

// Queries (graph only)
graph, _ := query.BuildGraph(tree)
node := graph.GetIndividual("@I1@")
spouses := node.Spouses()  // ✅ Public method, edge-based

// Or via graph convenience method
spouses, _ := graph.GetSpouses("@I1@")  // ✅ Also works
```

---

## Summary

**What You're Saying:**
- Records = Simple data containers for validation
- **Validation after parsing** = Validate record structure
- Graph = Query engine for relationships
- **Validation after graph building** = Validate graph integrity
- Clear separation, no duplication

**What Needs to Change:**
1. ❌ Remove relationship methods from records
2. ✅ Make graph the only way to query relationships
3. ✅ Ensure records are referenced, not duplicated (already done)
4. ✅ Simplify GedcomTree to validation-only
5. ✅ Update BuildGraph to be explicit about references
6. ✅ **Add graph validation after BuildGraph** (NEW)

**Validation Flow:**
```
Parse → Validate Records → Build Graph → Validate Graph → Use Graph
```

**Impact:**
- Breaking changes for code using record relationship methods
- Tests need to move to query package
- **New graph validator needed** (to be created)
- Clearer, simpler architecture
- Better performance and maintainability
- **Better data integrity** with two-stage validation

This is a **much cleaner architecture** than the current state. The separation of data (records) and queries (graph) is a solid design principle, and adding validation at both stages ensures data integrity throughout the pipeline.

