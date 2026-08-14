// Package gotermstyle is the house terminal style for the Go command-line tools.
//
// It is the Go counterpart to two things that already exist: dotfiles'
// appcore/formatting.py for the Python apps, and ~/.local/shell/formatting.sh
// for the bash ones. The parallel is the point — a Go screen, a Python screen
// and a bash screen sitting next to each other on PATH should be
// indistinguishable, which is a property no third-party styling library can
// provide, because each brings its own aesthetic.
//
// # Scope
//
// The palette, the section header, the help grammar, and the rules about when
// any of it is emitted at all. Not domain logic, not configuration, not
// anything a single tool could own instead.
//
// The membership test: it belongs here only if two tools would otherwise write
// it identically *and* writing it differently would be visible on screen.
// Everything else stays in the tool. menucore was named for one family and
// ended up holding the house style for all of them; the split that fixed it is
// the reason this package is named for the language rather than for a caller.
//
// # Dependencies
//
// None, and that is a constraint rather than a coincidence — the same one
// goselfupdate holds. Terminal detection is os.Stat plus os.ModeCharDevice, not
// golang.org/x/term, so a tool importing this downloads nothing and its own Go
// floor is unaffected.
//
// # Using it
//
// Resolve a Screen against the writer the output is going to, then render:
//
//	s := gotermstyle.New(cmd.OutOrStdout())
//	s.HelpHeader("menu labs", "Small experiments worth revisiting.")
//	s.HelpUsage("menu labs <verb>")
//
//	s.HelpSection("Commands")
//	s.HelpRow("menu labs list", "", "List every lab")
//	s.HelpRow("menu labs show", "<id>", "Show one lab")
//
//	s.HelpEnd()
//
// HelpRow buffers rather than writing, so the flush sizes the description
// column from the longest row in the section. No call site passes a width or a
// color. Use HelpText rather than writing directly for prose between rows, so
// pending rows flush ahead of it.
//
// For output that is not a help screen, Palette paints and Clip fits:
//
//	p := gotermstyle.PaletteFor(os.Stdout)
//	fmt.Println(p.Cyan(name), p.Command(path))
//
// # Where the state lives, and why it differs from the Python
//
// pytermstyle and formatting.sh keep the buffered rows and the section counter
// in globals, which is affordable in a one-shot CLI and costs pytermstyle a
// reset helper its tests have to call. Go runs its tests in one process and its
// programs concurrently, so that state lives on a Screen here. The rendered
// output is identical; only its owner differs.
//
// The design, the measurements behind it, and the libraries that were
// considered and rejected are in .planning/design.md.
package gotermstyle
