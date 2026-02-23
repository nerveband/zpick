package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/creativeprojects/go-selfupdate"
)

const (
	repoOwner = "nerveband"
	repoName  = "zpick"
	configDir = ".zpick"
	cacheFile = "update_cache.json"
)

// Cache stores the last version check result.
type Cache struct {
	LastCheck      time.Time `json:"last_check"`
	LatestVersion  string    `json:"latest_version"`
	UpdateRequired bool      `json:"update_required"`
}

// CheckResult holds the result of an update check.
type CheckResult struct {
	HasUpdate     bool
	LatestVersion string
	Err           error
}

// CheckAsync runs an update check in the background.
// Returns a channel that will receive the result.
func CheckAsync(currentVersion string) <-chan CheckResult {
	ch := make(chan CheckResult, 1)
	go func() {
		hasUpdate, latestVersion, err := Check(currentVersion)
		ch <- CheckResult{
			HasUpdate:     hasUpdate,
			LatestVersion: latestVersion,
			Err:           err,
		}
	}()
	return ch
}

// Check checks if a new version is available (with 24h cache).
func Check(currentVersion string) (hasUpdate bool, latestVersion string, err error) {
	if currentVersion == "dev" {
		return false, "", nil
	}

	// Check cache first
	cached, err := loadCache()
	if err == nil && time.Since(cached.LastCheck) < 24*time.Hour {
		return cached.UpdateRequired, cached.LatestVersion, nil
	}

	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return false, "", err
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return false, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	latest, found, err := updater.DetectLatest(ctx, selfupdate.NewRepositorySlug(repoOwner, repoName))
	if err != nil || !found {
		return false, "", err
	}

	hasUpdate = latest.GreaterThan(currentVersion)
	latestVer := latest.Version()

	saveCache(Cache{
		LastCheck:      time.Now(),
		LatestVersion:  latestVer,
		UpdateRequired: hasUpdate,
	})

	return hasUpdate, latestVer, nil
}

// Upgrade downloads and installs the latest version.
func Upgrade(currentVersion string) error {
	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Println("Checking for updates...")

	if currentVersion == "dev" {
		fmt.Println("Running dev build — use 'go install' or 'make install' to update.")
		return nil
	}

	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return fmt.Errorf("failed to create update source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.NewRepositorySlug(repoOwner, repoName))
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !found {
		fmt.Println("No releases found")
		return nil
	}

	if latest.LessOrEqual(currentVersion) {
		fmt.Printf("Already up to date (latest: %s)\n", latest.Version())
		return nil
	}

	fmt.Printf("New version available: %s\n", latest.Version())
	fmt.Printf("Downloading for %s/%s...\n", runtime.GOOS, runtime.GOARCH)

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Printf("Successfully upgraded to %s\n", latest.Version())

	// Clear cache so next check sees new version
	saveCache(Cache{
		LastCheck:      time.Now(),
		LatestVersion:  latest.Version(),
		UpdateRequired: false,
	})

	return nil
}

// FormatNotice returns a formatted update notification string, or empty if no update.
func FormatNotice(result CheckResult) string {
	if result.Err != nil || !result.HasUpdate {
		return ""
	}
	return fmt.Sprintf("\nUpdate available: %s\nRun 'zp upgrade' to update\n\n", result.LatestVersion)
}

func cachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configDir, cacheFile)
}

func loadCache() (*Cache, error) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return nil, err
	}
	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func saveCache(cache Cache) error {
	dir := filepath.Dir(cachePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(), data, 0644)
}
