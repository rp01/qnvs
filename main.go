package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// CONSTANTS & VERSION
// =============================================================================

const (
	QNVS_DIR_NAME = ".qnvs"
	VERSION       = "1.0.0"
)

// Global flag for insecure mode (skip TLS verification)
var insecureMode = false

// getHTTPClient returns an HTTP client, optionally skipping TLS verification
func getHTTPClient() *http.Client {
	if insecureMode {
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}
	return http.DefaultClient
}

// =============================================================================
// NODE VERSION SWITCHER
// =============================================================================

// QuickNodeVersionSwitcher manages Node.js versions
type QuickNodeVersionSwitcher struct {
	HomeDir     string
	QNVSDir     string
	VersionsDir string
	BinDir      string
	CurrentLink string
}

// NewQuickNodeVersionSwitcher creates a new instance
func NewQuickNodeVersionSwitcher() *QuickNodeVersionSwitcher {
	homeDir := getHomeDir()
	qnvsDir := filepath.Join(homeDir, QNVS_DIR_NAME)

	return &QuickNodeVersionSwitcher{
		HomeDir:     homeDir,
		QNVSDir:     qnvsDir,
		VersionsDir: filepath.Join(qnvsDir, "versions"),
		BinDir:      filepath.Join(qnvsDir, "bin"),
		CurrentLink: filepath.Join(qnvsDir, "current"),
	}
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	return "."
}

// =============================================================================
// INITIALIZATION
// =============================================================================

// Init creates the directory structure and installs the binary
func (qnvs *QuickNodeVersionSwitcher) Init() error {
	dirs := []string{qnvs.QNVSDir, qnvs.VersionsDir, qnvs.BinDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return qnvs.installSelf()
}

// installSelf copies the running executable to ~/.qnvs/bin
func (qnvs *QuickNodeVersionSwitcher) installSelf() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	targetName := "qnvs"
	if runtime.GOOS == "windows" {
		targetName = "qnvs.exe"
	}
	targetPath := filepath.Join(qnvs.BinDir, targetName)

	// Skip if already installed at this location
	if executable == targetPath {
		return nil
	}

	// Windows: Cannot overwrite running executable, rename old one first
	if _, err := os.Stat(targetPath); err == nil {
		oldPath := targetPath + ".old"
		os.Remove(oldPath)
		if err := os.Rename(targetPath, oldPath); err != nil {
			fmt.Printf("⚠️  Warning: Could not move existing binary\n")
		}
	}

	// Copy file
	src, err := os.Open(executable)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	if runtime.GOOS != "windows" {
		os.Chmod(targetPath, 0755)
	}

	fmt.Printf("✅ QNVS installed to %s\n", targetPath)
	return qnvs.showPathSetup()
}

// showPathSetup displays PATH configuration instructions
func (qnvs *QuickNodeVersionSwitcher) showPathSetup() error {
	fmt.Println("\n📋 PATH Setup Instructions")
	fmt.Println(strings.Repeat("─", 40))

	if runtime.GOOS == "windows" {
		fmt.Println("\nFor PowerShell (current session), run:")
		fmt.Printf("  $env:Path = \"%s;%s;\" + $env:Path\n", qnvs.BinDir, qnvs.CurrentLink)
		fmt.Println("\nTo make permanent, add these to your PATH environment variable:")
		fmt.Printf("  %s\n", qnvs.BinDir)
		fmt.Printf("  %s\n", qnvs.CurrentLink)
		fmt.Println("\n💡 Tip: If junctions don't work in your environment, QNVS will")
		fmt.Println("   automatically use shim mode (only bin directory needed in PATH).")
	} else {
		shell := filepath.Base(os.Getenv("SHELL"))
		profile := ".bashrc"
		if shell == "zsh" {
			profile = ".zshrc"
		}

		exportLine := fmt.Sprintf("export PATH=\"$HOME/%s/bin:$HOME/%s/current/bin:$PATH\"",
			QNVS_DIR_NAME, QNVS_DIR_NAME)

		fmt.Printf("\nAdd this to your ~/%s:\n", profile)
		fmt.Printf("  %s\n", exportLine)

		// Try to auto-append
		rcPath := filepath.Join(qnvs.HomeDir, profile)
		if content, err := os.ReadFile(rcPath); err == nil {
			if strings.Contains(string(content), QNVS_DIR_NAME) {
				fmt.Printf("\n✅ Already configured in ~/%s\n", profile)
				return nil
			}
		}

		fmt.Printf("\nAttempting to add automatically... ")
		f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			_, err = f.WriteString(fmt.Sprintf("\n# QNVS - Quick Node Version Switcher\n%s\n", exportLine))
			if err == nil {
				fmt.Println("✅ Done!")
				fmt.Println("👉 Restart your terminal or run: source ~/" + profile)
				return nil
			}
		}
		fmt.Println("Failed. Please add manually.")
	}

	return nil
}

