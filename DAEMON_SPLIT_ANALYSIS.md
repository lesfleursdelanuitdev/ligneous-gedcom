# Daemon/Service Considerations After API Split

**Date:** 2025-01-27  
**Question:** What changes are needed for the systemd daemon after splitting the API?  
**Status:** Analysis & Recommendations

---

## Executive Summary

**Current State:**
- API server runs as systemd daemon (`ligneous-api.service`)
- Background goroutine for graph persistence (`saveGraphToHybridStorage()`)
- Service file points to `/usr/local/bin/ligneous-api-server`
- Binary built from `cmd/api/main.go` in monorepo

**After Split:**
- Service file needs to point to new API project binary
- Background goroutine code moves to API project (no changes needed)
- Service configuration stays mostly the same
- Deployment process updates to build from new repository

**Changes Required:** **Minimal** - Mostly path and build updates

---

## 1. Current Daemon Setup

### 1.1 Systemd Service

**Service File:** `/etc/systemd/system/ligneous-api.service`

**Key Configuration:**
```ini
[Unit]
Description=Ligneous GEDCOM API Server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=ligneous
Group=ligneous
WorkingDirectory=/var/lib/ligneous

Environment="DATABASE_URL=postgres://..."
Environment="STORAGE_DIR=/var/lib/ligneous/storage"
Environment="PORT=8090"

ExecStart=/usr/local/bin/ligneous-api-server -port 8090
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

**Binary Location:** `/usr/local/bin/ligneous-api-server`  
**Source:** Built from `cmd/api/main.go` in monorepo

### 1.2 Background Goroutine

**Code Location:** `api/files_background.go`

**Functionality:**
- Runs as goroutine after file upload
- Persists graph to PostgreSQL + BadgerDB asynchronously
- Updates `FileMetadata` when complete
- Logs errors but doesn't block upload response

**Code:**
```go
// In api/files.go (line 183)
go s.saveGraphToHybridStorage(metadata, tree, fileID, graphStoragePath)
```

---

## 2. What Changes After Split

### 2.1 Binary Location

**Before (Monorepo):**
```bash
# Build from monorepo
cd /apps/gedcom-go
go build -o /usr/local/bin/ligneous-api-server ./cmd/api
```

**After (Separate API Project):**
```bash
# Build from API project
cd /apps/ligneous-gedcom-api
go build -o /usr/local/bin/ligneous-api-server ./cmd/api
```

**Service File:** No change needed (same binary path)

### 2.2 Background Goroutine Code

**Before:**
- Location: `api/files_background.go` in monorepo
- Imports: `github.com/lesfleursdelanuitdev/ligneous-gedcom/query`

**After:**
- Location: `internal/storage/graph_persistence.go` in API project
- Imports: `github.com/lesfleursdelanuitdev/ligneous-gedcom/query` (same!)

**Code Changes:** **None** - Just moves to new location, imports stay the same

### 2.3 Service Configuration

**Environment Variables:** No changes needed
- `DATABASE_URL` - Still needed (PostgreSQL connection)
- `STORAGE_DIR` - Still needed (file storage)
- `PORT` - Still needed (server port)

**Dependencies:** No changes needed
- Still requires PostgreSQL
- Still needs storage directory
- Still needs network

**Security Settings:** No changes needed
- Same user (`ligneous`)
- Same directories (`/var/lib/ligneous`)
- Same security hardening

---

## 3. Migration Steps

### 3.1 Step 1: Build New Binary

**From API Project:**
```bash
cd /apps/ligneous-gedcom-api
go build -o /usr/local/bin/ligneous-api-server ./cmd/api
```

**Verify:**
```bash
/usr/local/bin/ligneous-api-server --version
```

### 3.2 Step 2: Update Service File (Optional)

**If binary name changes:**
```ini
# Old
ExecStart=/usr/local/bin/ligneous-api-server -port 8090

# New (if renamed)
ExecStart=/usr/local/bin/ligneous-gedcom-api -port 8090
```

**If keeping same name:** No change needed

### 3.3 Step 3: Reload and Restart

```bash
# Reload systemd (if service file changed)
sudo systemctl daemon-reload

# Restart service
sudo systemctl restart ligneous-api

# Verify
sudo systemctl status ligneous-api
```

### 3.4 Step 4: Verify Functionality

```bash
# Test health endpoint
curl http://localhost:8090/health

# Test file upload (triggers background goroutine)
curl -X POST http://localhost:8090/api/v1/files \
  -F "file=@test.ged" \
  -F "name=test.ged"

# Check logs for background persistence
sudo journalctl -u ligneous-api -f
```

---

## 4. Background Goroutine After Split

### 4.1 Code Location

**Before:**
```
ligneous-gedcom/
└── api/
    ├── files.go
    └── files_background.go  ← Background goroutine
```

**After:**
```
ligneous-gedcom-api/
└── internal/
    └── storage/
        ├── graph_persistence.go  ← Background goroutine (moved)
        └── filestore.go
