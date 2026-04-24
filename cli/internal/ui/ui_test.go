package ui

import (
	"testing"
)

func TestExtractChoiceLetters(t *testing.T) {
	choices := []string{
		"[d] Show diff",
		"[r] Retry",
		"[q] Quit",
	}

	valid := make(map[string]bool)
	for _, c := range choices {
		if len(c) >= 3 && c[0] == '[' && c[2] == ']' {
			valid[string(c[1])] = true
		}
	}

	if !valid["d"] || !valid["r"] || !valid["q"] {
		t.Error("expected d, r, q to be valid choices")
	}

	if valid["w"] || valid["x"] || valid["s"] {
		t.Error("w, x, s should not be valid for this choice set")
	}
}

func TestPrintKeystrokeDoesNotPanic(t *testing.T) {
	PrintKeystroke("a")
	PrintKeystroke("<Esc>")
	PrintKeystroke("<CR>")
	PrintKeystroke("")
}
