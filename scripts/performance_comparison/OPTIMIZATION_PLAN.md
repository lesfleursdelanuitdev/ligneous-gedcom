# Optimization Plan Based on Test Results

## Test Results Analysis

### Performance by File Size

| File Size | Files | ParallelHierarchicalParser Performance |
|-----------|-------|----------------------------------------|
| **Large (400-500 KB)** | royal92 variants | ✅ **4.9-6.4% faster** |
| **Very Large (1.1 MB)** | pres2020 | ⚠️ **5.2% slower** |
| **Medium (4.6 KB)** | comprehensive | ⚠️ **96% slower** |
| **Small (< 1 KB)** | minimal files | ⚠️ **8-9x slower** |

### Key Findings

1. **Small files are the biggest problem:** 8-9x slower for files < 1KB
2. **Medium files also struggle:** 96% slower for 4.6KB files
3. **Large files perform well:** 4.9-6.4% faster for 400-500KB files
4. **Very large files need optimization:** 5.2% slower for 1.1MB files

## Root Causes

### 1. ParseLine Performance (Critical - Small Files)
- Uses `strings.SplitN` multiple times
- Creates multiple string slices per line
- For small files, this overhead dominates

### 2. Parser Overhead (Critical - Small Files)
- ParallelHierarchicalParser has goroutine overhead
- Channel operations add latency
- For small files, overhead > benefit

### 3. Factory Creation (Medium Impact)
- Creates new RecordFactory for each record
- Unnecessary allocations

### 4. String Operations (Medium Impact)
- Redundant TrimSpace calls
- String concatenation with `+=`

## Optimization Strategy

### Phase 1: Quick Wins (High Impact, Low Risk)

1. **Use ParseLineFast** (Expected: 2-3x improvement for small files)
   - Replace `ParseLine` with `ParseLineFast` in all parsers
   - Remove redundant TrimSpace from ParseLineFast callers

2. **Reuse RecordFactory** (Expected: 5-10% improvement)
   - Create factory once per parser instance
   - Reuse for all records

3. **Remove Redundant TrimSpace** (Expected: 5-10% improvement)
   - Line is trimmed in Parse() and ParseLine()
   - Remove from ParseLineFast (already assumes trimmed)

### Phase 2: String Optimization (Medium Impact)

4. **Use strings.Builder for CONC/CONT** (Expected: 10-20% for files with continuations)
   - Replace `+=` concatenation
   - Pre-allocate builder capacity

### Phase 3: Parser Selection (High Impact for Small Files)

5. **Smart Parser Selection** (Expected: 8-9x improvement for small files)
   - Use HierarchicalParser for files < 10KB
   - Use ParallelHierarchicalParser for files >= 10KB
   - Auto-detect file size and choose parser

### Phase 4: Advanced Optimizations

6. **Optimize Scanner Buffer** (Expected: 5-10% for large files)
   - Use larger buffer (256KB-1MB) for large files

7. **Profile Stack Operations** (Variable impact)
   - Check FindParent performance
   - Optimize if needed

## Implementation Priority

### Priority 1: ParseLineFast (Critical)
- **Impact:** 2-3x improvement, especially for small files
- **Risk:** Low (already created, just needs integration)
- **Effort:** Low (replace ParseLine calls)

### Priority 2: Smart Parser Selection (Critical for Small Files)
- **Impact:** 8-9x improvement for small files
- **Risk:** Low (just add file size check)
- **Effort:** Low

### Priority 3: Reuse RecordFactory (High Value)
- **Impact:** 5-10% improvement
- **Risk:** Low
- **Effort:** Low

### Priority 4: Remove Redundant TrimSpace (Easy Win)
- **Impact:** 5-10% improvement
- **Risk:** Very Low
- **Effort:** Very Low

### Priority 5: strings.Builder for CONC/CONT (Medium Value)
- **Impact:** 10-20% for files with continuations
- **Risk:** Low
- **Effort:** Medium

## Expected Results After Optimization

| File Size | Current | After Optimization | Improvement |
|-----------|---------|-------------------|-------------|
| Small (< 1 KB) | 8-9x slower | 1-2x slower | **4-5x improvement** |
| Medium (4-5 KB) | 96% slower | 20-30% slower | **3-4x improvement** |
| Large (400-500 KB) | 4.9-6.4% faster | 10-15% faster | **Maintain + improve** |
| Very Large (1.1 MB) | 5.2% slower | 5-10% faster | **10-15% improvement** |

## Success Criteria

- Small files (< 1KB): Within 2x of cacack parser
- Medium files (4-10KB): Within 1.5x of cacack parser
- Large files (400KB+): Maintain or improve current advantage
- Very large files (1MB+): Match or exceed cacack parser

