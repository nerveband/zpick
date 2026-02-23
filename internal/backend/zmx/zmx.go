package zmx

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	zmoshpkg "github.com/nerveband/zpick/internal/backend/zmosh"
)

func init() {
	backend.Register("zmx", func() backend.Backend { return New() })
}

// Zmx implements the Backend interface for zmx (standalone zmx without zmosh).
type Zmx struct{}

func New() *Zmx { return &Zmx{} }

func (z *Zmx) Name() string         { return "zmx" }
func (z *Zmx) BinaryName() string   { return "zmx" }
func (z *Zmx) SessionEnvVar() string { return "ZMX_SESSION" }

func (z *Zmx) InSession() bool {
	return os.Getenv("ZMX_SESSION") != ""
}

func (z *Zmx) Available() (bool, error) {
	_, err := exec.LookPath("zmx")
	if err != nil {
		return false, fmt.Errorf("zmx not found in PATH")
	}
	return true, nil
}

func (z *Zmx) Version() (string, error) {
	out, err := exec.Command("zmx", "version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run zmx version: %w", err)
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

func (z *Zmx) List() ([]backend.Session, error) {
	out, err := exec.Command("zmx", "list").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run zmx list: %w", err)
	}
	// zmx uses the same output format as zmosh
	return zmoshpkg.ParseSessions(string(out)), nil
}

func (z *Zmx) FastList() ([]backend.Session, error) {
	dir, err := zmoshpkg.ResolveZmxDir()
	if err != nil {
		return z.List()
	}
	sessions, err := zmoshpkg.FastListDir(dir)
	if err != nil {
		return z.List()
	}
	return sessions, nil
}

func (z *Zmx) Attach(name string) error {
	zmxPath, err := exec.LookPath("zmx")
	if err != nil {
		return fmt.Errorf("zmx not found: %w", err)
	}
	return backend.ExecCommand(zmxPath, []string{"zmx", "attach", name})
}

func (z *Zmx) AttachCommand(name, dir string) string {
	cmd := fmt.Sprintf(`zmx attach "%s"`, name)
	if dir != "" {
		return fmt.Sprintf(`cd "%s" && %s`, dir, cmd)
	}
	return cmd
}

func (z *Zmx) Kill(name string) error {
	return exec.Command("zmx", "kill", name).Run()
}
