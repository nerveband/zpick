package main

import (
	"fmt"
	"os"

	"github.com/nerveband/zpick/internal/hook"
)

func runInstallHook() error {
	for _, arg := range os.Args[2:] {
		switch arg {
		case "--remove":
			fmt.Fprintln(os.Stderr, "  note: --remove is deprecated, use 'zp remove-hook' instead")
			return hook.Remove()
		case "--guard":
			fmt.Fprintln(os.Stderr, "  note: --guard is deprecated, use 'zp install-guard' instead")
			return hook.InstallGuard()
		}
	}
	return hook.Install(false)
}

func runInstallGuard() error {
	return hook.InstallGuard()
}

func runRemoveHook() error {
	return hook.Remove()
}

func runRemoveGuard() error {
	return hook.RemoveGuard()
}

func runPostUpgradeHook(args []string) error {
	withGuard := false
	for _, arg := range args {
		if arg == "--guard" {
			withGuard = true
		}
	}
	return hook.PromptAndApplyHookUpdate(withGuard)
}
