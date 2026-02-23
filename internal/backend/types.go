package backend

import (
	"os"
	"syscall"
)

// Session represents a session from any backend.
type Session struct {
	Name      string `json:"name"`
	PID       int    `json:"pid,omitempty"`
	Clients   int    `json:"clients"`
	StartedIn string `json:"started_in"`
	Active    bool   `json:"active"`
}

// Backend is the interface that all session managers implement.
type Backend interface {
	// Identity
	Name() string        // "zmosh", "zmx", "tmux", "shpool"
	BinaryName() string  // binary to look up in PATH
	SessionEnvVar() string // env var set inside a session

	// Probing
	InSession() bool
	Available() (bool, error)
	Version() (string, error)

	// Runtime
	List() ([]Session, error)
	FastList() ([]Session, error)
	Attach(name string) error
	AttachCommand(name, dir string) string
	Kill(name string) error
}

// AllSessionEnvVars returns env var names from all known backends.
// Used by hook generation to check if we're inside any session.
func AllSessionEnvVars() []string {
	return []string{"ZMX_SESSION", "TMUX", "SHPOOL_SESSION_NAME"}
}

// ExecCommand replaces the current process with the given command.
func ExecCommand(path string, argv []string) error {
	return syscall.Exec(path, argv, os.Environ())
}
