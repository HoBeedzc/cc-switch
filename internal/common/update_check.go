package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultCheckInterval is the default interval between update checks
	DefaultCheckInterval = 24 * time.Hour

	// GitHubAPIURL is the GitHub API endpoint for latest release
	GitHubAPIURL = "https://api.github.com/repos/HoBeedzc/cc-switch/releases/latest"
)

// UpdateCheckCache stores the cached update check result
type UpdateCheckCache struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
	CheckInterval string    `json:"check_interval,omitempty"` // e.g., "24h", "7d"
}

// UpdateCheckResult contains the result of an update check
type UpdateCheckResult struct {
	HasUpdate      bool
	CurrentVersion string
	LatestVersion  string
	Error          error
}

// getUpdateCacheFile returns the path to the update check cache file
// The cache file is stored under profiles/ directory along with other cc-switch data
func getUpdateCacheFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claude", "profiles", ".update_check"), nil
}

// loadUpdateCache loads the cached update check result
func loadUpdateCache() (*UpdateCheckCache, error) {
	cachePath, err := getUpdateCacheFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No cache exists
		}
		return nil, err
	}

	var cache UpdateCheckCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveUpdateCache saves the update check result to cache
func saveUpdateCache(cache *UpdateCheckCache) error {
	cachePath, err := getUpdateCacheFile()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// getCheckInterval parses the check interval from cache or returns default
func getCheckInterval(cache *UpdateCheckCache) time.Duration {
	if cache == nil || cache.CheckInterval == "" {
		return DefaultCheckInterval
	}

	// Parse interval string like "24h", "7d", "1w"
	interval := cache.CheckInterval
	if strings.HasSuffix(interval, "d") {
		days := 1
		fmt.Sscanf(interval, "%dd", &days)
		return time.Duration(days) * 24 * time.Hour
	}
	if strings.HasSuffix(interval, "w") {
		weeks := 1
		fmt.Sscanf(interval, "%dw", &weeks)
		return time.Duration(weeks) * 7 * 24 * time.Hour
	}

	d, err := time.ParseDuration(interval)
	if err != nil {
		return DefaultCheckInterval
	}
	return d
}

// ShouldCheckUpdate determines if we should check for updates
func ShouldCheckUpdate() bool {
	cache, err := loadUpdateCache()
	if err != nil || cache == nil {
		return true // No cache or error, should check
	}

	interval := getCheckInterval(cache)
	return time.Since(cache.LastCheck) > interval
}

// CheckUpdateBackground performs an update check in the background
// It returns immediately and the check runs asynchronously
// The callback is called with the result (can be nil for no callback)
func CheckUpdateBackground(callback func(UpdateCheckResult)) {
	go func() {
		result := checkUpdateSync()
		if callback != nil {
			callback(result)
		}
	}()
}

// checkUpdateSync performs a synchronous update check
func checkUpdateSync() UpdateCheckResult {
	result := UpdateCheckResult{
		CurrentVersion: Version,
	}

	// Fetch latest version from GitHub
	latestVersion, err := fetchLatestVersion()
	if err != nil {
		result.Error = err
		return result
	}

	result.LatestVersion = latestVersion

	// Save to cache regardless of result
	cache := &UpdateCheckCache{
		LastCheck:     time.Now(),
		LatestVersion: latestVersion,
	}
	saveUpdateCache(cache) // Ignore error, cache is optional

	// Compare versions
	result.HasUpdate = IsNewerVersion(latestVersion, Version)

	return result
}

// fetchLatestVersion fetches the latest version from GitHub API
func fetchLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second, // Short timeout to not block
	}

	req, err := http.NewRequest("GET", GitHubAPIURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "cc-switch/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

// IsNewerVersion compares two semantic versions
// Returns true if v1 > v2
func IsNewerVersion(v1, v2 string) bool {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	// Pad to same length
	for len(v1Parts) < 3 {
		v1Parts = append(v1Parts, "0")
	}
	for len(v2Parts) < 3 {
		v2Parts = append(v2Parts, "0")
	}

	for i := 0; i < 3; i++ {
		var n1, n2 int
		fmt.Sscanf(v1Parts[i], "%d", &n1)
		fmt.Sscanf(v2Parts[i], "%d", &n2)

		if n1 > n2 {
			return true
		}
		if n1 < n2 {
			return false
		}
	}

	return false
}

// GetCachedUpdateInfo returns cached update info without making a network request
func GetCachedUpdateInfo() *UpdateCheckResult {
	cache, err := loadUpdateCache()
	if err != nil || cache == nil {
		return nil
	}

	if cache.LatestVersion == "" {
		return nil
	}

	return &UpdateCheckResult{
		CurrentVersion: Version,
		LatestVersion:  cache.LatestVersion,
		HasUpdate:      IsNewerVersion(cache.LatestVersion, Version),
	}
}

// PrintUpdateNotice prints an update notice if a new version is available
func PrintUpdateNotice() {
	info := GetCachedUpdateInfo()
	if info != nil && info.HasUpdate {
		fmt.Fprintf(os.Stderr, "\n💡 New version available: %s → %s\n", info.CurrentVersion, info.LatestVersion)
		fmt.Fprintf(os.Stderr, "   Run 'cc-switch update' to upgrade.\n\n")
	}
}

// ClearUpdateCache removes the update check cache file
// This should be called after a successful update
func ClearUpdateCache() error {
	cachePath, err := getUpdateCacheFile()
	if err != nil {
		return err
	}

	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// SaveUpdateCache saves the latest version to cache
// This can be called by the update command after checking
func SaveUpdateCache(latestVersion string) error {
	cache := &UpdateCheckCache{
		LastCheck:     time.Now(),
		LatestVersion: latestVersion,
	}
	return saveUpdateCache(cache)
}

// TriggerBackgroundCheck triggers a background update check if needed
// This is meant to be called at the start of command execution
// It will not block and will update the cache for future notice display
func TriggerBackgroundCheck() {
	if !ShouldCheckUpdate() {
		return
	}

	// Run check in background, don't wait for result
	go func() {
		checkUpdateSync()
	}()
}
