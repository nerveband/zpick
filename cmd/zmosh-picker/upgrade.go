package main

import (
	"github.com/nerveband/zmosh-picker/internal/update"
)

func runUpgrade() error {
	return update.Upgrade(version)
}
