# Persistence Architecture Analysis: Core vs API

**Date:** 2025-01-27  
**Question:** After splitting the API, where should data persistence code live?  
**Status:** Analysis & Recommendation

---

## Executive Summary

**Recommendation: Keep persistence capabilities in core, but make it optional and interface-based.**

**Rationale:**
- Core library should provide capabilities, not limitations
- Different consumers have different persistence needs
- API needs persistence, but CLI and other tools might too
- Separation of concerns: Core provides "how", API decides "when/where"

**Approach:**
- **Core Library:** Provides in-memory (default) + optional persistence interfaces
- **API Project:** Implements persistence strategy using core interfaces
- **CLI:** Uses in-memory (no persistence needed)

---

## 1. Current State Analysis

### 1.1 What Persistence Code Exists

**In Core Library (`query/` package):**
- `BuildGraph()` - In-memory graph (default)
- `BuildGraphHybrid()` - SQLite + BadgerDB persistence
- `BuildGraphHybridPostgres()` - PostgreSQL + BadgerDB persistence
- `BuildGraphLazy()` - Lazy loading (in-memory)
- `HybridStorage` - SQLite storage implementation
- `HybridStoragePostgres` - PostgreSQL storage implementation
- `HybridQueryHelpers` - Query helpers for SQLite
- `HybridQueryHelpersPostgres` - Query helpers for PostgreSQL
- ~15+ files related to hybrid storage

**In API (`api/` package):**
- Uses `BuildGraphHybridPostgres()` for persistence
- Manages file storage paths
- Background graph persistence logic

### 1.2 Who Uses What

**Current Usage:**
- **CLI (`cmd/gedcom/`):** Uses `BuildGraph()` (in-memory only)
- **API (`api/`):** Uses `BuildGraph()` (in-memory) + `BuildGraphHybridPostgres()` (background persistence)
- **Tests:** Use both in-memory and hybrid storage

---

## 2. Use Case Analysis

### 2.1 API Server Needs

**Requirements:**
- ✅ Persistence across server restarts
- ✅ Handle multiple files simultaneously
- ✅ Fast queries (indexed storage)
- ✅ Memory efficiency (lazy loading)
- ✅ Scalability (PostgreSQL for multi-server)

**Current Solution:**
- Uses `BuildGraphHybridPostgres()` 
- Background persistence after upload
- In-memory graph for immediate queries
- Hybrid graph for persistent queries

**Verdict:** **API definitely needs persistence**

### 2.2 CLI Tool Needs

**Requirements:**
- ✅ Fast startup (no database connection)
- ✅ Process and exit (no persistence needed)
- ✅ Simple deployment (single binary)
- ✅ No external dependencies

**Current Solution:**
- Uses `BuildGraph()` (in-memory)
- No persistence needed

**Verdict:** **CLI doesn't need persistence**

### 2.3 Other Potential Consumers

**Potential Use Cases:**
1. **Desktop Application:**
   - Might want to save/load graphs for performance
   - Could use SQLite hybrid storage
   - **Needs persistence**

2. **Batch Processing Script:**
   - Process multiple files
   - Might want to cache graphs
   - **Might need persistence**

3. **Web Application (different from API):**
   - Similar to API needs
   - **Needs persistence**

4. **Library Consumer:**
   - Embed in their own application
   - Might have their own persistence strategy
   - **Might need persistence interfaces**

**Verdict:** **Other consumers might need persistence**

---

## 3. Architecture Options

### Option 1: Keep All Persistence in Core (Recommended)

**Structure:**
```
Core Library (ligneous-gedcom)
├── query/
│   ├── graph.go              # Core graph (in-memory)
│   ├── builder.go            # BuildGraph() - in-memory
│   ├── hybrid_builder.go     # BuildGraphHybrid() - SQLite
│   ├── hybrid_builder_postgres.go  # BuildGraphHybridPostgres() - PostgreSQL
│   ├── hybrid_storage.go     # SQLite implementation
│   ├── hybrid_storage_postgres.go  # PostgreSQL implementation
│   └── ... (all hybrid code)
│
API Project (ligneous-gedcom-api)
├── internal/
│   └── storage/
│       └── graph_persistence.go  # Uses core's BuildGraphHybridPostgres()
```

**Pros:**
- ✅ Core provides full capabilities
- ✅ Other consumers can use persistence
- ✅ API just uses core functions
- ✅ No code duplication
- ✅ Single source of truth

**Cons:**
- ⚠️ Core has database dependencies (SQLite, PostgreSQL drivers)
- ⚠️ Core is slightly larger
- ⚠️ CLI doesn't need persistence but it's there

**Dependencies in Core:**
```go
// Core library go.mod
require (
    github.com/mattn/go-sqlite3      // SQLite driver
    github.com/jackc/pgx/v5         // PostgreSQL driver
    github.com/dgraph-io/badger/v4  // BadgerDB
)
```

