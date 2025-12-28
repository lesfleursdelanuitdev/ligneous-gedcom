package parser

import (
	"testing"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// TestIntegration_SpecializedRecords verifies that specialized record types are created
func TestIntegration_SpecializedRecords(t *testing.T) {
	sampleFile := findTestDataFile("royal92.ged")
	if sampleFile == "" {
		t.Skipf("Test file not found (tried royal92.ged)")
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(sampleFile)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify HEAD is HeaderRecord
	header := tree.GetHeader()
	if header == nil {
		t.Fatal("Expected HEAD record")
	}
	if _, ok := header.(*types.HeaderRecord); !ok {
		t.Error("Expected HEAD to be HeaderRecord type")
	}
	headerRecord := header.(*types.HeaderRecord)
	version := headerRecord.GetGedcomVersion()
	if version != "" {
		t.Logf("GEDCOM version: %s", version)
	}

	// Verify INDI is IndividualRecord
	allIndis := tree.GetAllIndividuals()
	if len(allIndis) == 0 {
		t.Fatal("Expected at least one individual")
	}
	
	// Check first individual
	var indi1 types.Record
	for _, indi := range allIndis {
		indi1 = indi
		break
	}
	
	if _, ok := indi1.(*types.IndividualRecord); !ok {
		t.Error("Expected INDI to be IndividualRecord type")
	} else {
		indiRecord := indi1.(*types.IndividualRecord)
		name := indiRecord.GetName()
		if name != "" {
			t.Logf("Individual name: %s", name)
		}
		birthDate := indiRecord.GetBirthDate()
		if birthDate != "" {
			t.Logf("Individual birth date: %s", birthDate)
		}
	}

	// Verify FAM is FamilyRecord
	allFams := tree.GetAllFamilies()
	if len(allFams) > 0 {
		var fam1 types.Record
		for _, fam := range allFams {
			fam1 = fam
			break
		}
		
		if _, ok := fam1.(*types.FamilyRecord); !ok {
			t.Error("Expected FAM to be FamilyRecord type")
		} else {
			famRecord := fam1.(*types.FamilyRecord)
			husband := famRecord.GetHusband()
			if husband != "" {
				t.Logf("Family husband: %s", husband)
			}
			marriageDate := famRecord.GetMarriageDate()
			if marriageDate != "" {
				t.Logf("Family marriage date: %s", marriageDate)
			}
		}
	}

	// Verify SOUR is SourceRecord (if present and is actually a source)
	// Note: @S1@ might be a SubmitterRecord in some files, so check the actual type
	sour1 := tree.GetRecordByXref("@S1@")
	if sour1 != nil {
		if sourceRecord, ok := sour1.(*types.SourceRecord); ok {
			title := sourceRecord.GetTitle()
			t.Logf("Source @S1@ title: %s", title)
		} else if submitterRecord, ok := sour1.(*types.SubmitterRecord); ok {
			// @S1@ might be a SubmitterRecord in some files
			name := submitterRecord.GetName()
			t.Logf("Submitter @S1@ name: %s", name)
		} else {
			// Just log the type - don't fail
			t.Logf("Record @S1@ is of type %T (not SourceRecord or SubmitterRecord)", sour1)
		}
	}

	// Verify REPO is RepositoryRecord (if present)
	repo1 := tree.GetRecordByXref("@R1@")
	if repo1 != nil {
		if _, ok := repo1.(*types.RepositoryRecord); !ok {
			t.Error("Expected REPO to be RepositoryRecord type")
		} else {
			repoRecord := repo1.(*types.RepositoryRecord)
			name := repoRecord.GetName()
			t.Logf("Repository @R1@ name: %s", name)
		}
	}

	// Verify SUBM is SubmitterRecord (if present)
	subm1 := tree.GetRecordByXref("@U1@")
	if subm1 != nil {
		if _, ok := subm1.(*types.SubmitterRecord); !ok {
			t.Error("Expected SUBM to be SubmitterRecord type")
		} else {
			submRecord := subm1.(*types.SubmitterRecord)
			name := submRecord.GetName()
			t.Logf("Submitter @U1@ name: %s", name)
		}
	}

	// Verify we can find at least one of each record type
	allSources := tree.GetAllSources()
	if len(allSources) > 0 {
		t.Logf("Found %d source records", len(allSources))
		for xref, source := range allSources {
			if _, ok := source.(*types.SourceRecord); ok {
				t.Logf("Source %s is SourceRecord type", xref)
				break
			}
		}
	}

	allSubmitters := tree.GetAllSubmitters()
	if len(allSubmitters) > 0 {
		t.Logf("Found %d submitter records", len(allSubmitters))
		for xref, subm := range allSubmitters {
			if _, ok := subm.(*types.SubmitterRecord); ok {
				t.Logf("Submitter %s is SubmitterRecord type", xref)
				break
			}
		}
	}
}



