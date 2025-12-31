# Combine Tool Specification

## Overview
Merge multiple PDFs into one. Default view shows a reorderable file list. On-demand page-level controls allow merging pages between two files or reordering pages within one file.

---

## User Flow

### Happy Path - Basic Combine
1. User clicks "Combine" on home screen
2. App transitions to Combine view
3. User adds PDFs via:
   - Click "Add Files" → native file picker (multi-select)
   - OR drag-and-drop multiple files onto drop zone
4. Files appear in reorderable list showing:
   - File name
   - Page count
   - File size
5. User drags to reorder files
6. User clicks "Combine" button
7. Progress section shows combining status
8. Native save dialog appears
9. Success state with "Open" button

### Page-Level: Merge Two Files
1. In file list, user selects exactly 2 files (checkboxes)
2. "Merge Pages" button becomes active
3. User clicks "Merge Pages"
4. Dialog asks: "How should pages be merged?"
   - Interleave (A1, B1, A2, B2...)
   - Append (All of A, then all of B)
5. User picks option
6. Two files are replaced by single merged file in list
7. User continues with normal combine flow

### Page-Level: Reorder Within File
1. In file list, user selects exactly 1 file
2. "Edit Pages" button becomes active
3. User clicks "Edit Pages"
4. Modal opens with thumbnail grid of all pages
5. User drags pages to reorder
6. User clicks "Done"
7. File in list is updated with new page order
8. User continues with normal combine flow

---

## UI Components

### Header
- Back button (←) - top left
- Title: "Combine PDFs"

### File Input Area (empty state)
```
┌─────────────────────────────────────────┐
│                                         │
│     📄 Drag & drop PDFs here            │
│        or                               │
│     [Add Files]                         │
│                                         │
└─────────────────────────────────────────┘
```

### File List (populated)
```
┌─────────────────────────────────────────┐
│  [Add More]                             │
│                                         │
│  ☐ ≡ 📄 document1.pdf                   │
│       8 pages • 2.1 MB                  │
│                                         │
│  ☐ ≡ 📄 report.pdf                      │
│       12 pages • 4.5 MB                 │
│                                         │
│  ☐ ≡ 📄 appendix.pdf                    │
│       3 pages • 0.8 MB                  │
│                                         │
│  ─────────────────────────────────────  │
│  Total: 23 pages • 7.4 MB               │
└─────────────────────────────────────────┘

Legend:
  ☐ = checkbox for selection
  ≡ = drag handle for reordering
```

### Selection Actions Bar (appears when files selected)
```
┌─────────────────────────────────────────┐
│  2 files selected                       │
│  [Merge Pages]  [Remove]  [Clear]       │
└─────────────────────────────────────────┘
```

When 1 file selected:
```
┌─────────────────────────────────────────┐
│  1 file selected                        │
│  [Edit Pages]   [Remove]  [Clear]       │
└─────────────────────────────────────────┘
```

### Merge Mode Dialog
```
┌─────────────────────────────────────────┐
│  How should pages be merged?            │
│                                         │
│  ○ Interleave                           │
│    A1 → B1 → A2 → B2 → A3 → B3...       │
│    (Good for double-sided scans)        │
│                                         │
│  ○ Append                               │
│    All of A → then all of B             │
│    (Standard merge)                     │
│                                         │
│  [Cancel]              [Merge]          │
└─────────────────────────────────────────┘
```

### Edit Pages Modal
```
┌─────────────────────────────────────────────────────┐
│  Edit Pages: document1.pdf                     [✕]  │
│  ───────────────────────────────────────────────────│
│                                                     │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐               │
│  │  1  │  │  2  │  │  3  │  │  4  │               │
│  │     │  │     │  │     │  │     │               │
│  │ 📄  │  │ 📄  │  │ 📄  │  │ 📄  │               │
│  └─────┘  └─────┘  └─────┘  └─────┘               │
│                                                     │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐               │
│  │  5  │  │  6  │  │  7  │  │  8  │               │
│  │     │  │     │  │     │  │     │               │
│  │ 📄  │  │ 📄  │  │ 📄  │  │ 📄  │               │
│  └─────┘  └─────┘  └─────┘  └─────┘               │
│                                                     │
│  Drag pages to reorder                              │
│                                                     │
│  [Cancel]                              [Done]       │
└─────────────────────────────────────────────────────┘
```

### Action Button
```
[        Combine        ]
```
- Disabled until ≥2 files in list
- Shows total page count: "Combine (23 pages)"

### Progress Section
```
┌─────────────────────────────────────────┐
│  Combining 3 files...                   │
│  ████████████░░░░░░░░  60%              │
│                                         │
│  ▼ Details                              │
│  ┌─────────────────────────────────────┐│
│  │ Processing document1.pdf...         ││
│  │ Adding pages 1-8...                 ││
│  │ Processing report.pdf...            ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### Success State
```
┌─────────────────────────────────────────┐
│  ✓ Combined Successfully                │
│                                         │
│  3 files → 1 PDF                        │
│  23 pages total                         │
│  Output size: 7.2 MB                    │
│                                         │
│  [Open]  [Combine More]                 │
└─────────────────────────────────────────┘
```

---

## Go Backend Functions

### Types
```go
type PDFDocument struct {
    ID        string `json:"id"`        // unique identifier for this session
    Path      string `json:"path"`
    Name      string `json:"name"`
    PageCount int    `json:"pageCount"`
    Size      int64  `json:"size"`
    SizeText  string `json:"sizeText"`
    // For page-level edits, tracks custom page order
    PageOrder []int  `json:"pageOrder,omitempty"` // nil = default order
}

