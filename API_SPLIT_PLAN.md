# Plan: Splitting REST API into Separate Project

**Date:** 2025-01-27  
**Purpose:** Extract the REST API server into its own independent project  
**Status:** Planning Phase  
**Repository Name:** `ligneous-gedcom-api` ✅ (Confirmed)

---

## Executive Summary

This plan outlines the strategy to split the REST API (`api/` package and `cmd/api/`) into a separate project that depends on the core `ligneous-gedcom` library. This separation will provide:

- **Independent versioning** and release cycles
- **Clearer dependencies** and separation of concerns
- **Better maintainability** with focused repositories
- **Flexibility** for different deployment strategies
- **Easier contribution** with smaller, focused codebases

---

## 1. Current State Analysis

### 1.1 What Needs to be Extracted

**Files to Move:**
```
api/
├── server.go              # HTTP server implementation
├── files.go               # File upload/management endpoints
├── files_background.go    # Background graph persistence
├── individuals.go         # Individual endpoints (478 lines)
├── relationships.go       # Relationship endpoints (732 lines)
├── validation.go         # Validation endpoints
├── README.md             # API documentation
├── IMPLEMENTATION_STATUS.md
├── HYBRID_STORAGE_USAGE_ANALYSIS.md
├── PERFORMANCE_ANALYSIS.md
└── ... (other docs)

cmd/api/
└── main.go               # API server entry point
```

**Dependencies on Core Library:**
- `github.com/lesfleursdelanuitdev/ligneous-gedcom/parser`
- `github.com/lesfleursdelanuitdev/ligneous-gedcom/query`
- `github.com/lesfleursdelanuitdev/ligneous-gedcom/types`

**What Stays in Core:**
- All core packages (parser, query, types, validator, exporter, duplicate, diff)
- CLI application (`cmd/gedcom/`)
- All test data and test files

### 1.2 Current API Structure

**Endpoints Implemented:**
- ✅ File upload (`POST /api/v1/files`)
- ✅ File listing (`GET /api/v1/files`)
- ✅ File info (`GET /api/v1/files/{file_id}`)
- ✅ File deletion (`DELETE /api/v1/files/{file_id}`)
- ✅ File validation (`POST /api/v1/files/{file_id}/validate`)
- ✅ Individual queries (`GET /api/v1/files/{file_id}/individuals`)
- ✅ Individual search (`POST /api/v1/files/{file_id}/individuals/search`)
- ✅ Relationship queries (parents, children, siblings, spouses, etc.)
- ✅ Health check (`GET /health`)

**Features:**
- In-memory file storage
- Background graph persistence (PostgreSQL + BadgerDB)
- Graph building (in-memory + hybrid)
- Query execution via graph

---

## 2. Proposed Structure

### 2.1 New Repository Structure

**Repository Name:** `ligneous-gedcom-api`

**Directory Structure:**
```
ligneous-gedcom-api/
├── cmd/
│   └── api/
│       └── main.go              # Server entry point
├── internal/
│   ├── server/
│   │   ├── server.go            # HTTP server
│   │   ├── routes.go            # Route registration
│   │   └── middleware.go        # Middleware (auth, logging, etc.)
│   ├── handlers/
│   │   ├── files.go             # File management handlers
│   │   ├── individuals.go       # Individual query handlers
│   │   ├── relationships.go      # Relationship query handlers
│   │   ├── validation.go        # Validation handlers
│   │   └── health.go            # Health check handler
│   ├── storage/
│   │   ├── filestore.go         # File metadata storage
│   │   └── graph_persistence.go # Graph persistence logic
│   └── models/
│       ├── response.go          # API response models
│       └── request.go           # API request models
├── pkg/
│   └── api/                     # Public API types (if needed)
├── config/
│   └── config.go                # Configuration management
├── docs/
│   ├── API.md                   # API documentation
│   ├── DEPLOYMENT.md            # Deployment guide
│   └── DEVELOPMENT.md           # Development guide
├── docker/
│   ├── Dockerfile               # Docker image
│   └── docker-compose.yml       # Local development setup
├── .github/
│   └── workflows/
│       └── ci.yml               # CI/CD pipeline
├── go.mod                       # Dependencies
├── go.sum
├── README.md
├── LICENSE
└── .gitignore
```

