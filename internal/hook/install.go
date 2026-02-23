package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	"github.com/nerveband/zpick/internal/guard"
)

// Old-style markers for migration (all generations)
const hookMarker = "zpick: session launcher"

// Legacy markers from before the rename (v1.x zmosh-picker era, v2.x zpick era)
var legacyMarkers = []string{
	"zmosh-picker: session launcher",
	"zmosh-picker: auto-launch session picker",
	"command zpick", // v2.x zpick hooks that reference the old binary name
}

// Block markers for new guard-based hook
const (
	blockStart = "# >>> zpick guard >>>"
	blockEnd   = "# <<< zpick guard <<<"
)

const termLine = `# zpick: terminal fix — ensures colors work in zmosh sessions
export TERM=xterm-ghostty`

const termMarker = "zpick: terminal fix"

// sessionEnvCheck builds the bash/zsh condition checking all session env vars.
func sessionEnvCheck() string {
	vars := backend.AllSessionEnvVars()
	var parts []string
	for _, v := range vars {
		parts = append(parts, fmt.Sprintf(`-z "$%s"`, v))
	}
	return strings.Join(parts, " && ")
}

// GenerateHookBlock builds the shell hook block from the guard config.
func GenerateHookBlock(apps []string) string {
	var b strings.Builder
	b.WriteString(blockStart)
	b.WriteByte('\n')

	// Autorun: launch saved command when entering a new session
	b.WriteString("# Auto-run: launch saved command when entering a new session\n")
	b.WriteString("if [[ -n \"$ZPICK_AUTORUN\" ]]; then\n")
	b.WriteString("  command zp autorun\n")
	b.WriteString("  unset ZPICK_AUTORUN\n")
	b.WriteString("fi\n")

	// Guard function — checks ALL backend session env vars
	envCheck := sessionEnvCheck()
	b.WriteString("_zpick_guard() {\n")
	fmt.Fprintf(&b, "  if [[ %s ]] && command -v zp &>/dev/null; then\n", envCheck)
	b.WriteString("    local _r\n")
	b.WriteString("    _r=$(command zp guard -- \"$@\")\n")
	b.WriteString("    if [[ -n \"$_r\" ]]; then eval \"$_r\"; return; fi\n")
	b.WriteString("  fi\n")
	b.WriteString("  command \"$@\"\n")
	b.WriteString("}\n")

	// Per-app shell functions
	for _, app := range apps {
		if err := guard.ValidateName(app); err != nil {
			continue
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
	case "fish":
		return installFish()
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
	case "fish":
		return removeFish()
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

// installShell installs the guard block into a shell config file.
func installShell(path string) error {
	guard.EnsureConfig()

	apps, err := guard.ReadConfig()
	if err != nil {
		return fmt.Errorf("cannot read guard config: %w", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	block := GenerateHookBlock(apps)
	if strings.Contains(content, blockStart) {
		content = removeBlock(content)
	}

	// Migration: remove old-style hook marker if present (v2.0+ zpick era)
	if strings.Contains(content, hookMarker) {
		content = removeOldHook(content, hookMarker)
		fmt.Printf("  migrated old hook from %s\n", path)
	}

	// Migration: remove legacy hooks from v1.x zmosh-picker era
	for _, marker := range legacyMarkers {
		if strings.Contains(content, marker) {
			content = removeOldHook(content, marker)
			fmt.Printf("  migrated legacy zmosh-picker hook from %s\n", path)
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

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

	before := content[:startIdx]
	after := content[endIdx+len(blockEnd):]
	if strings.HasPrefix(after, "\n") {
		after = after[1:]
	}
	return before + after
}

// removeOldHook removes an old-style hook (comment + command line) by marker.
func removeOldHook(content string, marker string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, marker) {
			skip = true
			continue
		}
		if skip {
			skip = false
			if strings.Contains(line, "zpick") || strings.Contains(line, "zmosh-picker") {
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

// hasHook checks if any zpick/zmosh-picker hook (any generation) is in the file.
func hasHook(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := string(data)
	if strings.Contains(content, hookMarker) || strings.Contains(content, blockStart) {
		return true
	}
	for _, marker := range legacyMarkers {
		if strings.Contains(content, marker) {
			return true
		}
	}
	return false
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
	hasLegacy := false
	for _, marker := range legacyMarkers {
		if strings.Contains(content, marker) {
			hasLegacy = true
			break
		}
	}

	if !hasOldHook && !hasNewHook && !hasTermMarker && !hasLegacy {
		fmt.Printf("  hook not found in %s\n", path)
		return nil
	}

	if hasNewHook {
		content = removeBlock(content)
	}

	if hasOldHook {
		content = removeOldHook(content, hookMarker)
	}

	for _, marker := range legacyMarkers {
		if strings.Contains(content, marker) {
			content = removeOldHook(content, marker)
		}
	}

	// Remove TERM fix — detect both bash/zsh and fish patterns
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
				if strings.Contains(line, "export TERM=") || strings.Contains(line, "set -gx TERM") {
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
