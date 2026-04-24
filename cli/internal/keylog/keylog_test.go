package keylog

import (
	"strings"
	"testing"
	"time"
)

func displays(strokes []Keystroke) []string {
	out := make([]string, len(strokes))
	for i, s := range strokes {
		out[i] = s.Display
	}
	return out
}

func TestSingleByteChars(t *testing.T) {
	strokes := Parse([]byte("hello"))
	if len(strokes) != 5 {
		t.Fatalf("expected 5 strokes, got %d", len(strokes))
	}
	for i, ch := range "hello" {
		if strokes[i].Display != string(ch) {
			t.Errorf("stroke %d: expected %q, got %q", i, string(ch), strokes[i].Display)
		}
	}
}

func TestSpecialSingleBytes(t *testing.T) {
	tests := []struct {
		input byte
		want  string
	}{
		{0x1b, "<Esc>"},
		{0x0d, "<CR>"},
		{0x0a, "<NL>"},
		{0x09, "<Tab>"},
		{0x01, "<C-A>"},
		{0x04, "<C-D>"},
		{0x12, "<C-R>"},
	}
	for _, tt := range tests {
		strokes := Parse([]byte{tt.input})
		if len(strokes) != 1 {
			t.Fatalf("byte 0x%02x: expected 1 stroke, got %d", tt.input, len(strokes))
		}
		if strokes[0].Display != tt.want {
			t.Errorf("byte 0x%02x: expected %q, got %q", tt.input, tt.want, strokes[0].Display)
		}
	}
}

func TestMultiByteKeys(t *testing.T) {
	tests := []struct {
		b2, b3 byte
		want   string
	}{
		{'k', 'u', "<Up>"},
		{'k', 'd', "<Down>"},
		{'k', 'l', "<Left>"},
		{'k', 'r', "<Right>"},
		{'k', '1', "<F1>"},
		{'k', 'b', "<BS>"},
		{'k', 'D', "<Del>"},
		{'k', 'h', "<Home>"},
		{'@', '7', "<End>"},
	}
	for _, tt := range tests {
		data := []byte{0x80, tt.b2, tt.b3}
		strokes := Parse(data)
		if len(strokes) != 1 {
			t.Fatalf("key %c%c: expected 1 stroke, got %d", tt.b2, tt.b3, len(strokes))
		}
		if strokes[0].Display != tt.want {
			t.Errorf("key %c%c: expected %q, got %q", tt.b2, tt.b3, tt.want, strokes[0].Display)
		}
	}
}

func TestFocusEventsIgnored(t *testing.T) {
	data := []byte{
		'a',
		0x80, 0xfd, 0x62, // Focus Gained
		'b',
		0x80, 0xfd, 0x63, // Focus Lost
		'c',
	}
	strokes := Parse(data)
	d := displays(strokes)
	result := strings.Join(d, "")
	if result != "abc" {
		t.Errorf("expected 'abc', got %q", result)
	}
}

func TestKEIgnore(t *testing.T) {
	data := []byte{0x80, 0xfd, 0x35, 'x'}
	strokes := Parse(data)
	if len(strokes) != 1 || strokes[0].Display != "x" {
		t.Errorf("expected just 'x', got %v", displays(strokes))
	}
}

func TestScore(t *testing.T) {
	// "dd" = 2 keystrokes
	if s := Score([]byte("dd")); s != 2 {
		t.Errorf("expected score 2, got %d", s)
	}

	// Escape + "dd" + ":wq" + CR = 1+2+3+1 = 7
	data := []byte{0x1b, 'd', 'd', ':', 'w', 'q', 0x0d}
	if s := Score(data); s != 7 {
		t.Errorf("expected score 7, got %d", s)
	}
}

func TestMixedKeystrokes(t *testing.T) {
	data := []byte{
		'i',                // insert mode
		'h', 'i',          // type "hi"
		0x1b,              // escape
		0x80, 'k', 'u',   // up arrow
		':',  'w', 'q',   // :wq
		0x0d,              // enter
	}
	strokes := Parse(data)
	d := displays(strokes)
	expected := []string{"i", "h", "i", "<Esc>", "<Up>", ":", "w", "q", "<CR>"}
	if len(d) != len(expected) {
		t.Fatalf("expected %d strokes, got %d: %v", len(expected), len(d), d)
	}
	for i := range expected {
		if d[i] != expected[i] {
			t.Errorf("stroke %d: expected %q, got %q", i, expected[i], d[i])
		}
	}
}

func TestKESniffDateRange(t *testing.T) {
	// Within the no-sniff date range, 0xfd 0x38 should be shifted to 0x39
	inRange := time.Date(2016, 6, 1, 0, 0, 0, 0, time.UTC)
	data := []byte{0x80, 0xfd, 0x38}

	strokesIn := ParseWithTime(data, inRange)

	// Outside the date range, 0xfd 0x38 stays as-is
	outRange := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	strokesOut := ParseWithTime(data, outRange)

	if len(strokesIn) != 1 || len(strokesOut) != 1 {
		t.Fatalf("expected 1 stroke each, got %d and %d", len(strokesIn), len(strokesOut))
	}

	// Inside range: byte is shifted, so the lookup differs from outside range
	if strokesIn[0].Display == strokesOut[0].Display {
		t.Logf("in-range: %q, out-range: %q", strokesIn[0].Display, strokesOut[0].Display)
		t.Error("expected different results inside vs outside KE_SNIFF date range")
	}
}

func TestUnknownMultiByte(t *testing.T) {
	data := []byte{0x80, 0xaa, 0xbb}
	strokes := Parse(data)
	if len(strokes) != 1 {
		t.Fatalf("expected 1 stroke, got %d", len(strokes))
	}
	if strokes[0].Display != "<aa-bb>" {
		t.Errorf("expected '<aa-bb>', got %q", strokes[0].Display)
	}
}
