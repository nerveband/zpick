package main

import (
	"os"
	"os/exec"

	"github.com/nerveband/zpick/internal/hook"
	"github.com/nerveband/zpick/internal/update"
)

func runUpgrade() error {
	hasHook, hasGuard := hook.HookStatus()

	upgraded, err := update.Upgrade(version)
	if err != nil {
		return err
	}
	if upgraded {
		if hasHook {
			return runPostUpgradeHookPrompt(hasGuard)
		} else {
			hook.CheckSymlink()
		}
	}
	return nil
}

func runPostUpgradeHookPrompt(hasGuard bool) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{"post-upgrade-hook"}
	if hasGuard {
		args = append(args, "--guard")
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
