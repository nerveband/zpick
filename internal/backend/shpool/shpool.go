package shpool

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
)

func init() {
	backend.Register("shpool", func() backend.Backend { return New() })
}

// Shpool implements the Backend interface for shpool.
type Shpool struct{}

func New() *Shpool { return &Shpool{} }

func (s *Shpool) Name() string         { return "shpool" }
func (s *Shpool) BinaryName() string   { return "shpool" }
func (s *Shpool) SessionEnvVar() string { return "SHPOOL_SESSION_NAME" }

func (s *Shpool) InSession() bool {
	return os.Getenv("SHPOOL_SESSION_NAME") != ""
}

func (s *Shpool) Available() (bool, error) {
	_, err := exec.LookPath("shpool")
	if err != nil {
		return false, fmt.Errorf("shpool not found in PATH")
	}
	// Also check that daemon is running
	if err := exec.Command("shpool", "status").Run(); err != nil {
		return false, fmt.Errorf("shpool daemon not running (start with: shpool daemon)")
	}
	return true, nil
}

func (s *Shpool) Version() (string, error) {
	out, err := exec.Command("shpool", "version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run shpool version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *Shpool) List() ([]backend.Session, error) {
	out, err := exec.Command("shpool", "list").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run shpool list: %w", err)
	}
	return parseShpoolSessions(string(out)), nil
}

// FastList is the same as List for shpool (no shortcut).
func (s *Shpool) FastList() ([]backend.Session, error) {
	return s.List()
}

func (s *Shpool) Attach(name string) error {
	shpoolPath, err := exec.LookPath("shpool")
	if err != nil {
		return fmt.Errorf("shpool not found: %w", err)
	}
	return backend.ExecCommand(shpoolPath, []string{"shpool", "attach", name})
}

func (s *Shpool) AttachCommand(name, dir string) string {
	cmd := fmt.Sprintf(`shpool attach "%s"`, name)
	if dir != "" {
		return fmt.Sprintf(`cd "%s" && %s`, dir, cmd)
	}
	return cmd
}

func (s *Shpool) Kill(name string) error {
	return exec.Command("shpool", "kill", name).Run()
}

// parseShpoolSessions parses the output of shpool list.
// Each line is a session name.
func parseShpoolSessions(output string) []backend.Session {
	active := os.Getenv("SHPOOL_SESSION_NAME")
	var sessions []backend.Session
	for _, line := range strings.Split(output, "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
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
