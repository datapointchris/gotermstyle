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
// # Status
//
// Scaffolded, not implemented. The design, the measurements behind it, and the
// libraries that were considered and rejected are in .planning/design.md.
package gotermstyle
