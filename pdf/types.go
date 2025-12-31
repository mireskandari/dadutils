package pdf

// FileInfo represents basic file information
type FileInfo struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	SizeText string `json:"sizeText"`
}

// PDFDocument represents a PDF with metadata
type PDFDocument struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	PageCount int    `json:"pageCount"`
	Size      int64  `json:"size"`
	SizeText  string `json:"sizeText"`
	PageOrder []int  `json:"pageOrder,omitempty"`
}

// CompressionPreset defines compression quality levels
type CompressionPreset string

const (
	PresetScreen   CompressionPreset = "screen"
	PresetEbook    CompressionPreset = "ebook"
	PresetPrinter  CompressionPreset = "printer"
	PresetPrepress CompressionPreset = "prepress"
	PresetDefault  CompressionPreset = "default"
)

// CompressionResult holds the result of a compression operation
type CompressionResult struct {
	Success        bool   `json:"success"`
	OriginalSize   int64  `json:"originalSize"`
	CompressedSize int64  `json:"compressedSize"`
	SavingsPercent int    `json:"savingsPercent"`
	OutputPath     string `json:"outputPath"`
	Error          string `json:"error,omitempty"`
}

// CombineResult holds the result of a combine operation
type CombineResult struct {
	Success    bool   `json:"success"`
	FileCount  int    `json:"fileCount"`
	PageCount  int    `json:"pageCount"`
	OutputSize int64  `json:"outputSize"`
	OutputPath string `json:"outputPath"`
	Error      string `json:"error,omitempty"`
}

// MergeMode defines how to merge two PDFs
type MergeMode string

const (
	MergeModeInterleave MergeMode = "interleave"
	MergeModeAppend     MergeMode = "append"
)

// ProgressUpdate represents a progress event
type ProgressUpdate struct {
	Percent int    `json:"percent"`
	Message string `json:"message"`
}
