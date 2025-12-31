package pdf

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ThumbnailResult represents a generated page thumbnail
type ThumbnailResult struct {
	PageIndex int    `json:"pageIndex"` // 0-based page index
	ImageData string `json:"imageData"` // base64 PNG data URL
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// ThumbnailConfig holds configuration for thumbnail generation
type ThumbnailConfig struct {
	Width   int           // Thumbnail width in pixels
	Height  int           // Thumbnail height in pixels
	Timeout time.Duration // Max time for generation
}

// DefaultThumbnailConfig returns sensible defaults
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		Width:   150,
		Height:  200,
		Timeout: 60 * time.Second,
	}
}

// GenerateAllThumbnails generates thumbnails for all pages in a PDF
// Uses a single Ghostscript process for efficiency
func GenerateAllThumbnails(ctx context.Context, pdfPath string, width, height int) ([]*ThumbnailResult, error) {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Validate file exists
	if _, err := os.Stat(pdfPath); err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	// Get page count
	pageCount, err := getPageCount(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF: %w", err)
	}

	// Safety limit
	if pageCount > 500 {
		return nil, fmt.Errorf("PDF has too many pages (%d), max is 500", pageCount)
	}

	if pageCount == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	// Check cache
	cacheDir := getThumbnailCacheDir(pdfPath)
	if cached := loadFromCache(cacheDir, pageCount, width, height); cached != nil {
		safeEmit(ctx, "thumbnail:log", "Loaded from cache")
		return cached, nil
	}

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create cache directory: %w", err)
	}

	// Get Ghostscript path
	gsPath, err := GetGhostscriptPath()
	if err != nil {
		return nil, fmt.Errorf("ghostscript not available: %w", err)
	}

	safeEmit(ctx, "thumbnail:progress", ProgressUpdate{Percent: 0, Message: "Generating thumbnails..."})

	// Generate all pages in ONE call
	outPattern := filepath.Join(cacheDir, "page_%03d.png")
	args := []string{
		"-dSAFER",
		"-dNOPAUSE",
		"-dBATCH",
		"-sDEVICE=png16m",
		"-r96", // 96 DPI for better quality thumbnails
		fmt.Sprintf("-g%dx%d", width, height),
		"-dPDFFitPage",
		"-dTextAlphaBits=4",     // Anti-aliasing for text
		"-dGraphicsAlphaBits=4", // Anti-aliasing for graphics
		fmt.Sprintf("-sOutputFile=%s", outPattern),
		pdfPath,
	}

	cmd := exec.CommandContext(ctx, gsPath, args...)
	hideWindow(cmd) // Hide console window on Windows
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Cleanup on failure
		os.RemoveAll(cacheDir)
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("thumbnail generation timed out")
		}
		errMsg := stderr.String()
		if errMsg != "" {
			return nil, fmt.Errorf("ghostscript failed: %s", errMsg)
		}
		return nil, fmt.Errorf("ghostscript failed: %w", err)
	}

	safeEmit(ctx, "thumbnail:progress", ProgressUpdate{Percent: 80, Message: "Loading thumbnails..."})

	// Load generated thumbnails
	results, err := loadGeneratedThumbnails(cacheDir, pageCount, width, height)
	if err != nil {
		return nil, err
	}

	safeEmit(ctx, "thumbnail:progress", ProgressUpdate{Percent: 100, Message: "Done"})

	return results, nil
}

