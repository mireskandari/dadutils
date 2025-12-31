# Compress Tool Specification

## Overview
Single-file PDF compression using Ghostscript-style quality presets. User selects a PDF, picks a compression level, and saves a smaller version.

---

## User Flow

### Happy Path
1. User clicks "Compress" on home screen
2. App transitions to Compress view (back button appears top-left)
3. User adds PDF via:
   - Click "Select PDF" button → native file picker opens
   - OR drag-and-drop file onto drop zone
4. File loads, app displays:
   - File name
   - Original file size (e.g., "12.4 MB")
5. User selects compression preset from 5 options
6. User clicks "Compress" button
7. Progress section appears:
   - Progress bar with percentage
   - Collapsible "Details" section (verbose logs)
8. On completion:
   - Native save dialog appears
   - User picks destination and filename
9. Success state shows:
   - Original size → Compressed size
   - Savings percentage (e.g., "Saved 67%")
   - "Open" button to view result
10. User can:
    - Compress another file (reset view)
    - Click back to return home

### Error Paths
- **Password-protected PDF**: Show password input dialog, retry with password
- **Corrupted/invalid PDF**: Show error message, allow user to try different file
- **Compression fails**: Show error with log details, suggest trying different preset
- **User cancels save dialog**: Return to success state, allow retry save

---

## UI Components

### Header
- Back button (← icon) - top left
- Title: "Compress PDF"

### File Input Area
```
┌─────────────────────────────────────────┐
│                                         │
│     📄 Drag & drop PDF here             │
│        or                               │
│     [Select PDF]                        │
│                                         │
└─────────────────────────────────────────┘
```

### File Loaded State
```
┌─────────────────────────────────────────┐
│  📄 my-document.pdf                     │
│     Size: 12.4 MB                       │
│     [✕ Remove]                          │
└─────────────────────────────────────────┘
```

### Compression Presets
```
┌─────────────────────────────────────────┐
│  Select Quality:                        │
│                                         │
│  ○ Screen    - Smallest, low quality    │
│  ○ eBook     - Small, medium quality    │
│  ● Printer   - Balanced (default)       │
│  ○ Prepress  - Large, high quality      │
│  ○ Default   - Minimal compression      │
└─────────────────────────────────────────┘
```

### Action Button
```
[        Compress        ]
```
- Disabled until file is loaded
- Changes to spinner during processing

### Progress Section (appears during processing)
```
┌─────────────────────────────────────────┐
│  Compressing...                         │
│  ████████████░░░░░░░░  60%              │
│                                         │
│  ▼ Details                              │
│  ┌─────────────────────────────────────┐│
│  │ Processing page 12/20...            ││
│  │ Recompressing images...             ││
│  │ Optimizing fonts...                 ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### Success State
```
┌─────────────────────────────────────────┐
│  ✓ Compression Complete                 │
│                                         │
│  Original:   12.4 MB                    │
│  Compressed:  4.1 MB                    │
│  Saved:      67% ↓                      │
│                                         │
│  [Open]  [Compress Another]             │
└─────────────────────────────────────────┘
```

### Password Dialog (modal)
```
┌─────────────────────────────────────────┐
│  🔒 PDF is password-protected           │
│                                         │
│  Password: [________________]           │
│                                         │
│  [Cancel]              [Unlock]         │
└─────────────────────────────────────────┘
```

---

## Go Backend Functions

### Types
```go
type CompressionPreset string

const (
    PresetScreen   CompressionPreset = "screen"
    PresetEbook    CompressionPreset = "ebook"
    PresetPrinter  CompressionPreset = "printer"
    PresetPrepress CompressionPreset = "prepress"
    PresetDefault  CompressionPreset = "default"
)

type FileInfo struct {
    Path     string `json:"path"`
    Name     string `json:"name"`
    Size     int64  `json:"size"`      // bytes
    SizeText string `json:"sizeText"`  // human-readable, e.g., "12.4 MB"
}

