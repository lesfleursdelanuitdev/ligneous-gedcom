package parser

import (
	"os"
	"testing"
)

func BenchmarkParse_SampleGed(b *testing.B) {
	sampleFile := "../../../family-tree/flask-backend/gedcom/sample.ged"
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		b.Skipf("Sample file not found: %s", sampleFile)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewHierarchicalParser()
		_, err := parser.Parse(sampleFile)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

func BenchmarkParse_GracisGed(b *testing.B) {
	file := "../../../family-tree/gedcom/gracis.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		b.Skipf("File not found: %s", file)
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

func BenchmarkParse_XavierGed(b *testing.B) {
	file := "../../../family-tree/gedcom/xavier.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		b.Skipf("File not found: %s", file)
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

func BenchmarkParse_Tree1Ged(b *testing.B) {
	file := "../../../family-tree/gedcom/tree1.ged"
	if _, err := os.Stat(file); os.IsNotExist(err) {
		b.Skipf("File not found: %s", file)
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



