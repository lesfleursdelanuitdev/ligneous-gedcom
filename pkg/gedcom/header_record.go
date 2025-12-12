package gedcom

// HeaderRecord represents a Header (HEAD) record with metadata methods.
type HeaderRecord struct {
	*BaseRecord
}

// NewHeaderRecord creates a new HeaderRecord from a GedcomLine.
func NewHeaderRecord(line *GedcomLine) *HeaderRecord {
	return &HeaderRecord{
		BaseRecord: NewBaseRecord(line),
	}
}

// GetGedcomVersion returns the GEDCOM version (GEDC.VERS).
func (hr *HeaderRecord) GetGedcomVersion() string {
	return hr.GetValue("GEDC.VERS")
}

// GetGedcomForm returns the GEDCOM form (GEDC.FORM).
func (hr *HeaderRecord) GetGedcomForm() string {
	return hr.GetValue("GEDC.FORM")
}

// GetCharacterEncoding returns the character encoding (CHAR).
func (hr *HeaderRecord) GetCharacterEncoding() string {
	return hr.GetValue("CHAR")
}

// GetSourceName returns the source name (SOUR.NAME).
func (hr *HeaderRecord) GetSourceName() string {
	return hr.GetValue("SOUR.NAME")
}

// GetSourceVersion returns the source version (SOUR.VERS).
func (hr *HeaderRecord) GetSourceVersion() string {
	return hr.GetValue("SOUR.VERS")
}

// GetSourceCorporation returns the source corporation (SOUR.CORP).
func (hr *HeaderRecord) GetSourceCorporation() string {
	return hr.GetValue("SOUR.CORP")
}

// GetSubmissionXref returns the submitter xref (SUBM).
func (hr *HeaderRecord) GetSubmissionXref() string {
	return hr.GetValue("SUBM")
}

// GetFile returns the file name (FILE).
func (hr *HeaderRecord) GetFile() string {
	return hr.GetValue("FILE")
}

// GetLanguage returns the language (LANG).
func (hr *HeaderRecord) GetLanguage() string {
	return hr.GetValue("LANG")
}

// GetDate returns the header date (DATE).
func (hr *HeaderRecord) GetDate() string {
	return hr.GetValue("DATE")
}

