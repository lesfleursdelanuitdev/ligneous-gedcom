// Package gedcom provides core data structures for GEDCOM files.
package gedcom

import "sync"

// GedcomTree represents the entire GEDCOM file structure.
// It serves as the root container for all parsed records and provides
// thread-safe access to individuals, families, notes, sources, and other
// record types.
//
// The tree maintains:
//   - Separate maps for each record type (individuals, families, etc.)
//   - A cross-reference index for fast lookups by xref ID
//   - Metadata such as encoding and version
//
// All methods are thread-safe and can be called concurrently.
type GedcomTree struct {
	mu sync.RWMutex

	// Records organized by type
	header       Record
	individuals  map[string]Record // key: xref_id
	families     map[string]Record
	notes        map[string]Record
	sources      map[string]Record
	repositories map[string]Record
	submitters   map[string]Record
	multimedia   map[string]Record

	// Cross-reference index (all records by xref_id)
	xrefIndex map[string]Record

	// Metadata
	encoding string
	version  string
}

// NewGedcomTree creates a new empty GedcomTree.
func NewGedcomTree() *GedcomTree {
	return &GedcomTree{
		individuals:  make(map[string]Record),
		families:     make(map[string]Record),
		notes:        make(map[string]Record),
		sources:      make(map[string]Record),
		repositories: make(map[string]Record),
		submitters:   make(map[string]Record),
		multimedia:   make(map[string]Record),
		xrefIndex:    make(map[string]Record),
	}
}

// AddRecord adds a record to the tree.
// For Step 1.5, we only handle level 0 records.
func (gt *GedcomTree) AddRecord(record Record) {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	recordType := record.Type()
	xrefID := record.XrefID()

	// Add to xref index if it has an xref
	if xrefID != "" {
		gt.xrefIndex[xrefID] = record
	}

	// Add to appropriate map based on type
	switch recordType {
	case RecordTypeHEAD:
		gt.header = record
	case RecordTypeINDI:
		if xrefID != "" {
			gt.individuals[xrefID] = record
		}
	case RecordTypeFAM:
		if xrefID != "" {
			gt.families[xrefID] = record
		}
	case RecordTypeNOTE:
		if xrefID != "" {
			gt.notes[xrefID] = record
		}
	case RecordTypeSOUR:
		if xrefID != "" {
			gt.sources[xrefID] = record
		}
	case RecordTypeREPO:
		if xrefID != "" {
			gt.repositories[xrefID] = record
		}
	case RecordTypeSUBM:
		if xrefID != "" {
			gt.submitters[xrefID] = record
		}
	case RecordTypeOBJE:
		if xrefID != "" {
			gt.multimedia[xrefID] = record
		}
	case RecordTypeTRLR:
		// TRLR doesn't need to be stored separately
		break
	}
}

// GetHeader returns the header record.
func (gt *GedcomTree) GetHeader() Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.header
}

// GetIndividual returns an individual record by xref ID.
func (gt *GedcomTree) GetIndividual(xrefID string) Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.individuals[xrefID]
}

// GetFamily returns a family record by xref ID.
func (gt *GedcomTree) GetFamily(xrefID string) Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.families[xrefID]
}

// GetAllIndividuals returns all individual records.
func (gt *GedcomTree) GetAllIndividuals() map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record)
	for k, v := range gt.individuals {
		result[k] = v
	}
	return result
}

// GetAllFamilies returns all family records.
func (gt *GedcomTree) GetAllFamilies() map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record)
	for k, v := range gt.families {
		result[k] = v
	}
	return result
}

// GetAllNotes returns all note records.
func (gt *GedcomTree) GetAllNotes() map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record)
	for k, v := range gt.notes {
		result[k] = v
	}
	return result
}

// GetAllSources returns all source records.
func (gt *GedcomTree) GetAllSources() map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record)
	for k, v := range gt.sources {
		result[k] = v
	}
	return result
}

// GetAllRepositories returns all repository records.
func (gt *GedcomTree) GetAllRepositories() map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record)
	for k, v := range gt.repositories {
		result[k] = v
	}
	return result
}

// GetAllSubmitters returns all submitter records.
func (gt *GedcomTree) GetAllSubmitters() map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record)
	for k, v := range gt.submitters {
		result[k] = v
	}
	return result
}

// GetAllMultimedia returns all multimedia records.
func (gt *GedcomTree) GetAllMultimedia() map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record)
	for k, v := range gt.multimedia {
		result[k] = v
	}
	return result
}

// GetRecordByXref returns any record by its xref ID.
func (gt *GedcomTree) GetRecordByXref(xrefID string) Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.xrefIndex[xrefID]
}

// SetEncoding sets the file encoding.
func (gt *GedcomTree) SetEncoding(encoding string) {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	gt.encoding = encoding
}

// GetEncoding returns the file encoding.
func (gt *GedcomTree) GetEncoding() string {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.encoding
}

// SetVersion sets the GEDCOM version.
func (gt *GedcomTree) SetVersion(version string) {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	gt.version = version
}

// GetVersion returns the GEDCOM version.
func (gt *GedcomTree) GetVersion() string {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.version
}
