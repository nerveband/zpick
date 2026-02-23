package guard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateName(t *testing.T) {
	valid := []string{"claude", "codex", "opencode", "my-app", "app2", "A_tool"}
	for _, name := range valid {
		if err := ValidateName(name); err != nil {
			t.Errorf("expected %q to be valid, got: %v", name, err)
		}
	}

	invalid := []string{"", "123", "foo bar", "a/b", "$cmd", "-flag", ".hidden"}
	for _, name := range invalid {
		if err := ValidateName(name); err == nil {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}

func TestParseConfig(t *testing.T) {
	content := `# Apps guarded by zpick
claude
codex

# Another comment
opencode
claude
`
	apps := parseConfig(content)

	if len(apps) != 3 {
		t.Fatalf("expected 3 apps (deduped), got %d: %v", len(apps), apps)
	}
	if apps[0] != "claude" || apps[1] != "codex" || apps[2] != "opencode" {
		t.Errorf("unexpected apps: %v", apps)
	}
}

func TestParseConfigEmpty(t *testing.T) {
	apps := parseConfig("")
	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestWriteAndReadConfig(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "zpick")
	configPath := filepath.Join(configDir, "guard.conf")

	// Override config path via XDG_CONFIG_HOME
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Verify config path uses our temp dir
	if ConfigPath() != configPath {
		t.Fatalf("expected config at %s, got %s", configPath, ConfigPath())
	}

	// Write config
	apps := []string{"claude", "codex", "aider"}
	if err := WriteConfig(apps); err != nil {
		t.Fatal(err)
	}

	// Read back
	got, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 apps, got %d", len(got))
	}
	if got[0] != "claude" || got[1] != "codex" || got[2] != "aider" {
		t.Errorf("unexpected: %v", got)
	}
}

func TestWriteConfigDeduplicates(t *testing.T) {
	dir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	if err := WriteConfig([]string{"claude", "codex", "claude"}); err != nil {
		t.Fatal(err)
	}
	got, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 apps (deduped), got %d: %v", len(got), got)
	}
}

func TestReadConfigDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "nonexistent"))
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	apps, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != len(DefaultApps) {
		t.Fatalf("expected %d default apps, got %d", len(DefaultApps), len(apps))
	}
}

func TestAddApp(t *testing.T) {
	dir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Create initial config
	WriteConfig([]string{"claude"})

	if err := AddApp("aider"); err != nil {
		t.Fatal(err)
	}

	apps, _ := ReadConfig()
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}

	// Adding duplicate should error
	if err := AddApp("aider"); err == nil {
		t.Error("adding duplicate should error")
	}

	// Adding invalid name should error
	if err := AddApp("bad name"); err == nil {
		t.Error("adding invalid name should error")
	}
}

func TestRemoveApp(t *testing.T) {
	dir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	WriteConfig([]string{"claude", "codex", "aider"})

	if err := RemoveApp("codex"); err != nil {
		t.Fatal(err)
	}

	apps, _ := ReadConfig()
	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}
	if apps[0] != "claude" || apps[1] != "aider" {
		t.Errorf("unexpected: %v", apps)
	}

	// Removing non-existent should error
	if err := RemoveApp("nothere"); err == nil {
		t.Error("removing non-existent should error")
	}
}

func TestFuncName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"claude", "claude"},
		{"my-app", "my_app"},
		{"open-code", "open_code"},
		{"a-b-c", "a_b_c"},
	}
	for _, tt := range tests {
		got := FuncName(tt.input)
		if got != tt.expected {
			t.Errorf("FuncName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestEnsureConfig(t *testing.T) {
	dir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Should create config
	if err := EnsureConfig(); err != nil {
		t.Fatal(err)
	}

	// Should exist now
	if _, err := os.Stat(ConfigPath()); err != nil {
		t.Error("config file should exist after EnsureConfig")
	}

	// Should be idempotent
	if err := EnsureConfig(); err != nil {
		t.Fatal(err)
	}
}
