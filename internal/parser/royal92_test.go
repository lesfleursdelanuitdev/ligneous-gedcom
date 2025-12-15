package parser

import (
	"os"
	"testing"
)

func TestRoyal92_SequentialParser(t *testing.T) {
	file := "/apps/gedcom-go/testdata/royal92.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = "testdata/royal92.ged"
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Skipf("Test file not found. Download it first.")
		}
	}

	parser := NewHierarchicalParser()
	tree, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify basic structure
	individuals := tree.GetAllIndividuals()
	families := tree.GetAllFamilies()
	header := tree.GetHeader()

	if header == nil {
		t.Error("Expected header record")
	}

	t.Logf("Parsed royal92.ged:")
	t.Logf("  Individuals: %d", len(individuals))
	t.Logf("  Families: %d", len(families))
	t.Logf("  Encoding: %s", tree.GetEncoding())

	// Check for errors
	errors := parser.GetErrors()
	if len(errors) > 0 {
		t.Logf("  Warnings/Errors: %d", len(errors))
		// Log first few errors
		for i, err := range errors {
			if i >= 5 {
				break
			}
			t.Logf("    - %s", err.Error())
		}
	}

	// Verify some expected records
	// Queen Victoria should be @I1@
	vic := individuals["@I1@"]
	if vic == nil {
		t.Error("Expected @I1@ (Queen Victoria) not found")
	} else {
		if vic.GetValue("NAME") == "" {
			t.Error("@I1@ should have a NAME")
		}
		if vic.GetValue("SEX") != "F" {
			t.Errorf("@I1@ SEX should be F, got %s", vic.GetValue("SEX"))
		}
		t.Logf("  @I1@: %s, SEX: %s", vic.GetValue("NAME"), vic.GetValue("SEX"))
	}

	// Check for Prince Albert @I2@
	albert := individuals["@I2@"]
	if albert == nil {
		t.Error("Expected @I2@ (Prince Albert) not found")
	} else {
		if albert.GetValue("SEX") != "M" {
			t.Errorf("@I2@ SEX should be M, got %s", albert.GetValue("SEX"))
		}
		t.Logf("  @I2@: %s, SEX: %s", albert.GetValue("NAME"), albert.GetValue("SEX"))
	}

	// Check for family @F1@ (Victoria and Albert)
	fam1 := families["@F1@"]
	if fam1 == nil {
		t.Error("Expected @F1@ (Victoria & Albert family) not found")
	} else {
		t.Logf("  @F1@: HUSB=%s, WIFE=%s", fam1.GetValue("HUSB"), fam1.GetValue("WIFE"))
	}
}

func TestRoyal92_TwoPhaseParser(t *testing.T) {
	file := "testdata/royal92.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = "../../testdata/royal92.ged"
		if _, err := os.Stat(file); os.IsNotExist(err) {
			file = "/apps/gedcom-go/testdata/royal92.ged"
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Skipf("Test file not found. Download it first.")
			}
		}
	}

	parser := NewTwoPhaseParser()
	tree, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify basic structure
	individuals := tree.GetAllIndividuals()
	families := tree.GetAllFamilies()

	t.Logf("Two-phase parser results:")
	t.Logf("  Individuals: %d", len(individuals))
	t.Logf("  Families: %d", len(families))

	// Should match sequential parser
	if len(individuals) == 0 {
		t.Error("Expected individuals to be parsed")
	}
	if len(families) == 0 {
		t.Error("Expected families to be parsed")
	}

	// Check for errors
	errors := parser.GetErrors()
	if len(errors) > 0 {
		t.Logf("  Warnings/Errors: %d", len(errors))
	}
}

func TestRoyal92_CompareParsers(t *testing.T) {
	file := "testdata/royal92.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = "../../testdata/royal92.ged"
		if _, err := os.Stat(file); os.IsNotExist(err) {
			file = "/apps/gedcom-go/testdata/royal92.ged"
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Skipf("Test file not found. Download it first.")
			}
		}
	}

	// Parse with sequential
	seqParser := NewHierarchicalParser()
	seqTree, err := seqParser.Parse(file)
	if err != nil {
		t.Fatalf("Sequential parse failed: %v", err)
	}

	// Parse with two-phase
	twoPhaseParser := NewTwoPhaseParser()
	twoPhaseTree, err := twoPhaseParser.Parse(file)
	if err != nil {
		t.Fatalf("Two-phase parse failed: %v", err)
	}

	// Compare results
	seqIndis := seqTree.GetAllIndividuals()
	twoPhaseIndis := twoPhaseTree.GetAllIndividuals()

	seqFams := seqTree.GetAllFamilies()
	twoPhaseFams := twoPhaseTree.GetAllFamilies()

	if len(seqIndis) != len(twoPhaseIndis) {
		t.Errorf("Individual count mismatch: sequential=%d, two-phase=%d", len(seqIndis), len(twoPhaseIndis))
	}

	if len(seqFams) != len(twoPhaseFams) {
		t.Errorf("Family count mismatch: sequential=%d, two-phase=%d", len(seqFams), len(twoPhaseFams))
	}

	t.Logf("Comparison:")
	t.Logf("  Sequential: %d individuals, %d families", len(seqIndis), len(seqFams))
	t.Logf("  Two-Phase:  %d individuals, %d families", len(twoPhaseIndis), len(twoPhaseFams))
	t.Logf("  Sequential errors: %d", len(seqParser.GetErrors()))
	t.Logf("  Two-Phase errors:  %d", len(twoPhaseParser.GetErrors()))
}

func BenchmarkRoyal92_Sequential(b *testing.B) {
	file := "testdata/royal92.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = "/apps/gedcom-go/testdata/royal92.ged"
		if _, err := os.Stat(file); os.IsNotExist(err) {
			b.Skipf("Test file not found")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewHierarchicalParser()
		_, err := parser.Parse(file)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

func BenchmarkRoyal92_TwoPhase(b *testing.B) {
	file := "testdata/royal92.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = "/apps/gedcom-go/testdata/royal92.ged"
		if _, err := os.Stat(file); os.IsNotExist(err) {
			b.Skipf("Test file not found")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewTwoPhaseParser()
		_, err := parser.Parse(file)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

func BenchmarkRoyal92_Comparison(b *testing.B) {
	file := "testdata/royal92.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = "/apps/gedcom-go/testdata/royal92.ged"
		if _, err := os.Stat(file); os.IsNotExist(err) {
			b.Skipf("Test file not found")
		}
	}

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parser := NewHierarchicalParser()
			_, err := parser.Parse(file)
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
		}
	})

	b.Run("TwoPhase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			parser := NewTwoPhaseParser()
			_, err := parser.Parse(file)
			if err != nil {
				b.Fatalf("Failed to parse: %v", err)
			}
		}
	})
}