// GenerateThumbnail generates a thumbnail for a single page
func GenerateThumbnail(ctx context.Context, pdfPath string, pageIndex int, width, height int) (*ThumbnailResult, error) {
	// Get page count to validate index
	pageCount, err := getPageCount(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF: %w", err)
	}

	if pageIndex < 0 || pageIndex >= pageCount {
		return nil, fmt.Errorf("page index %d out of range (0-%d)", pageIndex, pageCount-1)
	}

	// Check cache first
	cacheDir := getThumbnailCacheDir(pdfPath)
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("page_%03d.png", pageIndex+1))

	if data, err := os.ReadFile(cachePath); err == nil {
		return &ThumbnailResult{
			PageIndex: pageIndex,
			ImageData: "data:image/png;base64," + base64.StdEncoding.EncodeToString(data),
			Width:     width,
			Height:    height,
		}, nil
	}

	// Generate just this page
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	gsPath, err := GetGhostscriptPath()
	if err != nil {
		return nil, fmt.Errorf("ghostscript not available: %w", err)
	}

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create cache directory: %w", err)
	}

	// GS uses 1-based page numbers
	gsPageNum := pageIndex + 1

	args := []string{
		"-dSAFER",
		"-dNOPAUSE",
		"-dBATCH",
		"-sDEVICE=png16m",
		"-r96",
		fmt.Sprintf("-g%dx%d", width, height),
		"-dPDFFitPage",
		"-dTextAlphaBits=4",
		"-dGraphicsAlphaBits=4",
		fmt.Sprintf("-dFirstPage=%d", gsPageNum),
		fmt.Sprintf("-dLastPage=%d", gsPageNum),
		fmt.Sprintf("-sOutputFile=%s", cachePath),
		pdfPath,
	}

	cmd := exec.CommandContext(ctx, gsPath, args...)
	hideWindow(cmd) // Hide console window on Windows
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("thumbnail generation timed out")
		}
		errMsg := stderr.String()
		if errMsg != "" {
			return nil, fmt.Errorf("ghostscript failed: %s", errMsg)
		}
		return nil, fmt.Errorf("ghostscript failed: %w", err)
	}

	// Read the generated file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read generated thumbnail: %w", err)
	}

	return &ThumbnailResult{
		PageIndex: pageIndex,
		ImageData: "data:image/png;base64," + base64.StdEncoding.EncodeToString(data),
		Width:     width,
		Height:    height,
	}, nil
}

// getThumbnailCacheDir returns a unique cache directory for a PDF
func getThumbnailCacheDir(pdfPath string) string {
	// Include file modification time in hash for cache invalidation
	var modTime int64
	if info, err := os.Stat(pdfPath); err == nil {
		modTime = info.ModTime().UnixNano()
	}

	hashInput := fmt.Sprintf("%s:%d", pdfPath, modTime)
	hash := sha256.Sum256([]byte(hashInput))
	hashStr := hex.EncodeToString(hash[:8])

	return filepath.Join(os.TempDir(), "dadjoke_thumbs", hashStr)
}

// loadFromCache tries to load thumbnails from cache
func loadFromCache(cacheDir string, pageCount, width, height int) []*ThumbnailResult {
	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil
	}

	results := make([]*ThumbnailResult, 0, pageCount)

	for i := 0; i < pageCount; i++ {
		cachePath := filepath.Join(cacheDir, fmt.Sprintf("page_%03d.png", i+1))
		data, err := os.ReadFile(cachePath)
		if err != nil {
			// Cache miss - need to regenerate
			return nil
		}

		results = append(results, &ThumbnailResult{
			PageIndex: i,
			ImageData: "data:image/png;base64," + base64.StdEncoding.EncodeToString(data),
			Width:     width,
			Height:    height,
		})
	}

	return results
}

// loadGeneratedThumbnails loads freshly generated thumbnails from disk
func loadGeneratedThumbnails(cacheDir string, pageCount, width, height int) ([]*ThumbnailResult, error) {
	results := make([]*ThumbnailResult, 0, pageCount)

	for i := 0; i < pageCount; i++ {
		cachePath := filepath.Join(cacheDir, fmt.Sprintf("page_%03d.png", i+1))
		data, err := os.ReadFile(cachePath)
		if err != nil {
			return nil, fmt.Errorf("cannot read thumbnail for page %d: %w", i+1, err)
		}

		results = append(results, &ThumbnailResult{
			PageIndex: i,
			ImageData: "data:image/png;base64," + base64.StdEncoding.EncodeToString(data),
			Width:     width,
			Height:    height,
		})
	}

	return results, nil
}

// getPageCount returns the page count of a PDF
func getPageCount(pdfPath string) (int, error) {
	doc, err := GetPDFInfo(pdfPath)
	if err != nil {
		return 0, err
	}
	return doc.PageCount, nil
}

// CleanupThumbnailCache removes cached thumbnails for a PDF
func CleanupThumbnailCache(pdfPath string) error {
	cacheDir := getThumbnailCacheDir(pdfPath)
	return os.RemoveAll(cacheDir)
}

// CleanupAllThumbnailCache removes all cached thumbnails
func CleanupAllThumbnailCache() error {
	cacheDir := filepath.Join(os.TempDir(), "dadjoke_thumbs")
	return os.RemoveAll(cacheDir)
}