**Impact:** Minimal - these are already dependencies

---

### Option 2: Move All Persistence to API

**Structure:**
```
Core Library (ligneous-gedcom)
├── query/
│   ├── graph.go              # Core graph (in-memory only)
│   ├── builder.go            # BuildGraph() - in-memory only
│   └── ... (no hybrid code)
│
API Project (ligneous-gedcom-api)
├── internal/
│   └── storage/
│       ├── hybrid_storage.go      # SQLite implementation
│       ├── hybrid_storage_postgres.go  # PostgreSQL implementation
│       ├── hybrid_builder.go      # BuildGraphHybrid() - moved from core
│       └── graph_persistence.go   # API-specific persistence logic
```

**Pros:**
- ✅ Core is simpler (in-memory only)
- ✅ Core has no database dependencies
- ✅ Clear separation: Core = in-memory, API = persistence

**Cons:**
- ❌ Other consumers can't use persistence
- ❌ Code duplication if multiple consumers need persistence
- ❌ API has to reimplement core graph building logic
- ❌ API becomes tightly coupled to persistence implementation
- ❌ Harder to test persistence separately

**Verdict:** **Not recommended** - Too restrictive, loses flexibility

---

### Option 3: Interface-Based Approach (Hybrid)

**Structure:**
```
Core Library (ligneous-gedcom)
├── query/
│   ├── graph.go              # Core graph
│   ├── builder.go            # BuildGraph() - in-memory
│   ├── storage/
│   │   ├── interface.go      # Storage interface definitions
│   │   ├── in_memory.go      # In-memory implementation (default)
│   │   └── hybrid.go         # Hybrid storage implementation
│   └── ... (hybrid code)
│
API Project (ligneous-gedcom-api)
├── internal/
│   └── storage/
│       └── graph_persistence.go  # Uses core's storage interfaces
```

**Pros:**
- ✅ Core provides interfaces + implementations
- ✅ API can implement custom storage if needed
- ✅ Clear separation of concerns
- ✅ Flexible for different consumers

**Cons:**
- ⚠️ More complex architecture
- ⚠️ Might be over-engineering for current needs

**Verdict:** **Good for future, but not necessary now**

---

## 4. Recommendation: Option 1 (Keep in Core)

### 4.1 Why Keep Persistence in Core

**1. Core Should Provide Capabilities, Not Limitations**
- Core library is a toolkit
- Different consumers have different needs
- CLI doesn't need persistence, but others might
- Making persistence optional is better than removing it

**2. No Real Drawback**
- Database dependencies are already there
- Code is well-tested and working
- Size increase is minimal
- CLI can still use in-memory (default)

**3. API Benefits**
- API just uses core functions
- No need to reimplement
- Can focus on API-specific concerns
- Leverages existing, tested code

**4. Future Flexibility**
- Desktop apps can use SQLite hybrid storage
- Batch scripts can cache graphs
- Other web apps can use PostgreSQL hybrid storage
- Library consumers have options

### 4.2 How It Works After Split

**Core Library:**
```go
// query/builder.go
func BuildGraph(tree *types.GedcomTree) (*Graph, error) {
    // In-memory graph (default, no persistence)
}

// query/hybrid_builder.go
func BuildGraphHybrid(tree *types.GedcomTree, sqlitePath, badgerPath string, config *Config) (*Graph, error) {
    // SQLite + BadgerDB persistence (optional)
}

// query/hybrid_builder_postgres.go
func BuildGraphHybridPostgres(tree *types.GedcomTree, fileID, badgerPath, databaseURL string, config *Config) (*Graph, error) {
    // PostgreSQL + BadgerDB persistence (optional)
}
```

**API Project:**
```go
// internal/storage/graph_persistence.go
import "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"

func BuildPersistentGraph(tree *types.GedcomTree, fileID string) (*query.Graph, error) {
    // Use core's hybrid storage
    return query.BuildGraphHybridPostgres(
        tree,
        fileID,
        badgerPath,
        databaseURL,
        nil, // Use default config
    )
}
```

**CLI:**
```go
// cmd/gedcom/commands/interactive.go
import "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"

// Uses in-memory (no persistence)
graph, err := query.BuildGraph(tree)
```

### 4.3 Dependency Management

**Core Library go.mod:**
```go
module github.com/lesfleursdelanuitdev/ligneous-gedcom

require (
    // ... existing dependencies ...
    github.com/mattn/go-sqlite3 v1.14.32      // For SQLite hybrid storage
    github.com/jackc/pgx/v5 v5.8.0            // For PostgreSQL hybrid storage
    github.com/dgraph-io/badger/v4 v4.9.0     // For BadgerDB
)
```

**Note:** These are already dependencies, so no change needed.

