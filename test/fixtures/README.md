# Test Fixtures

PDF files for testing dadjoke PDF tools.

## Valid PDFs

| File | Size | Pages | Description |
|------|------|-------|-------------|
| `valid/simple-1page.pdf` | ~50KB | 1 | Simple single-page document |
| `valid/multi-page.pdf` | ~500KB | 10 | Multi-page document for combine tests |
| `valid/high-res-images.pdf` | ~5MB | 5 | High-resolution images for compression tests |
| `valid/large-document.pdf` | ~10MB | 50+ | Large document for stress testing |

## Edge Cases

| File | Description |
|------|-------------|
| `edge-cases/password-protected.pdf` | Password: `testpassword123` |
| `edge-cases/corrupted.pdf` | Truncated/invalid PDF data |
| `edge-cases/empty.pdf` | Zero-byte file |
| `edge-cases/not-a-pdf.pdf` | Plain text file with .pdf extension |

## Downloading Fixtures

Run from project root:

```bash
go run test/download_fixtures.go
```

This will download sample PDFs from public domain sources.

## Sources

- Simple/multi-page: Generated with pdfcpu
- High-res images: NASA public domain images
- Large document: Project Gutenberg public domain books
