package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got := configDir()
	want := filepath.Join(dir, "zpick")
	if got != want {
		t.Errorf("configDir() = %q, want %q", got, want)
	}
}

func TestConfigDirDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()

	got := configDir()
	want := filepath.Join(home, ".config", "zpick")
	if got != want {
		t.Errorf("configDir() = %q, want %q", got, want)
	}
}

func TestSetBackendAndReadBackend(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := SetBackend("tmux"); err != nil {
		t.Fatal(err)
	}

	got, err := readBackendConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got != "tmux" {
		t.Errorf("readBackendConfig() = %q, want %q", got, "tmux")
	}
}

func TestReadBackendConfigMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got, err := readBackendConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("readBackendConfig() = %q, want empty", got)
	}
}

func TestSetUDPConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := SetUDP(true, "myhost"); err != nil {
		t.Fatal(err)
	}

	enabled, host := ReadUDP()
	if !enabled {
		t.Error("expected UDP enabled")
	}
	if host != "myhost" {
		t.Errorf("host = %q, want %q", host, "myhost")
	}
}

func TestSetUDPConfigDisabled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := SetUDP(false, ""); err != nil {
		t.Fatal(err)
	}

	enabled, host := ReadUDP()
	if enabled {
		t.Error("expected UDP disabled")
	}
	if host != "" {
		t.Errorf("host = %q, want empty", host)
	}
}

func TestReadUDPDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	enabled, host := ReadUDP()
	if !enabled {
		t.Error("expected UDP enabled by default")
	}
	if host != "" {
		t.Errorf("host = %q, want empty by default", host)
	}
}

func TestDetectReturnsAvailableBackends(t *testing.T) {
	// Detect should return a list of available backend names.
	// We can't predict which are installed, but it should be a non-nil slice.
	backends := Detect()
	if backends == nil {
		t.Error("Detect() should return non-nil slice")
	}
}

func TestSetBackendValidation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := SetBackend("invalid-backend"); err == nil {
		t.Error("expected error for invalid backend name")
	}
}
