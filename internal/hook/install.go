package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	"github.com/nerveband/zpick/internal/guard"
	"golang.org/x/term"
)

// Block markers for the hook block in shell config
const (
	blockStart = "# >>> zpick hook >>>"
	blockEnd   = "# <<< zpick hook <<<"
)

// sessionEnvCheck builds the bash/zsh condition checking all session env vars.
// Includes ZPICK_SESSION (zpick's own marker) in addition to backend-specific vars.
func sessionEnvCheck() string {
	vars := append(backend.AllSessionEnvVars(), "ZPICK_SESSION")
	var parts []string
	for _, v := range vars {
		parts = append(parts, fmt.Sprintf(`-z "$%s"`, v))
	}
	return strings.Join(parts, " && ")
}

func generatePosixHookBlock(apps []string) string {
	var b strings.Builder
	b.WriteString(blockStart)
	b.WriteByte('\n')

	// Resolve the binary early so source-time autostart works before PATH setup.
	b.WriteString("_ZPICK_BIN=\n")
	b.WriteString("if command -v zp >/dev/null 2>&1; then\n")
	b.WriteString("  _ZPICK_BIN=zp\n")
	b.WriteString("elif [[ -x \"$HOME/.local/bin/zp\" ]]; then\n")
	b.WriteString("  _ZPICK_BIN=\"$HOME/.local/bin/zp\"\n")
	b.WriteString("elif [[ -x /usr/local/bin/zp ]]; then\n")
	b.WriteString("  _ZPICK_BIN=/usr/local/bin/zp\n")
	b.WriteString("fi\n")

	b.WriteString("_zpick_exec() {\n")
	b.WriteString("  if command -v zp >/dev/null 2>&1; then\n")
	b.WriteString("    command zp \"$@\"\n")
	b.WriteString("    return\n")
	b.WriteString("  fi\n")
	b.WriteString("  if [[ -n \"${_ZPICK_BIN:-}\" ]]; then\n")
	b.WriteString("    \"$_ZPICK_BIN\" \"$@\"\n")
	b.WriteString("    return\n")
	b.WriteString("  fi\n")
	b.WriteString("  return 127\n")
	b.WriteString("}\n")

	b.WriteString("_zpick_eval() {\n")
	b.WriteString("  local _zpick_out\n")
	b.WriteString("  _zpick_out=\"$(_zpick_exec \"$@\")\" || return $?\n")
	b.WriteString("  eval \"$_zpick_out\"\n")
	b.WriteString("}\n")

	// Picker launcher: eval the command zp outputs
	b.WriteString("zp() {\n")
	b.WriteString("  if [[ $# -eq 0 ]]; then\n")
	b.WriteString("    _zpick_eval\n")
	b.WriteString("    return\n")
	b.WriteString("  fi\n")
	b.WriteString("  _zpick_exec \"$@\"\n")
	b.WriteString("}\n")

	// Source-time autorun/resume/autostart so fresh shells drop into zp before
	// the prompt/plugin stack initializes.
	b.WriteString("if [[ -n \"$ZPICK_AUTORUN\" ]]; then\n")
	b.WriteString("  _zpick_exec autorun\n")
	b.WriteString("elif [[ -f \"$HOME/.cache/zpick/switch-target\" ]]; then\n")
	b.WriteString("  _zpick_eval resume\n")
	b.WriteString("elif [[ \"$-\" == *i* ]] && _zpick_exec should-autostart >/dev/null 2>&1; then\n")
	b.WriteString("  _zpick_eval\n")
	b.WriteString("fi\n")

	// Guard function + per-app wrappers (optional — only if apps configured)
	if len(apps) > 0 {
		envCheck := sessionEnvCheck()
		b.WriteString("_zpick_guard() {\n")
		fmt.Fprintf(&b, "  if [[ %s ]] && _zpick_exec version >/dev/null 2>&1; then\n", envCheck)
		b.WriteString("    local _r\n")
		b.WriteString("    _r=\"$(_zpick_exec guard -- \"$@\")\"\n")
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

// GenerateHookBlock builds the zsh hook block from the guard config.
func GenerateHookBlock(apps []string) string {
	return generatePosixHookBlock(apps)
}

// GenerateBashHookBlock builds the bash hook block from the guard config.
func GenerateBashHookBlock(apps []string) string {
	return generatePosixHookBlock(apps)
}

// Install adds the zpick shell hook to the appropriate shell config file.
// When withGuard is true, guard wrappers for configured apps are included.
func Install(withGuard bool) error {
	shell := backend.DetectShell()
	var err error
	switch shell {
	case "zsh":
		err = installShell(zshrcPath(), withGuard, "zsh")
	case "bash":
		err = installShell(bashrcPath(), withGuard, "bash")
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
func installShell(path string, withGuard bool, shell string) error {
	var apps []string
	if withGuard {
		apps, _ = guard.ReadConfig()
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	block := GenerateHookBlock(apps)
	if shell == "bash" {
		block = GenerateBashHookBlock(apps)
	}
	content = removeBlock(content)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	trimmed := strings.TrimLeft(content, "\n")
	fmt.Fprintf(f, "%s\n", block)
	if strings.TrimSpace(trimmed) != "" {
		fmt.Fprint(f, "\n")
		fmt.Fprint(f, trimmed)
		if !strings.HasSuffix(trimmed, "\n") {
			fmt.Fprint(f, "\n")
		}
	}

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
	path, ok := hookConfigPath()
	if !ok {
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

// HookUpdatePreview describes how the installed managed hook differs from the
// current binary's template.
type HookUpdatePreview struct {
	Path         string
	HasHook      bool
	WithGuard    bool
	CurrentBlock string
	DesiredBlock string
	Changed      bool
}

// PreviewManagedHookUpdate compares the currently installed managed hook block
// against the hook template generated by the current binary.
func PreviewManagedHookUpdate(withGuard bool) (HookUpdatePreview, error) {
	path, ok := hookConfigPath()
	if !ok {
		return HookUpdatePreview{}, fmt.Errorf("unsupported shell: %s", backend.DetectShell())
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return HookUpdatePreview{}, fmt.Errorf("cannot read %s: %w", path, err)
	}
	content := string(data)
	currentBlock, hasHook := extractBlock(content)

	var apps []string
	if withGuard {
		apps, _ = guard.ReadConfig()
	}
	desiredBlock := generateHookBlockForShell(apps)

	return HookUpdatePreview{
		Path:         path,
		HasHook:      hasHook,
		WithGuard:    withGuard,
		CurrentBlock: currentBlock,
		DesiredBlock: desiredBlock,
		Changed:      hasHook && currentBlock != desiredBlock,
	}, nil
}

// PromptAndApplyHookUpdate explains an available managed hook change and asks
// whether to apply it with install-hook/install-guard.
func PromptAndApplyHookUpdate(withGuard bool) error {
	preview, err := PreviewManagedHookUpdate(withGuard)
	if err != nil {
		return nil
	}
	if !preview.HasHook {
		CheckSymlink()
		return nil
	}
	if !preview.Changed {
		fmt.Printf("Shell hook in %s already matches this version.\n", preview.Path)
		return nil
	}

	fmt.Printf("Shell hook update available for %s\n\n", preview.Path)
	fmt.Println("This version would replace the managed zpick hook block with:")
	fmt.Println()
	fmt.Println(preview.DesiredBlock)
	fmt.Println()

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return printHookUpdateCommand(withGuard)
	}
	defer tty.Close()

	fmt.Fprintf(tty, "Apply this hook update now? %s[y/N]%s ", boldGreenTTY(), resetTTY())
	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return printHookUpdateCommand(withGuard)
	}
	defer term.Restore(int(tty.Fd()), oldState)

	buf := make([]byte, 3)
	n, err := tty.Read(buf)
	term.Restore(int(tty.Fd()), oldState)
	fmt.Fprintln(tty)
	if err != nil || n == 0 {
		return printHookUpdateCommand(withGuard)
	}

	if buf[0] == 'y' || buf[0] == 'Y' {
		if withGuard {
			return InstallGuard()
		}
		return Install(false)
	}

	fmt.Fprintln(tty, "Skipped hook update.")
	return printHookUpdateCommand(withGuard)
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

func hookConfigPath() (string, bool) {
	switch backend.DetectShell() {
	case "zsh":
		return zshrcPath(), true
	case "bash":
		return bashrcPath(), true
	case "fish":
		return currentFishConfigPath(), true
	default:
		return "", false
	}
}

func extractBlock(content string) (string, bool) {
	startIdx := strings.Index(content, blockStart)
	if startIdx < 0 {
		return "", false
	}
	endIdx := strings.Index(content[startIdx:], blockEnd)
	if endIdx < 0 {
		return "", false
	}
	endIdx += startIdx
	return content[startIdx : endIdx+len(blockEnd)], true
}

func generateHookBlockForShell(apps []string) string {
	switch backend.DetectShell() {
	case "fish":
		return GenerateFishHookBlock(apps)
	case "bash":
		return GenerateBashHookBlock(apps)
	default:
		return GenerateHookBlock(apps)
	}
}

func printHookUpdateCommand(withGuard bool) error {
	cmd := "zp install-hook"
	if withGuard {
		cmd = "zp install-guard"
	}
	fmt.Printf("Run '%s' to apply it.\n", cmd)
	return nil
}

func boldGreenTTY() string { return "\033[1;32m" }

func resetTTY() string { return "\033[0m" }
