package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
)

func init() {
	backend.Register("tmux", func() backend.Backend { return New() })
}

// Tmux implements the Backend interface for tmux.
type Tmux struct{}

func New() *Tmux { return &Tmux{} }

func (t *Tmux) Name() string         { return "tmux" }
func (t *Tmux) BinaryName() string   { return "tmux" }
func (t *Tmux) SessionEnvVar() string { return "TMUX" }

func (t *Tmux) InSession() bool {
	return os.Getenv("TMUX") != ""
}

func (t *Tmux) Available() (bool, error) {
	_, err := exec.LookPath("tmux")
	if err != nil {
		return false, fmt.Errorf("tmux not found in PATH")
	}
	return true, nil
}

func (t *Tmux) Version() (string, error) {
	out, err := exec.Command("tmux", "-V").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run tmux -V: %w", err)
	}
	// Output: "tmux 3.4" — return just the version
	ver := strings.TrimSpace(string(out))
	if strings.HasPrefix(ver, "tmux ") {
		ver = ver[5:]
	}
	return ver, nil
}

func (t *Tmux) List() ([]backend.Session, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F",
		"#{session_name}\t#{session_attached}\t#{pane_current_path}").Output()
	if err != nil {
		// tmux returns error when server not running (no sessions)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to run tmux list-sessions: %w", err)
	}
	return parseTmuxSessions(string(out)), nil
}

// FastList is the same as List for tmux (no socket shortcut).
func (t *Tmux) FastList() ([]backend.Session, error) {
	return t.List()
}

func (t *Tmux) Attach(name string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}
	return backend.ExecCommand(tmuxPath, []string{"tmux", "new-session", "-A", "-s", name})
}

func (t *Tmux) AttachCommand(name, dir string) string {
	if dir != "" {
		return fmt.Sprintf(`tmux new-session -A -s "%s" -c "%s"`, name, dir)
	}
	return fmt.Sprintf(`tmux new-session -A -s "%s"`, name)
}

func (t *Tmux) Kill(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

// parseTmuxSessions parses the tab-separated output of tmux list-sessions.
// Format: session_name\tsession_attached\tpane_current_path
func parseTmuxSessions(output string) []backend.Session {
	var sessions []backend.Session
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 1 {
			continue
		}
		s := backend.Session{
			Name:      fields[0],
			StartedIn: "~",
		}
		if len(fields) >= 2 {
			attached, _ := strconv.Atoi(fields[1])
			s.Clients = attached
			s.Active = attached > 0
		}
		if len(fields) >= 3 {
			s.StartedIn = fields[2]
		}
		sessions = append(sessions, s)
	}
	return sessions
}
