// Package gotermstyle is the house terminal style for the Go command-line tools.
//
// It is the Go counterpart to two things that already exist: pytermstyle
// (https://github.com/datapointchris/pytermstyle) for the Python tools, and
// formatting.sh for the bash ones. The README says where each lives. The
// parallel is the point —
// a Go screen, a Python screen and a bash screen sitting next to each other on
// PATH should be indistinguishable, which is a property no third-party styling
// library can provide, because each brings its own aesthetic.
//
// # Scope
//
// The palette, the section header, the help grammar, and the rules about when
// any of it is emitted at all. Not domain logic, not configuration, not
// anything a single tool could own instead.
//
// The membership test: it belongs here only if two tools would otherwise write
// it identically *and* writing it differently would be visible on screen.
// Everything else stays in the tool. The package is named for the language
// rather than for a caller, because a style package named for one family of
// tools ends up holding the house style for all of them.
//
// # Dependencies
//
// None, and that is a constraint rather than a coincidence. Terminal detection
// is os.Stat plus os.ModeCharDevice, not golang.org/x/term, so a tool importing
// this downloads no other module. The go line in this module's go.mod does
// become the importer's floor, and go get raises a lower one to match. This
// prints it:
//
//	go list -m -f '{{.GoVersion}}' github.com/datapointchris/gotermstyle@latest
//
// # Using it
//
// Resolve a Screen against the writer the output is going to, then render:
//
//	s := gotermstyle.New(cmd.OutOrStdout())
//	s.HelpHeader("lab", "Small experiments worth revisiting.")
//	s.HelpUsage("lab <verb>")
//
//	s.HelpSection("Commands")
//	s.HelpRow("lab list", "", "List every lab")
//	s.HelpRow("lab show", "<id>", "Show one lab")
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
// screen is the same; only the state's owner differs.
//
// The libraries that were considered and rejected, and why, are in the README.
package gotermstyle
