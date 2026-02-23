package tmux

import (
	"testing"

	"github.com/nerveband/zpick/internal/backend"
)

var _ backend.Backend = (*Tmux)(nil)

func TestTmuxName(t *testing.T) {
	b := New()
	if b.Name() != "tmux" {
		t.Errorf("Name() = %q, want %q", b.Name(), "tmux")
	}
}

func TestTmuxBinaryName(t *testing.T) {
	b := New()
	if b.BinaryName() != "tmux" {
		t.Errorf("BinaryName() = %q, want %q", b.BinaryName(), "tmux")
	}
}

func TestTmuxSessionEnvVar(t *testing.T) {
	b := New()
	if b.SessionEnvVar() != "TMUX" {
		t.Errorf("SessionEnvVar() = %q, want %q", b.SessionEnvVar(), "TMUX")
	}
}

func TestTmuxInSession(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-501/default,12345,0")
	b := New()
	if !b.InSession() {
		t.Error("InSession() should be true when TMUX is set")
	}
}

func TestTmuxInSessionFalse(t *testing.T) {
	t.Setenv("TMUX", "")
	b := New()
	if b.InSession() {
		t.Error("InSession() should be false when TMUX is empty")
	}
}

func TestTmuxAttachCommand(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "")
	want := `tmux new-session -A -s "my-session"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}

func TestTmuxAttachCommandWithDir(t *testing.T) {
	b := New()
	cmd := b.AttachCommand("my-session", "/tmp/foo")
	want := `tmux new-session -A -s "my-session" -c "/tmp/foo"`
	if cmd != want {
		t.Errorf("AttachCommand() = %q, want %q", cmd, want)
	}
}

func TestParseTmuxSessions(t *testing.T) {
	output := "work\t1\t/home/user/work\nplay\t0\t/home/user/play\n"
	sessions := parseTmuxSessions(output)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Name != "work" {
		t.Errorf("expected name work, got %s", sessions[0].Name)
	}
	if !sessions[0].Active {
		t.Error("expected work to be active (1 client)")
	}
	if sessions[1].Name != "play" {
		t.Errorf("expected name play, got %s", sessions[1].Name)
	}
	if sessions[1].Active {
		t.Error("expected play to be idle (0 clients)")
	}
	if sessions[1].StartedIn != "/home/user/play" {
		t.Errorf("expected started_in /home/user/play, got %s", sessions[1].StartedIn)
	}
}

func TestParseTmuxSessionsEmpty(t *testing.T) {
	sessions := parseTmuxSessions("")
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}
