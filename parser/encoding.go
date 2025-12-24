package parser

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Encoding represents the character encoding of a GEDCOM file
type Encoding string

const (
	EncodingUTF8  Encoding = "UTF-8"
	EncodingUTF16 Encoding = "UTF-16"
	EncodingANSEL Encoding = "ANSEL"
	EncodingASCII Encoding = "ASCII"
	EncodingANSI  Encoding = "ANSI"
)

// DetectEncoding detects the character encoding of a GEDCOM file by reading the BOM.
//
// Detection order:
// 1. Check for UTF-8 BOM (EF BB BF)
// 2. Check for UTF-16 BE BOM (FE FF)
// 3. Check for UTF-16 LE BOM (FF FE)
// 4. Default to UTF-8 if no BOM found
//
// Note: ANSEL detection requires reading the CHAR tag from the header,
// which will be done during parsing. This function only detects BOM-based encodings.
func DetectEncoding(filePath string) (Encoding, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 4 bytes to check for BOM
	bom := make([]byte, 4)
	n, err := file.Read(bom)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if n < 2 {
		// File too short, default to UTF-8
		return EncodingUTF8, nil
	}

	// Check for UTF-8 BOM (EF BB BF)
	if n >= 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		return EncodingUTF8, nil
	}

	// Check for UTF-16 BE BOM (FE FF)
	if bom[0] == 0xFE && bom[1] == 0xFF {
		return EncodingUTF16, nil
	}

	// Check for UTF-16 LE BOM (FF FE)
	if bom[0] == 0xFF && bom[1] == 0xFE {
		return EncodingUTF16, nil
	}

	// No BOM detected, default to UTF-8
	// (GEDCOM 5.5.1 spec says ANSEL is primary, but UTF-8 is most common in practice)
	return EncodingUTF8, nil
}

// GetReader returns an appropriate reader for the given encoding.
// The file should be positioned at the start (after BOM if present).
func GetReader(file *os.File, encoding Encoding) (io.Reader, error) {
	switch encoding {
	case EncodingUTF8:
		// Check if we need to skip BOM
		pos, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("failed to get file position: %w", err)
		}
		
		// If at start, check for BOM
		if pos == 0 {
			bom := make([]byte, 3)
			n, _ := file.Read(bom)
			if n == 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
				// BOM found, file is already positioned after it
				return file, nil
			}
			// No BOM, rewind
			file.Seek(0, io.SeekStart)
		}
		return file, nil

	case EncodingUTF16:
		// UTF-16 requires special handling - use binary reader
		// For now, we'll use a simple approach and let Go's encoding handle it
		// This is a placeholder - full UTF-16 support would need proper decoder
		return file, nil

	case EncodingANSEL:
		// ANSEL requires special decoder (not implemented yet)
		// For now, treat as UTF-8 and warn
		return file, nil

	case EncodingASCII:
		// ASCII is subset of UTF-8, can use UTF-8 reader
		return file, nil

	case EncodingANSI:
		// ANSI (Windows-1252) requires special handling
		// For now, treat as UTF-8 and warn
		return file, nil

	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

// ValidateEncoding checks if the detected encoding matches the CHAR tag from header.
// This is called after parsing the header to verify encoding consistency.
func ValidateEncoding(detected Encoding, declared Encoding) error {
	if declared == "" {
		// No CHAR tag, use detected encoding
		return nil
	}

	// Normalize encodings for comparison
	normalizedDetected := normalizeEncoding(detected)
	normalizedDeclared := normalizeEncoding(declared)

	if normalizedDetected != normalizedDeclared {
		return fmt.Errorf("encoding mismatch: detected %s, declared %s", detected, declared)
	}

	return nil
}

// normalizeEncoding normalizes encoding names for comparison (case-insensitive)
func normalizeEncoding(enc Encoding) Encoding {
	// Convert to uppercase for case-insensitive comparison
	upper := strings.ToUpper(string(enc))
	
	switch upper {
	case "UTF-8", "UTF8":
		return EncodingUTF8
	case "UTF-16", "UTF16", "UNICODE":
		return EncodingUTF16
	case "ANSEL":
		return EncodingANSEL
	case "ASCII":
		return EncodingASCII
	case "ANSI", "WINDOWS-1252":
		return EncodingANSI
	default:
		return enc
	}
}

// ReadBOM reads and returns the BOM bytes from a file without consuming them.
// The file position is restored after reading.
func ReadBOM(file *os.File) ([]byte, error) {
	// Save current position
	pos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("failed to get file position: %w", err)
	}

	// Read BOM
	bom := make([]byte, 4)
	n, err := file.Read(bom)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read BOM: %w", err)
	}

	// Restore position
	_, err = file.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to restore file position: %w", err)
	}

	return bom[:n], nil
}

// HasBOM checks if the given bytes contain a BOM
func HasBOM(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// UTF-8 BOM: EF BB BF
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return true
	}

	// UTF-16 BE BOM: FE FF
	if data[0] == 0xFE && data[1] == 0xFF {
		return true
	}

	// UTF-16 LE BOM: FF FE
	if data[0] == 0xFF && data[1] == 0xFE {
		return true
	}

	return false
}

// GetBOMType returns the type of BOM found in the data, or empty string if none
func GetBOMType(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "UTF-8"
	}

	if data[0] == 0xFE && data[1] == 0xFF {
		return "UTF-16-BE"
	}

	if data[0] == 0xFF && data[1] == 0xFE {
		return "UTF-16-LE"
	}

	return ""
}

// SkipBOM advances the file position past the BOM if present
func SkipBOM(file *os.File) error {
	bom, err := ReadBOM(file)
	if err != nil {
		return err
	}

	if HasBOM(bom) {
		skipBytes := len(bom)
		if bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
			skipBytes = 3 // UTF-8 BOM
		} else {
			skipBytes = 2 // UTF-16 BOM
		}
		_, err = file.Seek(int64(skipBytes), io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to skip BOM: %w", err)
		}
	}

	return nil
}

