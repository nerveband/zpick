package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const symlinkPath = "/usr/local/bin/zp"

// InstallSymlink creates a symlink from /usr/local/bin/zp to ~/.local/bin/zp.
// Skips if the symlink already points to the right place.
// On failure, prompts for sudo and retries automatically.
func InstallSymlink() {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return
	}

	home, _ := os.UserHomeDir()
	target := filepath.Join(home, ".local", "bin", "zp")

	// Check target binary exists
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return
	}

	// If symlink already points to the right place, nothing to do
	if dest, err := os.Readlink(symlinkPath); err == nil && dest == target {
		return
	}

	// Try to create (or replace) the symlink without sudo first
	os.Remove(symlinkPath) // remove stale symlink if any; ok to fail
	if err := os.Symlink(target, symlinkPath); err == nil {
		fmt.Printf("  symlinked %s -> %s\n", symlinkPath, target)
		return
	}

	// Need elevated permissions — try sudo
	fmt.Printf("  creating %s symlink (requires sudo)...\n", symlinkPath)
	cmd := exec.Command("sudo", "ln", "-sf", target, symlinkPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("  note: could not create symlink — run manually: sudo ln -sf %s %s\n", target, symlinkPath)
		return
	}
	fmt.Printf("  symlinked %s -> %s\n", symlinkPath, target)
}

// CheckSymlink prints a note if the symlink doesn't exist.
// Called by `zp upgrade` — never auto-creates.
func CheckSymlink() {
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		fmt.Println("  note: /usr/local/bin/zp not found — run 'zp install-hook' to add it")
	}
}
