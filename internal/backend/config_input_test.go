package backend

import "testing"

func TestBackendSelectionFromInput_SelectsChoice(t *testing.T) {
	available := []string{"zmosh", "tmux", "zellij"}

	selected, ok, cancelled := backendSelectionFromInput([]byte{'2'}, available)
	if cancelled {
		t.Fatal("selection should not be cancelled")
	}
	if !ok {
		t.Fatal("expected valid selection")
	}
	if selected != "tmux" {
		t.Fatalf("expected tmux, got %q", selected)
	}
}

func TestBackendSelectionFromInput_IgnoresInvalidChoice(t *testing.T) {
	available := []string{"zmosh", "tmux", "zellij"}

	if _, ok, cancelled := backendSelectionFromInput([]byte{'9'}, available); ok || cancelled {
		t.Fatal("out-of-range key should be ignored")
	}

	if _, ok, cancelled := backendSelectionFromInput([]byte{'x'}, available); ok || cancelled {
		t.Fatal("non-digit key should be ignored")
	}
}

func TestBackendSelectionFromInput_EscapeRequiresSingleByte(t *testing.T) {
	available := []string{"zmosh", "tmux"}

	if _, ok, cancelled := backendSelectionFromInput([]byte{27}, available); ok || !cancelled {
		t.Fatal("bare escape should cancel")
	}

	if _, ok, cancelled := backendSelectionFromInput([]byte{27, '[', 'A'}, available); ok || cancelled {
		t.Fatal("escape sequence should be ignored")
	}

	if _, ok, cancelled := backendSelectionFromInput([]byte{3}, available); ok || !cancelled {
		t.Fatal("ctrl-c should cancel")
	}
}
