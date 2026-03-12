package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookUsesResolvedBinaryWithoutPATHEntry(t *testing.T) {
	tests := []struct {
		name       string
		shell      string
		writeBlock func(t *testing.T, home string)
		args       func(shellPath, home string) []string
		env        func(home string) []string
	}{
		{
			name:  "zsh",
			shell: "zsh",
			writeBlock: func(t *testing.T, home string) {
				t.Helper()
				writeFile(t, filepath.Join(home, ".zshrc"), GenerateHookBlock(nil))
			},
			args: func(shellPath, home string) []string {
				return []string{shellPath, "-i", "-c", `zp version >"$HOME/version.txt"`}
			},
			env: func(home string) []string {
				return []string{"ZDOTDIR=" + home}
			},
		},
		{
			name:  "bash",
			shell: "bash",
			writeBlock: func(t *testing.T, home string) {
				t.Helper()
				writeFile(t, filepath.Join(home, ".bashrc"), GenerateBashHookBlock(nil))
			},
			args: func(shellPath, home string) []string {
				return []string{shellPath, "--rcfile", filepath.Join(home, ".bashrc"), "-i", "-c", `zp version >"$HOME/version.txt"`}
			},
			env: func(home string) []string {
				return nil
			},
		},
		{
			name:  "fish",
			shell: "fish",
			writeBlock: func(t *testing.T, home string) {
				t.Helper()
				confDir := filepath.Join(home, ".config", "fish", "conf.d")
				if err := os.MkdirAll(confDir, 0755); err != nil {
					t.Fatal(err)
				}
				writeFile(t, filepath.Join(confDir, "00-zp.fish"), GenerateFishHookBlock(nil))
			},
			args: func(shellPath, home string) []string {
				return []string{shellPath, "-i", "-c", `zp version > "$HOME/version.txt"`}
			},
			env: func(home string) []string {
				return []string{"XDG_CONFIG_HOME=" + filepath.Join(home, ".config")}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shellPath, err := exec.LookPath(tt.shell)
			if err != nil {
				t.Skipf("%s not installed", tt.shell)
			}

			home := t.TempDir()
			logPath := filepath.Join(home, "zp.log")
			writeStubZP(t, home)
			tt.writeBlock(t, home)

			env := append([]string{
				"HOME=" + home,
				"PATH=/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin",
				"TERM=xterm-256color",
				"ZP_SMOKE_LOG=" + logPath,
			}, tt.env(home)...)

			cmdArgs := tt.args(shellPath, home)
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", tt.name, err, string(out))
			}

			versionData, err := os.ReadFile(filepath.Join(home, "version.txt"))
			if err != nil {
				t.Fatalf("missing version output: %v", err)
			}
			if got := strings.TrimSpace(string(versionData)); got != "stub-version" {
				t.Fatalf("version output = %q, want stub-version", got)
			}

			logData, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("missing hook log: %v", err)
			}
			if got := strings.Fields(string(logData)); strings.Join(got, ",") != "should-autostart,BARE,version" {
				t.Fatalf("hook log = %q, want should-autostart,BARE,version", strings.Join(got, ","))
			}
		})
	}
}

func writeStubZP(t *testing.T, home string) {
	t.Helper()
	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}

	stub := strings.Join([]string{
		"#!/bin/sh",
		"log=\"${ZP_SMOKE_LOG:?}\"",
		"if [ $# -eq 0 ]; then",
		"  printf \"BARE\\n\" >>\"$log\"",
		"  printf \"true\\n\"",
		"  exit 0",
		"fi",
		"printf \"%s\\n\" \"$*\" >>\"$log\"",
		"case \"$1\" in",
		"  should-autostart) exit 0 ;;",
		"  version) printf \"stub-version\\n\"; exit 0 ;;",
		"  autorun|resume) printf \"true\\n\"; exit 0 ;;",
		"esac",
		"exit 0",
		"",
	}, "\n")

	writeFile(t, filepath.Join(binDir, "zp"), stub)
	if err := os.Chmod(filepath.Join(binDir, "zp"), 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
}
