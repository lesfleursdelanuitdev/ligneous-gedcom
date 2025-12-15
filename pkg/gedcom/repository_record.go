package gedcom

// RepositoryRecord represents a Repository (REPO) record.
type RepositoryRecord struct {
	*BaseRecord
}

// NewRepositoryRecord creates a new RepositoryRecord from a GedcomLine.
func NewRepositoryRecord(line *GedcomLine) *RepositoryRecord {
	return &RepositoryRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetName returns the repository name (NAME).
func (rr *RepositoryRecord) GetName() string {
	return rr.GetValue("NAME")
}

// GetAddress returns the address lines (ADDR).
func (rr *RepositoryRecord) GetAddress() []string {
	return rr.GetValues("ADDR")
}



