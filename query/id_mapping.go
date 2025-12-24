package query

// getOrCreateID gets the uint32 ID for an XREF string, creating it if it doesn't exist.
// Must be called with lock held.
func (g *Graph) getOrCreateID(xrefID string) uint32 {
	if id, exists := g.xrefToID[xrefID]; exists {
		return id
	}
	// Create new ID
	id := g.nextID
	g.nextID++
	g.xrefToID[xrefID] = id
	g.idToXref[id] = xrefID
	return id
}

// getID gets the uint32 ID for an XREF string, returning 0 if not found.
// Must be called with lock held.
func (g *Graph) getID(xrefID string) uint32 {
	return g.xrefToID[xrefID]
}

// getXref gets the XREF string for a uint32 ID, returning empty string if not found.
// Must be called with lock held.
func (g *Graph) getXref(id uint32) string {
	return g.idToXref[id]
}

// GetNodeByID returns a node by uint32 ID (internal use).
func (g *Graph) GetNodeByID(id uint32) GraphNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.nodes[id]
}

// GetNodeID returns the uint32 ID for an XREF string, or 0 if not found.
func (g *Graph) GetNodeID(xrefID string) uint32 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.xrefToID[xrefID]
}

// GetXrefFromID returns the XREF string for a uint32 ID, or empty string if not found.
func (g *Graph) GetXrefFromID(id uint32) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.idToXref[id]
}

