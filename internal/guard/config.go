package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// validName matches valid shell function names (letters, digits, underscores, hyphens).
var validName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// DefaultApps are the apps guarded by default.
var DefaultApps = []string{"claude", "codex", "opencode"}

// ConfigPath returns the path to the guard config file.
func ConfigPath() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "zpick", "guard.conf")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "zpick", "guard.conf")
}

// ReadConfig reads the guard config file and returns the list of guarded app names.
// Returns DefaultApps if the file does not exist.
func ReadConfig() ([]string, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultApps, nil
		}
		return nil, fmt.Errorf("cannot read guard config: %w", err)
	}
	return parseConfig(string(data)), nil
}

// parseConfig extracts app names from config content, skipping comments and blanks.
func parseConfig(content string) []string {
	var apps []string
	seen := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !seen[line] {
			apps = append(apps, line)
			seen[line] = true
		}
	}
	return apps
}

// WriteConfig writes the app list to the config file, creating directories as needed.
func WriteConfig(apps []string) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	// Deduplicate
	seen := map[string]bool{}
	var deduped []string
	for _, app := range apps {
		if !seen[app] {
			deduped = append(deduped, app)
			seen[app] = true
		}
	}

	var buf strings.Builder
	buf.WriteString("# Apps guarded by zpick (one per line)\n")
	for _, app := range deduped {
		buf.WriteString(app)
		buf.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(buf.String()), 0644)
}

// ValidateName checks if a name is valid for use as a guarded app.
func ValidateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid app name %q: must match [a-zA-Z][a-zA-Z0-9_-]*", name)
	}
	return nil
}

// AddApp adds an app to the config. Returns an error if the name is invalid.
func AddApp(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	apps, err := ReadConfig()
	if err != nil {
		return err
	}
	for _, a := range apps {
		if a == name {
			return fmt.Errorf("%q is already guarded", name)
		}
	}
	apps = append(apps, name)
	return WriteConfig(apps)
}

// RemoveApp removes an app from the config.
func RemoveApp(name string) error {
	apps, err := ReadConfig()
	if err != nil {
		return err
	}
	var filtered []string
	found := false
	for _, a := range apps {
		if a == name {
			found = true
			continue
		}
		filtered = append(filtered, a)
	}
	if !found {
		return fmt.Errorf("%q is not in the guard list", name)
	}
	return WriteConfig(filtered)
}

// FuncName converts an app name to a valid shell function name.
// Hyphens are replaced with underscores.
func FuncName(app string) string {
	return strings.ReplaceAll(app, "-", "_")
}

// EnsureConfig creates the config file with defaults if it doesn't exist.
func EnsureConfig() error {
	path := ConfigPath()
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	return WriteConfig(DefaultApps)
}
