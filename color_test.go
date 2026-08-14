package gotermstyle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// clearGate removes every variable the gate reads, so a test states its own
// conditions rather than inheriting the shell that ran `go test`.
func clearGate(t *testing.T) {
	t.Helper()
	for _, name := range []string{"NO_COLOR", "FORCE_COLOR", "TERM"} {
		t.Setenv(name, "")
		_ = os.Unsetenv(name)
	}
}

// redirected is a real *os.File that is not a terminal, which is what a redirect
// gives a command and what the gate has to recognize.
func redirected(t *testing.T) *os.File {
	t.Helper()
	f, err := os.Create(filepath.Join(t.TempDir(), "out"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func TestNoColorOutranksForceColor(t *testing.T) {
	// The two answer different questions. NO_COLOR is the user's preference and
	// FORCE_COLOR only overrides the terminal detection, so a FORCE_COLOR
	// inherited from a CI image must not overrule someone who asked for none.
	clearGate(t)
	t.Setenv("NO_COLOR", "1")
	t.Setenv("FORCE_COLOR", "1")
	if Enabled(os.Stdout) {
		t.Error("NO_COLOR lost to FORCE_COLOR")
	}
}

func TestForceColorBeatsANonTerminal(t *testing.T) {
	clearGate(t)
	t.Setenv("FORCE_COLOR", "1")
	if !Enabled(redirected(t)) {
		t.Error("FORCE_COLOR did not override the terminal detection")
	}
}

func TestADumbTerminalGetsNoColor(t *testing.T) {
	clearGate(t)
	t.Setenv("TERM", "dumb")
	if Enabled(os.Stdout) {
		t.Error("TERM=dumb still emitted color")
	}
}

func TestAFileIsNotATerminal(t *testing.T) {
	// `tool --help > notes.txt` has to write text, not escape sequences.
	clearGate(t)
	if Enabled(redirected(t)) {
		t.Error("a redirect was treated as a terminal")
	}
}

func TestAWriterThatIsNotAFileIsNotATerminal(t *testing.T) {
	clearGate(t)
	t.Setenv("TERM", "xterm-256color")
	if Enabled(&bytes.Buffer{}) {
		t.Error("a buffer was treated as a terminal")
	}
}

func TestADisabledPaletteEmitsNoEscapes(t *testing.T) {
	p := NewPalette(false)
	for _, got := range []string{
		p.Cyan("x"), p.Yellow("x"), p.Green("x"), p.Red("x"), p.Blue("x"),
		p.Magenta("x"), p.White("x"), p.Command("x"), p.Bold("x"),
		p.Paint(CodeCyan, "x"),
	} {
		if got != "x" {
			t.Errorf("disabled palette returned %q", got)
		}
	}
	if p.Code(CodeCyan) != "" {
		t.Errorf("Code returned %q with color off", p.Code(CodeCyan))
	}
}

func TestAnEnabledPaletteWrapsAndResets(t *testing.T) {
	p := NewPalette(true)
	if got, want := p.Cyan("x"), CodeCyan+"x"+CodeReset; got != want {
		t.Errorf("Cyan = %q, want %q", got, want)
	}
	if got := p.Code(CodeCyan); got != CodeCyan {
		t.Errorf("Code = %q, want the sequence", got)
	}
}

func TestPaintWithNoCodeLeavesTextAlone(t *testing.T) {
	// A continuation row carries no name and so gets no color, and it must not
	// pick up a bare reset either.
	if got := NewPalette(true).Paint("", "x"); got != "x" {
		t.Errorf("Paint with an empty code returned %q", got)
	}
}

func TestTheUniversalSectionRolesHoldOneColorWhereverTheySit(t *testing.T) {
	// A color is learnable across tools only if it never shifts with position.
	for _, c := range []struct {
		title string
		index int
		want  string
	}{
		{"Commands", 0, CodeCyan},
		{"Commands", 7, CodeCyan},
		{"Options", 3, CodeMagenta},
		{"Examples", 5, CodeYellow},
		{"Environment Variables", 2, CodeMagenta},
	} {
		if got := SectionColor(c.title, c.index); got != c.want {
			t.Errorf("SectionColor(%q, %d) = %q, want %q", c.title, c.index, got, c.want)
		}
	}
}

func TestSectionColorsMatchWhateverCaseTheHeadingUses(t *testing.T) {
	if SectionColor("OPTIONS", 0) != CodeMagenta {
		t.Error("uppercase heading missed the fixed color")
	}
	if SectionColor("examples", 0) != CodeYellow {
		t.Error("lowercase heading missed the fixed color")
	}
}

func TestAppSpecificSectionsRotateSoNeighborsDiffer(t *testing.T) {
	// A single fallback color rendered a screen whose headings were all
	// app-specific entirely monochrome.
	var rotated []string
	for i := range 4 {
		rotated = append(rotated, SectionColor("Whatever", i))
	}
	if rotated[0] == rotated[1] || rotated[1] == rotated[2] {
		t.Errorf("neighbors share a color: %q", rotated)
	}
	if rotated[3] != rotated[0] {
		t.Error("the rotation is three long and should wrap")
	}
}

func TestTheHouseCodesAreTheOnesDotfilesExports(t *testing.T) {
	// Three implementations emit this palette. Changing one without the other
	// two is the failure this library exists to prevent, so the codes are
	// asserted literally rather than derived.
	for _, c := range []struct{ got, want, name string }{
		{CodeCyan, "\x1b[0;96m", "cyan"},
		{CodeYellow, "\x1b[0;93m", "yellow"},
		{CodeGreen, "\x1b[0;92m", "green"},
		{CodeRed, "\x1b[0;91m", "red"},
		{CodeBlue, "\x1b[0;94m", "blue"},
		{CodeMagenta, "\x1b[0;95m", "magenta"},
		{CodeWhite, "\x1b[0;97m", "white"},
		{CodeCommand, "\x1b[0;36m", "command"},
		{CodeBold, "\x1b[1m", "bold"},
		{CodeReset, "\x1b[0m", "reset"},
	} {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestNothingInThePaletteIsDim(t *testing.T) {
	// standards/cli-design.md bans it outright for being unreadable against
	// half the terminal themes in use.
	for _, code := range []string{
		CodeCyan, CodeYellow, CodeGreen, CodeRed, CodeBlue,
		CodeMagenta, CodeWhite, CodeCommand, CodeBold, CodeReset,
	} {
		if strings.Contains(code, "[2m") {
			t.Errorf("%q is dim", code)
		}
	}
}

func TestTheBarIsFiftyColumns(t *testing.T) {
	// formatting.sh draws its rule at 50, so a Go screen and a bash screen
	// sitting next to each other draw the same one.
	if got := len([]rune(Bar)); got != 50 {
		t.Errorf("Bar is %d columns, want 50", got)
	}
}
