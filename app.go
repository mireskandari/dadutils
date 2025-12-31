package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"

	"dadjoke/pdf"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ============================================================================
// File Dialog Methods
// ============================================================================

// SelectPDFFile opens a native file picker for selecting a single PDF
func (a *App) SelectPDFFile() (*pdf.FileInfo, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select PDF File",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return nil, err
	}
	if selection == "" {
		return nil, nil // User cancelled
	}

	return pdf.GetFileInfo(selection)
}

// SelectPDFFiles opens a native file picker for selecting multiple PDFs
func (a *App) SelectPDFFiles() ([]pdf.PDFDocument, error) {
	selections, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select PDF Files",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(selections) == 0 {
		return nil, nil // User cancelled
	}

	var documents []pdf.PDFDocument
	for _, path := range selections {
		doc, err := pdf.GetPDFInfo(path)
		if err != nil {
			// Skip invalid files but log error
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("Skipped %s: %v", filepath.Base(path), err))
			continue
		}
		documents = append(documents, *doc)
	}

	return documents, nil
}

// SaveFile opens a save dialog and copies the temp file to the chosen location
func (a *App) SaveFile(tempPath string, suggestedName string) (string, error) {
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save PDF",
		DefaultFilename: suggestedName,
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return "", err
	}
	if savePath == "" {
		return "", nil // User cancelled
	}

	// Ensure .pdf extension
	if filepath.Ext(savePath) != ".pdf" {
		savePath += ".pdf"
	}

	// Copy temp file to save location
	if err := copyFile(tempPath, savePath); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Clean up temp file
	pdf.CleanupTempFiles(tempPath)

	return savePath, nil
}

// OpenFile opens a file with the system's default application
func (a *App) OpenFile(path string) error {
	var cmd *exec.Cmd

	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default: // linux and others
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

// ============================================================================
// PDF Information Methods
// ============================================================================

// LoadPDFInfo loads metadata for a PDF file
func (a *App) LoadPDFInfo(path string) (*pdf.PDFDocument, error) {
	return pdf.GetPDFInfo(path)
}

// ValidatePDF checks if a file is a valid PDF
func (a *App) ValidatePDF(path string) error {
	return pdf.ValidatePDF(path)
}

// ============================================================================
// Compress Methods
// ============================================================================

// CompressPDF compresses a PDF with the given preset
func (a *App) CompressPDF(path string, preset string) (*pdf.CompressionResult, error) {
	return pdf.CompressPDF(a.ctx, path, pdf.CompressionPreset(preset))
}

// ============================================================================
// Combine Methods
// ============================================================================

// CombinePDFs merges multiple PDFs into one
func (a *App) CombinePDFs(documents []pdf.PDFDocument) (*pdf.CombineResult, error) {
	return pdf.CombinePDFs(a.ctx, documents)
}

// MergeTwoFiles merges two PDFs with the specified mode
func (a *App) MergeTwoFiles(pathA, pathB string, mode string) (*pdf.PDFDocument, error) {
	return pdf.MergeTwoFiles(a.ctx, pathA, pathB, pdf.MergeMode(mode))
}

// ReorderPages creates a new PDF with pages in the specified order
func (a *App) ReorderPages(path string, pageOrder []int) (*pdf.PDFDocument, error) {
	return pdf.ReorderPages(a.ctx, path, pageOrder)
}

// ============================================================================
// Helper Functions
// ============================================================================

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
