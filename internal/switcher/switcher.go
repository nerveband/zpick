package switcher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Target describes what session to switch to after detaching.
type Target struct {
	Action string `json:"action"` // "attach" or "new"
	Name   string `json:"name"`
	Dir    string `json:"dir,omitempty"`
}

// filePath is the switch-target file location.
var filePath string

// defaultPath returns the default switch-target file path.
func defaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "zpick", "switch-target")
}

// path returns the active file path, using the override if set.
func path() string {
	if filePath != "" {
		return filePath
	}
	return defaultPath()
}

// SetPath overrides the switch-target file path (for testing).
func SetPath(p string) {
	filePath = p
}

// Write saves the switch target to disk as JSON.
func Write(t Target) error {
	p := path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("switcher: mkdir: %w", err)
	}
	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("switcher: marshal: %w", err)
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("switcher: write: %w", err)
	}
	return nil
}

// maxAge is the maximum age of a switch-target file before it's considered stale.
const maxAge = 30 * time.Second

// Read reads and deletes the switch target file.
// Returns error if file is missing or stale (>30s old).
func Read() (Target, error) {
	p := path()

	info, err := os.Stat(p)
	if err != nil {
		return Target{}, fmt.Errorf("switcher: %w", err)
	}

	if time.Since(info.ModTime()) > maxAge {
		os.Remove(p)
		return Target{}, fmt.Errorf("switcher: file is stale (older than %v)", maxAge)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return Target{}, fmt.Errorf("switcher: read: %w", err)
	}

	os.Remove(p)

	var t Target
	if err := json.Unmarshal(data, &t); err != nil {
		return Target{}, fmt.Errorf("switcher: unmarshal: %w", err)
	}
	return t, nil
}