type CompressionResult struct {
    Success        bool   `json:"success"`
    OriginalSize   int64  `json:"originalSize"`
    CompressedSize int64  `json:"compressedSize"`
    SavingsPercent int    `json:"savingsPercent"`
    OutputPath     string `json:"outputPath"`
    Error          string `json:"error,omitempty"`
}

type ProgressUpdate struct {
    Percent int    `json:"percent"`  // 0-100
    Message string `json:"message"`  // current operation
}
```

### Functions
```go
// SelectPDFFile opens native file picker, returns selected file info
// Returns nil if user cancels
func (a *App) SelectPDFFile() (*FileInfo, error)

// ValidatePDF checks if file is valid PDF, returns error if password-protected
func (a *App) ValidatePDF(path string) error

// UnlockPDF attempts to unlock with password, returns unlocked file path
func (a *App) UnlockPDF(path string, password string) (string, error)

// CompressPDF compresses the PDF with given preset
// Emits progress events during processing
// Returns temp file path on success (user hasn't saved yet)
func (a *App) CompressPDF(path string, preset CompressionPreset) (*CompressionResult, error)

// SaveFile opens save dialog, copies temp file to chosen location
// Returns final path or empty string if cancelled
func (a *App) SaveFile(tempPath string, suggestedName string) (string, error)

// OpenFile opens file in system default PDF viewer
func (a *App) OpenFile(path string) error

// GetFileInfo returns size info for a file path
func (a *App) GetFileInfo(path string) (*FileInfo, error)
```

### Events (Go → Frontend)
```go
// Emit progress updates during compression
runtime.EventsEmit(ctx, "compress:progress", ProgressUpdate{
    Percent: 45,
    Message: "Optimizing images...",
})

// Emit log lines for verbose details
runtime.EventsEmit(ctx, "compress:log", "Recompressing page 5/20")
```

---

## Frontend ↔ Backend Contract

### Frontend Calls (JavaScript)
```javascript
// File selection
const fileInfo = await SelectPDFFile();
// Returns: { path, name, size, sizeText } or null

// Validate before processing
await ValidatePDF(fileInfo.path);
// Throws if invalid or password-protected

// Handle password-protected files
const unlockedPath = await UnlockPDF(fileInfo.path, password);

// Compress
const result = await CompressPDF(fileInfo.path, "printer");
// Returns: { success, originalSize, compressedSize, savingsPercent, outputPath }

// Save to user location
const finalPath = await SaveFile(result.outputPath, "document_compressed.pdf");

// Open result
await OpenFile(finalPath);
```

### Frontend Event Listeners
```javascript
// Progress updates
runtime.EventsOn("compress:progress", (update) => {
    setProgress(update.percent);
    setStatus(update.message);
});

// Verbose log lines
runtime.EventsOn("compress:log", (line) => {
    appendLog(line);
});
```

---

## Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| File picker cancelled | Return null, no-op |
| Non-PDF file selected | Show error: "Please select a PDF file" |
| PDF is password-protected | Show password dialog |
| Wrong password entered | Show error, allow retry |
| PDF is corrupted | Show error with details from library |
| Compression produces larger file | Warn user, still allow save |
| Disk full during compression | Show error, clean up temp files |
| Save dialog cancelled | Keep result, allow retry |
| User clicks back during processing | Confirm dialog: "Cancel compression?" |

---

## Acceptance Criteria

- [ ] Can select PDF via file picker
- [ ] Can select PDF via drag-and-drop
- [ ] Shows file name and size after selection
- [ ] All 5 compression presets selectable
- [ ] Compress button disabled until file selected
- [ ] Progress bar updates during compression
- [ ] Verbose log section is collapsible
- [ ] Save dialog appears on completion
- [ ] Shows before/after size comparison
- [ ] Shows savings percentage
- [ ] "Open" button launches system PDF viewer
- [ ] Can compress another file without going home
- [ ] Back button returns to home screen
- [ ] Password dialog appears for protected PDFs
- [ ] Error messages are clear and actionable
- [ ] Follows system light/dark theme
