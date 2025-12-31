package pdf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// gsPresetSettings maps compression presets to Ghostscript -dPDFSETTINGS values
var gsPresetSettings = map[CompressionPreset]struct {
	setting     string
	description string
}{
	PresetScreen:   {setting: "/screen", description: "Screen quality (72 dpi, smallest)"},
	PresetEbook:    {setting: "/ebook", description: "eBook quality (150 dpi)"},
	PresetPrinter:  {setting: "/printer", description: "Printer quality (300 dpi)"},
	PresetPrepress: {setting: "/prepress", description: "Prepress quality (300 dpi, color preserving)"},
	PresetDefault:  {setting: "/default", description: "Default quality"},
}

// CompressPDF compresses a PDF file using Ghostscript with the given preset
func CompressPDF(ctx context.Context, inputPath string, preset CompressionPreset) (*CompressionResult, error) {
	// Get original file size
	originalInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}
	originalSize := originalInfo.Size()

	// Validate file is not empty
	if originalSize == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	// Emit initial progress
	safeEmit(ctx, "compress:progress", ProgressUpdate{
		Percent: 10,
		Message: "Preparing compression...",
	})
	safeEmit(ctx, "compress:log", fmt.Sprintf("Input file: %s (%s)", originalInfo.Name(), FormatFileSize(originalSize)))

	// Find Ghostscript
	gsPath, err := GetGhostscriptPath()
	if err != nil {
		return nil, fmt.Errorf("ghostscript not available: %w. %s", err, GhostscriptInstallInstructions())
	}
	safeEmit(ctx, "compress:log", fmt.Sprintf("Using Ghostscript: %s", gsPath))

	// Get preset config
	config, ok := gsPresetSettings[preset]
	if !ok {
		config = gsPresetSettings[PresetDefault]
	}
	safeEmit(ctx, "compress:log", fmt.Sprintf("Using preset: %s", config.description))

	// Create temp output file
	outputPath, err := CreateTempFile("compressed", ".pdf")
	if err != nil {
		return nil, fmt.Errorf("cannot create temp file: %w", err)
	}

	safeEmit(ctx, "compress:progress", ProgressUpdate{
		Percent: 20,
		Message: "Running Ghostscript compression...",
	})

	// Build Ghostscript command
	args := []string{
		"-q",                // Quiet mode
		"-dNOPAUSE",         // Don't pause between pages
		"-dBATCH",           // Exit after processing
		"-dSAFER",           // Restrict file operations
		"-sDEVICE=pdfwrite", // Output device
		"-dCompatibilityLevel=1.4",
		fmt.Sprintf("-dPDFSETTINGS=%s", config.setting),
		"-dEmbedAllFonts=true",
		"-dSubsetFonts=true",
		"-dCompressFonts=true",
		"-dColorImageDownsampleType=/Bicubic",
		"-dGrayImageDownsampleType=/Bicubic",
		"-dMonoImageDownsampleType=/Bicubic",
		fmt.Sprintf("-sOutputFile=%s", outputPath),
		inputPath,
	}

	safeEmit(ctx, "compress:log", "Running Ghostscript...")

	// Execute Ghostscript
	cmd := exec.CommandContext(ctx, gsPath, args...)
	hideWindow(cmd) // Hide console window on Windows
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	safeEmit(ctx, "compress:progress", ProgressUpdate{
		Percent: 50,
		Message: "Compressing PDF...",
	})

	if err := cmd.Run(); err != nil {
		CleanupTempFiles(outputPath)
		errMsg := stderr.String()
		if errMsg != "" {
			return nil, fmt.Errorf("ghostscript failed: %s", errMsg)
		}
		return nil, fmt.Errorf("ghostscript failed: %w", err)
	}

	safeEmit(ctx, "compress:progress", ProgressUpdate{
		Percent: 90,
		Message: "Finalizing...",
	})

	// Get compressed file size
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		CleanupTempFiles(outputPath)
		return nil, fmt.Errorf("cannot read compressed file: %w", err)
	}
	compressedSize := compressedInfo.Size()

	// Calculate savings
	savingsPercent := 0
	if originalSize > 0 {
		savingsPercent = int(100 - (compressedSize * 100 / originalSize))
	}

	safeEmit(ctx, "compress:log", fmt.Sprintf("Original: %s, Compressed: %s", FormatFileSize(originalSize), FormatFileSize(compressedSize)))

	if savingsPercent < 0 {
		safeEmit(ctx, "compress:log", "Warning: Compressed file is larger than original")
	} else {
		safeEmit(ctx, "compress:log", fmt.Sprintf("Saved: %d%%", savingsPercent))
	}

	safeEmit(ctx, "compress:progress", ProgressUpdate{
		Percent: 100,
		Message: "Complete",
	})

	return &CompressionResult{
		Success:        true,
		OriginalSize:   originalSize,
		CompressedSize: compressedSize,
		SavingsPercent: savingsPercent,
		OutputPath:     outputPath,
	}, nil
}
