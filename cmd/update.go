package cmd

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"cc-switch/internal/common"

	"github.com/spf13/cobra"
)

const (
	githubAPIURL  = "https://api.github.com/repos/HoBeedzc/cc-switch/releases/latest"
	githubRepoURL = "https://github.com/HoBeedzc/cc-switch"
)

// GitHubRelease represents the GitHub release API response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and update cc-switch",
	Long: `Check for new versions of cc-switch and optionally update to the latest version.

Modes:
  cc-switch update          Check for updates and prompt for confirmation
  cc-switch update -y       Check and automatically update without prompting
  cc-switch update -c       Only check for updates, don't update

The update is downloaded from GitHub Releases and replaces the current executable.`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolP("yes", "y", false, "Automatically update without prompting")
	updateCmd.Flags().BoolP("check", "c", false, "Only check for updates, don't update")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Skip the automatic update notice for update command
	skipUpdateNotice = true

	yesFlag, _ := cmd.Flags().GetBool("yes")
	checkFlag, _ := cmd.Flags().GetBool("check")

	// Validate flag combinations
	if yesFlag && checkFlag {
		return fmt.Errorf("cannot use -y/--yes and -c/--check together")
	}

	// Get current version
	currentVersion := common.Version

	// Fetch latest release info
	fmt.Println("🔍 Checking for updates...")
	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// Parse latest version (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Save to cache for future background checks
	common.SaveUpdateCache(latestVersion)

	// Display version info
	fmt.Printf("\n📦 Current version: %s\n", currentVersion)
	fmt.Printf("🚀 Latest version:  %s\n", latestVersion)

	// Compare versions
	if currentVersion == latestVersion {
		fmt.Println("\n✅ You are already using the latest version!")
		return nil
	}

	// Check if update is available
	if isNewerVersion(latestVersion, currentVersion) {
		fmt.Printf("\n🎉 A new version is available: %s → %s\n", currentVersion, latestVersion)
	} else {
		fmt.Println("\n✅ You are using a newer or development version.")
		return nil
	}

	// If check-only mode, stop here
	if checkFlag {
		fmt.Printf("\nRun 'cc-switch update' or 'cc-switch update -y' to update.\n")
		return nil
	}

	// Prompt for confirmation unless -y flag is set
	if !yesFlag {
		fmt.Print("\nDo you want to update? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("Update cancelled.")
			return nil
		}
	}

	// Perform update
	return performUpdate(release)
}

func fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "cc-switch/"+common.Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

