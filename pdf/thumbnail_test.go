package pdf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	// Disable Wails runtime calls during tests
	SetTestMode(true)
}

func TestGenerateAllThumbnails(t *testing.T) {
	ctx := context.Background()

	// Use multi-page fixture
	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "multi-page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Test fixture not found, run: go run test/download_fixtures.go")
	}

	// Get expected page count
	doc, err := GetPDFInfo(fixturePath)
	if err != nil {
		t.Fatalf("Failed to get PDF info: %v", err)
	}

	// Generate thumbnails
	results, err := GenerateAllThumbnails(ctx, fixturePath, 150, 200)
	if err != nil {
		t.Fatalf("GenerateAllThumbnails failed: %v", err)
	}

	// Verify results
	if len(results) != doc.PageCount {
		t.Errorf("Expected %d thumbnails, got %d", doc.PageCount, len(results))
	}

	for i, thumb := range results {
		// Check page index
		if thumb.PageIndex != i {
			t.Errorf("Thumbnail %d: expected PageIndex %d, got %d", i, i, thumb.PageIndex)
		}

		// Check image data is base64 PNG
		if !strings.HasPrefix(thumb.ImageData, "data:image/png;base64,") {
			t.Errorf("Thumbnail %d: invalid image data format", i)
		}

		// Check dimensions
		if thumb.Width != 150 || thumb.Height != 200 {
			t.Errorf("Thumbnail %d: expected 150x200, got %dx%d", i, thumb.Width, thumb.Height)
		}
	}

	// Test cache hit (second call should be faster)
	results2, err := GenerateAllThumbnails(ctx, fixturePath, 150, 200)
	if err != nil {
		t.Fatalf("Second GenerateAllThumbnails failed: %v", err)
	}

	if len(results2) != len(results) {
		t.Errorf("Cache returned different number of results")
	}
}

func TestGenerateThumbnail_SinglePage(t *testing.T) {
	ctx := context.Background()

	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Test fixture not found")
	}

	// Generate thumbnail for first page
	result, err := GenerateThumbnail(ctx, fixturePath, 0, 100, 140)
	if err != nil {
		t.Fatalf("GenerateThumbnail failed: %v", err)
	}

	if result.PageIndex != 0 {
		t.Errorf("Expected PageIndex 0, got %d", result.PageIndex)
	}

	if !strings.HasPrefix(result.ImageData, "data:image/png;base64,") {
		t.Error("Invalid image data format")
	}
}

func TestGenerateThumbnail_InvalidPageIndex(t *testing.T) {
	ctx := context.Background()

	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Test fixture not found")
	}

	// Try invalid page index
	_, err := GenerateThumbnail(ctx, fixturePath, 999, 100, 140)
	if err == nil {
		t.Error("Expected error for invalid page index")
	}

	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("Expected 'out of range' error, got: %v", err)
	}
}

func TestGenerateAllThumbnails_InvalidFile(t *testing.T) {
	ctx := context.Background()

	_, err := GenerateAllThumbnails(ctx, "/nonexistent/file.pdf", 150, 200)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestGenerateAllThumbnails_EmptyPDF(t *testing.T) {
	ctx := context.Background()

	fixturePath := filepath.Join("..", "test", "fixtures", "edge-cases", "empty.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Test fixture not found")
	}

	_, err := GenerateAllThumbnails(ctx, fixturePath, 150, 200)
	if err == nil {
		t.Error("Expected error for empty PDF")
	}
}

func TestCleanupThumbnailCache(t *testing.T) {
	ctx := context.Background()

	fixturePath := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("Test fixture not found")
	}

	// Generate to create cache
	_, err := GenerateAllThumbnails(ctx, fixturePath, 150, 200)
	if err != nil {
		t.Fatalf("GenerateAllThumbnails failed: %v", err)
	}

	// Get cache dir
	cacheDir := getThumbnailCacheDir(fixturePath)

	// Verify cache exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Fatal("Cache directory should exist after generation")
	}

	// Cleanup
	if err := CleanupThumbnailCache(fixturePath); err != nil {
		t.Fatalf("CleanupThumbnailCache failed: %v", err)
	}

	// Verify cache is gone
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Error("Cache directory should be removed after cleanup")
	}
}

func TestThumbnailCacheDir_DifferentForModifiedFile(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Get cache dir
	cacheDir1 := getThumbnailCacheDir(tmpFile.Name())

	// Explicitly change modification time to ensure it's different
	futureTime := time.Now().Add(time.Hour)
	if err := os.Chtimes(tmpFile.Name(), futureTime, futureTime); err != nil {
		t.Fatal(err)
	}

	// Get cache dir again
	cacheDir2 := getThumbnailCacheDir(tmpFile.Name())

	// Should be different due to different modification time
	if cacheDir1 == cacheDir2 {
		t.Error("Cache dir should change when file is modified")
	}
}
