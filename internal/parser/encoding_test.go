package parser

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectEncoding(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		fileData []byte
		want     Encoding
		wantErr  bool
	}{
		{
			name:     "UTF-8 with BOM",
			fileData: []byte{0xEF, 0xBB, 0xBF, '0', ' ', 'H', 'E', 'A', 'D'},
			want:     EncodingUTF8,
			wantErr:  false,
		},
		{
			name:     "UTF-8 without BOM",
			fileData: []byte{'0', ' ', 'H', 'E', 'A', 'D'},
			want:     EncodingUTF8,
			wantErr:  false,
		},
		{
			name:     "UTF-16 BE with BOM",
			fileData: []byte{0xFE, 0xFF, 0x00, '0', 0x00, ' ', 0x00, 'H'},
			want:     EncodingUTF16,
			wantErr:  false,
		},
		{
			name:     "UTF-16 LE with BOM",
			fileData: []byte{0xFF, 0xFE, '0', 0x00, ' ', 0x00, 'H', 0x00},
			want:     EncodingUTF16,
			wantErr:  false,
		},
		{
			name:     "empty file defaults to UTF-8",
			fileData: []byte{},
			want:     EncodingUTF8,
			wantErr:  false,
		},
		{
			name:     "single byte defaults to UTF-8",
			fileData: []byte{'0'},
			want:     EncodingUTF8,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile := filepath.Join(tmpDir, "test_"+tt.name+".ged")
			err := os.WriteFile(tmpFile, tt.fileData, 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			got, err := DetectEncoding(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectEncoding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectEncoding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectEncoding_FileErrors(t *testing.T) {
	tests := []struct {
		name    string
		filePath string
		wantErr bool
	}{
		{
			name:    "non-existent file",
			filePath: "/nonexistent/file.ged",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DetectEncoding(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectEncoding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasBOM(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "UTF-8 BOM",
			data: []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'},
			want: true,
		},
		{
			name: "UTF-16 BE BOM",
			data: []byte{0xFE, 0xFF, 0x00, 'H'},
			want: true,
		},
		{
			name: "UTF-16 LE BOM",
			data: []byte{0xFF, 0xFE, 'H', 0x00},
			want: true,
		},
		{
			name: "no BOM",
			data: []byte{'H', 'e', 'l', 'l', 'o'},
			want: false,
		},
		{
			name: "empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "single byte",
			data: []byte{'H'},
			want: false,
		},
		{
			name: "partial UTF-8 BOM",
			data: []byte{0xEF, 0xBB},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasBOM(tt.data)
			if got != tt.want {
				t.Errorf("HasBOM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBOMType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "UTF-8 BOM",
			data: []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'},
			want: "UTF-8",
		},
		{
			name: "UTF-16 BE BOM",
			data: []byte{0xFE, 0xFF, 0x00, 'H'},
			want: "UTF-16-BE",
		},
		{
			name: "UTF-16 LE BOM",
			data: []byte{0xFF, 0xFE, 'H', 0x00},
			want: "UTF-16-LE",
		},
		{
			name: "no BOM",
			data: []byte{'H', 'e', 'l', 'l', 'o'},
			want: "",
		},
		{
			name: "empty data",
			data: []byte{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBOMType(tt.data)
			if got != tt.want {
				t.Errorf("GetBOMType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeEncoding(t *testing.T) {
	tests := []struct {
		name string
		enc  Encoding
		want Encoding
	}{
		{"UTF-8 uppercase", "UTF-8", EncodingUTF8},
		{"UTF-8 lowercase", "utf-8", EncodingUTF8},
		{"UTF-8 mixed", "Utf-8", EncodingUTF8},
		{"UTF-8 no dash", "UTF8", EncodingUTF8},
		{"UTF-16 uppercase", "UTF-16", EncodingUTF16},
		{"UTF-16 lowercase", "utf-16", EncodingUTF16},
		{"UNICODE", "UNICODE", EncodingUTF16},
		{"unicode lowercase", "unicode", EncodingUTF16},
		{"ANSEL", "ANSEL", EncodingANSEL},
		{"ansel lowercase", "ansel", EncodingANSEL},
		{"ASCII", "ASCII", EncodingASCII},
		{"ANSI", "ANSI", EncodingANSI},
		{"Windows-1252", "Windows-1252", EncodingANSI},
		{"unknown encoding", "UNKNOWN", Encoding("UNKNOWN")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeEncoding(tt.enc)
			if got != tt.want {
				t.Errorf("normalizeEncoding(%q) = %q, want %q", tt.enc, got, tt.want)
			}
		})
	}
}

func TestValidateEncoding(t *testing.T) {
	tests := []struct {
		name      string
		detected  Encoding
		declared  Encoding
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "matching encodings",
			detected: EncodingUTF8,
			declared: EncodingUTF8,
			wantErr:  false,
		},
		{
			name:     "matching with different case",
			detected: EncodingUTF8,
			declared: "utf-8",
			wantErr:  false,
		},
		{
			name:     "no declared encoding",
			detected: EncodingUTF8,
			declared: "",
			wantErr:  false,
		},
		{
			name:     "mismatched encodings",
			detected: EncodingUTF8,
			declared: EncodingUTF16,
			wantErr:  true,
			errMsg:   "encoding mismatch",
		},
		{
			name:     "UNICODE normalized to UTF-16",
			detected: EncodingUTF16,
			declared: "UNICODE",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEncoding(tt.detected, tt.declared)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEncoding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateEncoding() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestReadBOM(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		fileData []byte
		wantBOM  []byte
		wantErr  bool
	}{
		{
			name:     "UTF-8 BOM",
			fileData: []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'},
			wantBOM:  []byte{0xEF, 0xBB, 0xBF},
			wantErr:  false,
		},
		{
			name:     "UTF-16 BE BOM",
			fileData: []byte{0xFE, 0xFF, 0x00, 'H'},
			wantBOM:  []byte{0xFE, 0xFF},
			wantErr:  false,
		},
		{
			name:     "no BOM",
			fileData: []byte{'H', 'e', 'l', 'l', 'o'},
			wantBOM:  []byte{'H', 'e', 'l', 'l', 'o'}, // Reads first 4 bytes
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(tmpDir, "test_"+tt.name+".ged")
			err := os.WriteFile(tmpFile, tt.fileData, 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			file, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("failed to open file: %v", err)
			}
			defer file.Close()

			got, err := ReadBOM(file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadBOM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that we got at least the BOM bytes
			if len(tt.wantBOM) > 0 && len(got) >= len(tt.wantBOM) {
				if !bytes.Equal(got[:len(tt.wantBOM)], tt.wantBOM) {
					t.Errorf("ReadBOM() = %v, want starts with %v", got, tt.wantBOM)
				}
			}

			// Verify file position was restored
			pos, _ := file.Seek(0, io.SeekCurrent)
			if pos != 0 {
				t.Errorf("ReadBOM() did not restore file position, current pos = %d", pos)
			}
		})
	}
}

func TestSkipBOM(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		fileData []byte
		wantPos  int64
		wantErr  bool
	}{
		{
			name:     "UTF-8 BOM",
			fileData: []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'},
			wantPos:  3, // After BOM
			wantErr:  false,
		},
		{
			name:     "UTF-16 BE BOM",
			fileData: []byte{0xFE, 0xFF, 0x00, 'H'},
			wantPos:  2, // After BOM
			wantErr:  false,
		},
		{
			name:     "no BOM",
			fileData: []byte{'H', 'e', 'l', 'l', 'o'},
			wantPos:  0, // No skip
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(tmpDir, "test_"+tt.name+".ged")
			err := os.WriteFile(tmpFile, tt.fileData, 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			file, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("failed to open file: %v", err)
			}
			defer file.Close()

			err = SkipBOM(file)
			if (err != nil) != tt.wantErr {
				t.Errorf("SkipBOM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			pos, _ := file.Seek(0, io.SeekCurrent)
			if pos != tt.wantPos {
				t.Errorf("SkipBOM() position = %d, want %d", pos, tt.wantPos)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 bytes.Contains([]byte(s), []byte(substr))))
}