func performUpdate(release *GitHubRelease) error {
	// Determine platform and architecture
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map to release asset naming
	var assetName string
	switch {
	case goos == "darwin" && goarch == "amd64":
		assetName = "cc-switch-darwin-x64.tar.gz"
	case goos == "darwin" && goarch == "arm64":
		assetName = "cc-switch-darwin-arm64.tar.gz"
	case goos == "linux" && goarch == "amd64":
		assetName = "cc-switch-linux-x64.tar.gz"
	case goos == "linux" && goarch == "arm64":
		assetName = "cc-switch-linux-arm64.tar.gz"
	case goos == "windows" && goarch == "amd64":
		assetName = "cc-switch-windows-x64.zip"
	default:
		return fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}

	// Find download URL
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("could not find release asset for %s", assetName)
	}

	fmt.Printf("\n📥 Downloading %s...\n", assetName)

	// Get current executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get real path
	realPath, err := filepath.EvalSymlinks(executable)
	if err != nil {
		realPath = executable
	}

	// Check write permission
	if err := checkWritePermission(realPath); err != nil {
		return err
	}

	// Create temp directory for download
	tempDir, err := os.MkdirTemp("", "cc-switch-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the archive
	archivePath := filepath.Join(tempDir, assetName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	fmt.Println("📦 Extracting...")

	// Extract binary
	var binaryPath string
	if strings.HasSuffix(assetName, ".tar.gz") {
		binaryPath, err = extractTarGz(archivePath, tempDir)
	} else if strings.HasSuffix(assetName, ".zip") {
		binaryPath, err = extractZip(archivePath, tempDir)
	} else {
		return fmt.Errorf("unknown archive format: %s", assetName)
	}

	if err != nil {
		return fmt.Errorf("failed to extract update: %w", err)
	}

	fmt.Println("🔄 Installing update...")

	// Replace the current executable
	if err := replaceExecutable(realPath, binaryPath); err != nil {
		return err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Clear update cache after successful update
	if err := common.ClearUpdateCache(); err != nil {
		// Non-fatal, just log it
		fmt.Fprintf(os.Stderr, "   Note: Could not clear update cache: %v\n", err)
	}

	fmt.Printf("\n✅ Successfully updated to version %s!\n", latestVersion)
	fmt.Println("   Run 'cc-switch --version' to verify.")

	return nil
}

func checkWritePermission(path string) error {
	// Try to open the file for writing
	dir := filepath.Dir(path)

	// Check if we can write to the directory
	testFile := filepath.Join(dir, ".cc-switch-update-test")
	f, err := os.Create(testFile)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: cannot write to %s\nTry running with sudo: sudo cc-switch update", dir)
		}
		return fmt.Errorf("cannot write to %s: %w", dir, err)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var binaryPath string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the cc-switch binary
		if header.Typeflag == tar.TypeReg && (header.Name == "cc-switch" || filepath.Base(header.Name) == "cc-switch") {
			binaryPath = filepath.Join(destDir, "cc-switch-new")
			outFile, err := os.Create(binaryPath)
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()

			// Set executable permission
			if err := os.Chmod(binaryPath, 0755); err != nil {
				return "", err
			}
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("cc-switch binary not found in archive")
	}

	return binaryPath, nil
}

func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var binaryPath string
	for _, f := range r.File {
		// Look for the cc-switch.exe binary
		if f.Name == "cc-switch.exe" || filepath.Base(f.Name) == "cc-switch.exe" {
			binaryPath = filepath.Join(destDir, "cc-switch-new.exe")

			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			outFile, err := os.Create(binaryPath)
			if err != nil {
				rc.Close()
				return "", err
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				outFile.Close()
				rc.Close()
				return "", err
			}
			outFile.Close()
			rc.Close()
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("cc-switch.exe binary not found in archive")
	}

	return binaryPath, nil
}

func replaceExecutable(oldPath, newPath string) error {
	// On Unix systems, we can replace a running executable
	// On Windows, we need a different approach

	if runtime.GOOS == "windows" {
		return replaceExecutableWindows(oldPath, newPath)
	}

	// Unix: atomic rename
	// First, create backup
	backupPath := oldPath + ".backup"
	if err := os.Rename(oldPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Move new binary to target location
	if err := os.Rename(newPath, oldPath); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, oldPath)
		return fmt.Errorf("failed to install new version: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	// Ensure executable permission
	if err := os.Chmod(oldPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

func replaceExecutableWindows(oldPath, newPath string) error {
	// On Windows, we can't replace a running executable directly
	// We rename the current one and copy the new one

	backupPath := oldPath + ".old"

	// Remove old backup if exists
	os.Remove(backupPath)

	// Rename current executable to .old
	if err := os.Rename(oldPath, backupPath); err != nil {
		return fmt.Errorf("failed to rename current executable: %w\nTry closing all cc-switch processes and run again", err)
	}

	// Copy new executable
	if err := copyFile(newPath, oldPath); err != nil {
		// Try to restore
		os.Rename(backupPath, oldPath)
		return fmt.Errorf("failed to install new version: %w", err)
	}

	// Note: We leave the .old file, user can delete it manually
	// or it will be cleaned up on next update
	fmt.Printf("   Note: Old version saved as %s (can be deleted)\n", backupPath)

	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// isNewerVersion compares two semantic versions
// Returns true if v1 > v2
func isNewerVersion(v1, v2 string) bool {
	return common.IsNewerVersion(v1, v2)
}

