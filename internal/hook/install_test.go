package hook

import (
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

func TestInsertBeforeP10k(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	content := "# Enable Powerlevel10k instant prompt. Should stay close to the top of ~/.zshrc.\nif [[ -r ... ]]; then\n  source ...\nfi\n\n# rest of config\n"
	os.WriteFile(path, []byte(content), 0644)

	if err := insertBeforeP10k(path, content); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	result := string(data)

	hookIdx := strings.Index(result, hookMarker)
	p10kIdx := strings.Index(result, p10kInstantPromptMarker)

	if hookIdx == -1 {
		t.Fatal("hook not found in output")
	}
	if p10kIdx == -1 {
		t.Fatal("p10k marker not found in output")
	}
	if hookIdx >= p10kIdx {
		t.Error("hook should appear before p10k instant prompt")
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