### 2.2 Dependency Management

**go.mod (New Project):**
```go
module github.com/lesfleursdelanuitdev/ligneous-gedcom-api

go 1.24.0

require (
    // Core library dependency
    github.com/lesfleursdelanuitdev/ligneous-gedcom v1.0.0
    
    // HTTP server dependencies
    github.com/google/uuid v1.6.0
    
    // Optional: Router (if upgrading from net/http)
    // github.com/gorilla/mux v1.8.1
    // OR
    // github.com/go-chi/chi/v5 v5.0.10
    
    // Optional: Authentication
    // github.com/golang-jwt/jwt/v5 v5.2.0
    
    // Optional: Rate limiting
    // golang.org/x/time/rate v0.5.0
    
    // Optional: Redis caching
    // github.com/redis/go-redis/v9 v9.3.0
)
```

**Version Strategy:**
- Use semantic versioning for the API project
- Pin core library version (e.g., `v1.0.0`)
- Update core library dependency as needed
- API version can evolve independently

---

## 3. Migration Plan

### Phase 1: Preparation (Week 1)

**Tasks:**
1. ✅ **Create new repository** `ligneous-gedcom-api`
2. ✅ **Set up basic structure** (directories, go.mod, README)
3. ✅ **Copy API files** from `api/` to new project
4. ✅ **Copy cmd/api/main.go** to new project
5. ✅ **Update imports** to use core library dependency
6. ✅ **Test basic compilation** with core library dependency

**Deliverables:**
- New repository with basic structure
- All API code copied and imports updated
- Project compiles successfully

### Phase 2: Refactoring (Week 2)

**Tasks:**
1. **Reorganize code structure**
   - Move handlers to `internal/handlers/`
   - Move server logic to `internal/server/`
   - Move storage logic to `internal/storage/`
   - Extract models to `internal/models/`

2. **Improve code organization**
   - Separate concerns (handlers, business logic, storage)
   - Add configuration management
   - Standardize error handling

3. **Add missing features**
   - Configuration file support (YAML/JSON)
   - Environment variable support
   - Logging improvements
   - Request ID tracking

**Deliverables:**
- Clean, organized code structure
- Configuration management
- Improved error handling

### Phase 3: Testing & Documentation (Week 3)

**Tasks:**
1. **Add tests**
   - Unit tests for handlers
   - Integration tests for endpoints
   - Test with real GEDCOM files

2. **Documentation**
   - API documentation (OpenAPI/Swagger)
   - Deployment guide
   - Development guide
   - Migration guide (from monorepo)

3. **CI/CD setup**
   - GitHub Actions workflow
   - Automated testing
   - Docker image building

**Deliverables:**
- Test suite with good coverage
- Complete documentation
- CI/CD pipeline

### Phase 4: Deployment & Migration (Week 4)

**Tasks:**
1. **Docker setup**
   - Dockerfile for API server
   - docker-compose.yml for local development
   - Production deployment configs

2. **Update core library**
   - Remove `api/` package from core library
   - Remove `cmd/api/` from core library
   - Update core library README
   - Add note about API project

3. **Release**
   - Tag first version of API project (v1.0.0)
   - Update core library to remove API code
   - Announce split in both repositories

**Deliverables:**
- Docker images
- Updated core library (API removed)
- Both projects released

---

## 4. Detailed Migration Steps

### 4.1 Step 1: Create New Repository

