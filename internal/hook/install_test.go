package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHookBlock(t *testing.T) {
	block := GenerateHookBlock([]string{"claude", "codex", "opencode"})

	if !strings.Contains(block, blockStart) {
		t.Error("block should contain start marker")
	}
	if !strings.Contains(block, blockEnd) {
		t.Error("block should contain end marker")
	}
	if !strings.Contains(block, "_zpick_guard") {
		t.Error("block should contain guard function")
	}
	if !strings.Contains(block, "ZPICK_AUTORUN") {
		t.Error("block should contain autorun check")
	}
	if !strings.Contains(block, `claude() { _zpick_guard claude "$@"; }`) {
		t.Error("block should contain claude function")
	}
	if !strings.Contains(block, `codex() { _zpick_guard codex "$@"; }`) {
		t.Error("block should contain codex function")
	}
	if !strings.Contains(block, `opencode() { _zpick_guard opencode "$@"; }`) {
		t.Error("block should contain opencode function")
	}
	// Verify all backend session env vars are checked
	if !strings.Contains(block, `"$ZMX_SESSION"`) {
		t.Error("block should check ZMX_SESSION")
	}
	if !strings.Contains(block, `"$TMUX"`) {
		t.Error("block should check TMUX")
	}
	if !strings.Contains(block, `"$SHPOOL_SESSION_NAME"`) {
		t.Error("block should check SHPOOL_SESSION_NAME")
	}
}

func TestGenerateHookBlockSkipsInvalidNames(t *testing.T) {
	block := GenerateHookBlock([]string{"claude", "bad name", "codex"})

	if !strings.Contains(block, "claude()") {
		t.Error("valid name should be included")
	}
	if strings.Contains(block, "bad name") {
		t.Error("invalid name should be skipped")
	}
	if !strings.Contains(block, "codex()") {
		t.Error("valid name should be included")
	}
}

func TestGenerateHookBlockHyphenConversion(t *testing.T) {
	block := GenerateHookBlock([]string{"my-app"})

	// Function name should use underscore
	if !strings.Contains(block, "my_app()") {
		t.Error("hyphen should be converted to underscore in function name")
	}
	// Command name should keep hyphen
	if !strings.Contains(block, "_zpick_guard my-app") {
		t.Error("original app name should be used in guard call")
	}
}

func TestHasHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	// File doesn't exist
	if hasHook(path) {
		t.Error("should return false for non-existent file")
	}

	// File without hook
	os.WriteFile(path, []byte("# some config\n"), 0644)
	if hasHook(path) {
		t.Error("should return false for file without hook")
	}

	// File with old-style hook
	os.WriteFile(path, []byte("# some config\n# zpick: session launcher\neval...\n"), 0644)
	if !hasHook(path) {
		t.Error("should return true for file with old-style hook")
	}

	// File with new-style block
	os.WriteFile(path, []byte("# some config\n"+blockStart+"\nstuff\n"+blockEnd+"\n"), 0644)
	if !hasHook(path) {
		t.Error("should return true for file with new-style block")
	}
}

func TestRemoveBlock(t *testing.T) {
	block := GenerateHookBlock([]string{"claude", "codex"})
	content := "# before\n\n" + block + "\n# after\n"

	result := removeBlock(content)

	if strings.Contains(result, blockStart) {
		t.Error("block start should be removed")
	}
	if strings.Contains(result, blockEnd) {
		t.Error("block end should be removed")
	}
	if strings.Contains(result, "_zpick_guard") {
		t.Error("guard function should be removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before block should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after block should remain")
	}
}

func TestRemoveBlockStartOnly(t *testing.T) {
	content := "# before\n" + blockStart + "\nsome stuff\n# after\n"

	result := removeBlock(content)

	if strings.Contains(result, blockStart) {
		t.Error("start marker should be removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after should remain")
	}
}

