package gedcom

// SourceRecord represents a Source (SOUR) record.
type SourceRecord struct {
	*BaseRecord
}

// NewSourceRecord creates a new SourceRecord from a GedcomLine.
func NewSourceRecord(line *GedcomLine) *SourceRecord {
	return &SourceRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetTitle returns the source title (TITL).
func (sr *SourceRecord) GetTitle() string {
	return sr.GetValue("TITL")
}

// GetAbbreviation returns the source abbreviation (ABBR).
func (sr *SourceRecord) GetAbbreviation() string {
	return sr.GetValue("ABBR")
}

// GetRepository returns the repository xref (REPO).
func (sr *SourceRecord) GetRepository() string {
	return sr.GetValue("REPO")
}

