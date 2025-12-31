package pdf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func TestCombinePDFs_TwoFiles(t *testing.T) {
	fixture1 := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	fixture2 := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")

	if _, err := os.Stat(fixture1); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	doc1, err := GetPDFInfo(fixture1)
	if err != nil {
		t.Fatalf("Failed to get PDF info: %v", err)
	}
	doc2, err := GetPDFInfo(fixture2)
	if err != nil {
		t.Fatalf("Failed to get PDF info: %v", err)
	}

	docs := []PDFDocument{*doc1, *doc2}
	result, err := CombinePDFs(mockContext(), docs)
	if err != nil {
		t.Fatalf("CombinePDFs() error = %v", err)
	}
	defer CleanupTempFiles(result.OutputPath)

	if !result.Success {
		t.Error("CombinePDFs().Success = false, want true")
	}

	if result.FileCount != 2 {
		t.Errorf("CombinePDFs().FileCount = %d, want 2", result.FileCount)
	}

	// Verify output is valid PDF
	if err := api.ValidateFile(result.OutputPath, nil); err != nil {
		t.Errorf("Output is not valid PDF: %v", err)
	}

	// Verify page count is sum of inputs
	outputPages, _ := api.PageCountFile(result.OutputPath)
	expectedPages := doc1.PageCount + doc2.PageCount
	if outputPages != expectedPages {
		t.Errorf("Output page count = %d, want %d", outputPages, expectedPages)
	}
}

func TestCombinePDFs_MultipleFiles(t *testing.T) {
	fixture := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	// Create 5 document references
	var docs []PDFDocument
	for i := 0; i < 5; i++ {
		doc, err := GetPDFInfo(fixture)
		if err != nil {
			t.Fatalf("Failed to get PDF info: %v", err)
		}
		docs = append(docs, *doc)
	}

	result, err := CombinePDFs(mockContext(), docs)
	if err != nil {
		t.Fatalf("CombinePDFs() error = %v", err)
	}
	defer CleanupTempFiles(result.OutputPath)

	if result.FileCount != 5 {
		t.Errorf("CombinePDFs().FileCount = %d, want 5", result.FileCount)
	}

	// Verify page count
	outputPages, _ := api.PageCountFile(result.OutputPath)
	expectedPages := docs[0].PageCount * 5
	if outputPages != expectedPages {
		t.Errorf("Output page count = %d, want %d", outputPages, expectedPages)
	}
}

func TestCombinePDFs_EmptyList(t *testing.T) {
	_, err := CombinePDFs(mockContext(), []PDFDocument{})
	if err == nil {
		t.Error("CombinePDFs() should return error for empty list")
	}
}

func TestCombinePDFs_SingleFile(t *testing.T) {
	fixture := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	doc, err := GetPDFInfo(fixture)
	if err != nil {
		t.Fatalf("Failed to get PDF info: %v", err)
	}

	_, err = CombinePDFs(mockContext(), []PDFDocument{*doc})
	if err == nil {
		t.Error("CombinePDFs() should return error for single file")
	}
}

func TestMergeTwoFiles_Append(t *testing.T) {
	fixture1 := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	fixture2 := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")

	if _, err := os.Stat(fixture1); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	merged, err := MergeTwoFiles(mockContext(), fixture1, fixture2, MergeModeAppend)
	if err != nil {
		t.Fatalf("MergeTwoFiles() error = %v", err)
	}
	defer CleanupTempFiles(merged.Path)

	// Verify output is valid PDF
	if err := api.ValidateFile(merged.Path, nil); err != nil {
		t.Errorf("Output is not valid PDF: %v", err)
	}

	// Verify page count is sum
	pages1, _ := api.PageCountFile(fixture1)
	pages2, _ := api.PageCountFile(fixture2)
	if merged.PageCount != pages1+pages2 {
		t.Errorf("Merged page count = %d, want %d", merged.PageCount, pages1+pages2)
	}
}

