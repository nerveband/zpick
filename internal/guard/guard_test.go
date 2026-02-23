package guard

import (
	"testing"
)

func TestEncodeDecodeArgv(t *testing.T) {
	tests := []struct {
		name string
		argv []string
	}{
		{"simple", []string{"claude"}},
		{"with args", []string{"claude", "--model", "opus"}},
		{"with spaces", []string{"my-tool", "arg with spaces"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encodeArgv(tt.argv)
			if encoded == "" {
				t.Fatal("encodeArgv returned empty")
			}

			decoded, err := DecodeArgv(encoded)
			if err != nil {
				t.Fatal(err)
			}

			if len(decoded) != len(tt.argv) {
				t.Fatalf("expected %d args, got %d", len(tt.argv), len(decoded))
			}
			for i, arg := range tt.argv {
				if decoded[i] != arg {
					t.Errorf("arg[%d]: expected %q, got %q", i, arg, decoded[i])
				}
			}
		})
	}
}

func TestEncodeArgvEmpty(t *testing.T) {
	if encodeArgv(nil) != "" {
		t.Error("nil argv should return empty")
	}
	if encodeArgv([]string{}) != "" {
		t.Error("empty argv should return empty")
	}
}

func TestDecodeArgvInvalid(t *testing.T) {
	if _, err := DecodeArgv("not-base64!!!"); err == nil {
		t.Error("invalid base64 should error")
	}

	// Valid base64 but not JSON
	if _, err := DecodeArgv("aGVsbG8="); err == nil {
		t.Error("non-JSON should error")
	}

	// Valid base64 JSON but empty array
	if _, err := DecodeArgv("W10="); err == nil {
		t.Error("empty array should error")
	}
}

func TestFormatArgv(t *testing.T) {
	tests := []struct {
		argv     []string
		expected string
	}{
		{nil, ""},
		{[]string{"claude"}, "claude"},
		{[]string{"claude", "--model", "opus"}, "claude --model opus"},
	}

	for _, tt := range tests {
		got := formatArgv(tt.argv)
		if got != tt.expected {
			t.Errorf("formatArgv(%v) = %q, want %q", tt.argv, got, tt.expected)
		}
	}
}
