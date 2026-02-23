package shpool

import (
	"testing"

	"github.com/nerveband/zpick/internal/backend"
)

var _ backend.Backend = (*Shpool)(nil)

func TestShpoolName(t *testing.T) {
	b := New()
	if b.Name() != "shpool" {
		t.Errorf("Name() = %q, want %q", b.Name(), "shpool")
	}
}

func TestShpoolBinaryName(t *testing.T) {
	b := New()
	if b.BinaryName() != "shpool" {
		t.Errorf("BinaryName() = %q, want %q", b.BinaryName(), "shpool")
	}
}

func TestShpoolSessionEnvVar(t *testing.T) {
	b := New()
	if b.SessionEnvVar() != "SHPOOL_SESSION_NAME" {
		t.Errorf("SessionEnvVar() = %q, want %q", b.SessionEnvVar(), "SHPOOL_SESSION_NAME")
	}
}

func TestShpoolInSession(t *testing.T) {
	t.Setenv("SHPOOL_SESSION_NAME", "test")
	b := New()
	if !b.InSession() {
		t.Error("InSession() should be true")
	}
}

func TestShpoolInSessionFalse(t *testing.T) {
	t.Setenv("SHPOOL_SESSION_NAME", "")
	b := New()
	if b.InSession() {
		t.Error("InSession() should be false")
	}
}

func TestShpoolAttachCommand(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "")
	want := `shpool attach "my-session"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}

func TestShpoolAttachCommandWithDir(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "/tmp/foo")
	want := `cd "/tmp/foo" && shpool attach "my-session"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}

func TestParseShpoolSessions(t *testing.T) {
	output := "work\nplay\ndev\n"
	sessions := parseShpoolSessions(output)
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}
	if sessions[0].Name != "work" {
		t.Errorf("expected name work, got %s", sessions[0].Name)
	}
	if sessions[2].Name != "dev" {
		t.Errorf("expected name dev, got %s", sessions[2].Name)
	}
}

func TestParseShpoolSessionsEmpty(t *testing.T) {
	sessions := parseShpoolSessions("")
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestParseShpoolSessionsActive(t *testing.T) {
	t.Setenv("SHPOOL_SESSION_NAME", "work")
	output := "work\nplay\n"
	sessions := parseShpoolSessions(output)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if !sessions[0].Active {
		t.Error("expected work to be active")
	}
	if sessions[1].Active {
		t.Error("expected play to be inactive")
	}
}
