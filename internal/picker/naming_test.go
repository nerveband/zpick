package picker

import (
	"strings"
	"testing"

	"github.com/nerveband/zpick/internal/backend"
)

func TestCounterName_NoConflict(t *testing.T) {
	name := CounterName("projects", nil)
	if name != "projects" {
		t.Errorf("expected 'projects', got '%s'", name)
	}
}

func TestCounterName_WithConflict(t *testing.T) {
	existing := []backend.Session{
		{Name: "projects"},
	}
	name := CounterName("projects", existing)
	if name != "projects-2" {
		t.Errorf("expected 'projects-2', got '%s'", name)
	}
}

func TestCounterName_MultipleConflicts(t *testing.T) {
	existing := []backend.Session{
		{Name: "projects"},
		{Name: "projects-2"},
		{Name: "projects-3"},
	}
	name := CounterName("projects", existing)
	if name != "projects-4" {
		t.Errorf("expected 'projects-4', got '%s'", name)
	}
}

func TestCounterName_FromPath(t *testing.T) {
	name := CounterName("/Users/nerveband/Documents/GitHub/my-project", nil)
	if name != "my-project" {
		t.Errorf("expected 'my-project', got '%s'", name)
	}
}

func TestDateName(t *testing.T) {
	name := DateName("/Users/nerveband/projects")
	if !strings.HasPrefix(name, "projects-") {
		t.Errorf("expected projects-MMDD format, got '%s'", name)
	}
	if len(name) != len("projects-0000") {
		t.Errorf("expected projects-MMDD length, got '%s' (len=%d)", name, len(name))
	}
}
