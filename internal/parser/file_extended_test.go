package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileInfo_Extended(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with valid file
	tmpFile := filepath.Join(tmpDir, "test.ged")
	content := []byte("0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 TRLR\n")
	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileInfo, err := GetFileInfo(tmpFile)
	if err != nil {
		t.Fatalf("GetFileInfo failed: %v", err)
	}

	if fileInfo == nil {
		t.Fatal("GetFileInfo returned nil")
	}

	if fileInfo.Path != tmpFile {
		t.Errorf("Expected path %s, got %s", tmpFile, fileInfo.Path)
	}

	if fileInfo.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), fileInfo.Size)
	}

	if fileInfo.IsDir {
		t.Error("Expected IsDir to be false")
	}

	if fileInfo.Mode == 0 {
		t.Error("Expected non-zero mode")
	}

	if fileInfo.ModTime == 0 {
		t.Error("Expected non-zero ModTime")
	}
}

func TestGetFileInfo_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.ged")

	fileInfo, err := GetFileInfo(nonExistentFile)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if fileInfo != nil {
		t.Error("Expected nil fileInfo for non-existent file")
	}
}

func TestGetFileInfo_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	fileInfo, err := GetFileInfo(tmpDir)
	if err == nil {
		t.Error("Expected error for directory")
	}
	if fileInfo != nil {
		t.Error("Expected nil fileInfo for directory")
	}
}

func TestGetFileInfo_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.ged")
	err := os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	fileInfo, err := GetFileInfo(emptyFile)
	if err == nil {
		t.Error("Expected error for empty file")
	}
	if fileInfo != nil {
		t.Error("Expected nil fileInfo for empty file")
	}
}