func TestRemoveOldHook(t *testing.T) {
	content := "# before\n# zpick: session launcher\n[[ -z \"$ZMX_SESSION\" ]] && command -v zpick &>/dev/null && eval \"$(zpick)\"\n# after\n"

	result := removeOldHook(content, hookMarker)

	if strings.Contains(result, hookMarker) {
		t.Error("old hook marker should be removed")
	}
	if strings.Contains(result, "eval") {
		t.Error("old hook command should be removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after should remain")
	}
}

func TestRemoveLegacyHook(t *testing.T) {
	// v1.x zmosh-picker format (Go rewrite era)
	content := "# before\n# zmosh-picker: session launcher\n[[ -z \"$ZMX_SESSION\" ]] && command -v zmosh-picker &>/dev/null && zmosh-picker\n# after\n"

	result := removeOldHook(content, "zmosh-picker: session launcher")

	if strings.Contains(result, "zmosh-picker") {
		t.Error("legacy hook should be removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after should remain")
	}
}

func TestRemoveLegacyInstallScriptHook(t *testing.T) {
	// v1.0 install.sh format
	content := "# before\n# zmosh-picker: auto-launch session picker\n[[ -z \"$ZMX_SESSION\" ]] && command -v zmosh-picker &>/dev/null && zmosh-picker\n# after\n"

	result := removeOldHook(content, "zmosh-picker: auto-launch session picker")

	if strings.Contains(result, "zmosh-picker") {
		t.Error("legacy install.sh hook should be removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after should remain")
	}
}

func TestHasHookDetectsAllGenerations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	// Gen 1: install.sh
	os.WriteFile(path, []byte("# zmosh-picker: auto-launch session picker\nzmosh-picker\n"), 0644)
	if !hasHook(path) {
		t.Error("should detect gen 1 (install.sh) hook")
	}

	// Gen 2: Go rewrite
	os.WriteFile(path, []byte("# zmosh-picker: session launcher\nzmosh-picker\n"), 0644)
	if !hasHook(path) {
		t.Error("should detect gen 2 (Go rewrite) hook")
	}

	// Gen 3: zpick rename
	os.WriteFile(path, []byte("# zpick: session launcher\neval...\n"), 0644)
	if !hasHook(path) {
		t.Error("should detect gen 3 (zpick) hook")
	}

	// Gen 4: guard block
	os.WriteFile(path, []byte(blockStart+"\nstuff\n"+blockEnd+"\n"), 0644)
	if !hasHook(path) {
		t.Error("should detect gen 4 (guard) hook")
	}

	// No hook
	os.WriteFile(path, []byte("# just config\n"), 0644)
	if hasHook(path) {
		t.Error("should not detect hook when none present")
	}
}

func TestRemoveFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	block := GenerateHookBlock([]string{"claude"})
	content := "# before\n\n" + block + "\n# after\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := removeFromFile(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	result := string(data)
	if strings.Contains(result, blockStart) {
		t.Error("block should have been removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before hook should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after hook should remain")
	}
}

func TestRemoveFromFileOldStyle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	content := "# before\n# zpick: session launcher\n[[ -z \"$ZMX_SESSION\" ]] && command -v zpick &>/dev/null && eval \"$(zpick)\"\n# after\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := removeFromFile(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	result := string(data)
	if strings.Contains(result, hookMarker) {
		t.Error("old hook should have been removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after should remain")
	}
}

func TestRemoveFromFileWithTermFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	block := GenerateHookBlock([]string{"claude"})
	content := "# before\n" + block + "\n" + termLine + "\n# after\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := removeFromFile(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	result := string(data)
	if strings.Contains(result, blockStart) {
		t.Error("block should have been removed")
	}
	if strings.Contains(result, termMarker) {
		t.Error("term fix should have been removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after should remain")
	}
}