func TestMergeTwoFiles_Interleave(t *testing.T) {
	fixture1 := filepath.Join("..", "test", "fixtures", "valid", "multi-page.pdf")
	fixture2 := filepath.Join("..", "test", "fixtures", "valid", "multi-page.pdf")

	if _, err := os.Stat(fixture1); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	merged, err := MergeTwoFiles(mockContext(), fixture1, fixture2, MergeModeInterleave)
	if err != nil {
		t.Fatalf("MergeTwoFiles() error = %v", err)
	}
	defer CleanupTempFiles(merged.Path)

	// Verify output is valid PDF
	if err := api.ValidateFile(merged.Path, nil); err != nil {
		t.Errorf("Output is not valid PDF: %v", err)
	}

	// Page count should still be sum of both
	pages1, _ := api.PageCountFile(fixture1)
	pages2, _ := api.PageCountFile(fixture2)
	if merged.PageCount != pages1+pages2 {
		t.Errorf("Merged page count = %d, want %d", merged.PageCount, pages1+pages2)
	}
}

func TestReorderPages_Simple(t *testing.T) {
	fixture := filepath.Join("..", "test", "fixtures", "valid", "multi-page.pdf")
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	originalPages, err := api.PageCountFile(fixture)
	if err != nil {
		t.Fatalf("Failed to get page count: %v", err)
	}

	if originalPages < 3 {
		t.Skip("Need at least 3 pages for this test")
	}

	// Reorder first 3 pages: [3, 1, 2]
	reordered, err := ReorderPages(mockContext(), fixture, []int{3, 1, 2})
	if err != nil {
		t.Fatalf("ReorderPages() error = %v", err)
	}
	defer CleanupTempFiles(reordered.Path)

	// Verify output is valid PDF
	if err := api.ValidateFile(reordered.Path, nil); err != nil {
		t.Errorf("Output is not valid PDF: %v", err)
	}

	// Verify page count matches requested order
	if reordered.PageCount != 3 {
		t.Errorf("Reordered page count = %d, want 3", reordered.PageCount)
	}
}

func TestReorderPages_AllPages(t *testing.T) {
	fixture := filepath.Join("..", "test", "fixtures", "valid", "multi-page.pdf")
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	originalPages, err := api.PageCountFile(fixture)
	if err != nil {
		t.Fatalf("Failed to get page count: %v", err)
	}

	// Create reversed order
	var order []int
	for i := originalPages; i >= 1; i-- {
		order = append(order, i)
	}

	reordered, err := ReorderPages(mockContext(), fixture, order)
	if err != nil {
		t.Fatalf("ReorderPages() error = %v", err)
	}
	defer CleanupTempFiles(reordered.Path)

	// Verify page count is preserved
	if reordered.PageCount != originalPages {
		t.Errorf("Reordered page count = %d, want %d", reordered.PageCount, originalPages)
	}
}

func TestReorderPages_EmptyOrder(t *testing.T) {
	fixture := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	_, err := ReorderPages(mockContext(), fixture, []int{})
	if err == nil {
		t.Error("ReorderPages() should return error for empty order")
	}
}

func TestCombinePDFs_OutputSize(t *testing.T) {
	fixture := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	doc, err := GetPDFInfo(fixture)
	if err != nil {
		t.Fatalf("Failed to get PDF info: %v", err)
	}

	docs := []PDFDocument{*doc, *doc}
	result, err := CombinePDFs(mockContext(), docs)
	if err != nil {
		t.Fatalf("CombinePDFs() error = %v", err)
	}
	defer CleanupTempFiles(result.OutputPath)

	if result.OutputSize == 0 {
		t.Error("CombinePDFs().OutputSize should not be 0")
	}
}

func BenchmarkCombinePDFs(b *testing.B) {
	fixture := filepath.Join("..", "test", "fixtures", "valid", "simple-1page.pdf")
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		b.Skip("Fixture not found. Run 'go run test/download_fixtures.go' first")
	}

	doc, _ := GetPDFInfo(fixture)
	docs := []PDFDocument{*doc, *doc, *doc}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := CombinePDFs(mockContext(), docs)
		if err != nil {
			b.Fatalf("CombinePDFs() error = %v", err)
		}
		CleanupTempFiles(result.OutputPath)
	}
}