**API Project go.mod:**
```go
module github.com/lesfleursdelanuitdev/ligneous-gedcom-api

require (
    github.com/lesfleursdelanuitdev/ligneous-gedcom v1.0.0
    // API doesn't need to import database drivers directly
    // They come transitively from core library
)
```

### 4.4 What API Owns

**API-Specific Persistence Concerns:**
1. **When to persist:** API decides (on upload, background, etc.)
2. **Where to store:** API manages file paths, storage directories
3. **How to organize:** API organizes by file_id, manages cleanup
4. **Configuration:** API configures database URLs, storage paths
5. **Lifecycle:** API manages graph lifecycle (create, load, delete)

**Core Provides:**
1. **How to persist:** Core provides the persistence mechanism
2. **Storage backends:** Core provides SQLite, PostgreSQL, BadgerDB
3. **Query interfaces:** Core provides query APIs that work with persisted graphs
4. **Graph building:** Core provides graph building with/without persistence

---

## 5. Alternative: Make Persistence Optional in Core

### 5.1 Build Tags Approach

**Option:** Use Go build tags to make persistence optional

```go
// query/builder.go
// +build !no_persistence

func BuildGraphHybrid(...) { ... }
```

**Pros:**
- CLI can build without database dependencies
- Smaller binary for CLI
- Core still provides persistence for those who need it

**Cons:**
- More complex build process
- Two versions of core library
- API still needs persistence

**Verdict:** **Not necessary** - Database dependencies are small, and CLI doesn't mind them

### 5.2 Interface-Based (Future Enhancement)

**Option:** Define storage interfaces, make implementations optional

```go
// query/storage/interface.go
type GraphStorage interface {
    SaveGraph(*Graph) error
    LoadGraph(fileID string) (*Graph, error)
    Close() error
}

// query/storage/in_memory.go
type InMemoryStorage struct { ... }

// query/storage/hybrid.go
type HybridStorage struct { ... }
```

**Pros:**
- Very flexible
- Clear separation
- Easy to add new storage backends

**Cons:**
- Requires refactoring
- More complex
- Not needed for current use cases

**Verdict:** **Good for future, but not needed now**

---

## 6. Comparison Matrix

| Aspect | Keep in Core | Move to API | Interface-Based |
|--------|-------------|-------------|-----------------|
| **Core simplicity** | ⚠️ Has DB deps | ✅ Simple | ✅ Simple |
| **API simplicity** | ✅ Uses core | ⚠️ Reimplements | ✅ Uses interfaces |
| **Other consumers** | ✅ Can use | ❌ Can't use | ✅ Can use |
| **Code duplication** | ✅ None | ❌ Possible | ✅ None |
| **Flexibility** | ✅ High | ❌ Low | ✅ Very High |
| **Testability** | ✅ Easy | ⚠️ Harder | ✅ Easy |
| **Maintenance** | ✅ Single place | ⚠️ Multiple places | ✅ Single place |
| **Current effort** | ✅ None | ❌ High | ⚠️ Medium |

**Winner:** **Keep in Core** - Best balance of simplicity, flexibility, and maintainability

---

## 7. Implementation After Split

### 7.1 Core Library (No Changes Needed)

**Structure:**
```
ligneous-gedcom/
├── query/
│   ├── graph.go                    # Core graph
│   ├── builder.go                   # BuildGraph() - in-memory
│   ├── hybrid_builder.go            # BuildGraphHybrid() - SQLite
│   ├── hybrid_builder_postgres.go   # BuildGraphHybridPostgres() - PostgreSQL
│   ├── hybrid_storage.go            # SQLite implementation
│   ├── hybrid_storage_postgres.go   # PostgreSQL implementation
│   └── ... (all hybrid code stays)
```

**No changes needed** - Everything stays as-is

### 7.2 API Project (Uses Core)

**Structure:**
```
ligneous-gedcom-api/
├── internal/
│   └── storage/
│       ├── graph_persistence.go     # Wraps core's BuildGraphHybridPostgres()
│       ├── file_metadata.go          # API-specific metadata storage
│       └── cleanup.go                # Cleanup logic
```

**API Code:**
```go
// internal/storage/graph_persistence.go
package storage

import (
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func BuildPersistentGraph(tree *types.GedcomTree, fileID, badgerPath, databaseURL string) (*query.Graph, error) {
    // Use core's hybrid storage
    return query.BuildGraphHybridPostgres(
        tree,
        fileID,
        badgerPath,
        databaseURL,
        nil, // Use default config
    )
}

func LoadPersistentGraph(fileID, badgerPath, databaseURL string) (*query.Graph, error) {
    // Load from core's hybrid storage
    // (Implementation depends on core's load capabilities)
}
```

### 7.3 CLI (No Changes)

