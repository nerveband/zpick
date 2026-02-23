package check

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
)

// DepStatus represents the installation status of a dependency.
type DepStatus struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
}

// Result represents the full dependency check result.
type Result struct {
	Zmosh             DepStatus `json:"zmosh"`
	Zoxide            DepStatus `json:"zoxide"`
	Fzf               DepStatus `json:"fzf"`
	Shell             string    `json:"shell"`
	OS                string    `json:"os"`
	Arch              string    `json:"arch"`
	Backend           string    `json:"backend,omitempty"`
	AvailableBackends []string  `json:"available_backends,omitempty"`
}

// JSON returns the result as indented JSON.
func (r Result) JSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	return string(b), err
}

// Run checks all dependencies and returns the result.
func Run() Result {
	r := Result{
		Shell: detectShell(),
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
	}
	r.Zmosh = checkDep("zmosh", "version")
	r.Zoxide = checkDep("zoxide", "--version")
	r.Fzf = checkDep("fzf", "--version")

	// Backend info
	r.AvailableBackends = backend.Detect()
	if name, err := backend.ReadBackendName(); err == nil && name != "" {
		r.Backend = name
	} else if len(r.AvailableBackends) == 1 {
		r.Backend = r.AvailableBackends[0]
	}

	return r
}

func checkDep(name, versionFlag string) DepStatus {
	path, err := exec.LookPath(name)
	if err != nil {
		return DepStatus{Installed: false}
	}
	status := DepStatus{Installed: true, Path: path}
	if out, err := exec.Command(name, versionFlag).Output(); err == nil {
		ver := strings.TrimSpace(string(out))
		if idx := strings.IndexByte(ver, '\n'); idx >= 0 {
			ver = strings.TrimSpace(ver[:idx])
		}
		if name == "zmosh" {
			fields := strings.Fields(ver)
			if len(fields) >= 2 {
				ver = fields[len(fields)-1]
			}
		}
		status.Version = ver
	}
	return status
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}
	return "unknown"
}

// HasBrew checks if Homebrew is available.
func HasBrew() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// PrintHuman prints the check result in a human-readable format.
func (r Result) PrintHuman() {
	printDep("zmosh", r.Zmosh, true)
	printDep("zoxide", r.Zoxide, false)
	printDep("fzf", r.Fzf, false)
	fmt.Printf("\nPlatform: %s/%s, Shell: %s\n", r.OS, r.Arch, r.Shell)
	if r.Backend != "" {
		fmt.Printf("Backend: %s\n", r.Backend)
	}
	if len(r.AvailableBackends) > 0 {
		fmt.Printf("Available: %s\n", strings.Join(r.AvailableBackends, ", "))
	}
}

// PrintGuide prints a guided installation walkthrough for missing dependencies.
func (r Result) PrintGuide() bool {
	hasBrew := HasBrew()
	missing := false

	fmt.Println("\n  Dependency check:")
	fmt.Println()

	// Show current backend
	if r.Backend != "" {
		fmt.Printf("  \033[1;36mBackend:\033[0m %s\n", r.Backend)
	}
	if len(r.AvailableBackends) > 0 {
		fmt.Printf("  \033[2mAvailable:\033[0m %s\n\n", strings.Join(r.AvailableBackends, ", "))
	}

	// zmosh (required only if zmosh is the backend)
	isZmoshBackend := r.Backend == "zmosh" || r.Backend == "zmx" || r.Backend == ""
	if r.Zmosh.Installed {
		fmt.Printf("  \033[32m\u2713\033[0m zmosh %s\n", r.Zmosh.Version)
	} else if isZmoshBackend {
		missing = true
		fmt.Printf("  \033[1;31m\u2717\033[0m zmosh \033[2m(required)\033[0m\n")
		fmt.Println()
		if hasBrew {
			fmt.Println("    Install with Homebrew:")
			fmt.Println("      brew install mmonad/tap/zmosh")
		} else if r.OS == "darwin" {
			fmt.Println("    First install Homebrew:")
			fmt.Println("      /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"")
			fmt.Println("    Then:")
			fmt.Println("      brew install mmonad/tap/zmosh")
		} else {
			fmt.Println("    Build from source:")
			fmt.Println("      git clone https://github.com/mmonad/zmosh.git")
			fmt.Println("      cd zmosh && zig build -Doptimize=ReleaseSafe")
			fmt.Println("      sudo cp zig-out/bin/zmosh /usr/local/bin/")
		}
		fmt.Println()
		fmt.Println("    More info: https://github.com/mmonad/zmosh")
		fmt.Println()
	} else {
		fmt.Printf("  \033[33m\u25CB\033[0m zmosh \033[2m(not using this backend)\033[0m\n")
	}

	// zoxide (optional)
	if r.Zoxide.Installed {
		fmt.Printf("  \033[32m\u2713\033[0m zoxide %s \033[2m(optional — directory picker)\033[0m\n", r.Zoxide.Version)
	} else {
		fmt.Printf("  \033[33m\u25CB\033[0m zoxide \033[2m(optional — enables 'z' directory picker)\033[0m\n")
		fmt.Println()
		if hasBrew {
			fmt.Println("    Install with Homebrew:")
			fmt.Println("      brew install zoxide")
		} else {
			fmt.Println("    Install:")
			fmt.Println("      curl -sSfL https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | sh")
		}
		fmt.Println()
		fmt.Println("    More info: https://github.com/ajeetdsouza/zoxide")
		fmt.Println()
	}

	// fzf (optional)
	if r.Fzf.Installed {
		fmt.Printf("  \033[32m\u2713\033[0m fzf %s \033[2m(optional — fuzzy finder for zoxide)\033[0m\n", r.Fzf.Version)
	} else {
		fmt.Printf("  \033[33m\u25CB\033[0m fzf \033[2m(optional — fuzzy finder used by zoxide)\033[0m\n")
		fmt.Println()
		if hasBrew {
			fmt.Println("    Install with Homebrew:")
			fmt.Println("      brew install fzf")
		} else {
			fmt.Println("    Install:")
			fmt.Println("      git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf")
			fmt.Println("      ~/.fzf/install")
		}
		fmt.Println()
		fmt.Println("    More info: https://github.com/junegunn/fzf")
		fmt.Println()
	}

	fmt.Printf("\n  Platform: %s/%s, Shell: %s\n", r.OS, r.Arch, r.Shell)

	if missing {
		fmt.Println("\n  \033[1;31mRequired dependencies missing.\033[0m Install them and run again.")
	} else {
		fmt.Println("\n  \033[32mAll set!\033[0m Run 'zp install-hook' to add the shell hook.")
	}

	return missing
}

func printDep(name string, d DepStatus, required bool) {
	status := "\u2713"
	if !d.Installed {
		if required {
			status = "\u2717"
		} else {
			status = "\u25CB"
		}
	}
	label := " (optional)"
	if required {
		label = " (required)"
	}
	if d.Installed {
		fmt.Printf("  %s %s%s \u2014 %s\n", status, name, label, d.Version)
	} else {
		fmt.Printf("  %s %s%s \u2014 not found\n", status, name, label)
	}
}
