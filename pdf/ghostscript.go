package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetGhostscriptPath returns the path to the Ghostscript binary.
// It checks for bundled binaries first, then falls back to system PATH.
func GetGhostscriptPath() (string, error) {
	// First, try to find bundled Ghostscript
	bundledPath, err := getBundledGSPath()
	if err == nil {
		if _, statErr := os.Stat(bundledPath); statErr == nil {
			return bundledPath, nil
		}
	}

	// Fall back to system PATH
	gsName := "gs"
	if runtime.GOOS == "windows" {
		gsName = "gswin64c.exe"
	}

	path, err := exec.LookPath(gsName)
	if err != nil {
		return "", fmt.Errorf("ghostscript not found: %w. Please install Ghostscript", err)
	}

	return path, nil
}

// getBundledGSPath returns the expected path for bundled Ghostscript
func getBundledGSPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: .app/Contents/MacOS/binary -> .app/Contents/Resources/gs
		macOSDir := filepath.Dir(exePath)     // Contents/MacOS
		contentsDir := filepath.Dir(macOSDir) // Contents
		resourcesDir := filepath.Join(contentsDir, "Resources")
		return filepath.Join(resourcesDir, "gs"), nil

	case "windows":
		// Windows: {exe_dir}/gs/gswin64c.exe
		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, "gs", "gswin64c.exe"), nil

	case "linux":
		// Linux: use system gs (no bundling)
		return "", fmt.Errorf("linux uses system ghostscript")

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// CheckGhostscriptInstalled verifies Ghostscript is available and returns version info
func CheckGhostscriptInstalled() (string, error) {
	gsPath, err := GetGhostscriptPath()
	if err != nil {
		return "", err
	}

	// Run gs --version to check it works
	cmd := exec.Command(gsPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ghostscript found but failed to run: %w", err)
	}

	return string(output), nil
}

// GhostscriptInstallInstructions returns platform-specific install instructions
func GhostscriptInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Ghostscript with: brew install ghostscript"
	case "windows":
		return "Download Ghostscript from: https://ghostscript.com/releases/gsdnld.html"
	case "linux":
		return "Install Ghostscript with: sudo apt install ghostscript"
	default:
		return "Please install Ghostscript for your platform"
	}
}
