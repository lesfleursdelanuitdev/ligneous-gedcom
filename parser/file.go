package parser

import (
	"fmt"
	"os"
)

// ValidateFile validates that a file exists, is readable, and is not empty.
//
// Checks performed:
// 1. File exists
// 2. Path is a file (not a directory)
// 3. File is readable
// 4. File is not empty
//
// Returns an error if any validation fails, with a descriptive message.
func ValidateFile(filePath string) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if path is a file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// Check if file is readable
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("file is not readable (permission denied): %s", filePath)
		}
		return fmt.Errorf("cannot open file: %w", err)
	}
	file.Close()

	// Check if file is empty
	if info.Size() == 0 {
		return fmt.Errorf("file is empty: %s", filePath)
	}

	return nil
}

// FileInfo holds information about a validated file
type FileInfo struct {
	Path    string
	Size    int64
	Mode    os.FileMode
	IsDir   bool
	ModTime int64
}

// GetFileInfo returns information about a file after validation.
// The file must pass ValidateFile first.
func GetFileInfo(filePath string) (*FileInfo, error) {
	if err := ValidateFile(filePath); err != nil {
		return nil, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Path:    filePath,
		Size:    info.Size(),
		Mode:    info.Mode(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime().Unix(),
	}, nil
}

// IsReadable checks if a file is readable without fully validating it.
// This is a lighter check than ValidateFile.
func IsReadable(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// FileExists checks if a file exists (may be a directory).
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// IsFile checks if the path exists and is a file (not a directory).
func IsFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDirectory checks if the path exists and is a directory.
func IsDirectory(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileSize returns the size of a file in bytes.
// Returns error if file doesn't exist or cannot be accessed.
func FileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("cannot get file size: %w", err)
	}
	return info.Size(), nil
}



