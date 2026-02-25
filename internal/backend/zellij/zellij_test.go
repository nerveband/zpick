package zellij

import (
	"testing"

	"github.com/nerveband/zpick/internal/backend"
)

// Verify Zellij implements the Backend interface.
var _ backend.Backend = (*Zellij)(nil)

func TestZellijName(t *testing.T) {
	b := New()
	if b.Name() != "zellij" {
		t.Errorf("Name() = %q, want %q", b.Name(), "zellij")
	}
}

func TestZellijBinaryName(t *testing.T) {
	b := New()
	if b.BinaryName() != "zellij" {
		t.Errorf("BinaryName() = %q, want %q", b.BinaryName(), "zellij")
	}
}

func TestZellijSessionEnvVar(t *testing.T) {
	b := New()
	if b.SessionEnvVar() != "ZELLIJ" {
		t.Errorf("SessionEnvVar() = %q, want %q", b.SessionEnvVar(), "ZELLIJ")
	}
}

func TestZellijInSessionTrue(t *testing.T) {
	t.Setenv("ZELLIJ", "0")
	b := New()
	if !b.InSession() {
		t.Error("InSession() should return true when ZELLIJ is set")
	}
}

func TestZellijInSessionFalse(t *testing.T) {
	t.Setenv("ZELLIJ", "")
	b := New()
	if b.InSession() {
		t.Error("InSession() should return false when ZELLIJ is empty")
	}
}

func TestZellijAttachCommand(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "")
	want := `zellij attach "my-session"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}

func TestZellijAttachCommandWithDir(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "/tmp/foo")
	want := `cd "/tmp/foo" && zellij attach "my-session"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}

func TestParseSessionsBasic(t *testing.T) {
	input := "work\nplay\n"
	sessions := parseSessions(input)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Name != "work" {
		t.Errorf("expected name work, got %s", sessions[0].Name)
	}
	if sessions[1].Name != "play" {
		t.Errorf("expected name play, got %s", sessions[1].Name)
	}
}

func TestParseSessionsEmpty(t *testing.T) {
	sessions := parseSessions("")
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestParseSessionsWithStatus(t *testing.T) {
	// Zellij may annotate sessions with status info
	input := "work [Created 2h ago]\nplay (EXITED - 3 hours ago)\ntest\n"
	sessions := parseSessions(input)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions (exited filtered), got %d", len(sessions))
	}
	if sessions[0].Name != "work" {
		t.Errorf("expected name work, got %s", sessions[0].Name)
	}
	if sessions[1].Name != "test" {
		t.Errorf("expected name test, got %s", sessions[1].Name)
	}
}

func TestParseSessionsActive(t *testing.T) {
	t.Setenv("ZELLIJ_SESSION_NAME", "work")
	input := "work\nplay\n"
	sessions := parseSessions(input)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if !sessions[0].Active {
		t.Error("expected work session to be active")
	}
	if sessions[1].Active {
		t.Error("expected play session to not be active")
	}
}

func TestParseSessionsNoActive(t *testing.T) {
	input := "No active zellij sessions found.\n"
	sessions := parseSessions(input)
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}
