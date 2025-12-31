# Split Tool Specification

## Overview
Extract specific pages from a PDF into a new document. User selects a PDF, views thumbnail grid of all pages, clicks to select pages, and exports selected pages as a new PDF.

---

## User Flow

### Happy Path
1. User clicks "Split" on home screen
2. App transitions to Split view
3. User adds PDF via:
   - Click "Select PDF" → native file picker
   - OR drag-and-drop file onto drop zone
4. Loading state while thumbnails generate
5. Thumbnail grid appears showing all pages
6. User clicks pages to select (toggle)
7. Selection count updates: "5 of 12 pages selected"
8. User clicks "Extract Selected"
9. Progress bar during extraction
10. Native save dialog appears
11. Success state with "Open" button

### Selection Shortcuts
- Click individual pages to toggle
- Shift+click for range selection
- "Select All" / "Deselect All" buttons
- "Select Odd" / "Select Even" for common patterns

---

## UI Components

### Header
- Back button (←) - top left
- Title: "Split PDF"

### File Input Area (empty state)
```
┌─────────────────────────────────────────┐
│                                         │
│     📄 Drag & drop PDF here             │
│        or                               │
│     [Select PDF]                        │
│                                         │
└─────────────────────────────────────────┘
```

### Loading State (generating thumbnails)
```
┌─────────────────────────────────────────┐
│  📄 my-document.pdf                     │
│     12 pages • 4.2 MB                   │
│                                         │
│  ⟳ Generating page previews...          │
│  ████████░░░░░░░░░░░░  40%              │
└─────────────────────────────────────────┘
```

### Page Grid (main view)
```
┌─────────────────────────────────────────────────────┐
│  📄 my-document.pdf                                 │
│     12 pages • 4.2 MB                [Change File]  │
│─────────────────────────────────────────────────────│
│                                                     │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐               │
│  │ ✓ 1 │  │  2  │  │ ✓ 3 │  │  4  │               │
│  │     │  │     │  │     │  │     │               │
│  │ 📄  │  │ 📄  │  │ 📄  │  │ 📄  │               │
│  └─────┘  └─────┘  └─────┘  └─────┘               │
│  selected          selected                         │
│                                                     │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐               │
│  │ ✓ 5 │  │  6  │  │  7  │  │  8  │               │
│  │     │  │     │  │     │  │     │               │
│  │ 📄  │  │ 📄  │  │ 📄  │  │ 📄  │               │
│  └─────┘  └─────┘  └─────┘  └─────┘               │
│  selected                                           │
│                                                     │
│  ... (scrollable if many pages)                     │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Selection Toolbar
```
┌─────────────────────────────────────────────────────┐
│  3 of 12 pages selected                             │
│                                                     │
│  [Select All] [Deselect All] [Odd] [Even]          │
└─────────────────────────────────────────────────────┘
```

### Action Button
```
[     Extract Selected (3 pages)     ]
```
- Disabled until ≥1 page selected
- Shows count of selected pages

### Progress Section
```
┌─────────────────────────────────────────┐
│  Extracting pages...                    │
│  ████████████░░░░░░░░  60%              │
│                                         │
│  ▼ Details                              │
│  ┌─────────────────────────────────────┐│
│  │ Extracting page 1...                ││
│  │ Extracting page 3...                ││
│  │ Extracting page 5...                ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### Success State
```
┌─────────────────────────────────────────┐
│  ✓ Extraction Complete                  │
│                                         │
│  Extracted 3 pages                      │
│  Output size: 1.2 MB                    │
│                                         │
│  [Open]  [Extract More from Same PDF]   │
│          [Start Over]                   │
└─────────────────────────────────────────┘
```

---

## Go Backend Functions

### Types
```go
type PDFInfo struct {
    Path      string `json:"path"`
    Name      string `json:"name"`
    PageCount int    `json:"pageCount"`
    Size      int64  `json:"size"`
    SizeText  string `json:"sizeText"`
}

type PageThumbnail struct {
    PageNum   int    `json:"pageNum"`   // 1-indexed
    Thumbnail string `json:"thumbnail"` // base64 encoded PNG
    Width     int    `json:"width"`     // thumbnail dimensions
    Height    int    `json:"height"`
}

type SplitResult struct {
    Success       bool   `json:"success"`
    PagesExtracted int   `json:"pagesExtracted"`
    OutputSize    int64  `json:"outputSize"`
    OutputPath    string `json:"outputPath"`
    Error         string `json:"error,omitempty"`
}
```

### Functions
```go
// SelectPDFFile opens native file picker for single PDF
func (a *App) SelectPDFFile() (*PDFInfo, error)

// LoadPDFInfo gets metadata for a file path
func (a *App) LoadPDFInfo(path string) (*PDFInfo, error)

// GenerateThumbnails creates preview images for all pages
// Emits progress events during generation
// Returns array of thumbnails ordered by page number
func (a *App) GenerateThumbnails(path string) ([]PageThumbnail, error)

// ExtractPages creates new PDF with only specified pages
// pages is array of 1-indexed page numbers, e.g., [1, 3, 5]
// Order in pages array determines order in output
func (a *App) ExtractPages(path string, pages []int) (*SplitResult, error)

// SaveFile opens save dialog
func (a *App) SaveFile(tempPath string, suggestedName string) (string, error)

// OpenFile opens in system viewer
func (a *App) OpenFile(path string) error
```

