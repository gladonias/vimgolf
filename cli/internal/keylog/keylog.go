package keylog

import (
	"fmt"
	"time"
)

var (
	noSniffStart = time.Date(2016, 4, 1, 0, 0, 0, 0, time.UTC)
	noSniffEnd   = time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC)
)

type Keystroke struct {
	Display string
}

func Parse(data []byte) []Keystroke {
	return ParseWithTime(data, time.Now().UTC())
}

func ParseWithTime(data []byte, t time.Time) []Keystroke {
	var strokes []Keystroke
	i := 0

	for i < len(data) {
		if data[i] == 0x80 && i+2 < len(data) {
			b2 := data[i+1]
			b3 := data[i+2]

			if b2 == 0xfd && b3 >= 0x38 && !t.Before(noSniffStart) && !t.After(noSniffEnd) {
				b3++
			}

			key := [2]byte{b2, b3}
			if name, ok := kcMbyte[key]; ok {
				if name != "" {
					strokes = append(strokes, Keystroke{Display: name})
				}
			} else {
				strokes = append(strokes, Keystroke{Display: fmt.Sprintf("<%02x-%02x>", b2, b3)})
			}
			i += 3
		} else {
			strokes = append(strokes, Keystroke{Display: kc1byte[data[i]]})
			i++
		}
	}

	return strokes
}

func Score(data []byte) int {
	return len(Parse(data))
}

var kc1byte [256]string

func init() {
	for n := 0; n < 256; n++ {
		kc1byte[n] = fmt.Sprintf("<%#04x>", n)
	}
	for n := 1; n < 128; n++ {
		kc1byte[n] = fmt.Sprintf("<C-%c>", byte(n^0x40))
	}
	for c := 32; c <= 126; c++ {
		kc1byte[c] = string(byte(c))
	}
	kc1byte[0x1b] = "<Esc>"
	kc1byte[0x0d] = "<CR>"
	kc1byte[0x0a] = "<NL>"
	kc1byte[0x09] = "<Tab>"
}

