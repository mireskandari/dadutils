.PHONY: dev build build-mac build-windows build-linux clean test

# Development
dev:
	wails dev

# Run tests
test:
	go test -v ./...

# Build for current platform
build:
	wails build

# Build for macOS (universal binary) with Ghostscript bundled
build-mac:
	wails build -platform darwin/universal
	@echo "Bundling Ghostscript for macOS..."
	@mkdir -p build/bin/DadsPDFTools.app/Contents/Resources
	@if [ -f binaries/darwin/gs ]; then \
		cp binaries/darwin/gs build/bin/DadsPDFTools.app/Contents/Resources/gs; \
		chmod +x build/bin/DadsPDFTools.app/Contents/Resources/gs; \
		echo "Ghostscript bundled successfully"; \
	else \
		echo "Warning: binaries/darwin/gs not found. Download it first."; \
	fi

# Build for Windows (requires binaries/windows/gswin64c.exe and gsdll64.dll)
build-windows:
	wails build -platform windows/amd64 -nsis
	@echo "Ghostscript will be bundled by NSIS installer"

# Build for Linux (uses system Ghostscript)
build-linux:
	wails build -platform linux/amd64
	@echo "Linux build complete. Users need: apt install ghostscript"

# Download Ghostscript binaries (manual step)
download-gs:
	@echo "Download Ghostscript binaries manually:"
	@echo ""
	@echo "macOS (Universal):"
	@echo "  Download from: https://pages.uoregon.edu/koch/"
	@echo "  Extract 'gs' binary to: binaries/darwin/gs"
	@echo ""
	@echo "Windows (x64):"
	@echo "  Download from: https://ghostscript.com/releases/gsdnld.html"
	@echo "  Extract to: binaries/windows/gswin64c.exe"
	@echo "            : binaries/windows/gsdll64.dll"

# Clean build artifacts
clean:
	rm -rf build/bin
	rm -rf frontend/dist
