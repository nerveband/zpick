package main

import (
	"encoding/json"
	"fmt"

	"github.com/nerveband/zpick/internal/backend"
)

// ListResult is the JSON output format for `zpick list --json`.
// Keeps zmosh_version for backwards compatibility, adds backend_version.
type ListResult struct {
	Sessions       []backend.Session `json:"sessions"`
	Count          int               `json:"count"`
	ZmoshVersion   string            `json:"zmosh_version,omitempty"`
	BackendVersion string            `json:"backend_version,omitempty"`
}

func runList() error {
	jsonOutput := hasJSONFlag()

	// Use non-interactive for --json paths
	b, err := loadBackend(!jsonOutput)
	if err != nil {
		return err
	}

	if jsonOutput {
		sessions, err := b.List()
		if err != nil {
			return err
		}
		result := ListResult{
			Sessions: sessions,
			Count:    len(sessions),
		}
		if ver, err := b.Version(); err == nil {
			result.BackendVersion = ver
			// Keep zmosh_version for backwards compat
			if b.Name() == "zmosh" || b.Name() == "zmx" {
				result.ZmoshVersion = ver
			}
		}
		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}

	sessions, err := b.List()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("  no sessions")
		return nil
	}

	for _, s := range sessions {
		status := "."
		if s.Active {
			status = "*"
		}
		fmt.Printf("  %s%s  (%d clients)  %s\n", status, s.Name, s.Clients, s.StartedIn)
	}
	return nil
}
