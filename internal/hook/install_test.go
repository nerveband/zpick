package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookLineContent(t *testing.T) {
	if !strings.Contains(hookLine, "zpick") {
		t.Error("hook line should contain zpick")
	}
	if !strings.Contains(hookLine, "ZMX_SESSION") {
		t.Error("hook line should check ZMX_SESSION")
	}
	if !strings.Contains(hookLine, "eval") {
		t.Error("hook line should use eval")
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

	// File with hook
	os.WriteFile(path, []byte("# some config\n"+hookLine+"\n"), 0644)
	if !hasHook(path) {
		t.Error("should return true for file with hook")
	}
}

func TestRemoveFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	content := "# before\n" + hookLine + "\n# after\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := removeFromFile(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), hookMarker) {
		t.Error("hook should have been removed")
	}
	if !strings.Contains(string(data), "# before") {
		t.Error("content before hook should remain")
	}
	if !strings.Contains(string(data), "# after") {
		t.Error("content after hook should remain")
	}
}

func TestHookAppendsToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	// Even with P10k content, hook should be appended at the end
	// (so it runs after PATH is configured)
	content := "# Enable Powerlevel10k instant prompt.\nexport PATH=\"$HOME/.local/bin:$PATH\"\n# rest of config\n"
	os.WriteFile(path, []byte(content), 0644)

	// Manually append hook like installZsh does
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	fmt.Fprintf(f, "\n%s\n", hookLine)
	f.Close()

	data, _ := os.ReadFile(path)
	result := string(data)

	if !strings.Contains(result, hookMarker) {
		t.Fatal("hook not found in output")
	}

	// Hook should come after the PATH line
	pathIdx := strings.Index(result, "PATH")
	hookIdx := strings.Index(result, hookMarker)
	if hookIdx <= pathIdx {
		t.Error("hook should appear after PATH configuration")
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
	// Save and restore env
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("TERM", origTerm)
	}()

	// Neither set to ghostty
	os.Setenv("TERM_PROGRAM", "iTerm2")
	os.Setenv("TERM", "xterm-256color")
	if isGhostty() {
		t.Error("should return false when not Ghostty")
	}

	// TERM_PROGRAM = Ghostty
	os.Setenv("TERM_PROGRAM", "Ghostty")
	os.Setenv("TERM", "xterm-256color")
	if !isGhostty() {
		t.Error("should return true when TERM_PROGRAM is Ghostty")
	}

	// TERM contains ghostty
	os.Setenv("TERM_PROGRAM", "")
	os.Setenv("TERM", "xterm-ghostty")
	if !isGhostty() {
		t.Error("should return true when TERM contains ghostty")
	}
}

func TestHasTermFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	// File doesn't exist
	if hasTermFix(path) {
		t.Error("should return false for non-existent file")
	}

	// File without term fix
	os.WriteFile(path, []byte("# some config\n"), 0644)
	if hasTermFix(path) {
		t.Error("should return false for file without term fix")
	}

	// File with term fix
	os.WriteFile(path, []byte("# some config\n"+termLine+"\n"), 0644)
	if !hasTermFix(path) {
		t.Error("should return true for file with term fix")
	}
}

func TestRemoveFromFileWithTermFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	content := "# before\n" + hookLine + "\n" + termLine + "\n# after\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := removeFromFile(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	result := string(data)
	if strings.Contains(result, hookMarker) {
		t.Error("hook should have been removed")
	}
	if strings.Contains(result, termMarker) {
		t.Error("term fix should have been removed")
	}
	if !strings.Contains(result, "# before") {
		t.Error("content before hook should remain")
	}
	if !strings.Contains(result, "# after") {
		t.Error("content after hook should remain")
	}
}

func TestRemoveFromFileTermFixOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	content := "# before\n" + termLine + "\n# after\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := removeFromFile(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	result := string(data)
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
