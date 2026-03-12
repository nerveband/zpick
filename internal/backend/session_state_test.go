package backend

import "testing"

func TestCommandHasSessionBackend(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{command: "", want: false},
		{command: "/opt/homebrew/bin/tmux new-session -A -s work", want: true},
		{command: "tmux: client (/dev/ttys012)", want: true},
		{command: "/usr/local/bin/zmosh attach work", want: true},
		{command: "/usr/local/bin/zmx attach work", want: true},
		{command: "/opt/homebrew/bin/zellij attach work", want: true},
		{command: "/usr/local/bin/shpool attach work", want: true},
		{command: "/bin/zsh -l", want: false},
		{command: "/usr/local/bin/zpick version", want: false},
	}

	for _, tt := range tests {
		if got := commandHasSessionBackend(tt.command); got != tt.want {
			t.Errorf("commandHasSessionBackend(%q) = %v, want %v", tt.command, got, tt.want)
		}
	}
}

func TestShouldAutostartFromState(t *testing.T) {
	tests := []struct {
		name      string
		ttyReady  bool
		inSession bool
		termName  string
		want      bool
	}{
		{name: "interactive tty", ttyReady: true, termName: "xterm-256color", want: true},
		{name: "missing tty", ttyReady: false, termName: "xterm-256color", want: false},
		{name: "inside session", ttyReady: true, inSession: true, termName: "xterm-256color", want: false},
		{name: "dumb term", ttyReady: true, termName: "dumb", want: false},
		{name: "empty term", ttyReady: true, termName: "", want: false},
	}

	for _, tt := range tests {
		if got := shouldAutostartFromState(tt.ttyReady, tt.inSession, tt.termName); got != tt.want {
			t.Errorf("shouldAutostartFromState(%v, %v, %q) = %v, want %v", tt.ttyReady, tt.inSession, tt.termName, got, tt.want)
		}
	}
}
