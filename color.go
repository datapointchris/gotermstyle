package gotermstyle

import (
	"io"
	"os"
	"strings"
)

// The house palette, as the escape sequences themselves. These are the codes
// `~/dotfiles/configs/common/.local/shell/colors.sh` exports and pytermstyle
// interpolates, so a Go screen, a Python screen and a bash screen sitting next
// to each other on PATH are the same product.
//
// They are terminal-palette codes rather than absolute RGB, which is why no
// color-profile degradation is needed: a 16-color terminal renders them
// correctly, and TERM=dumb is handled by the gate instead.
//
// There is deliberately no dim. See standards/cli-design.md § "Never use
// `[dim]`", which rules it out for being unreadable against half the terminal
// themes in use.
const (
	CodeCyan    = "\033[0;96m"
	CodeYellow  = "\033[0;93m"
	CodeGreen   = "\033[0;92m"
	CodeRed     = "\033[0;91m"
	CodeBlue    = "\033[0;94m"
	CodeMagenta = "\033[0;95m"
	CodeWhite   = "\033[0;97m"
	// CodeCommand is the one non-bright entry. A command you would type reads
	// differently from a name in a listing, and that difference is the point.
	CodeCommand = "\033[0;36m"
	CodeBold    = "\033[1m"
	CodeReset   = "\033[0m"
)

// Bar is the rule a header draws above and below its title.
const Bar = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

// sectionUnderline is how much wider a section's rule is than its title.
const sectionUnderline = 15

// fixedSections are the roles that recur in nearly every tool. They hold one
// color regardless of position, because a color is only learnable across
// tools if it never shifts.
var fixedSections = map[string]string{
	"commands":              CodeCyan,
	"verbs":                 CodeCyan,
	"suites":                CodeCyan,
	"groups":                CodeCyan,
	"phases":                CodeCyan,
	"options":               CodeMagenta,
	"flags":                 CodeMagenta,
	"arguments":             CodeMagenta,
	"environment variables": CodeMagenta,
	"examples":              CodeYellow,
}

// rotatingSections color the headings a tool invents for itself. They rotate
// by position so neighbors differ; a single fallback rendered a screen whose
// sections were all app-specific entirely monochrome.
var rotatingSections = []string{CodeGreen, CodeBlue, CodeWhite}

// SectionColor is the code a section heading renders in, by name first and by
// position second. Matching is case-insensitive because a heading is written
// for a reader rather than for this lookup.
func SectionColor(title string, index int) string {
	if code, ok := fixedSections[strings.ToLower(title)]; ok {
		return code
	}
	if index < 0 {
		index = -index
	}
	return rotatingSections[index%len(rotatingSections)]
}

// Palette decides whether the codes above are emitted. Resolve it once, against
// the stream the output is going to, and pass it around — re-deciding per call
// site is how one screen ends up half-colored.
type Palette struct {
	enabled bool
}

// NewPalette returns a palette that emits color, or does not.
func NewPalette(enabled bool) Palette {
	return Palette{enabled: enabled}
}

// PaletteFor resolves the gate against w. This is the constructor to reach for;
// NewPalette is for a caller that has already decided, and for tests.
func PaletteFor(w io.Writer) Palette {
	return Palette{enabled: Enabled(w)}
}

// Enabled reports whether escape sequences written to w will reach something
// that can render them.
//
// NO_COLOR is the user saying they do not want color (no-color.org), so it
// wins outright. FORCE_COLOR answers a different question — "is this a
// terminal" — for a caller that knows the detection is wrong: a CI log that
// renders escapes, a pager invoked with -R, a test that needs the codes in
// order to assert on them. Preference beats detection, so NO_COLOR is first.
//
// The terminal check is os.Stat plus os.ModeCharDevice rather than
// golang.org/x/term, which would cost this package the zero-dependency
// guarantee it is built around. It is the idiom goselfupdate/autoupdate uses.
func Enabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// On reports whether this palette emits color. A caller building a string by
// interpolation rather than through Paint needs to ask.
func (p Palette) On() bool {
	return p.enabled
}

// Code returns the sequence when color is on and the empty string when it is
// not, so an interpolating caller can use the constants directly.
func (p Palette) Code(code string) string {
	if !p.enabled {
		return ""
	}
	return code
}

// Paint wraps text in code and resets after it. An empty code returns the text
// untouched, which is what a continuation row wants.
func (p Palette) Paint(code, text string) string {
	if !p.enabled || code == "" {
		return text
	}
	return code + text + CodeReset
}

// The named colors, for the call sites that read better without the constant.
func (p Palette) Cyan(text string) string    { return p.Paint(CodeCyan, text) }
func (p Palette) Yellow(text string) string  { return p.Paint(CodeYellow, text) }
func (p Palette) Green(text string) string   { return p.Paint(CodeGreen, text) }
func (p Palette) Red(text string) string     { return p.Paint(CodeRed, text) }
func (p Palette) Blue(text string) string    { return p.Paint(CodeBlue, text) }
func (p Palette) Magenta(text string) string { return p.Paint(CodeMagenta, text) }
func (p Palette) White(text string) string   { return p.Paint(CodeWhite, text) }
func (p Palette) Command(text string) string { return p.Paint(CodeCommand, text) }
func (p Palette) Bold(text string) string    { return p.Paint(CodeBold, text) }
