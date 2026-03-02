package main

import (
	"github.com/nerveband/zpick/internal/hook"
	"github.com/nerveband/zpick/internal/update"
)

func runUpgrade() error {
	upgraded, err := update.Upgrade(version)
	if err != nil {
		return err
	}
	if upgraded {
		// Auto-update the shell hook so it stays current
		if hook.HasHookInstalled() {
			hook.Install(hook.HasGuardInstalled())
		} else {
			hook.CheckSymlink()
		}
	}
	return nil
}
