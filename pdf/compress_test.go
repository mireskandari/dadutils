package pdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// mockContext returns a context that doesn't emit events (for testing)
func mockContext() context.Context {
	return context.Background()
}

func TestCompressPDF_AllPresets(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "multi-page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	presets := []CompressionPreset{
		PresetScreen,
		PresetEbook,
		PresetPrinter,
		PresetPrepress,
		PresetDefault,
	}

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			result, err := CompressPDF(mockContext(), fixturePath, preset)
			if err != nil {
				t.Fatalf("CompressPDF() error = %v", err)
			}
			defer CleanupTempFiles(result.OutputPath)

			if !result.Success {
				t.Errorf("CompressPDF().Success = false, want true")
			}

			if result.OutputPath == "" {
				t.Error("CompressPDF().OutputPath should not be empty")
			}

			// Verify output is valid PDF
			if err := api.ValidateFile(result.OutputPath, nil); err != nil {
				t.Errorf("Output is not valid PDF: %v", err)
			}

			// Verify original size is recorded
			if result.OriginalSize == 0 {
				t.Error("CompressPDF().OriginalSize should not be 0")
			}

			// Verify compressed size is recorded
			if result.CompressedSize == 0 {
				t.Error("CompressPDF().CompressedSize should not be 0")
			}
		})
	}
}

func TestCompressPDF_OutputIsValidPDF(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	result, err := CompressPDF(mockContext(), fixturePath, PresetPrinter)
	if err != nil {
		t.Fatalf("CompressPDF() error = %v", err)
	}
	defer CleanupTempFiles(result.OutputPath)

	// Verify output file exists
	if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
		t.Error("Output file does not exist")
	}

	// Verify it's a valid PDF
	if err := api.ValidateFile(result.OutputPath, nil); err != nil {
		t.Errorf("Output is not valid PDF: %v", err)
	}

	// Verify page count is preserved
	originalPages, _ := api.PageCountFile(fixturePath)
	compressedPages, _ := api.PageCountFile(result.OutputPath)
	if originalPages != compressedPages {
		t.Errorf("Page count changed: original=%d, compressed=%d", originalPages, compressedPages)
	}
}

func TestCompressPDF_SavingsCalculation(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "high-res-images.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	result, err := CompressPDF(mockContext(), fixturePath, PresetScreen)
	if err != nil {
		t.Fatalf("CompressPDF() error = %v", err)
	}
	defer CleanupTempFiles(result.OutputPath)

	// Verify savings percentage is calculated correctly
	expectedSavings := int(100 - (result.CompressedSize * 100 / result.OriginalSize))
	if result.SavingsPercent != expectedSavings {
		t.Errorf("SavingsPercent = %d, want %d", result.SavingsPercent, expectedSavings)
	}
}

func TestCompressPDF_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "large-document.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	result, err := CompressPDF(mockContext(), fixturePath, PresetEbook)
	if err != nil {
		t.Fatalf("CompressPDF() error = %v", err)
	}
	defer CleanupTempFiles(result.OutputPath)

	if !result.Success {
		t.Error("CompressPDF() should succeed for large files")
	}
}

func TestCompressPDF_InvalidInput(t *testing.T) {
	// Test with non-existent file
	_, err := CompressPDF(mockContext(), "/nonexistent/file.pdf", PresetPrinter)
	if err == nil {
		t.Error("CompressPDF() should return error for non-existent file")
	}
}

func TestCompressPDF_NotAPDF(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "edge-cases", "not-a-pdf.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	_, err := CompressPDF(mockContext(), fixturePath, PresetPrinter)
	if err == nil {
		t.Error("CompressPDF() should return error for non-PDF file")
	}
}

func TestCompressPDF_EmptyFile(t *testing.T) {
	fixturePath := filepath.Join("..", "test", "fixtures", "edge-cases", "empty.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	_, err := CompressPDF(mockContext(), fixturePath, PresetPrinter)
	if err == nil {
		t.Error("CompressPDF() should return error for empty file")
	}
}

func BenchmarkCompressPDF(b *testing.B) {
	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "multi-page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		b.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := CompressPDF(mockContext(), fixturePath, PresetPrinter)
		if err != nil {
			b.Fatalf("CompressPDF() error = %v", err)
		}
		CleanupTempFiles(result.OutputPath)
	}
}
