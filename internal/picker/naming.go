package picker

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/nerveband/zpick/internal/zmosh"
)

// CounterName generates a session name like "dirname" or "dirname-N".
// Matches the zsh script's _zmosh_pick_name_counter logic.
func CounterName(dir string, existing []zmosh.Session) string {
	base := filepath.Base(dir)
	names := make(map[string]bool)
	for _, s := range existing {
		names[s.Name] = true
	}

	// First try: bare name
	if !names[base] {
		return base
	}

	// Increment: base-2, base-3, ...
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if !names[candidate] {
			return candidate
		}
	}
}

// DateName generates a session name like "dirname-MMDD".
// Matches the zsh script's _zmosh_pick_name_date logic.
func DateName(dir string) string {
	base := filepath.Base(dir)
	return fmt.Sprintf("%s-%s", base, time.Now().Format("0102"))
}
