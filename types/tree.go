// Package gedcom provides core data structures for GEDCOM files.
package types

import (
	"sync"
)

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

	// UUID index (all records by UUID for fast lookup)
	uuidIndex map[string]Record

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
		uuidIndex:    make(map[string]Record),
	}
}

// AddRecord adds a record to the tree.
// For Step 1.5, we only handle level 0 records.
func (gt *GedcomTree) AddRecord(record Record) {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	recordType := record.Type()
	xrefID := record.XrefID()

	// Set tree reference on the record (if it's a BaseRecord)
	if br, ok := record.(interface{ setTree(*GedcomTree) }); ok {
		br.setTree(gt)
	}

	// Add to xref index if it has an xref
	if xrefID != "" {
		gt.xrefIndex[xrefID] = record
	}

	// Add to UUID index (all records have UUIDs)
	gt.uuidIndex[record.UUID()] = record

	// Add to appropriate map based on type
	// Use a helper to reduce code duplication
	gt.addToTypeMap(recordType, xrefID, record)
}

// addToTypeMap adds a record to the appropriate type-specific map.
// This helper reduces code duplication in AddRecord.
func (gt *GedcomTree) addToTypeMap(recordType RecordType, xrefID string, record Record) {
	switch recordType {
	case RecordTypeHEAD:
		gt.header = record
	case RecordTypeTRLR:
		// TRLR doesn't need to be stored separately
		return
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

// getAllRecords is a helper that safely copies a record map.
// Used by all GetAll* methods to reduce code duplication.
func (gt *GedcomTree) getAllRecords(source map[string]Record) map[string]Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	result := make(map[string]Record, len(source))
	for k, v := range source {
		result[k] = v
	}
	return result
}

// GetAllIndividuals returns all individual records.
func (gt *GedcomTree) GetAllIndividuals() map[string]Record {
	return gt.getAllRecords(gt.individuals)
}

// GetAllFamilies returns all family records.
func (gt *GedcomTree) GetAllFamilies() map[string]Record {
	return gt.getAllRecords(gt.families)
}

// GetAllNotes returns all note records.
func (gt *GedcomTree) GetAllNotes() map[string]Record {
	return gt.getAllRecords(gt.notes)
}

// GetAllSources returns all source records.
func (gt *GedcomTree) GetAllSources() map[string]Record {
	return gt.getAllRecords(gt.sources)
}

// GetAllRepositories returns all repository records.
func (gt *GedcomTree) GetAllRepositories() map[string]Record {
	return gt.getAllRecords(gt.repositories)
}

// GetAllSubmitters returns all submitter records.
func (gt *GedcomTree) GetAllSubmitters() map[string]Record {
	return gt.getAllRecords(gt.submitters)
}

// GetAllMultimedia returns all multimedia records.
func (gt *GedcomTree) GetAllMultimedia() map[string]Record {
	return gt.getAllRecords(gt.multimedia)
}

// GetRecordByXref returns any record by its xref ID.
func (gt *GedcomTree) GetRecordByXref(xrefID string) Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.xrefIndex[xrefID]
}

// GetRecordByUUID returns any record by its system-generated UUID.
func (gt *GedcomTree) GetRecordByUUID(uuid string) Record {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.uuidIndex[uuid]
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

