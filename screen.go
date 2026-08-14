package gotermstyle

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Screen renders one help screen to one writer.
//
// It carries the buffered rows and the section counter that pytermstyle and
// formatting.sh both keep in globals. Go runs its tests in one process and its
// programs concurrently, so a value is the same code without the reset helper
// those two need, and makes concurrent use a non-question. The output is
// identical; only where the state lives differs.
type Screen struct {
	w       io.Writer
	palette Palette
	pending []row
	section string
	index   int
}

type row struct {
	name        string
	args        string
	description string
}

// New returns a Screen writing to w, with color resolved once against w.
func New(w io.Writer) *Screen {
	return &Screen{w: w, palette: PaletteFor(w)}
}

// NewWith returns a Screen using a palette the caller has already resolved.
// A tool that renders to several streams resolves once and shares it; a test
// asserts on the codes by passing NewPalette(true).
func NewWith(w io.Writer, p Palette) *Screen {
	return &Screen{w: w, palette: p}
}

// Palette is the resolved palette, for the parts of a screen this grammar does
// not render.
func (s *Screen) Palette() Palette {
	return s.palette
}

func (s *Screen) print(text string) {
	_, _ = fmt.Fprintln(s.w, text)
}

// Header draws a boxed section header: rule, title, rule, blank line. This is
// the standalone header a command prints over its own output, not the one that
// opens a help screen.
func (s *Screen) Header(title string) {
	s.print(s.palette.Cyan(Bar))
	s.print(s.palette.Cyan(" " + title))
	s.print(s.palette.Cyan(Bar))
	s.print("")
}

// HelpHeader opens a help screen. The blank line leads rather than trails,
// which is what separates the screen from the command that asked for it.
func (s *Screen) HelpHeader(name, tagline string) {
	s.index = 0
	s.print("")
	s.print(s.palette.Cyan(Bar))
	s.print(s.palette.Cyan(" " + name))
	s.print(s.palette.Cyan(Bar))
	if tagline != "" {
		s.print(tagline)
	}
}

// HelpUsage prints the invocation lines. The "Usage: " label belongs to this
// library rather than to the caller, so every screen on the fleet labels it
// identically. Extra lines align under the first.
func (s *Screen) HelpUsage(lines ...string) {
	s.print("")
	prefix := "Usage: "
	for _, line := range lines {
		s.print(s.palette.Command(prefix + line))
		prefix = "       "
	}
}

// HelpSection opens a section, flushing whatever rows the previous one buffered.
func (s *Screen) HelpSection(title string) {
	s.flush()
	s.section = title
	code := SectionColor(title, s.index)
	s.index++
	s.print("")
	s.print(title)
	rule := strings.Repeat("─", utf8.RuneCountInString(title)+sectionUnderline)
	s.print(s.palette.Paint(code, rule))
}

// HelpRow buffers one row. Nothing is written until the section ends, because
// the description column is sized from the longest row in the section and no
// call site should have to pass a width.
//
// A row with an empty name is a continuation of the one above it.
func (s *Screen) HelpRow(name, args, description string) {
	s.pending = append(s.pending, row{name: name, args: args, description: description})
}

// HelpText prints prose between rows. It flushes first, so a line written after
// a row cannot appear above it.
func (s *Screen) HelpText(lines ...string) {
	s.flush()
	for _, line := range lines {
		s.print(line)
	}
}

// HelpEnd closes the screen and resets it for the next one.
func (s *Screen) HelpEnd() {
	s.flush()
	s.section = ""
	s.index = 0
	s.print("")
}

// flush writes the buffered rows, sizing the column from the longest left side.
//
// The width is measured on the uncolored text and the color applied after. A
// padded field counts the escape bytes otherwise, and pushes every description
// right by the length of the escape. Measured in runes rather than bytes, which
// is where this improves on the Python: an em-dash is three bytes and one
// column, so a byte count slides the row two places left.
func (s *Screen) flush() {
	if len(s.pending) == 0 {
		return
	}
	// An example is a command you would type, so it reads in the command color
	// rather than the name color a listing uses.
	code := CodeCyan
	if strings.EqualFold(s.section, "examples") {
		code = CodeCommand
	}

	width := 0
	for _, r := range s.pending {
		if n := utf8.RuneCountInString(r.left()); n > width {
			width = n
		}
	}
	width += 2

	for _, r := range s.pending {
		rowCode := code
		if r.name == "" {
			rowCode = ""
		}
		trailing := ""
		if r.args != "" {
			trailing = " " + r.args
		}
		pad := width - utf8.RuneCountInString(r.left())
		if pad < 0 {
			pad = 0
		}
		line := "  " + s.palette.Paint(rowCode, r.name) + trailing + strings.Repeat(" ", pad) + r.description
		s.print(strings.TrimRight(line, " "))
	}

	s.pending = nil
}

// left is the padded column's content: the name and its args together, because
// the name is colored and the args are not, and they share one column.
func (r row) left() string {
	if r.args == "" {
		return r.name
	}
	return strings.TrimSpace(r.name + " " + r.args)
}
