package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	"github.com/nerveband/zpick/internal/guard"
)

// fishConfigPath returns the path to the managed fish hook file.
// Uses conf.d/ for cleaner fish config, respects XDG_CONFIG_HOME, and prefixes
// the filename so fish loads zpick before slower prompt/plugin snippets.
func fishConfigPath() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "fish", "conf.d", "00-zp.fish")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fish", "conf.d", "00-zp.fish")
}

func legacyFishConfigPath() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "fish", "conf.d", "zp.fish")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fish", "conf.d", "zp.fish")
}

func currentFishConfigPath() string {
	if _, err := os.Stat(fishConfigPath()); err == nil {
		return fishConfigPath()
	}
	if _, err := os.Stat(legacyFishConfigPath()); err == nil {
		return legacyFishConfigPath()
	}
	return fishConfigPath()
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

	b.WriteString("set -g _ZPICK_BIN\n")
	b.WriteString("if command -sq zp\n")
	b.WriteString("  set _ZPICK_BIN (command -s zp)\n")
	b.WriteString("else if test -x \"$HOME/.local/bin/zp\"\n")
	b.WriteString("  set _ZPICK_BIN \"$HOME/.local/bin/zp\"\n")
	b.WriteString("else if test -x /usr/local/bin/zp\n")
	b.WriteString("  set _ZPICK_BIN /usr/local/bin/zp\n")
	b.WriteString("end\n")

	b.WriteString("function _zpick_exec\n")
	b.WriteString("  if test -n \"$_ZPICK_BIN\"\n")
	b.WriteString("    $_ZPICK_BIN $argv\n")
	b.WriteString("    return\n")
	b.WriteString("  end\n")
	b.WriteString("  return 127\n")
	b.WriteString("end\n")

	b.WriteString("function _zpick_eval\n")
	b.WriteString("  set -l _zpick_out (_zpick_exec $argv)\n")
	b.WriteString("  or return $status\n")
	b.WriteString("  eval $_zpick_out\n")
	b.WriteString("end\n")

	// Picker function: eval the command zp outputs
	b.WriteString("function zp\n")
	b.WriteString("  if test (count $argv) -eq 0\n")
	b.WriteString("    _zpick_eval\n")
	b.WriteString("    return\n")
	b.WriteString("  end\n")
	b.WriteString("  _zpick_exec $argv\n")
	b.WriteString("end\n")

	// Source-time autorun/resume/autostart so fresh shells hit zp before fish
	// prompt/plugins load.
	b.WriteString("if test -n \"$ZPICK_AUTORUN\"\n")
	b.WriteString("  _zpick_exec autorun\n")
	b.WriteString("else if test -f \"$HOME/.cache/zpick/switch-target\"\n")
	b.WriteString("  _zpick_eval resume\n")
	b.WriteString("else if status is-interactive\n")
	b.WriteString("  _zpick_exec should-autostart >/dev/null 2>&1\n")
	b.WriteString("  and _zpick_eval\n")
	b.WriteString("end\n")

	// Guard function + per-app wrappers (optional — only if apps configured)
	if len(apps) > 0 {
		envCheck := fishSessionEnvCheck()
		b.WriteString("function _zpick_guard\n")
		fmt.Fprintf(&b, "  if %s; and _zpick_exec version >/dev/null 2>&1\n", envCheck)
		b.WriteString("    set -l _r (_zpick_exec guard -- $argv)\n")
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

	legacyPath := legacyFishConfigPath()
	if legacyPath != path {
		_ = os.Remove(legacyPath)
	}

	printInstallSummary(path, apps)
	return nil
}

// removeFish removes the fish hook file.
func removeFish() error {
	paths := []string{fishConfigPath(), legacyFishConfigPath()}
	removed := false
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("cannot remove %s: %w", path, err)
		}
		fmt.Printf("  removed hook from %s\n", path)
		removed = true
	}
	if !removed {
		fmt.Printf("  fish hook not found at %s\n", fishConfigPath())
	}
	return nil
}
