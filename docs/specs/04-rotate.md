# Rotate Tool Specification

## Overview
Rotate pages in a PDF. Quick action to rotate all pages at once, or per-page control to rotate individual pages differently. User loads PDF, applies rotations, and saves the result.

---

## User Flow

### Happy Path - Rotate All
1. User clicks "Rotate" on home screen
2. App transitions to Rotate view
3. User adds PDF via:
   - Click "Select PDF" → native file picker
   - OR drag-and-drop file onto drop zone
4. Loading state while thumbnails generate
5. View shows:
   - "Rotate All" quick action buttons
   - Thumbnail grid of all pages
6. User clicks "Rotate All 90°" button
7. All thumbnails update to show rotation preview
8. User clicks "Apply & Save"
9. Progress bar during rotation
10. Native save dialog appears
11. Success state with "Open" button

### Happy Path - Per-Page Rotation
1. Steps 1-5 same as above
2. User clicks on individual page thumbnail
3. Page gets selected (highlighted)
4. Rotation buttons appear for selected pages
5. User clicks "90°" to rotate selected pages
6. Selected thumbnails update to show rotation
7. User can select other pages and rotate differently
8. User clicks "Apply & Save" when done
9. Continue from step 9 above

---

## UI Components

### Header
- Back button (←) - top left
- Title: "Rotate PDF"

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

### Loading State
```
┌─────────────────────────────────────────┐
│  📄 my-document.pdf                     │
│     12 pages • 4.2 MB                   │
│                                         │
│  ⟳ Generating page previews...          │
│  ████████░░░░░░░░░░░░  40%              │
└─────────────────────────────────────────┘
```

### Quick Rotate Actions
```
┌─────────────────────────────────────────────────────┐
│  📄 my-document.pdf                                 │
│     12 pages • 4.2 MB                [Change File]  │
│─────────────────────────────────────────────────────│
│                                                     │
│  Rotate All Pages:                                  │
│  [↻ 90°]  [↻ 180°]  [↻ 270°]  [Reset All]          │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Page Grid with Rotation Indicators
```
┌─────────────────────────────────────────────────────┐
│                                                     │
│  Or select pages for individual rotation:           │
│                                                     │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐               │
│  │  1  │  │  2  │  │  3  │  │  4  │               │
│  │     │  │ ↻90 │  │     │  │ ↻90 │               │
│  │ 📄  │  │ 📄↻ │  │ 📄  │  │ 📄↻ │               │
│  └─────┘  └─────┘  └─────┘  └─────┘               │
│           rotated          rotated                  │
│                                                     │
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐               │
│  │✓ 5  │  │✓ 6  │  │  7  │  │  8  │               │
│  │     │  │     │  │     │  │     │               │
│  │ 📄  │  │ 📄  │  │ 📄  │  │ 📄  │               │
│  └─────┘  └─────┘  └─────┘  └─────┘               │
│  selected selected                                  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Selection Actions Bar (when pages selected)
```
┌─────────────────────────────────────────────────────┐
│  2 pages selected                                   │
│                                                     │
│  Rotate selected:  [↻ 90°] [↻ 180°] [↻ 270°]       │
│                                                     │
│  [Clear Selection]                                  │
└─────────────────────────────────────────────────────┘
```

### Pending Changes Indicator
```
┌─────────────────────────────────────────────────────┐
│  ⚠️ 4 pages have pending rotations                  │
│                                                     │
│  [Reset All]           [Apply & Save]               │
└─────────────────────────────────────────────────────┘
```

### Progress Section
```
┌─────────────────────────────────────────┐
│  Applying rotations...                  │
│  ████████████░░░░░░░░  60%              │
│                                         │
│  ▼ Details                              │
│  ┌─────────────────────────────────────┐│
│  │ Rotating page 2 by 90°...           ││
│  │ Rotating page 4 by 90°...           ││
│  │ Rotating page 7 by 180°...          ││
│  └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### Success State
```
┌─────────────────────────────────────────┐
│  ✓ Rotation Complete                    │
│                                         │
│  4 pages rotated                        │
│  Output size: 4.1 MB                    │
│                                         │
│  [Open]  [Rotate Another]               │
└─────────────────────────────────────────┘
```

---

## Go Backend Functions

### Types
```go
type RotationAngle int