```

### 4.2 Code Changes

**Minimal Changes Needed:**

**Old Code (`api/files_background.go`):**
```go
package api

import (
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func (s *Server) saveGraphToHybridStorage(...) {
    // Uses query.BuildGraphHybridPostgres()
}
```

**New Code (`internal/storage/graph_persistence.go`):**
```go
package storage

import (
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/query"
    "github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

func (s *Server) saveGraphToHybridStorage(...) {
    // Same code, just different package
    // Still uses query.BuildGraphHybridPostgres()
}
```

**Key Point:** The function signature and logic stay the same, just moves to new package

### 4.3 Functionality

**No Changes:**
- ✅ Still runs as goroutine
- ✅ Still persists to PostgreSQL + BadgerDB
- ✅ Still updates metadata when complete
- ✅ Still logs errors
- ✅ Still doesn't block upload response

**Benefits:**
- ✅ Code is in API project (where it belongs)
- ✅ Clear separation of concerns
- ✅ Easier to maintain API-specific persistence logic

---

## 5. Service File Options

### Option 1: Keep Same Binary Name (Recommended)

**Advantage:** No service file changes needed

**Implementation:**
```bash
# Build with same name
go build -o /usr/local/bin/ligneous-api-server ./cmd/api
```

**Service File:** No changes

### Option 2: Use New Binary Name

**Advantage:** Clearer naming, reflects new project

**Implementation:**
```bash
# Build with new name
go build -o /usr/local/bin/ligneous-gedcom-api ./cmd/api
```

**Service File Update:**
```ini
# Update ExecStart
ExecStart=/usr/local/bin/ligneous-gedcom-api -port 8090
```

**Then:**
```bash
sudo systemctl daemon-reload
sudo systemctl restart ligneous-api
```

**Recommendation:** **Option 1** - Keep same name to minimize changes

---

## 6. Deployment Process Updates

### 6.1 Build Process

**Before (Monorepo):**
```bash
cd /apps/gedcom-go
go build -o /usr/local/bin/ligneous-api-server ./cmd/api
```

**After (API Project):**
```bash
cd /apps/ligneous-gedcom-api
go build -o /usr/local/bin/ligneous-api-server ./cmd/api
```

**Change:** Just different source directory

### 6.2 CI/CD Updates

**If using CI/CD:**

**Before:**
```yaml
# .github/workflows/api.yml
- name: Build API
  run: |
    cd cmd/api
    go build -o ligneous-api-server
```

**After:**
```yaml
# .github/workflows/api.yml (in API project)
- name: Build API
  run: |
    go build -o ligneous-api-server ./cmd/api
```

**Change:** Repository context changes, build command similar

### 6.3 Deployment Scripts

**Update deployment scripts to:**
1. Clone API repository (instead of monorepo)
2. Build from API project
3. Install binary (same location)
4. Restart service (same command)

---

## 7. Background Goroutine Considerations

### 7.1 Current Implementation

**How It Works:**
1. File upload completes
2. In-memory graph built (fast)
3. Response returned immediately
4. Background goroutine starts: `go s.saveGraphToHybridStorage(...)`
5. Goroutine persists graph to PostgreSQL + BadgerDB
6. Metadata updated when complete

**Benefits:**
- ✅ Fast upload response (20-70ms)
- ✅ Fast queries (in-memory graph ready)
- ✅ Persistent storage (background save)
- ✅ Non-blocking (doesn't slow upload)

### 7.2 After Split

**No Changes to Functionality:**
- ✅ Still runs as goroutine
- ✅ Still uses core library's `BuildGraphHybridPostgres()`
- ✅ Still updates metadata
- ✅ Still logs errors

**Code Organization:**
- ✅ Moves to `internal/storage/` in API project
- ✅ Better separation of concerns
- ✅ API owns persistence strategy

### 7.3 Potential Enhancements

**After Split, Could Add:**
1. **Job Queue:** Use Redis/PostgreSQL for background jobs
2. **Retry Logic:** Retry failed persistence attempts
3. **Progress Tracking:** Track persistence progress
4. **Webhooks:** Notify when persistence completes
5. **Monitoring:** Metrics for background jobs

**But:** Not required for split - current implementation works fine

---

## 8. Service Dependencies

### 8.1 Current Dependencies

**System Dependencies:**
- PostgreSQL (for graph persistence)
- Network (for API requests)
- Storage directory (for files)

**Service Dependencies:**
```ini
After=network.target postgresql.service
Requires=postgresql.service
```

### 8.2 After Split

**No Changes:**
- ✅ Still needs PostgreSQL
- ✅ Still needs network
- ✅ Still needs storage directory
- ✅ Same service dependencies

**Service File:** No changes needed

---

## 9. Testing After Split

### 9.1 Service Startup

```bash
# Stop old service
sudo systemctl stop ligneous-api

# Build new binary
cd /apps/ligneous-gedcom-api
go build -o /usr/local/bin/ligneous-api-server ./cmd/api

# Start service
sudo systemctl start ligneous-api

# Check status
sudo systemctl status ligneous-api
```

### 9.2 Background Goroutine Test

```bash
# Upload file (triggers background goroutine)
curl -X POST http://localhost:8090/api/v1/files \
  -F "file=@testdata/xavier.ged" \
  -F "name=xavier.ged"

# Check logs for background persistence
sudo journalctl -u ligneous-api -f | grep -i "hybrid\|persist\|background"
```

**Expected Log Output:**
```
In-memory graph built successfully in 45ms for file abc-123
[Background] Starting hybrid graph persistence for file abc-123
Hybrid graph built and persisted successfully in 1.2s for file abc-123
```

### 9.3 Verify Persistence

```bash
# Check if graph persisted
curl http://localhost:8090/api/v1/files/{file_id}

# Should show: "graph_persisted": true
```

---

## 10. Rollback Plan

### 10.1 If Issues Occur

**Option 1: Revert to Old Binary**
```bash
# Stop new service
sudo systemctl stop ligneous-api

# Restore old binary (if backed up)
sudo cp /backup/ligneous-api-server.old /usr/local/bin/ligneous-api-server

# Start service
sudo systemctl start ligneous-api
```

**Option 2: Revert to Old Repository**
```bash
# Build from old monorepo
cd /apps/gedcom-go
go build -o /usr/local/bin/ligneous-api-server ./cmd/api

# Restart service
sudo systemctl restart ligneous-api
```

### 10.2 Backup Before Migration

```bash
# Backup current binary
sudo cp /usr/local/bin/ligneous-api-server /backup/ligneous-api-server.old

# Backup service file
sudo cp /etc/systemd/system/ligneous-api.service /backup/ligneous-api.service.old
```

---

## 11. Summary of Changes

### 11.1 What Changes

| Component | Before | After | Change Required |
|-----------|--------|-------|-----------------|
| **Binary Source** | `cmd/api/main.go` (monorepo) | `cmd/api/main.go` (API project) | ✅ Build from new repo |
| **Binary Path** | `/usr/local/bin/ligneous-api-server` | `/usr/local/bin/ligneous-api-server` | ❌ No change |
| **Service File** | `/etc/systemd/system/ligneous-api.service` | Same | ❌ No change |
| **Background Code** | `api/files_background.go` | `internal/storage/graph_persistence.go` | ✅ Move code |
| **Environment Vars** | DATABASE_URL, STORAGE_DIR, PORT | Same | ❌ No change |
| **Dependencies** | PostgreSQL, network | Same | ❌ No change |

### 11.2 What Stays the Same

- ✅ Service file configuration
- ✅ Environment variables
- ✅ Service dependencies
- ✅ Binary location
- ✅ Background goroutine functionality
- ✅ Service management commands
- ✅ Logging (systemd journal)

### 11.3 Migration Checklist

- [ ] Build new binary from API project
- [ ] Test binary manually (`./ligneous-api-server`)
- [ ] Stop old service
- [ ] Replace binary (or build to same location)
- [ ] Start service
- [ ] Verify health endpoint
- [ ] Test file upload (triggers background goroutine)
- [ ] Verify background persistence in logs
- [ ] Test queries work correctly
- [ ] Monitor for 24 hours

---

## 12. Recommendations

### 12.1 Binary Naming

**Recommendation:** Keep same binary name (`ligneous-api-server`)

**Reasoning:**
- No service file changes needed
- Easier migration
- Less disruption
- Can rename later if desired

### 12.2 Background Goroutine

**Recommendation:** Move to `internal/storage/graph_persistence.go`

**Reasoning:**
- Better code organization
- Clear separation of concerns
- API owns persistence strategy
- Easier to enhance later

### 12.3 Service Configuration

**Recommendation:** No changes needed

**Reasoning:**
- Current configuration works
- Same dependencies
- Same environment variables
- No need to change

### 12.4 Deployment

**Recommendation:** Update build scripts, keep service same

**Reasoning:**
- Minimal changes
- Lower risk
- Easier rollback
- Standard practice

---

## 13. Conclusion

**The daemon/service setup requires minimal changes after the split:**

1. ✅ **Build Process:** Build from new API project (different directory)
2. ✅ **Background Goroutine:** Move code to API project (same functionality)
3. ❌ **Service File:** No changes needed (same binary path)
4. ❌ **Configuration:** No changes needed (same environment variables)
5. ❌ **Dependencies:** No changes needed (same requirements)

**The split is transparent to the systemd service** - it just runs a binary, which happens to be built from a different repository now.

**Background goroutine code moves** but functionality stays the same - it still uses core library's persistence functions, just from a different package location.

**Overall Impact:** **Low** - Mostly build process updates, service configuration unchanged

---

**Status:** Ready for Implementation  
**Risk Level:** Low  
**Migration Time:** ~30 minutes (build + test + deploy)

