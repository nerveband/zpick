package zellij

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
)

func init() {
	backend.Register("zellij", func() backend.Backend { return New() })
}

// Zellij implements the Backend interface for zellij.
type Zellij struct{}

func New() *Zellij { return &Zellij{} }

func (z *Zellij) Name() string         { return "zellij" }
func (z *Zellij) BinaryName() string   { return "zellij" }
func (z *Zellij) SessionEnvVar() string { return "ZELLIJ" }

func (z *Zellij) InSession() bool {
	return os.Getenv("ZELLIJ") != ""
}

func (z *Zellij) Available() (bool, error) {
	_, err := exec.LookPath("zellij")
	if err != nil {
		return false, fmt.Errorf("zellij not found in PATH")
	}
	return true, nil
}

func (z *Zellij) Version() (string, error) {
	out, err := exec.Command("zellij", "--version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run zellij --version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (z *Zellij) List() ([]backend.Session, error) {
	out, err := exec.Command("zellij", "list-sessions", "--short", "--no-formatting").CombinedOutput()
	if err != nil {
		// zellij list-sessions returns exit code 1 when no sessions exist
		if strings.Contains(string(out), "No active") || strings.TrimSpace(string(out)) == "" {
			return nil, nil
		}
		// Try without --short --no-formatting flags (older zellij versions)
		out, err = exec.Command("zellij", "list-sessions").CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to run zellij list-sessions: %w", err)
		}
	}
	return parseSessions(string(out)), nil
}

// FastList is the same as List for zellij (no shortcut).
func (z *Zellij) FastList() ([]backend.Session, error) {
	return z.List()
}

func (z *Zellij) Attach(name string) error {
	zellijPath, err := exec.LookPath("zellij")
	if err != nil {
		return fmt.Errorf("zellij not found: %w", err)
	}
	return backend.ExecCommand(zellijPath, []string{"zellij", "attach", name})
}

func (z *Zellij) AttachCommand(name, dir string) string {
	cmd := fmt.Sprintf(`zellij attach "%s"`, name)
	if dir != "" {
		return fmt.Sprintf(`cd "%s" && %s`, dir, cmd)
	}
	return cmd
}

func (z *Zellij) Kill(name string) error {
	return exec.Command("zellij", "kill-session", name).Run()
}

// parseSessions parses the output of zellij list-sessions.
// Output format varies by version. With --short --no-formatting, each line is a session name.
// Without those flags, lines may include status like "(current session)" or "EXITED".
func parseSessions(output string) []backend.Session {
	active := os.Getenv("ZELLIJ_SESSION_NAME")
	var sessions []backend.Session
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip lines that are purely informational
		if strings.HasPrefix(line, "No active") {
			continue
		}

		name := line
		isExited := false

		// Handle status markers: "session_name (EXITED ...)" or "session_name [Created ...]"
		if idx := strings.IndexAny(line, "(["); idx > 0 {
			name = strings.TrimSpace(line[:idx])
			lower := strings.ToLower(line[idx:])
			if strings.Contains(lower, "exited") {
				isExited = true
			}
		}

		// Skip exited/dead sessions
		if isExited {
			continue
		}

		sessions = append(sessions, backend.Session{
			Name:      name,
			StartedIn: "~",
			Active:    name == active,
		})
	}
	return sessions
}
