package gotermstyle

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

var escapes = regexp.MustCompile("\x1b\\[[0-9;]*m")

// plain strips the escapes, so an assertion about a column measures what the
// reader sees rather than what the bytes say.
func plain(s string) string { return escapes.ReplaceAllString(s, "") }

// render runs a screen with color on and returns its lines.
func render(t *testing.T, build func(*Screen)) []string {
	t.Helper()
	var b bytes.Buffer
	s := NewWith(&b, NewPalette(true))
	build(s)
	return strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
}

// column reports how many display columns precede needle, counted in runes so
// a multi-byte cell earlier in the row does not itself read as an offset.
func column(t *testing.T, line, needle string) int {
	t.Helper()
	at := strings.Index(line, needle)
	if at < 0 {
		t.Fatalf("%q does not contain %q", line, needle)
	}
	return utf8.RuneCountInString(line[:at])
}

// lineWith returns the one rendered line containing needle.
func lineWith(t *testing.T, lines []string, needle string) string {
	t.Helper()
	var hits []string
	for _, line := range lines {
		if strings.Contains(line, needle) {
			hits = append(hits, line)
		}
	}
	if len(hits) != 1 {
		t.Fatalf("wanted one line containing %q, got %d in:\n%s", needle, len(hits), strings.Join(lines, "\n"))
	}
	return hits[0]
}

func TestHelpRowsAlignTheirDescriptions(t *testing.T) {
	// The whole reason widths are measured on the uncolored text. A padded
	// field counting escape bytes pushes one description right by the length of
	// the escape.
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("short", "", "first")
		s.HelpRow("a-much-longer-name", "<arg>", "second")
		s.HelpEnd()
	})
	first := column(t, plain(lineWith(t, lines, "first")), "first")
	second := column(t, plain(lineWith(t, lines, "second")), "second")
	if first != second {
		t.Errorf("descriptions start at %d and %d:\n%s", first, second, strings.Join(lines, "\n"))
	}
}

func TestRowsStillAlignWhenColorIsOff(t *testing.T) {
	var b bytes.Buffer
	s := NewWith(&b, NewPalette(false))
	s.HelpSection("Commands")
	s.HelpRow("short", "", "first")
	s.HelpRow("a-much-longer-name", "<arg>", "second")
	s.HelpEnd()

	lines := strings.Split(b.String(), "\n")
	first := column(t, lineWith(t, lines, "first"), "first")
	second := column(t, lineWith(t, lines, "second"), "second")
	if first != second {
		t.Errorf("descriptions start at %d and %d:\n%s", first, second, b.String())
	}
}

func TestAMultiByteNameDoesNotShiftTheDescription(t *testing.T) {
	// Where this improves on the Python it ports. An em-dash is three bytes and
	// one column, so a byte-measured width slides the row two places left.
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("a—b", "", "first")
		s.HelpRow("abc", "", "second")
		s.HelpEnd()
	})
	first := column(t, plain(lineWith(t, lines, "first")), "first")
	second := column(t, plain(lineWith(t, lines, "second")), "second")
	if first != second {
		t.Errorf("descriptions start at %d and %d:\n%s", first, second, strings.Join(lines, "\n"))
	}
}

func TestArgsShareTheNameColumn(t *testing.T) {
	// The args are uncolored but sit inside the padded column, so a row with
	// args must not push its description past a row without them.
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("name", "<arg>", "described")
		s.HelpEnd()
	})
	row := plain(lineWith(t, lines, "described"))
	if !strings.HasPrefix(row, "  name <arg>") {
		t.Errorf("row began %q", row)
	}
	if !strings.HasSuffix(row, "described") {
		t.Errorf("row ended %q", row)
	}
}

func TestARowWithNoDescriptionCarriesNoTrailingPadding(t *testing.T) {
	// A section of bare commands should not trail invisible whitespace.
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("bare", "", "")
		s.HelpEnd()
	})
	if got := plain(lineWith(t, lines, "bare")); got != "  bare" {
		t.Errorf("row = %q, want %q", got, "  bare")
	}
}

func TestHelpTextFlushesPendingRowsFirst(t *testing.T) {
	// Rows buffer until flush, so prose written between them would otherwise
	// appear above rows declared before it.
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("a-row", "", "described")
		s.HelpText("  some prose")
		s.HelpEnd()
	})
	rowAt, proseAt := -1, -1
	for i, line := range lines {
		switch {
		case strings.Contains(line, "a-row"):
			rowAt = i
		case strings.Contains(line, "some prose"):
			proseAt = i
		}
	}
	if rowAt < 0 || proseAt < 0 || rowAt > proseAt {
		t.Errorf("row at %d, prose at %d:\n%s", rowAt, proseAt, strings.Join(lines, "\n"))
	}
}

func TestANewSectionFlushesTheOneBeforeIt(t *testing.T) {
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("first-row", "", "one")
		s.HelpSection("Options")
		s.HelpRow("second-row", "", "two")
		s.HelpEnd()
	})
	rowAt, headingAt := -1, -1
	for i, line := range lines {
		switch {
		case strings.Contains(line, "first-row"):
			rowAt = i
		case plain(line) == "Options":
			headingAt = i
		}
	}
	if rowAt < 0 || headingAt < 0 || rowAt > headingAt {
		t.Errorf("first row at %d, second heading at %d:\n%s", rowAt, headingAt, strings.Join(lines, "\n"))
	}
}

