# gotermstyle

A Go library that styles a command-line tool's output. It holds a fixed
palette, a boxed section header, a help-screen grammar, and the rule for when
any of it reaches the terminal. Its purpose is a Go tool whose help screen
looks the same as its Python and bash neighbors on `PATH`. Those neighbors are
styled by pytermstyle and `formatting.sh`. The section on the other two says
where each lives.

It renders and does nothing else. The caller keeps its command framework, its
execution and its help function, and calls this from inside them. Real layout
belongs to [lipgloss](https://github.com/charmbracelet/lipgloss), which has
tables, lists, layers and wide-rune width. None of those are here.

Something belongs in this package only if two tools would otherwise write it
identically and a difference would show on screen. Domain logic and
configuration stay in the tool. So does anything one tool could own alone.

The module requires nothing outside the standard library, by its own choice.

## Install

```bash
go get github.com/datapointchris/gotermstyle@latest
```

Releases are git tags on the 0.x line. Under semantic versioning, a 0.x minor
release may change the API.

The module's `go` line becomes the floor of every module that imports it.
`go get` raises a lower `go` line to match. This prints the floor:

```bash
go list -m -f '{{.GoVersion}}' github.com/datapointchris/gotermstyle@latest
```

## Starting from scratch

Save this as `main.go` in an empty directory:

```go
package main

import (
    "os"

    "github.com/datapointchris/gotermstyle"
)

func main() {
    s := gotermstyle.New(os.Stdout)

    s.HelpHeader("hello", "Greet people from the terminal.")
    s.HelpUsage("hello <verb>")

    s.HelpSection("Commands")
    s.HelpRow("hello greet", "<name>", "Greet one person")
    s.HelpRow("hello list", "", "List everyone greeted")

    s.HelpEnd()
}
```

Then, from that directory:

```bash
go mod init example.com/hello    # a module to import it into
go mod tidy                      # resolves the import; go.sum gains its lines and no others
go run .                         # cyan rules and names, because stdout is a terminal
go run . | cat                   # the same screen uncolored, because a pipe is not a terminal
NO_COLOR=1 go run .              # uncolored, because the user's preference outranks everything
FORCE_COLOR=1 go run . | cat -v  # the escape codes made visible, proving they reached the bytes
```

Color is emitted only where something will render it. `New` resolves that once,
against the writer it is given. `FORCE_COLOR` overrides the terminal detection
for a caller who knows it is wrong. A CI log that renders escapes is one case.
A pager run with `-R` is another. `NO_COLOR` still wins, because it is the
user's preference and `FORCE_COLOR` only corrects the detection.

A writer that is not an `*os.File`, such as a `bytes.Buffer`, never gets color
from `New`. A test that asserts on the codes builds its own palette:
`gotermstyle.NewWith(&buf, gotermstyle.NewPalette(true))`.

`go doc -all github.com/datapointchris/gotermstyle` prints the whole API. Run
it inside a module that requires the package.

## `Screen` — rows buffer per section, so no call site passes a width or a color

In a cobra tool, the screen is drawn from the function given to `SetHelpFunc`,
against `cmd.OutOrStdout()`:

```go
root.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
    s := gotermstyle.New(cmd.OutOrStdout())

    s.HelpHeader("lab", "Small experiments worth revisiting.")
    s.HelpUsage("lab <verb>")

    s.HelpSection("Commands")
    s.HelpRow("lab list", "", "List every lab")
    s.HelpRow("lab show", "<id>", "Show one lab")
    s.HelpRow("", "", "An id comes from list") // continues the row above

    s.HelpEnd()
})
```

`HelpRow` buffers rather than writing. The next `HelpSection`, `HelpText` or
`HelpEnd` flushes the buffer. The flush sizes the description column from the
longest name and arguments in the section. A row with an empty name continues
the row above it, with its description aligned and no color.

`HelpText` prints prose between rows. It flushes first, so pending rows land
above the prose rather than under it. `HelpEnd` closes the screen and resets
it for the next one.

The `Usage:` label belongs to the library, so every screen labels it the same.
Extra usage lines align under the first. `Header` draws a boxed header over a
command's own output, outside a help screen.

A section's underline takes its color from the heading's name first and its
position second. The headings most tools share keep one color everywhere,
because a color is learnable across tools only if it never shifts. `Commands`
is cyan, `Options` is magenta, and `Examples` is yellow. `fixedSections` in
`color.go` holds the full mapping. A heading the tool invents rotates through
green, blue and white by position, so neighbors differ. Rows under `Examples`
render in the command color, because an example is a command you would type.

## `Palette` — color is decided once per stream, never at a call site

```go
p := gotermstyle.PaletteFor(os.Stdout)     // decide once, against the stream
fmt.Println(p.Cyan(name), p.Command(path)) // paint with the decision
```

Re-deciding at each call site is how one screen ends up half colored. Resolve
the palette once and pass it around. Inside a help screen, `s.Palette()` is the
screen's own.

The gate runs in a fixed order. `NO_COLOR` turns color off. `FORCE_COLOR` turns
it on. `TERM=dumb` turns it off. Otherwise color is on only when the writer is
an `*os.File` on a character device. That check is `os.Stat` plus
`os.ModeCharDevice`, so it needs no `golang.org/x/term`.

The codes are terminal-palette codes rather than RGB. A 16-color terminal
renders them in its theme's colors, so nothing degrades a color profile. The
command color is the one non-bright code, so a command you would type reads
differently from a name in a listing. `Code` returns a code or the empty string
for a caller building a string by interpolation. `On` reports whether the
palette emits color.

There is no dim. Its contrast is whatever the terminal theme makes it, and on
many themes it is unreadable.

## `Clip` and alignment — widths are counted in runes, on the uncolored text

```go
rest := gotermstyle.Clip(path, utf8.RuneCountInString(prefix)) // measure plain text
fmt.Println(prefix + p.Command(rest))                          // color after clipping
```

`Clip(text, used)` shortens text to fit the terminal beside `used` other
columns. Pass the width of the uncolored surrounding text and color the result
afterwards. A clip then cannot land inside an escape sequence. With fewer than
two columns of room, `Clip` returns the text whole. An overflowing row says
more than a bare ellipsis.

`Columns` reads `$COLUMNS` first, then the terminal on stdout, then
`DefaultColumns`. That is the order Python's `shutil.get_terminal_size` uses.
The terminal query is a `TIOCGWINSZ` ioctl through the standard library's
`syscall`. It runs on Linux, macOS and the BSDs, Android and iOS included.
Everywhere else `$COLUMNS` or the default answers.

Pad a colored column by hand. `text/tabwriter` counts an escape sequence as
visible width, so one colored cell shifts every column after it. A padded
`fmt` verb has the same fault:

```go
fmt.Fprintf(w, "%-*s  %s\n", width, p.Cyan(name), address) // wrong: the width counts the escape

gap := strings.Repeat(" ", width-utf8.RuneCountInString(name))
fmt.Fprintf(w, "%s%s  %s\n", p.Cyan(name), gap, address)
```

Widths are runes, the unit Python's `len` counts on a `str`. A multibyte name
aligns the same here as in pytermstyle. Both assume a wide rune or a grapheme
cluster is one column, so CJK and emoji misalign. That is the one thing
lipgloss would buy, and this package accepts the gap deliberately.

## `cobrahelp` — one call renders every screen of a cobra tree

```go
import "github.com/datapointchris/gotermstyle/cobrahelp"

cobrahelp.Install(root)
```

Every screen then says what the command is, and lists every command a caller
can run beneath it. Each row is spelled from the screen's own command down to
the leaf, arguments included, so a command three levels down is on the root
rather than three screens away. The examples and flags follow. A section is a
group the command declares with `AddGroup`, and commands keep the order the
tree adds them, so `Install` turns off cobra's alphabetical sorting. A command
no group claims sits under "Other commands". Descriptions wrap at 80 columns,
and a command line wider than 40 sits alone with its description beneath.

It replaces cobra's help and usage templates outright rather than patching
their rows. Cobra's template gives a namespace a `Usage: <cmd> [command]`
block, which a first-time reader takes for the answer, and puts the examples
above the commands.

## Why this exists — cobra renders help with no notion of style

A Go tool on cobra prints its help monochrome. A Python tool on Typer and Rich
prints its help in color. So does a bash script that sources `formatting.sh`.

The gap follows the stack rather than the tool, so it wants one answer per
stack. Typer renders help through Rich, and color arrives with it. Cobra
renders help through `text/template`. A template emits plain strings and has no
notion of a style it could apply or strip, so nothing arrives for free.

## Why not a third-party library — none matches the neighbors at zero cost

Good libraries exist. lipgloss is the Go peer of Rich. Its tables, lists,
layers and ANSI-aware width do more than this package does.

The value bought here is narrower. It is a Go screen that looks the same as its
Python and bash neighbors on `PATH`, with no dependency. Three alternatives
were weighed against that:

- **`charmbracelet/fang`** owns command execution. `fang.Execute` wraps the
  root command's `ExecuteContext`, which a tool with a self-update wrapper
  already owns. It installs its own help function, so a help template or help
  function the tool set is bypassed. Its theme is Charm's house style, in
  lipgloss colors that switch between light and dark variants. Fixed ANSI codes
  are a different model. Its README calls it experimental.
- **`ivanpirog/coloredcobra`** is the closest drop-in and loses the least. It
  rewrites cobra's usage template with regular expressions, and a tool may
  already patch that template. It does nothing for grouping or row width. Two
  things mutating one template that cobra is free to change is worse than one
  thing owned outright.
- **lipgloss by hand** gives the most control for the most code. Its value here
  would be wide-rune width alone.

Typer is the counter-example that settled it. Its Rich renderer styles help
text, option defaults and panel borders dim. Dim therefore reaches every Typer
help screen, and no source file in the tool asks for it. A framework that
renders for you applies its opinions where no source grep will find them.

## Relationship to the other two — the same screen, not the same bytes

| Language | Library | Lives in |
| --- | --- | --- |
| Go | `gotermstyle` | this repo |
| Python | `pytermstyle` | [datapointchris/pytermstyle](https://github.com/datapointchris/pytermstyle) |
| bash | `formatting.sh` + `colors.sh` | `configs/common/.local/shell/` in [datapointchris/dotfiles](https://github.com/datapointchris/dotfiles) |

The three emit the same escape codes for the same roles. The Go tests assert
each code as a literal, and no test reads the other two libraries. A change to
the palette is a change to all three.

The bytes differ in two places. pytermstyle and `formatting.sh` open a help
row's color before its two-space indent, and Go opens it after. They also write
a reset on a continuation row, which opens no color, and Go writes none.

The API differs from the Python in one place. pytermstyle and `formatting.sh`
keep the buffered rows and the section counter in globals, so pytermstyle's
tests must reset them between screens. Go runs its tests in one process and its
programs concurrently, so that state lives on a `Screen` here.

## License

MIT.