const (
    Rotate0   RotationAngle = 0   // no rotation / reset
    Rotate90  RotationAngle = 90
    Rotate180 RotationAngle = 180
    Rotate270 RotationAngle = 270
)

type PDFInfo struct {
    Path      string `json:"path"`
    Name      string `json:"name"`
    PageCount int    `json:"pageCount"`
    Size      int64  `json:"size"`
    SizeText  string `json:"sizeText"`
}

type PageThumbnail struct {
    PageNum   int    `json:"pageNum"`
    Thumbnail string `json:"thumbnail"` // base64 PNG
    Width     int    `json:"width"`
    Height    int    `json:"height"`
}

type PageRotation struct {
    PageNum  int           `json:"pageNum"` // 1-indexed
    Rotation RotationAngle `json:"rotation"`
}

type RotateResult struct {
    Success      bool   `json:"success"`
    PagesRotated int    `json:"pagesRotated"`
    OutputSize   int64  `json:"outputSize"`
    OutputPath   string `json:"outputPath"`
    Error        string `json:"error,omitempty"`
}
```

### Functions
```go
// SelectPDFFile opens native file picker
func (a *App) SelectPDFFile() (*PDFInfo, error)

// LoadPDFInfo gets metadata for a file path
func (a *App) LoadPDFInfo(path string) (*PDFInfo, error)

// GenerateThumbnails creates preview images for all pages
func (a *App) GenerateThumbnails(path string) ([]PageThumbnail, error)

// GenerateRotatedThumbnail creates a preview of how a page will look when rotated
// Used for instant UI preview without modifying the actual PDF
func (a *App) GenerateRotatedThumbnail(path string, pageNum int, rotation RotationAngle) (*PageThumbnail, error)

// ApplyRotations applies all specified rotations and creates new PDF
// Only pages with non-zero rotation are modified
func (a *App) ApplyRotations(path string, rotations []PageRotation) (*RotateResult, error)

// SaveFile opens save dialog
func (a *App) SaveFile(tempPath string, suggestedName string) (string, error)

// OpenFile opens in system viewer
func (a *App) OpenFile(path string) error
```

### Events (Go → Frontend)
```go
// Thumbnail generation progress
runtime.EventsEmit(ctx, "rotate:thumbnail-progress", ProgressUpdate{
    Percent: 50,
    Message: "Generating preview for page 6/12...",
})

// Rotation progress
runtime.EventsEmit(ctx, "rotate:progress", ProgressUpdate{
    Percent: 75,
    Message: "Rotating page 4...",
})

// Log lines
runtime.EventsEmit(ctx, "rotate:log", "Rotating page 2 by 90°...")
```

---

## Frontend ↔ Backend Contract

### State Management (Frontend)
```javascript
const [pdfInfo, setPdfInfo] = useState(null);           // PDFInfo
const [thumbnails, setThumbnails] = useState([]);       // PageThumbnail[]
const [rotations, setRotations] = useState({});         // { [pageNum]: RotationAngle }
const [selectedPages, setSelectedPages] = useState([]); // number[]
const [isLoading, setIsLoading] = useState(false);
const [isProcessing, setIsProcessing] = useState(false);
const [progress, setProgress] = useState(0);
const [logs, setLogs] = useState([]);

// Derived
const hasChanges = Object.values(rotations).some(r => r !== 0);
const changedPageCount = Object.values(rotations).filter(r => r !== 0).length;
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

// Rotate all pages
const rotateAll = (angle) => {
    const newRotations = {};
    thumbnails.forEach(t => {
        newRotations[t.pageNum] = angle;
    });
    setRotations(newRotations);
    // Update thumbnail previews (can be done client-side with CSS transform)
};