// kcMbyte maps the 2nd and 3rd bytes of a 0x80-prefixed sequence to a
// display string. Empty string means the keystroke is intentionally
// suppressed (e.g. window focus events).
var kcMbyte = map[[2]byte]string{
	// Function keys
	{'k', '1'}: "<F1>",
	{'k', '2'}: "<F2>",
	{'k', '3'}: "<F3>",
	{'k', '4'}: "<F4>",
	{'k', '5'}: "<F5>",
	{'k', '6'}: "<F6>",
	{'k', '7'}: "<F7>",
	{'k', '8'}: "<F8>",
	{'k', '9'}: "<F9>",
	{'k', ';'}: "<F10>",
	{'F', '1'}: "<F11>",
	{'F', '2'}: "<F12>",
	{'F', '3'}: "<F13>",
	{'F', '4'}: "<F14>",
	{'F', '5'}: "<F15>",
	{'F', '6'}: "<F16>",
	{'F', '7'}: "<F17>",
	{'F', '8'}: "<F18>",
	{'F', '9'}: "<F19>",

	// Special keys
	{'%', '1'}: "<Help>",
	{'&', '8'}: "<Undo>",
	{'#', '2'}: "<S-Home>",
	{'*', '7'}: "<S-End>",
	{'K', '1'}: "<kHome>",
	{'K', '4'}: "<kEnd>",
	{'K', '3'}: "<kPageUp>",
	{'K', '5'}: "<kPageDown>",
	{'K', '6'}: "<kPlus>",
	{'K', '7'}: "<kMinus>",
	{'K', '8'}: "<kDivide>",
	{'K', '9'}: "<kMultiply>",
	{'K', 'A'}: "<kEnter>",
	{'K', 'B'}: "<kPoint>",
	{'K', 'C'}: "<k0>",
	{'K', 'D'}: "<k1>",
	{'K', 'E'}: "<k2>",
	{'K', 'F'}: "<k3>",
	{'K', 'G'}: "<k4>",
	{'K', 'H'}: "<k5>",
	{'K', 'I'}: "<k6>",
	{'K', 'J'}: "<k7>",
	{'K', 'K'}: "<k8>",
	{'K', 'L'}: "<k9>",

	// Navigation
	{'k', 'P'}: "<PageUp>",
	{'k', 'N'}: "<PageDown>",
	{'k', 'h'}: "<Home>",
	{'@', '7'}: "<End>",
	{'k', 'I'}: "<Insert>",
	{'k', 'D'}: "<Del>",
	{'k', 'b'}: "<BS>",

	// Arrows
	{'k', 'u'}: "<Up>",
	{'k', 'd'}: "<Down>",
	{'k', 'l'}: "<Left>",
	{'k', 'r'}: "<Right>",
	{'#', '4'}: "<S-Left>",
	{'%', 'i'}: "<S-Right>",

	// Misc
	{'k', 'B'}: "<S-Tab>",
	{0xff, 'X'}: "<C-@>",

	// Literal 0x80 escape
	{0xfe, 'X'}: "<0x80>",

	// Modifier prefixes
	{0xfc, 0x02}: "<S->",
	{0xfc, 0x04}: "<C->",
	{0xfc, 0x06}: "<C-S->",
	{0xfc, 0x08}: "<A->",
	{0xfc, 0x0a}: "<A-S->",
	{0xfc, 0x0c}: "<C-A>",
	{0xfc, 0x0e}: "<C-A-S->",
	{0xfc, 0x10}: "<M->",
	{0xfc, 0x12}: "<M-S->",
	{0xfc, 0x14}: "<M-C->",
	{0xfc, 0x16}: "<M-C-S->",
	{0xfc, 0x18}: "<M-A->",
	{0xfc, 0x1a}: "<M-A-S->",
	{0xfc, 0x1c}: "<M-C-A>",
	{0xfc, 0x1e}: "<M-C-A-S->",

	// KS_EXTRA keycodes (0x80 0xfd ...)
	{0xfd, 0x04}: "<S-Up>",
	{0xfd, 0x05}: "<S-Down>",
	{0xfd, 0x06}: "<S-F1>",
	{0xfd, 0x07}: "<S-F2>",
	{0xfd, 0x08}: "<S-F3>",
	{0xfd, 0x09}: "<S-F4>",
	{0xfd, 0x0a}: "<S-F5>",
	{0xfd, 0x0b}: "<S-F6>",
	{0xfd, 0x0c}: "<S-F7>",
	{0xfd, 0x0d}: "<S-F9>",
	{0xfd, 0x0e}: "<S-F10>",
	{0xfd, 0x0f}: "<S-F10>",
	{0xfd, 0x10}: "<S-F11>",
	{0xfd, 0x11}: "<S-F12>",
	{0xfd, 0x12}: "<S-F13>",
	{0xfd, 0x13}: "<S-F14>",
	{0xfd, 0x14}: "<S-F15>",
	{0xfd, 0x15}: "<S-F16>",
	{0xfd, 0x16}: "<S-F17>",
	{0xfd, 0x17}: "<S-F18>",
	{0xfd, 0x18}: "<S-F19>",
	{0xfd, 0x19}: "<S-F20>",
	{0xfd, 0x1a}: "<S-F21>",
	{0xfd, 0x1b}: "<S-F22>",
	{0xfd, 0x1c}: "<S-F23>",
	{0xfd, 0x1d}: "<S-F24>",
	{0xfd, 0x1e}: "<S-F25>",
	{0xfd, 0x1f}: "<S-F26>",
	{0xfd, 0x20}: "<S-F27>",
	{0xfd, 0x21}: "<S-F28>",
	{0xfd, 0x22}: "<S-F29>",
	{0xfd, 0x23}: "<S-F30>",
	{0xfd, 0x24}: "<S-F31>",
	{0xfd, 0x25}: "<S-F32>",
	{0xfd, 0x26}: "<S-F33>",
	{0xfd, 0x27}: "<S-F34>",
	{0xfd, 0x28}: "<S-F35>",
	{0xfd, 0x29}: "<S-F36>",
	{0xfd, 0x2a}: "<S-F37>",
	{0xfd, 0x2b}: "<Mouse>",
	{0xfd, 0x2c}: "<LeftMouse>",
	{0xfd, 0x2d}: "<LeftDrag>",
	{0xfd, 0x2e}: "<LeftRelease>",
	{0xfd, 0x2f}: "<MiddleMouse>",
	{0xfd, 0x30}: "<MiddleDrag>",
	{0xfd, 0x31}: "<MiddleRelease>",
	{0xfd, 0x32}: "<RightMouse>",
	{0xfd, 0x33}: "<RightDrag>",
	{0xfd, 0x34}: "<RightRelease>",
	{0xfd, 0x35}: "", // KE_IGNORE

	// Scroll wheel
	{0xfd, 0x4b}: "<ScrollWheelUp>",
	{0xfd, 0x4c}: "<ScrollWheelDown>",
	{0xfd, 0x4d}: "<ScrollWheelRight>",
	{0xfd, 0x4e}: "<ScrollWheelLeft>",
	{0xfd, 0x4f}: "<kInsert>",
	{0xfd, 0x50}: "<kDel>",
	{0xfd, 0x51}: "<0x9b>",
	{0xfd, 0x53}: "<C-Left>",
	{0xfd, 0x54}: "<C-Right>",
	{0xfd, 0x55}: "<C-Left>",
	{0xfd, 0x56}: "<C-Right>",
	{0xfd, 0x57}: "<C-Home>",
	{0xfd, 0x58}: "<C-End>",
	{0xfd, 0x5e}: "", // 7.2 compat

	// Focus events (GVIM) — suppress
	{0xfd, 0x60}: "", // 7.2 Focus Gained compat
	{0xfd, 0x61}: "", // Focus Gained (>7.4.1433)
	{0xfd, 0x62}: "", // Focus Gained
	{0xfd, 0x63}: "", // Focus Lost
}
