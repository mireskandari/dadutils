//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const fixturesDir = "test/fixtures"

// Sample PDFs from public domain sources
var downloads = map[string]string{
	// arXiv paper - multi-page scientific document
	"valid/multi-page.pdf": "https://arxiv.org/pdf/1706.03762.pdf",
	// arXiv paper with figures/images
	"valid/high-res-images.pdf": "https://arxiv.org/pdf/2312.00752.pdf",
	// Another arXiv paper - larger
	"valid/large-document.pdf": "https://arxiv.org/pdf/2303.08774.pdf",
}

func main() {
	fmt.Println("Downloading test fixtures...")

	// Create directories
	dirs := []string{
		filepath.Join(fixturesDir, "valid"),
		filepath.Join(fixturesDir, "edge-cases"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Generate simple 1-page PDF with pdfcpu
	fmt.Println("Generating simple-1page.pdf...")
	if err := generateSimplePDF(filepath.Join(fixturesDir, "valid/simple-1page.pdf")); err != nil {
		fmt.Printf("Error generating simple PDF: %v\n", err)
	} else {
		fmt.Println("  Created valid/simple-1page.pdf")
	}

	// Download PDFs from public sources
	for path, url := range downloads {
		fullPath := filepath.Join(fixturesDir, path)
		fmt.Printf("Downloading %s...\n", path)

		if err := downloadFile(fullPath, url); err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			info, _ := os.Stat(fullPath)
			fmt.Printf("  Downloaded %s (%.2f MB)\n", path, float64(info.Size())/(1024*1024))
		}
	}

	// Create edge case files
	fmt.Println("\nCreating edge case files...")

	// Empty file
	emptyPath := filepath.Join(fixturesDir, "edge-cases/empty.pdf")
	if f, err := os.Create(emptyPath); err == nil {
		f.Close()
		fmt.Println("  Created edge-cases/empty.pdf (0 bytes)")
	}

	// Not a PDF (text file with .pdf extension)
	notPdfPath := filepath.Join(fixturesDir, "edge-cases/not-a-pdf.pdf")
	if err := os.WriteFile(notPdfPath, []byte("This is not a PDF file.\nJust plain text.\n"), 0644); err == nil {
		fmt.Println("  Created edge-cases/not-a-pdf.pdf")
	}

	// Corrupted PDF (truncated)
	corruptedPath := filepath.Join(fixturesDir, "edge-cases/corrupted.pdf")
	corruptedData := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\nTRUNCATED")
	if err := os.WriteFile(corruptedPath, corruptedData, 0644); err == nil {
		fmt.Println("  Created edge-cases/corrupted.pdf")
	}

	// Password-protected PDF
	fmt.Println("  Creating password-protected.pdf...")
	if err := createPasswordProtectedPDF(); err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Println("  Created edge-cases/password-protected.pdf (password: testpassword123)")
	}

	fmt.Println("\nDone! Run 'go test ./...' to verify fixtures.")
}

func downloadFile(filepath string, url string) error {
	// Skip if already exists
	if _, err := os.Stat(filepath); err == nil {
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func generateSimplePDF(pdfPath string) error {
	// Skip if already exists
	if _, err := os.Stat(pdfPath); err == nil {
		return nil
	}

	// Use multi-page.pdf as source and collect just the first page
	sourcePath := filepath.Join(fixturesDir, "valid/multi-page.pdf")
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source PDF not found - download multi-page.pdf first")
	}

	// Collect page 1 into a new PDF
	if err := api.CollectFile(sourcePath, pdfPath, []string{"1"}, nil); err != nil {
		return fmt.Errorf("collect page failed: %w", err)
	}

	return nil
}

func createPasswordProtectedPDF() error {
	sourcePath := filepath.Join(fixturesDir, "valid/simple-1page.pdf")
	destPath := filepath.Join(fixturesDir, "edge-cases/password-protected.pdf")

	// Skip if already exists
	if _, err := os.Stat(destPath); err == nil {
		return nil
	}

	// Make sure source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		if err := generateSimplePDF(sourcePath); err != nil {
			return fmt.Errorf("cannot create source PDF: %w", err)
		}
	}

	// Encrypt with password
	conf := model.NewDefaultConfiguration()
	conf.UserPW = "testpassword123"
	conf.OwnerPW = "testpassword123"

	if err := api.EncryptFile(sourcePath, destPath, conf); err != nil {
		return err
	}

	return nil
}