```bash
# Create new directory
mkdir ligneous-gedcom-api
cd ligneous-gedcom-api

# Initialize git
git init
git remote add origin https://github.com/lesfleursdelanuitdev/ligneous-gedcom-api.git

# Create basic structure
mkdir -p cmd/api internal/{server,handlers,storage,models} config docs docker
```

### 4.2 Step 2: Set Up go.mod

```bash
# Initialize go module
go mod init github.com/lesfleursdelanuitdev/ligneous-gedcom-api

# Add core library dependency
go get github.com/lesfleursdelanuitdev/ligneous-gedcom@v1.0.0

# Add other dependencies
go get github.com/google/uuid@latest
```

### 4.3 Step 3: Copy and Refactor Code

**File Mapping:**
```
Old Location                          New Location
─────────────────────────────────────────────────────────
api/server.go              →  internal/server/server.go
api/files.go               →  internal/handlers/files.go
api/files_background.go    →  internal/storage/graph_persistence.go
api/individuals.go        →  internal/handlers/individuals.go
api/relationships.go      →  internal/handlers/relationships.go
api/validation.go          →  internal/handlers/validation.go
api/server.go (response)   →  internal/models/response.go
cmd/api/main.go            →  cmd/api/main.go
```

**Import Updates:**
```go
// Old (in monorepo)
import "github.com/lesfleursdelanuitdev/ligneous-gedcom/api"

// New (in API project)
import (
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)
```

### 4.4 Step 4: Update Core Library

**Remove from core library:**
```bash
# In ligneous-gedcom repository
rm -rf api/
rm -rf cmd/api/
```

**Update core library README:**
```markdown
## REST API

The REST API has been moved to a separate project:
- **Repository:** [ligneous-gedcom-api](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-api)
- **Documentation:** [API Documentation](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-api/blob/main/docs/API.md)

To use the API, install the API server:
```bash
go install github.com/lesfleursdelanuitdev/ligneous-gedcom-api/cmd/api@latest
```
```

---

## 5. Benefits of Separation

### 5.1 Independent Versioning

**Core Library:**
- Version: `v1.0.0`, `v1.1.0`, `v2.0.0` (breaking changes)
- Focus: Core functionality, performance, stability

**API Project:**
- Version: `v1.0.0`, `v1.1.0`, `v2.0.0` (API changes)
- Focus: API features, endpoints, deployment

### 5.2 Clearer Dependencies

**Before (Monorepo):**
- API code mixed with core library
- Hard to understand what's API vs core
- All code in one repository

**After (Separate):**
- Clear dependency: API → Core Library
- API project is a consumer of core library
- Easier to understand architecture

### 5.3 Better Maintainability

- **Smaller codebases** = easier to navigate
- **Focused issues** = API issues in API repo, core issues in core repo
- **Independent releases** = API can release without waiting for core
- **Clear ownership** = API team vs core team

### 5.4 Deployment Flexibility

- **API server** can be deployed independently
- **Core library** can be used by other projects (CLI, other APIs, etc.)
- **Different release cycles** = API can iterate faster

---

## 6. Challenges & Solutions

### 6.1 Challenge: Breaking Changes in Core Library

**Problem:** API project depends on core library. If core library has breaking changes, API needs updates.

**Solution:**
- Use semantic versioning in core library
- Pin specific version in API project
- Update API project when ready
- Use Go modules' versioning features

### 6.2 Challenge: Shared Code

**Problem:** Some code might be useful in both projects.

**Solution:**
- Keep shared utilities in core library
- Or create a shared utilities package
- Or duplicate if truly API-specific

### 6.3 Challenge: Testing

**Problem:** API tests need to test against core library.

**Solution:**
- Use core library as dependency in tests
- Test with real GEDCOM files
- Integration tests with actual core library

### 6.4 Challenge: Documentation

**Problem:** Documentation split across two repositories.

**Solution:**
- Clear links between repositories
- API docs in API repository
- Core docs in core repository
- Cross-reference where needed

---

