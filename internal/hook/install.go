package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nerveband/zpick/internal/guard"
)

// Old-style marker for migration
const hookMarker = "zpick: session launcher"

// Block markers for new guard-based hook
const (
	blockStart = "# >>> zpick guard >>>"
	blockEnd   = "# <<< zpick guard <<<"
)

const termLine = `# zpick: terminal fix — ensures colors work in zmosh sessions
export TERM=xterm-ghostty`

const termMarker = "zpick: terminal fix"

// GenerateHookBlock builds the shell hook block from the guard config.
func GenerateHookBlock(apps []string) string {
	var b strings.Builder
	b.WriteString(blockStart)
	b.WriteByte('\n')

	// Autorun: launch saved command when entering a new zmosh session
	b.WriteString("# Auto-run: launch saved command when entering a new zmosh session\n")
	b.WriteString("if [[ -n \"$ZPICK_AUTORUN\" ]]; then\n")
	b.WriteString("  command zpick autorun\n")
	b.WriteString("  unset ZPICK_AUTORUN\n")
	b.WriteString("fi\n")

	// Guard function
	b.WriteString("_zpick_guard() {\n")
	b.WriteString("  if [[ -z \"$ZMX_SESSION\" ]] && command -v zpick &>/dev/null; then\n")
	b.WriteString("    local _r\n")
	b.WriteString("    _r=$(command zpick guard -- \"$@\")\n")
	b.WriteString("    if [[ -n \"$_r\" ]]; then eval \"$_r\"; return; fi\n")
	b.WriteString("  fi\n")
	b.WriteString("  command \"$@\"\n")
	b.WriteString("}\n")

	// Per-app shell functions
	for _, app := range apps {
		if err := guard.ValidateName(app); err != nil {
			continue // skip invalid names
		}
		fname := guard.FuncName(app)
		fmt.Fprintf(&b, "%s() { _zpick_guard %s \"$@\"; }\n", fname, app)
	}

	b.WriteString(blockEnd)
	return b.String()
}

// Install adds the zpick guard hook to the appropriate shell config file.
func Install() error {
	shell := detectShell()
	switch shell {
	case "zsh":
		return installShell(zshrcPath())
	case "bash":
		return installShell(bashrcPath())
	default:
		apps, _ := guard.ReadConfig()
		block := GenerateHookBlock(apps)
		return fmt.Errorf("unsupported shell: %s\nManually add this to your shell config:\n\n%s", shell, block)
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

// installShell installs the guard block into a shell config file.
func installShell(path string) error {
	// Ensure config file exists with defaults
	guard.EnsureConfig()

	apps, err := guard.ReadConfig()
	if err != nil {
		return fmt.Errorf("cannot read guard config: %w", err)
	}

	// Read existing file content (or empty if doesn't exist)
	data, _ := os.ReadFile(path)
	content := string(data)

	// Check if new-style block already exists with same content
	block := GenerateHookBlock(apps)
	if strings.Contains(content, blockStart) {
		// Remove existing block first, then re-add (idempotent update)
		content = removeBlock(content)
	}

	// Migration: remove old-style hook marker if present
	if strings.Contains(content, hookMarker) {
		content = removeOldHook(content)
		fmt.Printf("  migrated old hook from %s\n", path)
	}

	// Append new block
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	// Write existing content + new block
	trimmed := strings.TrimRight(content, "\n")
	if trimmed != "" {
		fmt.Fprint(f, trimmed)
		fmt.Fprint(f, "\n\n")
	}
	fmt.Fprintf(f, "%s\n", block)

	fmt.Printf("  installed guard hook in %s\n", path)
	appendTermFix(path)
	return nil
}

// removeBlock removes the guard block from content, returning cleaned content.
func removeBlock(content string) string {
	startIdx := strings.Index(content, blockStart)
	if startIdx < 0 {
		return content
	}

	endIdx := strings.Index(content, blockEnd)
	if endIdx < 0 {
		// Start found but no end — remove only the start line + warn
		lines := strings.Split(content, "\n")
		var result []string
		for _, line := range lines {
			if strings.Contains(line, blockStart) {
				continue
			}
			result = append(result, line)
		}
		fmt.Println("  warning: found start marker but no end marker")
		return strings.Join(result, "\n")
	}

	// Remove everything from start to end (inclusive of the end line)
	before := content[:startIdx]
	after := content[endIdx+len(blockEnd):]
	// Clean up extra newlines
	if strings.HasPrefix(after, "\n") {
		after = after[1:]
	}
	return before + after
}

// removeOldHook removes the old-style hook (comment + command line).
func removeOldHook(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, hookMarker) {
			skip = true
			continue
		}
		if skip {
			skip = false
			if strings.Contains(line, "zpick") {
				continue
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}
	return "unknown"
}

// hasHook checks if any zpick hook (old or new style) is in the file.
func hasHook(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, hookMarker) || strings.Contains(content, blockStart)
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

// removeFromFile removes hook lines (old and new style) and TERM fix from a file.
func removeFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	content := string(data)
	hasOldHook := strings.Contains(content, hookMarker)
	hasNewHook := strings.Contains(content, blockStart)
	hasTermMarker := strings.Contains(content, termMarker)

	if !hasOldHook && !hasNewHook && !hasTermMarker {
		fmt.Printf("  hook not found in %s\n", path)
		return nil
	}

	// Remove new-style block
	if hasNewHook {
		content = removeBlock(content)
	}

	// Remove old-style hook
	if hasOldHook {
		content = removeOldHook(content)
	}

	// Remove TERM fix
	if hasTermMarker {
		lines := strings.Split(content, "\n")
		var result []string
		skip := false
		for _, line := range lines {
			if strings.Contains(line, termMarker) {
				skip = true
				continue
			}
			if skip {
				skip = false
				if strings.Contains(line, "export TERM=") {
					continue
				}
			}
			result = append(result, line)
		}
		content = strings.Join(result, "\n")
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	if hasOldHook || hasNewHook {
		fmt.Printf("  removed hook from %s\n", path)
	}
	if hasTermMarker {
		fmt.Printf("  removed TERM fix from %s\n", path)
	}
	return nil
}
