package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hookLine = `# zpick: session launcher
[[ -z "$ZMX_SESSION" ]] && command -v zpick &>/dev/null && eval "$(zpick)"`

const hookMarker = "zpick: session launcher"

// Install adds the zpick hook to the appropriate shell config file.
func Install() error {
	shell := detectShell()
	switch shell {
	case "zsh":
		return installZsh()
	case "bash":
		return installBash()
	default:
		return fmt.Errorf("unsupported shell: %s\nManually add this to your shell config:\n\n%s", shell, hookLine)
	}
}

// Remove removes the zpick hook from the shell config file.
func Remove() error {
	shell := detectShell()
	switch shell {
	case "zsh":
		return removeFromFile(zshrcPath())
	case "bash":
		return removeFromFile(bashrcPath())
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}
	return "unknown"
}

// hasHook checks if the hook is already installed in the given file.
func hasHook(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), hookMarker)
}

// removeFromFile removes the hook lines from a file.
func removeFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	content := string(data)
	if !strings.Contains(content, hookMarker) {
		fmt.Printf("  hook not found in %s\n", path)
		return nil
	}

	// Remove the hook block (comment + command line)
	lines := strings.Split(content, "\n")
	var result []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, hookMarker) {
			skip = true
			continue
		}
		if skip {
			// Skip the next non-empty line (the actual hook command)
			skip = false
			if strings.Contains(line, "zpick") {
				continue
			}
		}
		result = append(result, line)
	}

	if err := os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	fmt.Printf("  removed hook from %s\n", path)
	return nil
}
