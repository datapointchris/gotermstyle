package tuitheme

import (
	"fmt"
	"strconv"
	"strings"
)

type colorKind uint8

const (
	unset colorKind = iota
	slot
	rgb
)

// Color is one of a style's two colors: unset, which is the terminal's own, a
// slot of the terminal's 16, or an exact RGB value.
type Color struct {
	kind    colorKind
	slot    uint8
	r, g, b uint8
}

// parseColor reads the file's two color forms, and the empty string as unset.
func parseColor(s string) (Color, error) {
	if s == "" {
		return Color{}, nil
	}
	if strings.HasPrefix(s, "#") {
		v, err := strconv.ParseUint(s[1:], 16, 32)
		if len(s) != 7 || err != nil {
			return Color{}, fmt.Errorf("%q is neither #rrggbb nor a slot from 0 to 15", s)
		}
		return Color{kind: rgb, r: uint8(v >> 16), g: uint8(v >> 8), b: uint8(v)}, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 15 || strconv.Itoa(n) != s {
		return Color{}, fmt.Errorf("%q is neither #rrggbb nor a slot from 0 to 15", s)
	}
	return Color{kind: slot, slot: uint8(n)}, nil
}

// String is the color as the file writes it: "#rrggbb", a slot from "0" to
// "15", or "" when unset.
func (c Color) String() string {
	switch c.kind {
	case slot:
		return strconv.Itoa(int(c.slot))
	case rgb:
		return fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b)
	}
	return ""
}

// Slot is the terminal palette slot this color names, if it names one.
func (c Color) Slot() (int, bool) {
	return int(c.slot), c.kind == slot
}

// RGB is this color's exact value, if it has one.
func (c Color) RGB() (r, g, b uint8, ok bool) {
	return c.r, c.g, c.b, c.kind == rgb
}

// sgr is the color's SGR parameters, where base is 30 for a foreground and 40
// for a background.
func (c Color) sgr(base int) string {
	switch c.kind {
	case slot:
		if c.slot < 8 {
			return strconv.Itoa(base + int(c.slot))
		}
		return strconv.Itoa(base + 60 + int(c.slot) - 8)
	case rgb:
		return fmt.Sprintf("%d;2;%d;%d;%d", base+8, c.r, c.g, c.b)
	}
	return ""
}

// Style is what one role draws in.
type Style struct {
	Fg, Bg        Color
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Dim           bool
}

// Code is the style as one SGR sequence, for gotermstyle.Palette.Paint. It
// opens with a reset, as the house palette's codes do, so it replaces whatever
// came before. A style that sets nothing is "", which Paint leaves unpainted.
func (s Style) Code() string {
	var params []string
	for _, attr := range []struct {
		on   bool
		code string
	}{{s.Bold, "1"}, {s.Dim, "2"}, {s.Italic, "3"}, {s.Underline, "4"}, {s.Strikethrough, "9"}} {
		if attr.on {
			params = append(params, attr.code)
		}
	}
	if fg := s.Fg.sgr(30); fg != "" {
		params = append(params, fg)
	}
	if bg := s.Bg.sgr(40); bg != "" {
		params = append(params, bg)
	}
	if len(params) == 0 {
		return ""
	}
	return "\033[0;" + strings.Join(params, ";") + "m"
}