### Events (Go → Frontend)
```go
// Thumbnail generation progress
runtime.EventsEmit(ctx, "split:thumbnail-progress", ProgressUpdate{
    Percent: 50,
    Message: "Generating preview for page 6/12...",
})

// Extraction progress
runtime.EventsEmit(ctx, "split:progress", ProgressUpdate{
    Percent: 66,
    Message: "Extracting page 3...",
})

// Log lines
runtime.EventsEmit(ctx, "split:log", "Extracting page 5 of 12...")
```

---

## Frontend ↔ Backend Contract

### State Management (Frontend)
```javascript
const [pdfInfo, setPdfInfo] = useState(null);           // PDFInfo
const [thumbnails, setThumbnails] = useState([]);       // PageThumbnail[]
const [selectedPages, setSelectedPages] = useState([]); // number[] (1-indexed)
const [isLoading, setIsLoading] = useState(false);      // thumbnail generation
const [isProcessing, setIsProcessing] = useState(false);// extraction
const [progress, setProgress] = useState(0);
const [logs, setLogs] = useState([]);

// Derived
const canExtract = selectedPages.length > 0;
```

### Frontend Calls
```javascript
// Select file
const info = await SelectPDFFile();
setPdfInfo(info);

// Generate thumbnails
setIsLoading(true);
const thumbs = await GenerateThumbnails(info.path);
setThumbnails(thumbs);
setIsLoading(false);

// Toggle page selection
const togglePage = (pageNum) => {
    setSelectedPages(prev =>
        prev.includes(pageNum)
            ? prev.filter(p => p !== pageNum)
            : [...prev, pageNum].sort((a, b) => a - b)
    );
};

// Select all / deselect all
const selectAll = () => setSelectedPages(thumbnails.map(t => t.pageNum));
const deselectAll = () => setSelectedPages([]);

// Select odd/even
const selectOdd = () => setSelectedPages(
    thumbnails.map(t => t.pageNum).filter(p => p % 2 === 1)
);
const selectEven = () => setSelectedPages(
    thumbnails.map(t => t.pageNum).filter(p => p % 2 === 0)
);

// Range select (shift+click)
const rangeSelect = (startPage, endPage) => {
    const range = [];
    for (let i = Math.min(startPage, endPage); i <= Math.max(startPage, endPage); i++) {
        range.push(i);
    }
    setSelectedPages(prev => [...new Set([...prev, ...range])].sort((a, b) => a - b));
};

// Extract
setIsProcessing(true);
const result = await ExtractPages(pdfInfo.path, selectedPages);
const finalPath = await SaveFile(result.outputPath, "extracted.pdf");
await OpenFile(finalPath);
```

### Event Listeners
```javascript
// Thumbnail generation progress
runtime.EventsOn("split:thumbnail-progress", (update) => {
    setProgress(update.percent);
});

// Extraction progress
runtime.EventsOn("split:progress", (update) => {
    setProgress(update.percent);
});

// Logs
runtime.EventsOn("split:log", (line) => {
    setLogs(prev => [...prev, line]);
});
```

---

## Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| File picker cancelled | No-op |
| Non-PDF file | Show error message |
| Password-protected PDF | Show password dialog |
| Very large PDF (500+ pages) | Paginate thumbnail grid, lazy load |
| Thumbnail generation fails for a page | Show placeholder, allow continue |
| No pages selected | Extract button disabled |
| All pages selected | Works, but show hint: "To keep all pages, use Compress instead" |
| User changes file during extraction | Cancel current operation |
| Extraction fails | Show error, allow retry |
| Save dialog cancelled | Keep result, allow retry save |
| Back button during processing | Confirm: "Cancel extraction?" |

---

## Acceptance Criteria

### File Selection
- [ ] Can select PDF via file picker
- [ ] Can select PDF via drag-and-drop
- [ ] Shows file name, page count, size
- [ ] "Change File" button to select different PDF

### Thumbnail Grid
- [ ] Shows loading state while generating
- [ ] Progress bar during generation
- [ ] All pages displayed as thumbnails
- [ ] Page numbers visible on each thumbnail
- [ ] Selected pages clearly highlighted
- [ ] Scrollable for many pages

### Selection
- [ ] Click to toggle single page
- [ ] Shift+click for range selection
- [ ] "Select All" button works
- [ ] "Deselect All" button works
- [ ] "Odd" button selects odd pages only
- [ ] "Even" button selects even pages only
- [ ] Selection count updates in real-time

### Extraction
- [ ] Button shows selected page count
- [ ] Button disabled until ≥1 selected
- [ ] Progress bar during extraction
- [ ] Collapsible verbose logs
- [ ] Save dialog on completion
- [ ] Success shows output info
- [ ] "Open" button works
- [ ] Can extract more from same PDF
- [ ] Can start over with new file

### General
- [ ] Back button returns home
- [ ] Handles password-protected PDFs
- [ ] Follows system light/dark theme
- [ ] Clear error messages
