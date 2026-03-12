package backend

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var extraSearchDirs = []string{
	"/opt/homebrew/bin",
	"/opt/homebrew/sbin",
	"/usr/local/bin",
	"/usr/local/sbin",
}

// LookPath resolves a command from PATH first, then common install locations
// used before shell startup files have finished extending PATH.
func LookPath(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	for _, candidate := range extraLookupCandidates(name) {
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}
	return "", exec.ErrNotFound
}

// Command builds an exec.Cmd using LookPath, falling back to the raw name if
// no executable is found so the caller gets the usual os/exec error.
func Command(name string, args ...string) *exec.Cmd {
	path, err := LookPath(name)
	if err != nil {
		path = name
	}
	return exec.Command(path, args...)
}

// ShellCommand returns the shortest safe command token for shell eval output.
// If the binary is only discoverable from common fallback locations, it
// returns a quoted absolute path so early autostart shells can still exec it.
func ShellCommand(name string) string {
	if _, err := exec.LookPath(name); err == nil {
		return name
	}
	if path, err := LookPath(name); err == nil {
		return shellQuote(path)
	}
	return name
}

func extraLookupCandidates(name string) []string {
	seen := map[string]struct{}{}
	var candidates []string

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		for _, dir := range []string{
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".cargo", "bin"),
		} {
			candidate := filepath.Join(dir, name)
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			candidates = append(candidates, candidate)
		}
	}

	for _, dir := range extraSearchDirs {
		candidate := filepath.Join(dir, name)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	return candidates
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0o111 != 0
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
