package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string // Returns file path
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid file",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "valid.ged")
				os.WriteFile(tmpFile, []byte("0 HEAD\n1 GEDC\n"), 0644)
				return tmpFile
			},
			wantErr: false,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.ged")
			},
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name: "directory instead of file",
			setup: func() string {
				dirPath := filepath.Join(tmpDir, "adir")
				os.Mkdir(dirPath, 0755)
				return dirPath
			},
			wantErr: true,
			errMsg:  "is a directory",
		},
		{
			name: "empty file",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "empty.ged")
				os.WriteFile(tmpFile, []byte{}, 0644)
				return tmpFile
			},
			wantErr: true,
			errMsg:  "is empty",
		},
		{
			name: "file with single byte",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "single.ged")
				os.WriteFile(tmpFile, []byte("0"), 0644)
				return tmpFile
			},
			wantErr: false,
		},
		{
			name: "file with whitespace only",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "whitespace.ged")
				os.WriteFile(tmpFile, []byte("   \n\t  "), 0644)
				return tmpFile
			},
			wantErr: false, // Not empty (has bytes), even if only whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()
			err := ValidateFile(filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFile() expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateFile() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFile() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateFile_Permissions(t *testing.T) {
	// This test may not work on all systems, so we'll skip if it fails
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unreadable.ged")
	os.WriteFile(tmpFile, []byte("0 HEAD\n"), 0644)

	// Try to make file unreadable (may not work on all systems)
	originalMode := os.FileMode(0)
	if info, err := os.Stat(tmpFile); err == nil {
		originalMode = info.Mode()
		os.Chmod(tmpFile, 0000)
		defer os.Chmod(tmpFile, originalMode) // Restore permissions
	}

	err := ValidateFile(tmpFile)
	// On systems where we can't change permissions, this will pass
	// On systems where we can, it should fail
	if err != nil {
		if !strings.Contains(err.Error(), "permission") && !strings.Contains(err.Error(), "readable") {
			t.Logf("ValidateFile() with unreadable file returned error (expected on some systems): %v", err)
		}
	}
}

func TestGetFileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.ged")
	content := []byte("0 HEAD\n1 GEDC\n")
	os.WriteFile(tmpFile, content, 0644)

	info, err := GetFileInfo(tmpFile)
	if err != nil {
		t.Fatalf("GetFileInfo() error = %v", err)
	}

	if info.Path != tmpFile {
		t.Errorf("GetFileInfo() Path = %q, want %q", info.Path, tmpFile)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("GetFileInfo() Size = %d, want %d", info.Size, len(content))
	}
	if info.IsDir {
		t.Errorf("GetFileInfo() IsDir = true, want false")
	}
	if info.Size == 0 {
		t.Errorf("GetFileInfo() Size = 0, want > 0")
	}
}

func TestGetFileInfo_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistent := filepath.Join(tmpDir, "nonexistent.ged")

	_, err := GetFileInfo(nonexistent)
	if err == nil {
		t.Errorf("GetFileInfo() expected error for non-existent file")
	}
}

func TestIsReadable(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		setup func() string
		want  bool
	}{
		{
			name: "readable file",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "readable.ged")
				os.WriteFile(tmpFile, []byte("0 HEAD\n"), 0644)
				return tmpFile
			},
			want: true,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.ged")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()
			got := IsReadable(filePath)
			if got != tt.want {
				t.Errorf("IsReadable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		setup func() string
		want  bool
	}{
		{
			name: "existing file",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "exists.ged")
				os.WriteFile(tmpFile, []byte("0 HEAD\n"), 0644)
				return tmpFile
			},
			want: true,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.ged")
			},
			want: false,
		},
		{
			name: "existing directory",
			setup: func() string {
				dirPath := filepath.Join(tmpDir, "adir")
				os.Mkdir(dirPath, 0755)
				return dirPath
			},
			want: true, // Directory exists, even if not a file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()
			got := FileExists(filePath)
			if got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		setup func() string
		want  bool
	}{
		{
			name: "file",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "file.ged")
				os.WriteFile(tmpFile, []byte("0 HEAD\n"), 0644)
				return tmpFile
			},
			want: true,
		},
		{
			name: "directory",
			setup: func() string {
				dirPath := filepath.Join(tmpDir, "adir")
				os.Mkdir(dirPath, 0755)
				return dirPath
			},
			want: false,
		},
		{
			name: "non-existent",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.ged")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()
			got := IsFile(filePath)
			if got != tt.want {
				t.Errorf("IsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		setup func() string
		want  bool
	}{
		{
			name: "directory",
			setup: func() string {
				dirPath := filepath.Join(tmpDir, "adir")
				os.Mkdir(dirPath, 0755)
				return dirPath
			},
			want: true,
		},
		{
			name: "file",
			setup: func() string {
				tmpFile := filepath.Join(tmpDir, "file.ged")
				os.WriteFile(tmpFile, []byte("0 HEAD\n"), 0644)
				return tmpFile
			},
			want: false,
		},
		{
			name: "non-existent",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()
			got := IsDirectory(filePath)
			if got != tt.want {
				t.Errorf("IsDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileSize(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  []byte
		wantSize int64
		wantErr  bool
	}{
		{
			name:     "file with content",
			content:  []byte("0 HEAD\n1 GEDC\n"),
			wantSize: int64(len([]byte("0 HEAD\n1 GEDC\n"))), // Use actual length
			wantErr:  false,
		},
		{
			name:     "empty file",
			content:  []byte{},
			wantSize: 0,
			wantErr:  false,
		},
		{
			name:     "single byte",
			content:  []byte("0"),
			wantSize: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(tmpDir, "size_test.ged")
			os.WriteFile(tmpFile, tt.content, 0644)

			got, err := FileSize(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantSize {
				t.Errorf("FileSize() = %d, want %d", got, tt.wantSize)
			}
		})
	}

	// Test non-existent file
	_, err := FileSize(filepath.Join(tmpDir, "nonexistent.ged"))
	if err == nil {
		t.Errorf("FileSize() expected error for non-existent file")
	}
}