**CLI continues to use in-memory:**
```go
// cmd/gedcom/commands/interactive.go
graph, err := query.BuildGraph(tree)  // In-memory, no persistence
```

---

## 8. Data Persistence Strategy

### 8.1 What Gets Persisted

**Graph Data (Core Library):**
- Node metadata (name, dates, places, etc.) → SQLite/PostgreSQL
- Graph structure (nodes, edges) → BadgerDB
- Indexes for fast queries → SQLite/PostgreSQL

**API Metadata (API Project):**
- File metadata (name, size, upload time, etc.)
- User information (if added)
- API keys (if added)
- Export history (if added)

**Storage Location:**
- **Graph data:** Managed by core library (SQLite files, PostgreSQL tables, BadgerDB directories)
- **API metadata:** Managed by API project (separate database or file storage)

### 8.2 Separation of Concerns

**Core Library Owns:**
- How to persist graph data
- Graph storage backends (SQLite, PostgreSQL, BadgerDB)
- Graph query interfaces
- Graph building with persistence

**API Project Owns:**
- When to persist (on upload, background, etc.)
- Where to store (file paths, directories)
- API metadata storage
- Graph lifecycle management
- Cleanup and deletion

---

## 9. Benefits of Keeping Persistence in Core

### 9.1 For Core Library

1. **Complete Toolkit:** Core provides full capabilities
2. **Flexibility:** Consumers can choose in-memory or persistence
3. **No Limitations:** Core doesn't artificially restrict features
4. **Well-Tested:** Persistence code is already tested

### 9.2 For API Project

1. **Simple Integration:** Just use core functions
2. **No Reimplementation:** Leverage existing, tested code
3. **Focus on API:** Can focus on HTTP, routing, authentication
4. **Future-Proof:** If core adds new storage backends, API gets them

### 9.3 For Other Consumers

1. **Desktop Apps:** Can use SQLite hybrid storage
2. **Batch Scripts:** Can cache graphs for performance
3. **Web Apps:** Can use PostgreSQL hybrid storage
4. **Library Users:** Have options for their use case

---

## 10. Potential Concerns & Solutions

### 10.1 Concern: Core Has Database Dependencies

**Reality:**
- Dependencies are already there
- They're small (SQLite is embedded, PostgreSQL driver is small)
- CLI doesn't need to use them (just uses in-memory)
- No runtime cost if not used

**Solution:** No action needed - already acceptable

### 10.2 Concern: Core Is Larger

**Reality:**
- Persistence code is ~15 files, ~5,000 lines
- Database drivers are small
- Total size increase is minimal
- Benefits outweigh costs

**Solution:** Acceptable trade-off

### 10.3 Concern: CLI Doesn't Need Persistence

**Reality:**
- CLI uses in-memory (default)
- Persistence is optional
- No performance impact if not used
- CLI binary size increase is minimal

**Solution:** No problem - persistence is optional

---

## 11. Final Recommendation

### ✅ **Keep Persistence in Core Library**

**Rationale:**
1. Core should provide capabilities, not limitations
2. API benefits from using existing, tested code
3. Other consumers might need persistence
4. No real drawbacks (dependencies already exist)
5. Clear separation: Core provides "how", API decides "when/where"

**Implementation:**
- **Core Library:** Keep all hybrid storage code
- **API Project:** Use core's persistence functions
- **CLI:** Continue using in-memory (no changes)

**Result:**
- ✅ Core provides complete toolkit
- ✅ API focuses on API concerns
- ✅ Other consumers have options
- ✅ No code duplication
- ✅ Single source of truth

---

## 12. Future Considerations

### 12.1 If Persistence Becomes Problematic

**Option:** Extract to separate package
```
ligneous-gedcom/
├── query/              # Core (in-memory only)
└── query/persistence/  # Optional persistence package
```

**When:** Only if persistence code becomes very large or causes issues

### 12.2 If API Needs Custom Persistence

**Option:** API implements custom storage using core interfaces
- Core provides interfaces
- API implements custom backend
- Best of both worlds

**When:** If API has unique persistence requirements

### 12.3 If Other Consumers Need Different Storage

**Option:** Add new storage backends to core
- Core provides multiple options
- Consumers choose what they need
- All benefit from shared implementation

**When:** If new storage backends are needed

---

## 13. Conclusion

**Keep persistence in core library.** This provides:
- ✅ Maximum flexibility for all consumers
- ✅ Simple integration for API
- ✅ No code duplication
- ✅ Single source of truth
- ✅ Well-tested, working code

**The separation of concerns is:**
- **Core:** Provides "how" to persist (mechanisms)
- **API:** Decides "when/where" to persist (strategy)

This is the right balance between flexibility and simplicity.

---

**Status:** Ready for Implementation  
**Next Step:** Proceed with API split, keeping persistence in core