## 7. Implementation Checklist

### Phase 1: Preparation
- [ ] Create new repository `ligneous-gedcom-api`
- [ ] Set up basic directory structure
- [ ] Initialize go.mod with core library dependency
- [ ] Copy all API files from `api/` directory
- [ ] Copy `cmd/api/main.go`
- [ ] Update all imports to use core library
- [ ] Test compilation

### Phase 2: Refactoring
- [ ] Reorganize code into `internal/` structure
- [ ] Extract models to `internal/models/`
- [ ] Move handlers to `internal/handlers/`
- [ ] Move server logic to `internal/server/`
- [ ] Move storage logic to `internal/storage/`
- [ ] Add configuration management
- [ ] Improve error handling
- [ ] Add logging improvements

### Phase 3: Testing & Documentation
- [ ] Write unit tests for handlers
- [ ] Write integration tests
- [ ] Create API documentation (OpenAPI/Swagger)
- [ ] Write deployment guide
- [ ] Write development guide
- [ ] Set up CI/CD pipeline
- [ ] Add Docker support

### Phase 4: Migration
- [ ] Create Dockerfile
- [ ] Create docker-compose.yml
- [ ] Remove `api/` from core library
- [ ] Remove `cmd/api/` from core library
- [ ] Update core library README
- [ ] Tag first release of API project
- [ ] Announce split

---

## 8. File-by-File Migration Guide

### 8.1 Core Library Files to Remove

```bash
# Files to delete from ligneous-gedcom repository
rm -rf api/
rm -rf cmd/api/
```

### 8.2 API Project Files to Create

**cmd/api/main.go:**
```go
package main

import (
    "flag"
    "log"
    "os"
    
    "github.com/lesfleursdelanuitdev/ligneous-gedcom-api/internal/server"
)

func main() {
    port := flag.String("port", "8080", "Port to listen on")
    flag.Parse()
    
    if envPort := os.Getenv("PORT"); envPort != "" {
        port = &envPort
    }
    
    srv := server.NewServer(*port)
    if err := srv.Start(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

**internal/server/server.go:**
```go
package server

import (
    "github.com/lesfleursdelanuitdev/ligneous-gedcom-api/internal/handlers"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom-api/internal/storage"
)

type Server struct {
    // ... existing fields
    handlers *handlers.Handlers
    storage  *storage.FileStore
}

func NewServer(port string) *Server {
    // ... initialization
}
```

### 8.3 Import Updates

**Before (in monorepo):**
```go
import (
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/api"
)
```

**After (in API project):**
```go
import (
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/parser"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
    
    "github.com/lesfleursdelanuitdev/ligneous-gedcom-api/internal/handlers"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom-api/internal/storage"
)
```

---

## 9. Versioning Strategy

### 9.1 Core Library Versioning

**Current:** `v1.0.0`  
**Strategy:** Semantic versioning
- `v1.0.0` → `v1.1.0` (minor features)
- `v1.1.0` → `v2.0.0` (breaking changes)

### 9.2 API Project Versioning

**Initial:** `v1.0.0`  
**Strategy:** Semantic versioning
- `v1.0.0` → `v1.1.0` (new endpoints)
- `v1.1.0` → `v2.0.0` (breaking API changes)

### 9.3 Dependency Management

**API Project go.mod:**
```go
require (
    github.com/lesfleursdelanuitdev/ligneous-gedcom v1.0.0
)
```

**Update Strategy:**
- Pin to specific version initially
- Update when new core features needed
- Test thoroughly before updating

---

## 10. Testing Strategy

### 10.1 Unit Tests

**Test Structure:**
```
internal/
├── handlers/
│   ├── files_test.go
│   ├── individuals_test.go
│   └── relationships_test.go
└── storage/
    └── filestore_test.go