func TestEachSectionSizesItsOwnColumn(t *testing.T) {
	// Sizing across the screen would pad a short section to the width of the
	// longest name anywhere on it.
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("a-very-long-command-name", "", "one")
		s.HelpSection("Options")
		s.HelpRow("-x", "", "two")
		s.HelpEnd()
	})
	if got := column(t, plain(lineWith(t, lines, "two")), "two"); got > 10 {
		t.Errorf("short section padded to %d columns:\n%s", got, strings.Join(lines, "\n"))
	}
}

func TestExamplesRowsUseTheCommandColor(t *testing.T) {
	// An example is something you would type, so it reads as a command rather
	// than as a listing name.
	lines := render(t, func(s *Screen) {
		s.HelpSection("Examples")
		s.HelpRow("some command", "", "does a thing")
		s.HelpEnd()
	})
	if row := lineWith(t, lines, "does a thing"); !strings.Contains(row, CodeCommand) {
		t.Errorf("examples row was not in the command color: %q", row)
	}
}

func TestAContinuationRowGetsNoColorEscape(t *testing.T) {
	lines := render(t, func(s *Screen) {
		s.HelpSection("Commands")
		s.HelpRow("named", "", "one")
		s.HelpRow("", "", "carries on")
		s.HelpEnd()
	})
	if row := lineWith(t, lines, "carries on"); escapes.MatchString(row) {
		t.Errorf("continuation row carried an escape: %q", row)
	}
}

func TestHelpEndResetsTheScreenForTheNextOne(t *testing.T) {
	var b bytes.Buffer
	s := NewWith(&b, NewPalette(true))
	s.HelpSection("Examples")
	s.HelpRow("x", "", "y")
	s.HelpEnd()
	b.Reset()

	// A second screen starts at rotation zero and is no longer in Examples, so
	// its rows take the listing color rather than the command one.
	s.HelpSection("Commands")
	s.HelpRow("a", "", "b")
	s.HelpEnd()
	if row := lineWith(t, strings.Split(b.String(), "\n"), "b"); strings.Contains(row, CodeCommand) {
		t.Errorf("the Examples color survived HelpEnd: %q", row)
	}
}

func TestHelpHeaderResetsTheSectionRotation(t *testing.T) {
	var b bytes.Buffer
	s := NewWith(&b, NewPalette(true))
	s.HelpSection("Alpha")
	s.HelpSection("Beta")
	first := SectionColor("Alpha", 0)

	b.Reset()
	s.HelpHeader("tool", "")
	s.HelpSection("Alpha")
	if rule := lineWith(t, strings.Split(b.String(), "\n"), "───"); !strings.Contains(rule, first) {
		t.Error("HelpHeader did not reset the rotation to zero")
	}
}

func TestUsageLabelsEveryLineAndAlignsTheRest(t *testing.T) {
	lines := render(t, func(s *Screen) {
		s.HelpUsage("tool verb", "tool other")
	})
	first := plain(lineWith(t, lines, "tool verb"))
	second := plain(lineWith(t, lines, "tool other"))
	if !strings.HasPrefix(first, "Usage: ") {
		t.Errorf("first usage line = %q", first)
	}
	if column(t, first, "tool") != column(t, second, "tool") {
		t.Errorf("usage lines do not align:\n%q\n%q", first, second)
	}
}

func TestATaglineIsOptional(t *testing.T) {
	lines := render(t, func(s *Screen) { s.HelpHeader("tool", "") })
	for _, line := range lines {
		if plain(line) == "" {
			continue
		}
		if !strings.Contains(plain(line), "tool") && !strings.Contains(line, "━") {
			t.Errorf("unexpected line with no tagline: %q", plain(line))
		}
	}
}

func TestHeaderDrawsARuleAboveAndBelowItsTitle(t *testing.T) {
	lines := render(t, func(s *Screen) { s.Header("Results") })
	if plain(lines[0]) != Bar {
		t.Errorf("first line = %q", plain(lines[0]))
	}
	if plain(lines[1]) != " Results" {
		t.Errorf("title line = %q", plain(lines[1]))
	}
	if plain(lines[2]) != Bar {
		t.Errorf("third line = %q", plain(lines[2]))
	}
}

func TestTheSectionRuleOutrunsItsTitle(t *testing.T) {
	lines := render(t, func(s *Screen) { s.HelpSection("Commands") })
	rule := plain(lineWith(t, lines, "───"))
	if got, want := len([]rune(rule)), len("Commands")+sectionUnderline; got != want {
		t.Errorf("rule is %d columns, want %d", got, want)
	}
}

func TestNewResolvesTheGateAgainstItsWriter(t *testing.T) {
	// A buffer is not a terminal, so a Screen built on one writes plain text
	// without the caller having to say so.
	clearGate(t)
	var b bytes.Buffer
	s := New(&b)
	s.HelpSection("Commands")
	s.HelpRow("name", "", "described")
	s.HelpEnd()
	if escapes.MatchString(b.String()) {
		t.Errorf("color leaked onto a non-terminal:\n%q", b.String())
	}
}
