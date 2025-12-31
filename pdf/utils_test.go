package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{10485760, "10.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	// IDs should be 8 characters
	if len(id1) != 8 {
		t.Errorf("GenerateID() length = %d, want 8", len(id1))
	}

	// IDs should be unique
	if id1 == id2 {
		t.Error("GenerateID() should generate unique IDs")
	}
}

func TestGetFileInfo(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("Hello, World!")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	info, err := GetFileInfo(tmpFile.Name())
	if err != nil {
		t.Fatalf("GetFileInfo() error = %v", err)
	}

	if info.Path != tmpFile.Name() {
		t.Errorf("GetFileInfo().Path = %q, want %q", info.Path, tmpFile.Name())
	}

	if info.Size != int64(len(content)) {
		t.Errorf("GetFileInfo().Size = %d, want %d", info.Size, len(content))
	}

	if info.SizeText == "" {
		t.Error("GetFileInfo().SizeText should not be empty")
	}
}

func TestGetFileInfo_NotFound(t *testing.T) {
	_, err := GetFileInfo("/nonexistent/file.pdf")
	if err == nil {
		t.Error("GetFileInfo() should return error for nonexistent file")
	}
}

func TestGetPDFInfo(t *testing.T) {
	// This test requires fixture files
	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	info, err := GetPDFInfo(fixturePath)
	if err != nil {
		t.Fatalf("GetPDFInfo() error = %v", err)
	}

	if info.PageCount < 1 {
		t.Errorf("GetPDFInfo().PageCount = %d, want >= 1", info.PageCount)
	}

	if info.ID == "" {
		t.Error("GetPDFInfo().ID should not be empty")
	}
}

func TestGetPDFInfo_NotPDF(t *testing.T) {
	// Create a non-PDF file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = GetPDFInfo(tmpFile.Name())
	if err == nil {
		t.Error("GetPDFInfo() should return error for non-PDF file")
	}
}

func TestValidatePDF_Valid(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	err := ValidatePDF(fixturePath)
	if err != nil {
		t.Errorf("ValidatePDF() error = %v, want nil", err)
	}
}

func TestValidatePDF_NotPDF(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "edge-cases", "not-a-pdf.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	err := ValidatePDF(fixturePath)
	if err == nil {
		t.Error("ValidatePDF() should return error for non-PDF file")
	}
}

func TestValidatePDF_Empty(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "edge-cases", "empty.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	err := ValidatePDF(fixturePath)
	if err == nil {
		t.Error("ValidatePDF() should return error for empty file")
	}
}

func TestValidatePDF_Corrupted(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "edge-cases", "corrupted.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	err := ValidatePDF(fixturePath)
	if err == nil {
		t.Error("ValidatePDF() should return error for corrupted file")
	}
}

func TestValidatePDF_PasswordProtected(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "edge-cases", "password-protected.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	err := ValidatePDF(fixturePath)
	if err == nil {
		t.Error("ValidatePDF() should return error for password-protected file")
	}
	if err != nil && !strings.Contains(err.Error(), "password") {
		t.Errorf("ValidatePDF() error should mention password, got: %v", err)
	}
}

func TestCreateTempFile(t *testing.T) {
	path, err := CreateTempFile("test", ".pdf")
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}

	if !strings.HasSuffix(path, ".pdf") {
		t.Errorf("CreateTempFile() path should end with .pdf, got %q", path)
	}

	if !strings.Contains(path, "test_") {
		t.Errorf("CreateTempFile() path should contain prefix, got %q", path)
	}
}