```

**Test Approach:**
- Mock core library interfaces where possible
- Use real core library for integration tests
- Test with real GEDCOM files

### 10.2 Integration Tests

**Test Structure:**
```
tests/
├── integration/
│   ├── api_test.go
│   └── testdata/
│       └── *.ged
```

**Test Approach:**
- Full end-to-end API tests
- Use test GEDCOM files
- Test all endpoints
- Test error cases

---

## 11. Documentation Plan

### 11.1 API Project Documentation

**README.md:**
- Quick start guide
- Installation instructions
- Configuration
- API overview
- Links to detailed docs

**docs/API.md:**
- Complete API reference
- All endpoints documented
- Request/response examples
- Error codes

**docs/DEPLOYMENT.md:**
- Docker deployment
- Production setup
- Environment variables
- Database setup

**docs/DEVELOPMENT.md:**
- Development setup
- Running tests
- Contributing guidelines

### 11.2 Core Library Documentation Updates

**Update README.md:**
- Remove API section
- Add link to API project
- Update installation instructions

---

## 12. Deployment Considerations

### 12.1 Docker Support

**Dockerfile:**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api-server ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api-server .
EXPOSE 8080
CMD ["./api-server"]
```

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://...
      - STORAGE_DIR=/data
    volumes:
      - ./data:/data
```

### 12.2 Environment Variables

**Required:**
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `STORAGE_DIR` - File storage directory

**Optional:**
- `LOG_LEVEL` - Logging level
- `CORS_ORIGIN` - CORS origin
- `RATE_LIMIT` - Rate limiting config

---

## 13. Timeline

### Week 1: Preparation
- Create repository
- Copy code
- Update imports
- Basic testing

### Week 2: Refactoring
- Reorganize structure
- Improve code organization
- Add configuration

### Week 3: Testing & Documentation
- Write tests
- Create documentation
- Set up CI/CD

### Week 4: Migration & Release
- Docker setup
- Remove from core library
- Release both projects

**Total Time:** ~4 weeks

---

## 14. Success Criteria

### 14.1 Technical Success

- ✅ API project compiles and runs
- ✅ All endpoints work correctly
- ✅ Tests pass
- ✅ Documentation complete
- ✅ Docker images build

### 14.2 Migration Success

- ✅ Core library no longer contains API code
- ✅ API project is independent
- ✅ Both projects released
- ✅ Documentation updated

### 14.3 Long-term Success

- ✅ Independent versioning works
- ✅ Both projects maintainable
- ✅ Clear separation of concerns
- ✅ Easy to contribute to either project

---

## 15. Risks & Mitigation

### 15.1 Risk: Breaking Changes

**Risk:** Core library changes break API project.

**Mitigation:**
- Pin core library version
- Test thoroughly before updating
- Use semantic versioning
- Document breaking changes

### 15.2 Risk: Code Duplication

**Risk:** Some code duplicated between projects.

**Mitigation:**
- Keep shared code in core library
- Extract utilities to shared package if needed
- Document what belongs where

### 15.3 Risk: Maintenance Overhead

**Risk:** Two repositories to maintain.

**Mitigation:**
- Clear ownership
- Good documentation
- Automated testing
- CI/CD pipelines

---

## 16. Next Steps

1. ✅ **Repository name decided:** `ligneous-gedcom-api`
2. **Review this plan** with team
3. **Create new repository** `ligneous-gedcom-api`
4. **Start Phase 1** (Preparation)
5. **Iterate** based on feedback
6. **Complete migration** in 4 weeks

---

## 17. Questions & Decisions Needed

1. ✅ **Repository name:** `ligneous-gedcom-api` (decided)
2. **Versioning:** Start at `v1.0.0` or `v0.1.0`?
3. **Router:** Keep `net/http` or upgrade to `gorilla/mux` or `chi`?
4. **Authentication:** Add now or later?
5. **Documentation:** OpenAPI/Swagger now or later?

---

**Status:** Ready for Review  
**Next Action:** Create new repository and start Phase 1

