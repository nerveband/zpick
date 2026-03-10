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
// Includes ZPICK_SESSION (zpick's own marker) in addition to backend-specific vars.
func fishSessionEnvCheck() string {
	vars := append(backend.AllSessionEnvVars(), "ZPICK_SESSION")
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

	// Picker function: eval the command zp outputs
	b.WriteString("function zp\n  eval (command zp)\nend\n")

	// Autorun
	b.WriteString("# Auto-run: launch saved command when entering a new session\n")
	b.WriteString("if test -n \"$ZPICK_AUTORUN\"\n")
	b.WriteString("  command zp autorun\n")
	b.WriteString("  set -e ZPICK_AUTORUN\n")
	b.WriteString("end\n")

	// Switch-target: resume after in-session detach
	b.WriteString("if test -f \"$HOME/.cache/zpick/switch-target\"\n")
	b.WriteString("  eval (command zp resume)\n")
	b.WriteString("end\n")

	// Guard function + per-app wrappers (optional — only if apps configured)
	if len(apps) > 0 {
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

		for _, app := range apps {
			if err := guard.ValidateName(app); err != nil {
				continue
			}
			fname := guard.FuncName(app)
			fmt.Fprintf(&b, "function %s\n  _zpick_guard %s $argv\nend\n", fname, app)
		}
	}

	b.WriteString(blockEnd)
	return b.String()
}

// installFish installs the fish hook to conf.d/.
// Guard wrappers are only included when withGuard is true.
func installFish(withGuard bool) error {
	var apps []string
	if withGuard {
		apps, _ = guard.ReadConfig()
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

	printInstallSummary(path, apps)
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
