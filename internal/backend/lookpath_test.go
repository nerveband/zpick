package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookPathFallsBackToHomeBin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "/usr/bin:/bin")

	cmdPath := filepath.Join(home, ".local", "bin", "zp-test")
	if err := os.MkdirAll(filepath.Dir(cmdPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cmdPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := LookPath("zp-test")
	if err != nil {
		t.Fatalf("LookPath returned error: %v", err)
	}
	if got != cmdPath {
		t.Fatalf("LookPath = %q, want %q", got, cmdPath)
	}
}

func TestShellCommandUsesAbsolutePathWhenOnlyFallbackExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "/usr/bin:/bin")

	cmdPath := filepath.Join(home, ".cargo", "bin", "zp-test")
	if err := os.MkdirAll(filepath.Dir(cmdPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cmdPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	got := ShellCommand("zp-test")
	want := shellQuote(cmdPath)
	if got != want {
		t.Fatalf("ShellCommand = %q, want %q", got, want)
	}
}
