package gedcom

import (
	"testing"
)

func TestRecordFactory_CreateRecord(t *testing.T) {
	factory := NewRecordFactory()

	tests := []struct {
		name     string
		tag      string
		xrefID   string
		wantType RecordType
	}{
		{"INDI", "INDI", "@I1@", RecordTypeINDI},
		{"FAM", "FAM", "@F1@", RecordTypeFAM},
		{"HEAD", "HEAD", "", RecordTypeHEAD},
		{"NOTE", "NOTE", "@N1@", RecordTypeNOTE},
		{"SOUR", "SOUR", "@S1@", RecordTypeSOUR},
		{"REPO", "REPO", "@R1@", RecordTypeREPO},
		{"SUBM", "SUBM", "@U1@", RecordTypeSUBM},
		{"OBJE", "OBJE", "@O1@", RecordTypeOBJE},
		{"TRLR", "TRLR", "", RecordTypeTRLR},
		{"Unknown", "UNKNOWN", "", RecordType("UNKNOWN")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewGedcomLine(0, tt.tag, "", tt.xrefID)
			record := factory.CreateRecord(line)

			if record == nil {
				t.Fatal("Expected record to be created")
			}

			if record.Type() != tt.wantType {
				t.Errorf("Expected type %v, got %v", tt.wantType, record.Type())
			}

			if record.XrefID() != tt.xrefID {
				t.Errorf("Expected xref %q, got %q", tt.xrefID, record.XrefID())
			}

			// Verify specialized types
			switch tt.wantType {
			case RecordTypeINDI:
				if _, ok := record.(*IndividualRecord); !ok {
					t.Error("Expected IndividualRecord type")
				}
			case RecordTypeFAM:
				if _, ok := record.(*FamilyRecord); !ok {
					t.Error("Expected FamilyRecord type")
				}
			case RecordTypeHEAD:
				if _, ok := record.(*HeaderRecord); !ok {
					t.Error("Expected HeaderRecord type")
				}
			case RecordTypeNOTE:
				if _, ok := record.(*NoteRecord); !ok {
					t.Error("Expected NoteRecord type")
				}
			case RecordTypeSOUR:
				if _, ok := record.(*SourceRecord); !ok {
					t.Error("Expected SourceRecord type")
				}
			case RecordTypeREPO:
				if _, ok := record.(*RepositoryRecord); !ok {
					t.Error("Expected RepositoryRecord type")
				}
			case RecordTypeSUBM:
				if _, ok := record.(*SubmitterRecord); !ok {
					t.Error("Expected SubmitterRecord type")
				}
			case RecordTypeOBJE:
				if _, ok := record.(*MultimediaRecord); !ok {
					t.Error("Expected MultimediaRecord type")
				}
			}
		})
	}
}

func TestRecordFactory_CreateRecord_NilLine(t *testing.T) {
	factory := NewRecordFactory()
	record := factory.CreateRecord(nil)
	if record != nil {
		t.Error("Expected nil record for nil line")
	}
}



