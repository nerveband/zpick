package main

import (
	"fmt"

	"github.com/nerveband/zpick/internal/picker"
)

func runPicker() error {
	b, err := loadBackend(true)
	if err != nil {
		return err
	}
	cmd, err := picker.Run(b, version)
	if err != nil {
		return err
	}
	if cmd != "" {
		fmt.Print(cmd)
	}
	return nil
}
