package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	"github.com/nerveband/zpick/internal/guard"
)

// fishConfigPath returns the path to the fish hook file.
// Uses conf.d/ for cleaner fish config, respects XDG_CONFIG_HOME.
func fishConfigPath() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "fish", "conf.d", "zp.fish")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fish", "conf.d", "zp.fish")
}

// fishSessionEnvCheck builds the fish condition checking all session env vars.
func fishSessionEnvCheck() string {
	vars := backend.AllSessionEnvVars()
	var parts []string
	for _, v := range vars {
		parts = append(parts, fmt.Sprintf(`test -z "$%s"`, v))
	}
	return strings.Join(parts, "; and ")
}

// GenerateFishHookBlock builds the fish shell hook block.
func GenerateFishHookBlock(apps []string) string {
	var b strings.Builder
	b.WriteString(blockStart)
	b.WriteByte('\n')

	// Autorun
	b.WriteString("# Auto-run: launch saved command when entering a new session\n")
	b.WriteString("if test -n \"$ZPICK_AUTORUN\"\n")
	b.WriteString("  command zp autorun\n")
	b.WriteString("  set -e ZPICK_AUTORUN\n")
	b.WriteString("end\n")

	// Guard function
	envCheck := fishSessionEnvCheck()
	b.WriteString("function _zpick_guard\n")
	fmt.Fprintf(&b, "  if %s; and command -v zp &>/dev/null\n", envCheck)
	b.WriteString("    set -l _r (command zp guard -- $argv)\n")
	b.WriteString("    if test -n \"$_r\"\n")
	b.WriteString("      eval $_r\n")
	b.WriteString("      return\n")
	b.WriteString("    end\n")
	b.WriteString("  end\n")
	b.WriteString("  command $argv\n")
	b.WriteString("end\n")

	// Per-app functions
	for _, app := range apps {
		if err := guard.ValidateName(app); err != nil {
			continue
		}
		fname := guard.FuncName(app)
		fmt.Fprintf(&b, "function %s\n  _zpick_guard %s $argv\nend\n", fname, app)
	}

	b.WriteString(blockEnd)
	return b.String()
}

// installFish installs the fish hook to conf.d/.
func installFish() error {
	guard.EnsureConfig()

	apps, err := guard.ReadConfig()
	if err != nil {
		return fmt.Errorf("cannot read guard config: %w", err)
	}

	path := fishConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("cannot create %s: %w", filepath.Dir(path), err)
	}

	block := GenerateFishHookBlock(apps)

	if err := os.WriteFile(path, []byte(block+"\n"), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	fmt.Printf("  installed fish hook in %s\n", path)

	// TERM fix for Ghostty
	if isGhostty() && !hasTermFixFish(path) {
		appendTermFixFish(path)
	}

	return nil
}

// removeFish removes the fish hook file.
func removeFish() error {
	path := fishConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("  fish hook not found at %s\n", path)
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("cannot remove %s: %w", path, err)
	}

	fmt.Printf("  removed fish hook from %s\n", path)
	return nil
}

const fishTermLine = `# zpick: terminal fix — ensures colors work in zmosh sessions
set -gx TERM xterm-ghostty`

func hasTermFixFish(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "zpick: terminal fix")
}

func appendTermFixFish(path string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "\n%s\n", fishTermLine)
	fmt.Println("  added TERM=xterm-ghostty for fish (ensures colors work in sessions)")
}
