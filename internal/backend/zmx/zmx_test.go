package zmx

import (
	"testing"

	"github.com/nerveband/zpick/internal/backend"
)

var _ backend.Backend = (*Zmx)(nil)

func TestZmxName(t *testing.T) {
	b := New()
	if b.Name() != "zmx" {
		t.Errorf("Name() = %q, want %q", b.Name(), "zmx")
	}
}

func TestZmxBinaryName(t *testing.T) {
	b := New()
	if b.BinaryName() != "zmx" {
		t.Errorf("BinaryName() = %q, want %q", b.BinaryName(), "zmx")
	}
}

func TestZmxSessionEnvVar(t *testing.T) {
	b := New()
	if b.SessionEnvVar() != "ZMX_SESSION" {
		t.Errorf("SessionEnvVar() = %q, want %q", b.SessionEnvVar(), "ZMX_SESSION")
	}
}

func TestZmxInSession(t *testing.T) {
	t.Setenv("ZMX_SESSION", "test")
	b := New()
	if !b.InSession() {
		t.Error("InSession() should be true")
	}
}

func TestZmxAttachCommand(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "")
	want := `zmx attach "my-session"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}

func TestZmxAttachCommandWithDir(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "/tmp/foo")
	want := `cd "/tmp/foo" && zmx attach "my-session"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}
