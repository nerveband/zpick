package zmosh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
)

func init() {
	backend.Register("zmosh", func() backend.Backend { return New() })
}

// Zmosh implements the Backend interface for zmosh.
type Zmosh struct{}

// New creates a new zmosh backend.
func New() *Zmosh {
	return &Zmosh{}
}

func (z *Zmosh) Name() string        { return "zmosh" }
func (z *Zmosh) BinaryName() string  { return "zmosh" }
func (z *Zmosh) SessionEnvVar() string { return "ZMX_SESSION" }

func (z *Zmosh) InSession() bool {
	return os.Getenv("ZMX_SESSION") != ""
}

func (z *Zmosh) Available() (bool, error) {
	_, err := exec.LookPath("zmosh")
	if err != nil {
		return false, fmt.Errorf("zmosh not found in PATH")
	}
	return true, nil
}

func (z *Zmosh) Version() (string, error) {
	out, err := exec.Command("zmosh", "version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run zmosh version: %w", err)
	}
	ver := strings.TrimSpace(string(out))
	if idx := strings.IndexByte(ver, '\n'); idx >= 0 {
		ver = strings.TrimSpace(ver[:idx])
	}
	fields := strings.Fields(ver)
	if len(fields) >= 2 {
		ver = fields[len(fields)-1]
	}
	return ver, nil
}

func (z *Zmosh) List() ([]backend.Session, error) {
	out, err := exec.Command("zmosh", "list").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run zmosh list: %w", err)
	}
	return ParseSessions(string(out)), nil
}

func (z *Zmosh) FastList() ([]backend.Session, error) {
	dir, err := ResolveZmxDir()
	if err != nil {
		return z.List()
	}
	sessions, err := FastListDir(dir)
	if err != nil {
		return z.List()
	}
	return sessions, nil
}

func (z *Zmosh) Attach(name string) error {
	zmoshPath, err := exec.LookPath("zmosh")
	if err != nil {
		return fmt.Errorf("zmosh not found: %w", err)
	}
	return backend.ExecCommand(zmoshPath, []string{"zmosh", "attach", name})
}

func (z *Zmosh) AttachCommand(name, dir string) string {
	// Check UDP config for -r flag
	enabled, host := backend.ReadUDP()
	var attachCmd string
	if enabled && host != "" {
		attachCmd = fmt.Sprintf(`zmosh attach -r %s "%s"`, host, name)
	} else {
		attachCmd = fmt.Sprintf(`zmosh attach "%s"`, name)
	}

	if dir != "" {
		return fmt.Sprintf(`cd "%s" && %s`, dir, attachCmd)
	}
	return attachCmd
}

func (z *Zmosh) Kill(name string) error {
	return exec.Command("zmosh", "kill", name).Run()
}
