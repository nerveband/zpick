package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	"github.com/nerveband/zpick/internal/guard"
)

// Block markers for the hook block in shell config
const (
	blockStart = "# >>> zpick hook >>>"
	blockEnd   = "# <<< zpick hook <<<"
)

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

	// Picker launcher: eval the command zp outputs
	b.WriteString("zp() { eval \"$(command zp)\"; }\n")

	// Autorun: defer to precmd so it runs after shell init (avoids p10k instant prompt conflict)
	b.WriteString("if [[ -n \"$ZPICK_AUTORUN\" ]]; then\n")
	b.WriteString("  _zpick_autorun() {\n")
	b.WriteString("    precmd_functions=(${precmd_functions:#_zpick_autorun})\n")
	b.WriteString("    command zp autorun\n")
	b.WriteString("    unset ZPICK_AUTORUN\n")
	b.WriteString("  }\n")
	b.WriteString("  precmd_functions+=(_zpick_autorun)\n")
	b.WriteString("fi\n")

	// Switch-target: resume after in-session detach
	b.WriteString("if [[ -f \"$HOME/.cache/zpick/switch-target\" ]]; then\n")
	b.WriteString("  _zpick_switch() {\n")
	b.WriteString("    precmd_functions=(${precmd_functions:#_zpick_switch})\n")
	b.WriteString("    eval \"$(command zp resume)\"\n")
	b.WriteString("  }\n")
	b.WriteString("  precmd_functions+=(_zpick_switch)\n")
	b.WriteString("fi\n")

	// Guard function + per-app wrappers (optional — only if apps configured)
	if len(apps) > 0 {
		envCheck := sessionEnvCheck()
		b.WriteString("_zpick_guard() {\n")
		fmt.Fprintf(&b, "  if [[ %s ]] && command -v zp &>/dev/null; then\n", envCheck)
		b.WriteString("    local _r\n")
		b.WriteString("    _r=$(command zp guard -- \"$@\")\n")
		b.WriteString("    if [[ -n \"$_r\" ]]; then eval \"$_r\"; return; fi\n")
		b.WriteString("  fi\n")
		b.WriteString("  command \"$@\"\n")
		b.WriteString("}\n")

		for _, app := range apps {
			if err := guard.ValidateName(app); err != nil {
				continue
			}
			fname := guard.FuncName(app)
			fmt.Fprintf(&b, "%s() { _zpick_guard %s \"$@\"; }\n", fname, app)
		}
	}

	b.WriteString(blockEnd)
	return b.String()
}

// Install adds the zpick shell hook to the appropriate shell config file.
// When withGuard is true, guard wrappers for configured apps are included.
func Install(withGuard bool) error {
	shell := backend.DetectShell()
	var err error
	switch shell {
	case "zsh":
		err = installShell(zshrcPath(), withGuard)
	case "bash":
		err = installShell(bashrcPath(), withGuard)
	case "fish":
		err = installFish(withGuard)
	default:
		var apps []string
		if withGuard {
			apps, _ = guard.ReadConfig()
		}
		block := GenerateHookBlock(apps)
		return fmt.Errorf("unsupported shell: %s\nManually add this to your shell config:\n\n%s", shell, block)
	}
	if err == nil {
		InstallSymlink()
		fmt.Println("  to remove: zp remove-hook")
	}
	return err
}

// Remove removes the zpick hook from the shell config file.
func Remove() error {
	shell := backend.DetectShell()
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

// installShell installs the hook block into a shell config file.
// Guard wrappers are only included when withGuard is true.
func installShell(path string, withGuard bool) error {
	var apps []string
	if withGuard {
		apps, _ = guard.ReadConfig()
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	block := GenerateHookBlock(apps)
	content = removeBlock(content)

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

	printInstallSummary(path, apps)
	return nil
}

// removeBlock removes the hook block from content, returning cleaned content.
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

// hasHook checks if the zpick hook block is in the file.
func hasHook(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), blockStart)
}

// HookStatus reads the current shell's config once and reports whether the
// hook and guard wrappers are installed.
func HookStatus() (hasHook, hasGuard bool) {
	shell := backend.DetectShell()
	var path string
	switch shell {
	case "zsh":
		path = zshrcPath()
	case "bash":
		path = bashrcPath()
	case "fish":
		path = fishConfigPath()
	default:
		return false, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}
	content := string(data)
	hasHook = strings.Contains(content, blockStart)
	hasGuard = strings.Contains(content, "_zpick_guard")
	return
}

// HasHookInstalled checks if any zpick hook is installed in the current shell's config.
func HasHookInstalled() bool {
	h, _ := HookStatus()
	return h
}

// HasGuardInstalled checks if guard wrappers are present in the installed hook.
func HasGuardInstalled() bool {
	_, g := HookStatus()
	return g
}

// hasGuardInFile checks if the guard function is present in a file.
func hasGuardInFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "_zpick_guard")
}

// InstallGuard installs guard wrappers, adding the hook first if missing.
func InstallGuard() error {
	if !HasHookInstalled() {
		fmt.Println("  hook not installed, installing first...")
	}
	return Install(true)
}

// RemoveGuard removes guard wrappers but keeps the shell hook.
func RemoveGuard() error {
	if !HasHookInstalled() {
		fmt.Println("  hook not installed, nothing to do")
		return nil
	}
	if !HasGuardInstalled() {
		fmt.Println("  guard wrappers not installed, nothing to do")
		return nil
	}
	return Install(false)
}

// removeFromFile removes the hook block from a file.
func removeFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	content := string(data)
	if !strings.Contains(content, blockStart) {
		fmt.Printf("  hook not found in %s\n", path)
		return nil
	}

	content = removeBlock(content)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	fmt.Printf("  removed hook from %s\n", path)
	return nil
}

// printInstallSummary prints the install confirmation for both shell and fish.
func printInstallSummary(path string, apps []string) {
	fmt.Printf("  installed shell hook in %s\n", path)
	fmt.Println("    - zp() function (session picker launcher)")
	if len(apps) > 0 {
		fmt.Printf("    - guard wrappers for: %s\n", strings.Join(apps, ", "))
		fmt.Println("      (ensures these tools run inside a session)")
	}
}
