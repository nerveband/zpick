package main

import (
	"fmt"

	"github.com/nerveband/zpick/internal/picker"
)

func runPicker() error {
	cmd, err := picker.Run()
	if err != nil {
		return err
	}
	if cmd != "" {
		fmt.Print(cmd)
	}
	return nil
}
