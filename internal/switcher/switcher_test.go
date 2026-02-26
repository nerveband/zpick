package switcher

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestWriteAndRead(t *testing.T) {
	// Use a temp file so tests don't pollute the real cache.
	tmp := t.TempDir()
	path := tmp + "/switch-target"
	SetPath(path)

	want := Target{Action: "attach", Name: "work"}
	if err := Write(want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.Action != want.Action {
		t.Errorf("Action = %q, want %q", got.Action, want.Action)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
}

func TestReadDeletesFile(t *testing.T) {
	tmp := t.TempDir()
	p := tmp + "/switch-target"
	SetPath(p)

	if err := Write(Target{Action: "attach", Name: "dev"}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Confirm file exists before Read.
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file should exist after Write: %v", err)
	}

	if _, err := Read(); err != nil {
		t.Fatalf("Read: %v", err)
	}

	// File must be gone after Read.
	if _, err := os.Stat(p); err == nil {
		t.Errorf("file still exists after Read; expected it to be deleted")
	}
}

func TestReadStaleFile(t *testing.T) {
	tmp := t.TempDir()
	p := tmp + "/switch-target"
	SetPath(p)

	if err := Write(Target{Action: "attach", Name: "old"}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Backdate the file by 60 seconds to simulate staleness.
	old := time.Now().Add(-60 * time.Second)
	if err := os.Chtimes(p, old, old); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	_, err := Read()
	if err == nil {
		t.Fatal("expected error for stale file, got nil")
	}
	if !strings.Contains(err.Error(), "stale") {
		t.Errorf("error should mention 'stale', got: %v", err)
	}

	// Stale file should still be cleaned up.
	if _, err := os.Stat(p); err == nil {
		t.Errorf("stale file should be deleted after Read attempt")
	}
}

func TestReadMissingFile(t *testing.T) {
	tmp := t.TempDir()
	p := tmp + "/nonexistent"
	SetPath(p)

	_, err := Read()
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file") {
		t.Errorf("expected 'no such file' error, got: %v", err)
	}
}

func TestWriteWithDir(t *testing.T) {
	tmp := t.TempDir()
	p := tmp + "/switch-target"
	SetPath(p)

	want := Target{Action: "new", Name: "project", Dir: "/home/user/project"}
	if err := Write(want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.Action != want.Action {
		t.Errorf("Action = %q, want %q", got.Action, want.Action)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.Dir != want.Dir {
		t.Errorf("Dir = %q, want %q", got.Dir, want.Dir)
	}
}
