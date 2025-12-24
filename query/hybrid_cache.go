package query

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// HybridCache provides LRU caching for hybrid storage operations
type HybridCache struct {
	// Node cache: nodeID -> GraphNode
	nodeCache *lru.Cache[uint32, GraphNode]

	// XREF cache: xref -> nodeID (bidirectional)
	xrefToIDCache *lru.Cache[string, uint32]
	idToXrefCache *lru.Cache[uint32, string]

	// Query result cache: cacheKey -> []uint32 (node IDs)
	queryCache *lru.Cache[string, []uint32]

	mu sync.RWMutex
}

// NewHybridCache creates a new hybrid cache with specified sizes
func NewHybridCache(nodeCacheSize, xrefCacheSize, queryCacheSize int) (*HybridCache, error) {
	nodeCache, err := lru.New[uint32, GraphNode](nodeCacheSize)
	if err != nil {
		return nil, err
	}

	xrefToIDCache, err := lru.New[string, uint32](xrefCacheSize)
	if err != nil {
		return nil, err
	}

	idToXrefCache, err := lru.New[uint32, string](xrefCacheSize)
	if err != nil {
		return nil, err
	}

	queryCache, err := lru.New[string, []uint32](queryCacheSize)
	if err != nil {
		return nil, err
	}

	return &HybridCache{
		nodeCache:     nodeCache,
		xrefToIDCache: xrefToIDCache,
		idToXrefCache: idToXrefCache,
		queryCache:    queryCache,
	}, nil
}

// GetNode retrieves a node from cache
func (hc *HybridCache) GetNode(nodeID uint32) (GraphNode, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.nodeCache.Get(nodeID)
}

// SetNode stores a node in cache
func (hc *HybridCache) SetNode(nodeID uint32, node GraphNode) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.nodeCache.Add(nodeID, node)
}

// GetXrefToID retrieves node ID from XREF cache
func (hc *HybridCache) GetXrefToID(xref string) (uint32, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.xrefToIDCache.Get(xref)
}

// SetXrefToID stores XREF -> nodeID mapping
func (hc *HybridCache) SetXrefToID(xref string, nodeID uint32) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.xrefToIDCache.Add(xref, nodeID)
	hc.idToXrefCache.Add(nodeID, xref)
}

// GetIDToXref retrieves XREF from node ID cache
func (hc *HybridCache) GetIDToXref(nodeID uint32) (string, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.idToXrefCache.Get(nodeID)
}

// GetQuery retrieves query results from cache
func (hc *HybridCache) GetQuery(key string) ([]uint32, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.queryCache.Get(key)
}

// SetQuery stores query results in cache
func (hc *HybridCache) SetQuery(key string, nodeIDs []uint32) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	// Make a copy to avoid external modification
	ids := make([]uint32, len(nodeIDs))
	copy(ids, nodeIDs)
	hc.queryCache.Add(key, ids)
}

// Clear clears all caches
func (hc *HybridCache) Clear() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.nodeCache.Purge()
	hc.xrefToIDCache.Purge()
	hc.idToXrefCache.Purge()
	hc.queryCache.Purge()
}

// Stats returns cache statistics
func (hc *HybridCache) Stats() CacheStats {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return CacheStats{
		NodeCacheSize:      hc.nodeCache.Len(),
		NodeCacheCapacity:  hc.nodeCache.Len(), // LRU doesn't expose capacity, use current size
		XrefCacheSize:      hc.xrefToIDCache.Len(),
		XrefCacheCapacity:  hc.xrefToIDCache.Len(),
		QueryCacheSize:     hc.queryCache.Len(),
		QueryCacheCapacity: hc.queryCache.Len(),
	}
}

// CacheStats provides statistics about cache usage
type CacheStats struct {
	NodeCacheSize      int
	NodeCacheCapacity  int
	XrefCacheSize      int
	XrefCacheCapacity  int
	QueryCacheSize     int
	QueryCacheCapacity int
}

