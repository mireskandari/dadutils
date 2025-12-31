package pdf

import (
	"context"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// CombinePDFs merges multiple PDF files into one
func CombinePDFs(ctx context.Context, documents []PDFDocument) (*CombineResult, error) {
	if len(documents) < 2 {
		return nil, fmt.Errorf("need at least 2 files to combine")
	}

	safeEmit(ctx, "combine:progress", ProgressUpdate{
		Percent: 10,
		Message: fmt.Sprintf("Preparing to combine %d files...", len(documents)),
	})

	// Collect all input paths
	var inputPaths []string
	totalPages := 0

	for _, doc := range documents {
		safeEmit(ctx, "combine:log", fmt.Sprintf("Adding: %s (%d pages)", doc.Name, doc.PageCount))
		inputPaths = append(inputPaths, doc.Path)
		totalPages += doc.PageCount
	}

	// Create temp output file
	outputPath, err := CreateTempFile("combined", ".pdf")
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file: %w", err)
	}

	safeEmit(ctx, "combine:progress", ProgressUpdate{
		Percent: 30,
		Message: "Merging PDF files...",
	})

	// Merge the PDFs
	if err := api.MergeCreateFile(inputPaths, outputPath, false, nil); err != nil {
		CleanupTempFiles(outputPath)
		return nil, fmt.Errorf("merge failed: %w", err)
	}

	safeEmit(ctx, "combine:progress", ProgressUpdate{
		Percent: 80,
		Message: "Finalizing...",
	})

	// Get output file size
	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		CleanupTempFiles(outputPath)
		return nil, fmt.Errorf("cannot read output file: %w", err)
	}

	safeEmit(ctx, "combine:log", fmt.Sprintf("Combined %d files into %d pages", len(documents), totalPages))
	safeEmit(ctx, "combine:log", fmt.Sprintf("Output size: %s", FormatFileSize(outputInfo.Size())))

	safeEmit(ctx, "combine:progress", ProgressUpdate{
		Percent: 100,
		Message: "Complete",
	})

	return &CombineResult{
		Success:    true,
		FileCount:  len(documents),
		PageCount:  totalPages,
		OutputSize: outputInfo.Size(),
		OutputPath: outputPath,
	}, nil
}

// MergeTwoFiles merges two PDFs with the specified mode (interleave or append)
func MergeTwoFiles(ctx context.Context, pathA, pathB string, mode MergeMode) (*PDFDocument, error) {
	safeEmit(ctx, "combine:log", fmt.Sprintf("Merging two files with mode: %s", mode))

	// Create temp output file
	outputPath, err := CreateTempFile("merged", ".pdf")
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file: %w", err)
	}

	if mode == MergeModeInterleave {
		// Get page counts
		countA, err := api.PageCountFile(pathA)
		if err != nil {
			return nil, fmt.Errorf("cannot read first PDF: %w", err)
		}
		countB, err := api.PageCountFile(pathB)
		if err != nil {
			return nil, fmt.Errorf("cannot read second PDF: %w", err)
		}

		safeEmit(ctx, "combine:log", fmt.Sprintf("Interleaving %d + %d pages", countA, countB))

		// Efficient interleave: first merge both PDFs, then reorder pages
		// After merge, PDF A pages are 1..countA, PDF B pages are (countA+1)..(countA+countB)
		mergedPath, err := CreateTempFile("merged_temp", ".pdf")
		if err != nil {
			CleanupTempFiles(outputPath)
			return nil, fmt.Errorf("cannot create temp file: %w", err)
		}

		// Merge both PDFs
		if err := api.MergeCreateFile([]string{pathA, pathB}, mergedPath, false, nil); err != nil {
			CleanupTempFiles(outputPath, mergedPath)
			return nil, fmt.Errorf("merge failed: %w", err)
		}

		// Build interleaved page order
		// Pages from A: 1, 2, ..., countA
		// Pages from B: countA+1, countA+2, ..., countA+countB
		// Interleaved: 1, countA+1, 2, countA+2, ...
		var pageOrder []string
		maxPages := countA
		if countB > maxPages {
			maxPages = countB
		}

		for i := 1; i <= maxPages; i++ {
			if i <= countA {
				pageOrder = append(pageOrder, fmt.Sprintf("%d", i))
			}
			if i <= countB {
				pageOrder = append(pageOrder, fmt.Sprintf("%d", countA+i))
			}
		}

		// Use CollectFile to reorder pages
		if err := api.CollectFile(mergedPath, outputPath, pageOrder, nil); err != nil {
			CleanupTempFiles(outputPath, mergedPath)
			return nil, fmt.Errorf("interleave failed: %w", err)
		}

		CleanupTempFiles(mergedPath)
	} else {
		// Simple append
		return mergeTwoAppend(ctx, pathA, pathB, outputPath)
	}

	// Get info about merged file
	return GetPDFInfo(outputPath)
}

func mergeTwoAppend(ctx context.Context, pathA, pathB, outputPath string) (*PDFDocument, error) {
	safeEmit(ctx, "combine:log", "Appending files...")

	if err := api.MergeCreateFile([]string{pathA, pathB}, outputPath, false, nil); err != nil {
		CleanupTempFiles(outputPath)
		return nil, fmt.Errorf("merge failed: %w", err)
	}

	return GetPDFInfo(outputPath)
}

// ReorderPages creates a new PDF with pages in the specified order
func ReorderPages(ctx context.Context, path string, pageOrder []int) (*PDFDocument, error) {
	if len(pageOrder) == 0 {
		return nil, fmt.Errorf("page order cannot be empty")
	}

	safeEmit(ctx, "combine:log", fmt.Sprintf("Reordering %d pages", len(pageOrder)))

	// Create temp output file
	outputPath, err := CreateTempFile("reordered", ".pdf")
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file: %w", err)
	}

	// Build page selection string for pdfcpu
	var pageSelections []string
	for _, pageNum := range pageOrder {
		pageSelections = append(pageSelections, fmt.Sprintf("%d", pageNum))
	}

	// Use pdfcpu to collect pages in new order
	if err := api.CollectFile(path, outputPath, pageSelections, nil); err != nil {
		CleanupTempFiles(outputPath)
		return nil, fmt.Errorf("reorder failed: %w", err)
	}

	return GetPDFInfo(outputPath)
}
