# DadJoke - PDF Toolset

A lightweight, open-source PDF toolset built with [Wails](https://wails.io/) (Go + Svelte).

## Features

- **Compress**: Reduce PDF file size using Ghostscript presets (screen, ebook, printer, prepress)
- **Combine**: Merge multiple PDFs into one, with support for interleaving and page reordering
- **Split**: *(Coming Soon)* Split PDFs by page ranges
- **Rotate**: *(Coming Soon)* Rotate pages within a PDF

## Requirements

### Runtime Dependencies

**Ghostscript** is required for PDF compression:

- **macOS**: `brew install ghostscript`
- **Linux**: `sudo apt install ghostscript`
- **Windows**: Bundled with the installer

### Development

- Go 1.21+
- Node.js 18+ (or Bun)
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Development

```bash
# Install dependencies
cd frontend && bun install && cd ..

# Run in development mode
wails dev

# Run tests
go test -v ./...
```

## Building

```bash
# Build for current platform
wails build

# Build for specific platform
wails build -platform darwin/universal
wails build -platform windows/amd64
wails build -platform linux/amd64
```

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

This license is required because we use [Ghostscript](https://ghostscript.com/), which is also AGPL-licensed.

See [LICENSE](LICENSE) for the full license text.

## Acknowledgments

- [Ghostscript](https://ghostscript.com/) - PDF compression engine (AGPL)
- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - PDF processing library (Apache 2.0)
- [Wails](https://wails.io/) - Go + Web frontend framework (MIT)
