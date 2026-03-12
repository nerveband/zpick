package guard

import "testing"

func TestKeyActionForInput_EscapeRequiresSingleByte(t *testing.T) {
	if got := keyActionForInput([]byte{27}); got != keyEscape {
		t.Fatalf("expected keyEscape for bare esc, got %v", got)
	}

	if got := keyActionForInput([]byte{27, '[', 'A'}); got != keyOther {
		t.Fatalf("expected keyOther for arrow escape sequence, got %v", got)
	}

	if got := keyActionForInput([]byte{3}); got != keyEscape {
		t.Fatalf("expected ctrl-c to cancel, got %v", got)
	}
}

func TestReadMeaningfulKey_IgnoresOtherKeysUntilEnter(t *testing.T) {
	reader := scriptedReader(
		[]byte{'s'},
		[]byte{27, '[', 'A'},
		[]byte{13},
	)

	if got := readMeaningfulKey(reader); got != keyEnter {
		t.Fatalf("expected keyEnter, got %v", got)
	}
}

func scriptedReader(inputs ...[]byte) func([]byte) (int, error) {
	index := 0
	return func(buf []byte) (int, error) {
		if index >= len(inputs) {
			return 0, nil
		}
		input := inputs[index]
		index++
		copy(buf, input)
		return len(input), nil
	}
}