func TestGenerateFishHookBlock(t *testing.T) {
	block := GenerateFishHookBlock([]string{"claude", "codex"})

	if !strings.Contains(block, blockStart) {
		t.Error("fish block should contain start marker")
	}
	if !strings.Contains(block, blockEnd) {
		t.Error("fish block should contain end marker")
	}
	if !strings.Contains(block, "function _zpick_guard") {
		t.Error("fish block should contain guard function")
	}
	if !strings.Contains(block, "set -e ZPICK_AUTORUN") {
		t.Error("fish block should use set -e to unset ZPICK_AUTORUN")
	}
	if !strings.Contains(block, `test -z "$ZMX_SESSION"`) {
		t.Error("fish block should check ZMX_SESSION")
	}
	if !strings.Contains(block, `test -z "$TMUX"`) {
		t.Error("fish block should check TMUX")
	}
	if !strings.Contains(block, `test -z "$SHPOOL_SESSION_NAME"`) {
		t.Error("fish block should check SHPOOL_SESSION_NAME")
	}
	if !strings.Contains(block, "function claude\n  _zpick_guard claude $argv\nend") {
		t.Error("fish block should contain claude function")
	}
	if !strings.Contains(block, "command $argv") {
		t.Error("fish block should use $argv not \"$@\"")
	}
}

func TestFishConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got := fishConfigPath()
	want := filepath.Join(dir, "fish", "conf.d", "zp.fish")
	if got != want {
		t.Errorf("fishConfigPath() = %q, want %q", got, want)
	}
}

func TestInstallAndRemoveFish(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Write guard config so install can read it
	confDir := filepath.Join(dir, "zpick")
	os.MkdirAll(confDir, 0755)
	os.WriteFile(filepath.Join(confDir, "guard.conf"), []byte("claude\ncodex\n"), 0644)

	// Ensure not Ghostty (no TERM fix)
	t.Setenv("TERM_PROGRAM", "iTerm2")
	t.Setenv("TERM", "xterm-256color")

	if err := installFish(); err != nil {
		t.Fatal(err)
	}

	path := fishConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("fish hook file should exist after install")
	}
	content := string(data)

	if !strings.Contains(content, "function _zpick_guard") {
		t.Error("fish hook should contain guard function")
	}
	if !strings.Contains(content, "function claude") {
		t.Error("fish hook should contain claude function")
	}

	// Remove
	if err := removeFish(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("fish hook file should be removed after removeFish()")
	}
}

func TestDetectShell(t *testing.T) {
	shell := detectShell()
	if shell == "" || shell == "unknown" {
		t.Skip("SHELL not set")
	}
	// Should be just the basename
	if strings.Contains(shell, "/") {
		t.Errorf("expected basename only, got %s", shell)
	}
}

func TestTermLineContent(t *testing.T) {
	if !strings.Contains(termLine, "export TERM=xterm-ghostty") {
		t.Error("term line should export TERM=xterm-ghostty")
	}
	if !strings.Contains(termLine, termMarker) {
		t.Error("term line should contain term marker")
	}
}

func TestIsGhostty(t *testing.T) {
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("TERM", origTerm)
	}()

	os.Setenv("TERM_PROGRAM", "iTerm2")
	os.Setenv("TERM", "xterm-256color")
	if isGhostty() {
		t.Error("should return false when not Ghostty")
	}

	os.Setenv("TERM_PROGRAM", "Ghostty")
	os.Setenv("TERM", "xterm-256color")
	if !isGhostty() {
		t.Error("should return true when TERM_PROGRAM is Ghostty")
	}

	os.Setenv("TERM_PROGRAM", "")
	os.Setenv("TERM", "xterm-ghostty")
	if !isGhostty() {
		t.Error("should return true when TERM contains ghostty")
	}
}

func TestHasTermFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	if hasTermFix(path) {
		t.Error("should return false for non-existent file")
	}

	os.WriteFile(path, []byte("# some config\n"), 0644)
	if hasTermFix(path) {
		t.Error("should return false for file without term fix")
	}

	os.WriteFile(path, []byte("# some config\n"+termLine+"\n"), 0644)
	if !hasTermFix(path) {
		t.Error("should return true for file with term fix")
	}
}

func TestRemoveFromFileNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	os.WriteFile(path, []byte("# just config\n"), 0644)

	if err := removeFromFile(path); err != nil {
		t.Fatal(err)
	}
	// Should print "not found" message but not error
}
