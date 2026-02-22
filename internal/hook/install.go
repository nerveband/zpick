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

const termLine = `# zpick: terminal fix — ensures colors work in zmosh sessions
export TERM=xterm-ghostty`

const termMarker = "zpick: terminal fix"

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

// isGhostty returns true if the current terminal is Ghostty.
func isGhostty() bool {
	if strings.EqualFold(os.Getenv("TERM_PROGRAM"), "Ghostty") {
		return true
	}
	return strings.Contains(strings.ToLower(os.Getenv("TERM")), "ghostty")
}

// hasTermFix checks if the TERM fix is already in the given file.
func hasTermFix(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), termMarker)
}

// appendTermFix appends the TERM fix block to the given file if running in Ghostty.
func appendTermFix(path string) {
	if !isGhostty() {
		return
	}
	if hasTermFix(path) {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "\n%s\n", termLine)
	fmt.Println("  added TERM=xterm-ghostty (ensures colors work in zmosh sessions)")
}

// removeFromFile removes the hook lines and TERM fix from a file.
func removeFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	content := string(data)
	hasHookMarker := strings.Contains(content, hookMarker)
	hasTermMarker := strings.Contains(content, termMarker)

	if !hasHookMarker && !hasTermMarker {
		fmt.Printf("  hook not found in %s\n", path)
		return nil
	}

	// Remove the hook block and TERM fix block (comment + next line each)
	lines := strings.Split(content, "\n")
	var result []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, hookMarker) {
			skip = true
			continue
		}
		if strings.Contains(line, termMarker) {
			skip = true
			continue
		}
		if skip {
			skip = false
			// Skip the next line (the actual command) if it matches expected content
			if strings.Contains(line, "zpick") || strings.Contains(line, "export TERM=") {
				continue
			}
		}
		result = append(result, line)
	}

	if err := os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	if hasHookMarker {
		fmt.Printf("  removed hook from %s\n", path)
	}
	if hasTermMarker {
		fmt.Printf("  removed TERM fix from %s\n", path)
	}
	return nil
}