type PageThumbnail struct {
    PageNum   int    `json:"pageNum"`
    Thumbnail string `json:"thumbnail"` // base64 encoded image
}

type MergeMode string

const (
    MergeModeInterleave MergeMode = "interleave"
    MergeModeAppend     MergeMode = "append"
)

type CombineResult struct {
    Success    bool   `json:"success"`
    PageCount  int    `json:"pageCount"`
    OutputSize int64  `json:"outputSize"`
    OutputPath string `json:"outputPath"`
    Error      string `json:"error,omitempty"`
}
```

### Functions
```go
// SelectPDFFiles opens native file picker with multi-select
// Returns list of file infos, or empty if cancelled
func (a *App) SelectPDFFiles() ([]PDFDocument, error)

// LoadPDFInfo loads metadata for a dropped file path
func (a *App) LoadPDFInfo(path string) (*PDFDocument, error)

// GetPageThumbnails generates thumbnails for all pages in a PDF
// Used for the Edit Pages modal
func (a *App) GetPageThumbnails(path string) ([]PageThumbnail, error)

// MergeTwoFiles merges two PDFs with specified mode
// Returns a new temp PDF that replaces both in the list
func (a *App) MergeTwoFiles(pathA string, pathB string, mode MergeMode) (*PDFDocument, error)

// ReorderPages creates new PDF with pages in specified order
// pageOrder is array of 1-indexed page numbers, e.g., [3, 1, 2, 4]
func (a *App) ReorderPages(path string, pageOrder []int) (*PDFDocument, error)

// CombinePDFs merges all documents in order
// documents should include any custom pageOrder from edits
func (a *App) CombinePDFs(documents []PDFDocument) (*CombineResult, error)

// SaveFile opens save dialog with suggested name
func (a *App) SaveFile(tempPath string, suggestedName string) (string, error)

// OpenFile opens in system viewer
func (a *App) OpenFile(path string) error
```

### Events (Go → Frontend)
```go
// Progress during combine
runtime.EventsEmit(ctx, "combine:progress", ProgressUpdate{
    Percent: 33,
    Message: "Processing document1.pdf...",
})

// Log lines
runtime.EventsEmit(ctx, "combine:log", "Adding pages 1-8 from document1.pdf")
```

---

## Frontend ↔ Backend Contract

### State Management (Frontend)
```javascript
const [documents, setDocuments] = useState([]);     // PDFDocument[]
const [selectedIds, setSelectedIds] = useState([]); // string[]
const [isProcessing, setIsProcessing] = useState(false);
const [progress, setProgress] = useState(0);
const [logs, setLogs] = useState([]);

// Derived
const canMergePages = selectedIds.length === 2;
const canEditPages = selectedIds.length === 1;
const canCombine = documents.length >= 2;
```

### Frontend Calls
```javascript
// Add files via picker
const newDocs = await SelectPDFFiles();
setDocuments([...documents, ...newDocs]);

// Handle dropped files
for (const path of droppedPaths) {
    const doc = await LoadPDFInfo(path);
    setDocuments(prev => [...prev, doc]);
}

// Get thumbnails for edit modal
const thumbnails = await GetPageThumbnails(doc.path);

// Merge two selected files
const merged = await MergeTwoFiles(docA.path, docB.path, "interleave");
// Replace docA and docB with merged in documents array

// Reorder pages in one file
const reordered = await ReorderPages(doc.path, [3, 1, 2, 4]);
// Replace doc with reordered in documents array

// Final combine
const result = await CombinePDFs(documents);
const finalPath = await SaveFile(result.outputPath, "combined.pdf");
await OpenFile(finalPath);
```

---

## Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| Less than 2 files | Combine button disabled |
| File picker cancelled | No-op |
| Password-protected PDF | Show password dialog per file |
| Duplicate file added | Allow (user may want same file twice) |
| File removed from disk mid-session | Show error when attempting combine |
| Very large PDF (100+ pages) | Thumbnail generation shows loading state |
| Interleave with unequal page counts | Append remaining pages at end |
| User closes Edit Pages without saving | Discard changes, keep original order |
| Combine produces very large file | Show warning but proceed |
| Back button during processing | Confirm: "Cancel combine?" |

---

## Acceptance Criteria

### Basic Combine
- [ ] Can add files via file picker (multi-select)
- [ ] Can add files via drag-and-drop
- [ ] Files appear in list with name, page count, size
- [ ] Can reorder files via drag-and-drop
- [ ] Can remove individual files from list
- [ ] Total page count and size shown
- [ ] Combine button disabled until ≥2 files
- [ ] Progress bar during combine
- [ ] Collapsible verbose logs
- [ ] Save dialog on completion
- [ ] Success shows output info
- [ ] "Open" button works
- [ ] "Combine More" resets view

### Page-Level Controls
- [ ] Checkbox selection on each file
- [ ] Selection bar appears when files selected
- [ ] "Merge Pages" active when exactly 2 selected
- [ ] "Edit Pages" active when exactly 1 selected
- [ ] Merge dialog offers Interleave and Append
- [ ] Merged files replace originals in list
- [ ] Edit Pages shows thumbnail grid
- [ ] Can drag to reorder pages in grid
- [ ] Done button applies new order
- [ ] Cancel discards changes

### General
- [ ] Back button returns home
- [ ] Handles password-protected PDFs
- [ ] Follows system light/dark theme
- [ ] Clear error messages
