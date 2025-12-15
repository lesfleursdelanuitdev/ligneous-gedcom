package gedcom

import "testing"

func TestNewGedcomTree(t *testing.T) {
	tree := NewGedcomTree()
	if tree == nil {
		t.Fatal("NewGedcomTree returned nil")
	}
	if tree.individuals == nil {
		t.Error("individuals map should be initialized")
	}
	if tree.families == nil {
		t.Error("families map should be initialized")
	}
	if tree.xrefIndex == nil {
		t.Error("xrefIndex map should be initialized")
	}
}

func TestGedcomTree_AddRecord(t *testing.T) {
	tree := NewGedcomTree()

	// Add header
	headerLine := NewGedcomLine(0, "HEAD", "", "")
	header := NewHeaderRecord(headerLine)
	tree.AddRecord(header)

	if tree.GetHeader() != header {
		t.Error("Header not stored correctly")
	}

	// Add individual
	indiLine := NewGedcomLine(0, "INDI", "", "@I1@")
	indi := NewIndividualRecord(indiLine)
	tree.AddRecord(indi)

	if tree.GetIndividual("@I1@") != indi {
		t.Error("Individual not stored correctly")
	}

	// Add family
	famLine := NewGedcomLine(0, "FAM", "", "@F1@")
	fam := NewFamilyRecord(famLine)
	tree.AddRecord(fam)

	if tree.GetFamily("@F1@") != fam {
		t.Error("Family not stored correctly")
	}

	// Add note
	noteLine := NewGedcomLine(0, "NOTE", "Test note", "@N1@")
	note := NewNoteRecord(noteLine)
	tree.AddRecord(note)

	notes := tree.GetAllNotes()
	if len(notes) != 1 || notes["@N1@"] != note {
		t.Error("Note not stored correctly")
	}

	// Add source
	sourceLine := NewGedcomLine(0, "SOUR", "", "@S1@")
	source := NewSourceRecord(sourceLine)
	tree.AddRecord(source)

	sources := tree.GetAllSources()
	if len(sources) != 1 || sources["@S1@"] != source {
		t.Error("Source not stored correctly")
	}

	// Add repository
	repoLine := NewGedcomLine(0, "REPO", "", "@R1@")
	repo := NewRepositoryRecord(repoLine)
	tree.AddRecord(repo)

	repos := tree.GetAllRepositories()
	if len(repos) != 1 || repos["@R1@"] != repo {
		t.Error("Repository not stored correctly")
	}

	// Add submitter
	submLine := NewGedcomLine(0, "SUBM", "", "@U1@")
	subm := NewSubmitterRecord(submLine)
	tree.AddRecord(subm)

	submitters := tree.GetAllSubmitters()
	if len(submitters) != 1 || submitters["@U1@"] != subm {
		t.Error("Submitter not stored correctly")
	}

	// Add multimedia
	multLine := NewGedcomLine(0, "OBJE", "", "@M1@")
	mult := NewMultimediaRecord(multLine)
	tree.AddRecord(mult)

	multimedia := tree.GetAllMultimedia()
	if len(multimedia) != 1 || multimedia["@M1@"] != mult {
		t.Error("Multimedia not stored correctly")
	}

	// Test xref index
	record := tree.GetRecordByXref("@I1@")
	if record != indi {
		t.Error("Xref index not working correctly")
	}

	// Test TRLR (should not be stored)
	trlrLine := NewGedcomLine(0, "TRLR", "", "")
	trlr := NewBaseRecord(trlrLine)
	tree.AddRecord(trlr)
	// TRLR should not crash, but also shouldn't be stored
}

func TestGedcomTree_GetAllMethods(t *testing.T) {
	tree := NewGedcomTree()

	// Add multiple individuals
	for i := 1; i <= 3; i++ {
		xrefID := "@I" + string(rune('0'+i)) + "@"
		line := NewGedcomLine(0, "INDI", "", xrefID)
		indi := NewIndividualRecord(line)
		tree.AddRecord(indi)
	}

	individuals := tree.GetAllIndividuals()
	if len(individuals) != 3 {
		t.Errorf("Expected 3 individuals, got %d", len(individuals))
	}

	// Add multiple families
	for i := 1; i <= 2; i++ {
		xrefID := "@F" + string(rune('0'+i)) + "@"
		line := NewGedcomLine(0, "FAM", "", xrefID)
		fam := NewFamilyRecord(line)
		tree.AddRecord(fam)
	}

	families := tree.GetAllFamilies()
	if len(families) != 2 {
		t.Errorf("Expected 2 families, got %d", len(families))
	}
}

func TestGedcomTree_EncodingAndVersion(t *testing.T) {
	tree := NewGedcomTree()

	// Test encoding
	tree.SetEncoding("UTF-8")
	if tree.GetEncoding() != "UTF-8" {
		t.Errorf("Expected encoding 'UTF-8', got %q", tree.GetEncoding())
	}

	tree.SetEncoding("ANSEL")
	if tree.GetEncoding() != "ANSEL" {
		t.Errorf("Expected encoding 'ANSEL', got %q", tree.GetEncoding())
	}

	// Test version
	tree.SetVersion("5.5.5")
	if tree.GetVersion() != "5.5.5" {
		t.Errorf("Expected version '5.5.5', got %q", tree.GetVersion())
	}

	tree.SetVersion("5.5.1")
	if tree.GetVersion() != "5.5.1" {
		t.Errorf("Expected version '5.5.1', got %q", tree.GetVersion())
	}
}

func TestGedcomTree_GetRecordByXref(t *testing.T) {
	tree := NewGedcomTree()

	// Add record with xref
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	indi := NewIndividualRecord(line)
	tree.AddRecord(indi)

	// Get by xref
	record := tree.GetRecordByXref("@I1@")
	if record != indi {
		t.Error("GetRecordByXref failed")
	}

	// Non-existent xref
	record2 := tree.GetRecordByXref("@I999@")
	if record2 != nil {
		t.Error("Expected nil for non-existent xref")
	}
}

func TestGedcomTree_ThreadSafety(t *testing.T) {
	tree := NewGedcomTree()

	// Test that methods are thread-safe (they use locks)
	// This is a basic test - full concurrency testing would require more setup
	line := NewGedcomLine(0, "INDI", "", "@I1@")
	indi := NewIndividualRecord(line)
	tree.AddRecord(indi)

	// These should not panic
	_ = tree.GetHeader()
	_ = tree.GetIndividual("@I1@")
	_ = tree.GetAllIndividuals()
	_ = tree.GetEncoding()
	_ = tree.GetVersion()
}

