package picker

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/nerveband/zpick/internal/backend"
)

// CounterName generates a session name like "dirname" or "dirname-N".
func CounterName(dir string, existing []backend.Session) string {
	base := filepath.Base(dir)
	names := make(map[string]bool)
	for _, s := range existing {
		names[s.Name] = true
	}

	if !names[base] {
		return base
	}

	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if !names[candidate] {
			return candidate
		}
	}
}

// DateName generates a session name like "dirname-MMDD".
func DateName(dir string) string {
	base := filepath.Base(dir)
	return fmt.Sprintf("%s-%s", base, time.Now().Format("0102"))
}
