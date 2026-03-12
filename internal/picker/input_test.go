package picker

import (
	"testing"

	"github.com/nerveband/zpick/internal/backend"
)

func TestPickerActionForInput_AttachesAvailableSession(t *testing.T) {
	LoadKeyMode("letters")
	defer LoadKeyMode("numbers")

	sessions := []backend.Session{
		{Name: "alpha"},
		{Name: "beta"},
	}

	action := pickerActionForInput([]byte{'b'}, sessions)
	if action.Type != ActionAttach {
		t.Fatalf("expected ActionAttach, got %v", action.Type)
	}
	if action.Name != "beta" {
		t.Fatalf("expected beta, got %q", action.Name)
	}
}

func TestPickerActionForInput_OutOfRangeSessionKeyRetries(t *testing.T) {
	LoadKeyMode("letters")
	defer LoadKeyMode("numbers")

	sessions := []backend.Session{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "delta"},
		{Name: "echo"},
		{Name: "foxtrot"},
	}

	action := pickerActionForInput([]byte{'s'}, sessions)
	if action.Type != ActionRetry {
		t.Fatalf("expected ActionRetry for out-of-range key, got %v", action.Type)
	}
}

func TestPickerActionForInput_InvalidKeyRetries(t *testing.T) {
	action := pickerActionForInput([]byte{'!'}, nil)
	if action.Type != ActionRetry {
		t.Fatalf("expected ActionRetry for invalid key, got %v", action.Type)
	}
}

func TestPickerActionForInput_EscapeRequiresSingleByte(t *testing.T) {
	if action := pickerActionForInput([]byte{27}, nil); action.Type != ActionEscape {
		t.Fatalf("expected escape for bare esc, got %v", action.Type)
	}

	if action := pickerActionForInput([]byte{27, '[', 'A'}, nil); action.Type != ActionRetry {
		t.Fatalf("expected retry for escape sequence, got %v", action.Type)
	}

	if action := pickerActionForInput([]byte{3}, nil); action.Type != ActionEscape {
		t.Fatalf("expected escape for ctrl-c, got %v", action.Type)
	}
}