// =============================================================================
// VERSION RESOLUTION
// =============================================================================

// resolveVersion converts version aliases to actual versions
func (qnvs *QuickNodeVersionSwitcher) resolveVersion(requested string) (string, error) {
	fmt.Printf("🔎 Resolving version '%s'...\n", requested)

	resp, err := getHTTPClient().Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return "", fmt.Errorf("failed to fetch version index: %w", err)
	}
	defer resp.Body.Close()

	var versions []struct {
		Version string      `json:"version"`
		Lts     interface{} `json:"lts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", fmt.Errorf("failed to decode version index: %w", err)
	}

	cleanInput := strings.TrimPrefix(strings.ToLower(requested), "v")

	// Handle "latest" or "current"
	if cleanInput == "latest" || cleanInput == "current" {
		fmt.Printf("   → %s\n", versions[0].Version)
		return versions[0].Version, nil
	}

	// Handle "lts"
	if cleanInput == "lts" {
		for _, v := range versions {
			if _, ok := v.Lts.(string); ok {
				fmt.Printf("   → %s (LTS)\n", v.Version)
				return v.Version, nil
			}
		}
		return "", fmt.Errorf("no LTS version found")
	}

	// Exact match (e.g., "18.17.0")
	exactTarget := "v" + cleanInput
	for _, v := range versions {
		if v.Version == exactTarget {
			fmt.Printf("   → %s\n", v.Version)
			return v.Version, nil
		}
	}

	// Prefix match (e.g., "18" matches "v18.x.x")
	prefixTarget := "v" + cleanInput + "."
	for _, v := range versions {
		if strings.HasPrefix(v.Version, prefixTarget) {
			fmt.Printf("   → %s\n", v.Version)
			return v.Version, nil
		}
	}

	return "", fmt.Errorf("version '%s' not found", requested)
}

// =============================================================================
// INSTALL
// =============================================================================

// Install downloads and installs a Node.js version
func (qnvs *QuickNodeVersionSwitcher) Install(requestedVersion string) error {
	// Resolve version
	resolvedVersion, err := qnvs.resolveVersion(requestedVersion)
	if err != nil {
		return err
	}

	version := strings.TrimPrefix(resolvedVersion, "v")
	targetDir := filepath.Join(qnvs.VersionsDir, "v"+version)

	// Check if already installed
	if _, err := os.Stat(targetDir); err == nil {
		fmt.Printf("✅ Node.js v%s is already installed\n", version)
		return nil
	}

	// Determine platform and architecture
	osName := runtime.GOOS
	arch := runtime.GOARCH

	if arch == "amd64" {
		arch = "x64"
	} else if arch == "386" {
		arch = "x86"
	}

	extension := "tar.gz"
	if osName == "windows" {
		osName = "win"
		extension = "zip"
	}

	fileName := fmt.Sprintf("node-v%s-%s-%s.%s", version, osName, arch, extension)
	url := fmt.Sprintf("https://nodejs.org/dist/v%s/%s", version, fileName)

	// Download
	tmpFile := filepath.Join(qnvs.QNVSDir, "temp-"+fileName)
	defer os.Remove(tmpFile)

	fmt.Printf("📥 Downloading Node.js v%s...\n", version)
	if err := downloadFileWithProgress(url, tmpFile); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Extract
	fmt.Println("📦 Extracting...")
	extractTempDir := filepath.Join(qnvs.QNVSDir, "temp-extract-"+version)
	os.RemoveAll(extractTempDir)
	defer os.RemoveAll(extractTempDir)

	if extension == "zip" {
		if err := unzip(tmpFile, extractTempDir); err != nil {
			return fmt.Errorf("extraction failed: %w", err)
		}
	} else {
		if err := untar(tmpFile, extractTempDir); err != nil {
			return fmt.Errorf("extraction failed: %w", err)
		}
	}

	// Find and move the extracted folder
	files, _ := os.ReadDir(extractTempDir)
	var rootFolder string
	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), "node-") {
			rootFolder = filepath.Join(extractTempDir, f.Name())
			break
		}
	}

	if rootFolder == "" {
		rootFolder = extractTempDir
	}

	if err := os.Rename(rootFolder, targetDir); err != nil {
		return fmt.Errorf("failed to move extracted files: %w", err)
	}

	// Fix symlinks on Unix
	if runtime.GOOS != "windows" {
		qnvs.fixNpmSymlinks(targetDir)
	}

	fmt.Printf("✅ Installed Node.js v%s\n", version)
	return nil
}

// fixNpmSymlinks repairs npm/npx symlinks
func (qnvs *QuickNodeVersionSwitcher) fixNpmSymlinks(versionDir string) error {
	binDir := filepath.Join(versionDir, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return nil
	}

	links := map[string]string{
		"npm": "../lib/node_modules/npm/bin/npm-cli.js",
		"npx": "../lib/node_modules/npm/bin/npx-cli.js",
	}

	for name, target := range links {
		linkPath := filepath.Join(binDir, name)
		os.Remove(linkPath)
		if err := os.Symlink(target, linkPath); err != nil {
			return fmt.Errorf("failed to link %s: %w", name, err)
		}
	}

	return nil
}

// =============================================================================
// USE (SWITCH VERSION)
// =============================================================================

// Use switches to a specific Node.js version
func (qnvs *QuickNodeVersionSwitcher) Use(version string) error {
	version = strings.TrimPrefix(version, "v")
	targetDir := filepath.Join(qnvs.VersionsDir, "v"+version)

	// Try exact match first
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// Try fuzzy match
		files, _ := os.ReadDir(qnvs.VersionsDir)
		prefix := "v" + version + "."
		for _, f := range files {
			if strings.HasPrefix(f.Name(), prefix) {
				targetDir = filepath.Join(qnvs.VersionsDir, f.Name())
				version = strings.TrimPrefix(f.Name(), "v")
				break
			}
		}
	}

	// Check if version exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("version v%s is not installed. Run 'qnvs install %s' first", version, version)
	}

	// Create new link
	fmt.Printf("🔄 Switching to v%s...\n", version)

	if runtime.GOOS == "windows" {
		// Windows: Try junction first, fall back to shims if it fails
		if err := qnvs.useWindowsJunction(targetDir); err != nil {
			// Junction failed - use shim-based approach
			fmt.Println("⚠️  Junctions unavailable, using shim mode...")
			if err := qnvs.createWindowsShims(version, targetDir); err != nil {
				return fmt.Errorf("failed to switch version: %w", err)
			}
		}
	} else {
		// Unix: Standard symlink
		// Remove existing symlink first
		if _, err := os.Lstat(qnvs.CurrentLink); err == nil {
			if err := os.Remove(qnvs.CurrentLink); err != nil {
				return fmt.Errorf("failed to remove existing link: %w", err)
			}
		}
		if err := os.Symlink(targetDir, qnvs.CurrentLink); err != nil {
			return fmt.Errorf("symlink failed: %w", err)
		}
	}

	fmt.Printf("✅ Now using Node.js v%s\n", version)

	// Check PATH
	if !strings.Contains(os.Getenv("PATH"), QNVS_DIR_NAME) {
		fmt.Println("⚠️  QNVS is not in your PATH. Run 'qnvs setup' for instructions.")
	}

	return nil
}

// useWindowsJunction attempts to create a directory junction on Windows
func (qnvs *QuickNodeVersionSwitcher) useWindowsJunction(targetDir string) error {
	// Remove existing junction/directory
	if _, err := os.Lstat(qnvs.CurrentLink); err == nil {
		if err := os.Remove(qnvs.CurrentLink); err != nil {
			// Try rmdir for junctions
			cmd := exec.Command("cmd", "/c", "rmdir", qnvs.CurrentLink)
			cmd.Run()
		}
	}

	// Create junction
	cmd := exec.Command("cmd", "/c", "mklink", "/J", qnvs.CurrentLink, targetDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("junction failed: %s: %w", string(output), err)
	}

	return nil
}

// createWindowsShims creates .cmd shim files that redirect to the active Node.js version
func (qnvs *QuickNodeVersionSwitcher) createWindowsShims(version, targetDir string) error {
	// Save current version to a file (for reference)
	versionFile := filepath.Join(qnvs.QNVSDir, "current_version")
	if err := os.WriteFile(versionFile, []byte("v"+version), 0644); err != nil {
		return fmt.Errorf("failed to write version file: %w", err)
	}

	// Define shims to create
	// Map of shim name -> relative path within the Node.js version directory
	shims := map[string]string{
		"node": "node.exe",
		"npm":  "npm.cmd",
		"npx":  "npx.cmd",
	}

	// Check if corepack exists and add it
	corepackPath := filepath.Join(targetDir, "corepack.cmd")
	if _, err := os.Stat(corepackPath); err == nil {
		shims["corepack"] = "corepack.cmd"
	}

	// Create each shim
	for name, executable := range shims {
		shimPath := filepath.Join(qnvs.BinDir, name+".cmd")
		exePath := filepath.Join(targetDir, executable)

		// Check if the executable exists
		if _, err := os.Stat(exePath); os.IsNotExist(err) {
			continue // Skip if executable doesn't exist
		}

		// Create shim content
		// @echo off prevents command echoing
		// %* passes all arguments to the real executable
		shimContent := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", exePath)

		if err := os.WriteFile(shimPath, []byte(shimContent), 0644); err != nil {
			return fmt.Errorf("failed to create shim for %s: %w", name, err)
		}
	}

	return nil
}

// getWindowsCurrentVersion reads the current version from the version file (shim mode)
func (qnvs *QuickNodeVersionSwitcher) getWindowsCurrentVersion() string {
	versionFile := filepath.Join(qnvs.QNVSDir, "current_version")
	if content, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(content))
	}
	return ""
}

// removeWindowsShims removes the shim .cmd files and version file
func (qnvs *QuickNodeVersionSwitcher) removeWindowsShims() {
	shims := []string{"node.cmd", "npm.cmd", "npx.cmd", "corepack.cmd"}
	for _, shim := range shims {
		shimPath := filepath.Join(qnvs.BinDir, shim)
		os.Remove(shimPath)
	}
	// Remove version file
	versionFile := filepath.Join(qnvs.QNVSDir, "current_version")
	os.Remove(versionFile)
}

// =============================================================================
// LIST & CURRENT
// =============================================================================

// List shows all installed versions
func (qnvs *QuickNodeVersionSwitcher) List() error {
	files, err := os.ReadDir(qnvs.VersionsDir)
	if err != nil || len(files) == 0 {
		fmt.Println("📦 No versions installed")
		fmt.Println("   Run 'qnvs install <version>' to install one")
		return nil
	}

	// Try to get current version from junction first, then fall back to shim mode
	currentVersion := ""
	if target, err := filepath.EvalSymlinks(qnvs.CurrentLink); err == nil {
		currentVersion = filepath.Base(target)
	} else if runtime.GOOS == "windows" {
		// Shim mode - read from version file
		currentVersion = qnvs.getWindowsCurrentVersion()
	}

	fmt.Println("📦 Installed Node.js versions:")
	fmt.Println()

	for _, f := range files {
		if f.IsDir() {
			prefix := "   "
			suffix := ""
			if f.Name() == currentVersion {
				prefix = " ▸ "
				suffix = " (current)"
			}
			fmt.Printf("%s%s%s\n", prefix, f.Name(), suffix)
		}
	}

	return nil
}

// Current shows the currently active version
func (qnvs *QuickNodeVersionSwitcher) Current() error {
	version := ""

	// Try junction/symlink first
	if target, err := filepath.EvalSymlinks(qnvs.CurrentLink); err == nil {
		version = filepath.Base(target)
	} else if runtime.GOOS == "windows" {
		// Shim mode - read from version file
		version = qnvs.getWindowsCurrentVersion()
	}

	if version == "" {
		fmt.Println("No version currently selected")
		fmt.Println("Run 'qnvs use <version>' to select one")
		return nil
	}

	fmt.Printf("📍 Current: %s\n", version)
	return nil
}

// =============================================================================
// UNINSTALL
// =============================================================================

// Uninstall removes an installed version
func (qnvs *QuickNodeVersionSwitcher) Uninstall(version string) error {
	version = strings.TrimPrefix(version, "v")
	targetDir := filepath.Join(qnvs.VersionsDir, "v"+version)

	// Try exact match first
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// Try fuzzy match
		files, _ := os.ReadDir(qnvs.VersionsDir)
		prefix := "v" + version + "."
		for _, f := range files {
			if strings.HasPrefix(f.Name(), prefix) {
				targetDir = filepath.Join(qnvs.VersionsDir, f.Name())
				version = strings.TrimPrefix(f.Name(), "v")
				break
			}
		}
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("version v%s is not installed", version)
	}

	// Check if this is the current version (handle both junction and shim modes)
	isCurrent := false
	if currentTarget, err := filepath.EvalSymlinks(qnvs.CurrentLink); err == nil {
		isCurrent = (targetDir == currentTarget)
	} else if runtime.GOOS == "windows" {
		// Shim mode - check version file
		currentVersion := qnvs.getWindowsCurrentVersion()
		isCurrent = (currentVersion == "v"+version)
	}

	if isCurrent {
		// Remove junction if it exists
		os.Remove(qnvs.CurrentLink)
		// Remove shims if in shim mode on Windows
		if runtime.GOOS == "windows" {
			qnvs.removeWindowsShims()
		}
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	fmt.Printf("✅ Uninstalled Node.js v%s\n", version)
	return nil
}

// =============================================================================
// DOWNLOAD WITH PROGRESS
// =============================================================================

// downloadFileWithProgress downloads a file with a Charm progress bar
func downloadFileWithProgress(url string, dest string) error {
	// First, do a HEAD request to get content length
	headResp, err := getHTTPClient().Head(url)
	if err != nil {
		return err
	}
	headResp.Body.Close()
	totalBytes := headResp.ContentLength

	// Create progress bar
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Download with progress
	resp, err := getHTTPClient().Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	var currentBytes int64
	lastPercent := -1

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := f.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			currentBytes += int64(n)

			// Update progress display
			if totalBytes > 0 {
				percent := int(float64(currentBytes) / float64(totalBytes) * 100)
				if percent != lastPercent {
					lastPercent = percent
					progressView := prog.ViewAs(float64(currentBytes) / float64(totalBytes))
					mb := float64(totalBytes) / 1024 / 1024
					currentMb := float64(currentBytes) / 1024 / 1024
					fmt.Printf("\r  %s %.1f/%.1f MB", progressView, currentMb, mb)
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	fmt.Println() // New line after progress
	return nil
}

// =============================================================================
// ARCHIVE UTILITIES
// =============================================================================

func untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
			os.Chmod(target, os.FileMode(header.Mode))
		case tar.TypeSymlink:
			os.MkdirAll(filepath.Dir(target), 0755)
			os.Remove(target)
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Security check
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(fpath), os.ModePerm)

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// =============================================================================
// HELP
// =============================================================================

func printHelp() {
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	cmd := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	flag := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))

	fmt.Println()
	fmt.Println(title.Render("🚀 QNVS - Quick Node Version Switcher v" + VERSION))
	fmt.Println(help.Render("   A fast, lightweight Node.js version manager"))
	fmt.Println(help.Render("   No admin privileges required!"))
	fmt.Println()
	fmt.Println(title.Render("USAGE:"))
	fmt.Printf("   %s                       Launch interactive TUI\n", cmd.Render("qnvs"))
	fmt.Printf("   %s         Install a Node.js version\n", cmd.Render("qnvs install <version>"))
	fmt.Printf("   %s             Switch to an installed version\n", cmd.Render("qnvs use <version>"))
	fmt.Printf("   %s                   List installed versions\n", cmd.Render("qnvs list"))
	fmt.Printf("   %s                Show currently active version\n", cmd.Render("qnvs current"))
	fmt.Printf("   %s       Remove an installed version\n", cmd.Render("qnvs uninstall <version>"))
	fmt.Printf("   %s                  Initialize QNVS and configure PATH\n", cmd.Render("qnvs setup"))
	fmt.Printf("   %s                   Show this help message\n", cmd.Render("qnvs help"))
	fmt.Println()
	fmt.Println(title.Render("FLAGS:"))
	fmt.Printf("   %s              Skip TLS certificate verification\n", flag.Render("--insecure"))
	fmt.Println(help.Render("                         (Use if behind corporate VPN/proxy like Cato, Zscaler)"))
	fmt.Println()
	fmt.Println(title.Render("VERSION FORMATS:"))
	fmt.Println("   22, 20, 18         Latest version of that major release")
	fmt.Println("   22.1.0             Specific version")
	fmt.Println("   lts                Latest LTS version")
	fmt.Println("   latest             Latest available version")
	fmt.Println()
	fmt.Println(title.Render("EXAMPLES:"))
	fmt.Printf("   %s\n", cmd.Render("qnvs install 22"))
	fmt.Printf("   %s\n", cmd.Render("qnvs install lts"))
	fmt.Printf("   %s\n", cmd.Render("qnvs use 20"))
	fmt.Printf("   %s\n", cmd.Render("qnvs list"))
	fmt.Printf("   %s     %s\n", cmd.Render("qnvs install 22 --insecure"), help.Render("# For VPN/proxy issues"))
	fmt.Println()
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	qnvs := NewQuickNodeVersionSwitcher()

	// Parse global flags first
	args := os.Args[1:]
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--insecure" || arg == "-k" {
			insecureMode = true
			if insecureMode {
				fmt.Println("⚠️  Warning: TLS certificate verification disabled")
			}
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// No arguments - launch interactive TUI
	if len(filteredArgs) < 1 {
		RunInteractiveCLI()
		return
	}

	cmd := filteredArgs[0]
	args = filteredArgs[1:]

	switch cmd {
	case "install", "i":
		if len(args) < 1 {
			fmt.Println("❌ Error: version required")
			fmt.Println("Usage: qnvs install <version>")
			fmt.Println("Example: qnvs install 22")
			os.Exit(1)
		}
		if err := qnvs.Init(); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		if err := qnvs.Install(args[0]); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "use", "u":
		if len(args) < 1 {
			fmt.Println("❌ Error: version required")
			fmt.Println("Usage: qnvs use <version>")
			fmt.Println("Example: qnvs use 22")
			os.Exit(1)
		}
		if err := qnvs.Use(args[0]); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "list", "ls", "l":
		if err := qnvs.List(); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "current", "c":
		if err := qnvs.Current(); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "uninstall", "remove", "rm":
		if len(args) < 1 {
			fmt.Println("❌ Error: version required")
			fmt.Println("Usage: qnvs uninstall <version>")
			os.Exit(1)
		}
		if err := qnvs.Uninstall(args[0]); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "setup", "init":
		if err := qnvs.Init(); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "interactive", "tui":
		RunInteractiveCLI()

	case "help", "-h", "--help":
		printHelp()

	case "version", "-v", "--version":
		fmt.Printf("qnvs version %s\n", VERSION)

	default:
		fmt.Printf("❌ Unknown command: %s\n", cmd)
		fmt.Println()
		printHelp()
		os.Exit(1)
	}
}
