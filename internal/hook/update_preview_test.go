package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewManagedHookUpdate_NoChange(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/zsh")

	path := filepath.Join(home, ".zshrc")
	block := GenerateHookBlock(nil)
	if err := os.WriteFile(path, []byte(block+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	preview, err := PreviewManagedHookUpdate(false)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.HasHook {
		t.Fatal("expected hook to be detected")
	}
	if preview.Changed {
		t.Fatal("expected hook preview to report no changes")
	}
}

func TestPreviewManagedHookUpdate_DetectsChangedManagedBlock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/zsh")

	path := filepath.Join(home, ".zshrc")
	oldBlock := strings.Join([]string{
		blockStart,
		`zp() { eval "$(command zp)"; }`,
		blockEnd,
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(oldBlock), 0644); err != nil {
		t.Fatal(err)
	}

	preview, err := PreviewManagedHookUpdate(false)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.HasHook {
		t.Fatal("expected hook to be detected")
	}
	if !preview.Changed {
		t.Fatal("expected hook preview to report changes")
	}
	if !strings.Contains(preview.DesiredBlock, `"$_ZPICK_BIN" "$@"`) {
		t.Fatal("expected desired hook block to execute the resolved binary path")
	}
}