// Rotate selected pages
const rotateSelected = (angle) => {
    const newRotations = { ...rotations };
    selectedPages.forEach(pageNum => {
        const current = newRotations[pageNum] || 0;
        newRotations[pageNum] = (current + angle) % 360;
    });
    setRotations(newRotations);
};

// Reset all
const resetAll = () => {
    setRotations({});
    setSelectedPages([]);
};

// Apply and save
const apply = async () => {
    const rotationList = Object.entries(rotations)
        .filter(([_, angle]) => angle !== 0)
        .map(([pageNum, angle]) => ({
            pageNum: parseInt(pageNum),
            rotation: angle
        }));

    setIsProcessing(true);
    const result = await ApplyRotations(pdfInfo.path, rotationList);
    const finalPath = await SaveFile(result.outputPath, "rotated.pdf");
    await OpenFile(finalPath);
};

// Toggle page selection
const togglePage = (pageNum) => {
    setSelectedPages(prev =>
        prev.includes(pageNum)
            ? prev.filter(p => p !== pageNum)
            : [...prev, pageNum]
    );
};
```

### Event Listeners
```javascript
// Thumbnail progress
runtime.EventsOn("rotate:thumbnail-progress", (update) => {
    setProgress(update.percent);
});

// Rotation progress
runtime.EventsOn("rotate:progress", (update) => {
    setProgress(update.percent);
});

// Logs
runtime.EventsOn("rotate:log", (line) => {
    setLogs(prev => [...prev, line]);
});
```

### CSS Rotation Preview
```css
/* Apply rotation preview without regenerating thumbnails */
.thumbnail-90 { transform: rotate(90deg); }
.thumbnail-180 { transform: rotate(180deg); }
.thumbnail-270 { transform: rotate(270deg); }
```

---

## Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| File picker cancelled | No-op |
| Non-PDF file | Show error message |
| Password-protected PDF | Show password dialog |
| No rotations applied | "Apply & Save" disabled, show hint |
| Rotating already-rotated page | Stack rotations (90 + 90 = 180) |
| Reset after making changes | Clear all rotations, confirm if many changes |
| Large PDF (100+ pages) | Paginate grid, show counts |
| Rotation + save fails | Show error, allow retry |
| Save dialog cancelled | Keep changes, allow retry save |
| Back button with pending changes | Confirm: "Discard changes?" |
| Rotate all when some pages already rotated | Override all to same angle |

---

## Acceptance Criteria

### File Selection
- [ ] Can select PDF via file picker
- [ ] Can select PDF via drag-and-drop
- [ ] Shows file name, page count, size
- [ ] "Change File" button to select different PDF
- [ ] Confirm if changing file with pending rotations

### Thumbnail Grid
- [ ] Shows loading state while generating
- [ ] Progress bar during generation
- [ ] All pages displayed as thumbnails
- [ ] Page numbers visible on each thumbnail
- [ ] Rotation indicator shown on rotated pages
- [ ] Thumbnails visually rotate (CSS transform)

### Rotate All
- [ ] "Rotate All 90°" rotates all pages
- [ ] "Rotate All 180°" rotates all pages
- [ ] "Rotate All 270°" rotates all pages
- [ ] "Reset All" clears all rotations
- [ ] Thumbnail previews update immediately

### Per-Page Rotation
- [ ] Click to select individual pages
- [ ] Multi-select supported
- [ ] Selection bar appears when pages selected
- [ ] Can rotate selected pages 90°/180°/270°
- [ ] Each page can have different rotation
- [ ] Rotation stacks (clicking 90° twice = 180°)
- [ ] "Clear Selection" deselects all

### Apply Changes
- [ ] Shows count of pages with pending rotations
- [ ] "Apply & Save" disabled if no changes
- [ ] Progress bar during rotation
- [ ] Collapsible verbose logs
- [ ] Save dialog on completion
- [ ] Success shows output info
- [ ] "Open" button works
- [ ] "Rotate Another" resets for new file

### General
- [ ] Back button returns home (with confirm if changes)
- [ ] Handles password-protected PDFs
- [ ] Follows system light/dark theme
- [ ] Clear error messages
