package cmds

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
	PublishedAt time.Time `json:"published_at"`
}

// UpdateCommand returns the update command configuration
func UpdateCommand() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"up"},
		Usage:   "Update FamStack to the latest version",
		Subcommands: []*cli.Command{
			{
				Name:   "check",
				Usage:  "Check for available updates",
				Action: checkUpdate,
			},
			{
				Name:   "install",
				Usage:  "Install the latest version",
				Action: installUpdate,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Force update even if already on latest version",
					},
				},
			},
			{
				Name:   "version",
				Usage:  "Show current version",
				Action: showVersion,
			},
		},
	}
}

// Version is set at build time with -ldflags
var Version = "development"

// getCurrentVersion returns the current version
func getCurrentVersion() string {
	return Version
}

// checkUpdate checks for available updates
func checkUpdate(c *cli.Context) error {
	fmt.Println("Checking for updates...")

	currentVersion := getCurrentVersion()
	fmt.Printf("Current version: %s\n", currentVersion)

	latest, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Printf("Latest version: %s\n", latest.TagName)
	fmt.Printf("Released: %s\n", latest.PublishedAt.Format("2006-01-02 15:04:05"))

	if currentVersion == "development" {
		fmt.Println("⚠️  Running development version - update available")
		return nil
	}

	if latest.TagName != currentVersion {
		fmt.Printf("🆙 Update available: %s → %s\n", currentVersion, latest.TagName)
		fmt.Println("Run 'famstack update install' to update")
	} else {
		fmt.Println("✅ You're running the latest version")
	}

	return nil
}

// installUpdate installs the latest version
func installUpdate(c *cli.Context) error {
	force := c.Bool("force")

	fmt.Println("Installing latest version...")

	currentVersion := getCurrentVersion()

	latest, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	if !force && currentVersion == latest.TagName {
		fmt.Println("✅ Already running the latest version")
		return nil
	}

	// Find the appropriate asset for this platform
	asset, err := findAssetForPlatform(latest)
	if err != nil {
		return fmt.Errorf("failed to find asset for platform: %w", err)
	}

	fmt.Printf("Downloading %s...\n", asset.Name)

	// Download the asset
	resp, err := http.Get(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read download: %w", err)
	}

	// Verify checksum
	fmt.Println("Verifying checksum...")
	if checksumErr := verifyChecksum(latest, asset.Name, bodyBytes); checksumErr != nil {
		return fmt.Errorf("checksum verification failed: %w", checksumErr)
	}
	fmt.Println("✅ Checksum verified")

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create backup
	backupPath := execPath + ".backup"
	if err := copyFile(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	fmt.Println("Installing new version...")

	// Extract and install the new binary
	if err := extractAndInstall(bytes.NewReader(bodyBytes), execPath, asset.Name); err != nil {
		// Restore backup on failure
		if restoreErr := os.Rename(backupPath, execPath); restoreErr != nil {
			return fmt.Errorf("failed to install and failed to restore backup: install error: %w, restore error: %v", err, restoreErr)
		}
		return fmt.Errorf("failed to install: %w", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	fmt.Printf("✅ Successfully updated to %s\n", latest.TagName)
	fmt.Println("Restart FamStack to use the new version")

	return nil
}

// showVersion displays the current version
func showVersion(c *cli.Context) error {
	version := getCurrentVersion()
	fmt.Printf("FamStack version: %s\n", version)
	return nil
}

// getLatestRelease fetches the latest release from GitHub
func getLatestRelease() (*GitHubRelease, error) {
	url := "https://api.github.com/repos/chrisrob11/famstack/releases/latest"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found - check if GitHub releases are published")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// findAssetForPlatform finds the appropriate asset for the current platform
func findAssetForPlatform(release *GitHubRelease) (*struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}, error) {
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch names to common naming conventions
	if arch == "amd64" {
		arch = "amd64"
	}

	// Look for matching asset
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, platform) && strings.Contains(name, arch) {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no asset found for platform %s/%s", platform, arch)
}

// extractAndInstall extracts the binary from the downloaded archive and installs it
func extractAndInstall(reader io.Reader, targetPath, assetName string) error {
	if strings.HasSuffix(assetName, ".tar.gz") {
		return extractTarGz(reader, targetPath)
	} else if strings.HasSuffix(assetName, ".exe") {
		// For Windows .exe files, just copy directly
		return copyFromReader(reader, targetPath)
	}

	return fmt.Errorf("unsupported asset format: %s", assetName)
}

// extractTarGz extracts a tar.gz file and finds the binary
func extractTarGz(reader io.Reader, targetPath string) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for the main binary (usually named "famstack" or similar)
		if header.Typeflag == tar.TypeReg &&
			(strings.Contains(header.Name, "famstack") && !strings.Contains(header.Name, ".")) {

			// Create temporary file
			tempPath := targetPath + ".tmp"
			tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}

			// Copy binary content
			_, err = io.Copy(tempFile, tarReader)
			tempFile.Close()
			if err != nil {
				os.Remove(tempPath)
				return err
			}

			// Replace the original
			return os.Rename(tempPath, targetPath)
		}
	}

	return fmt.Errorf("binary not found in archive")
}

// copyFromReader copies data from reader to target file
func copyFromReader(reader io.Reader, targetPath string) error {
	tempPath := targetPath + ".tmp"
	tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(tempFile, reader)
	tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		return err
	}

	return os.Rename(tempPath, targetPath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// verifyChecksum downloads checksums.txt and verifies the binary checksum
func verifyChecksum(release *GitHubRelease, assetName string, data []byte) error {
	// Find checksums.txt in the release assets
	var checksumsURL string
	for _, asset := range release.Assets {
		if asset.Name == "checksums.txt" {
			checksumsURL = asset.BrowserDownloadURL
			break
		}
	}

	if checksumsURL == "" {
		return fmt.Errorf("checksums.txt not found in release assets")
	}

	// Download checksums.txt
	resp, err := http.Get(checksumsURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksums.txt: %s", resp.Status)
	}

	checksumsContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksums.txt: %w", err)
	}

	// Parse checksums.txt to find our asset
	expectedChecksum, err := parseChecksum(string(checksumsContent), assetName)
	if err != nil {
		return fmt.Errorf("failed to find checksum for %s: %w", assetName, err)
	}

	// Calculate SHA256 of downloaded data
	hasher := sha256.New()
	hasher.Write(data)
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Compare checksums
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// parseChecksum extracts the expected checksum for a specific file from checksums.txt
func parseChecksum(checksumsContent, filename string) (string, error) {
	lines := strings.Split(checksumsContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "checksum  filename" (two spaces between checksum and filename)
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == filename {
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum not found for file: %s", filename)
}
