package main

import (
	"github.com/nerveband/zpick/internal/update"
)

func runUpgrade() error {
	return update.Upgrade(version)
}
