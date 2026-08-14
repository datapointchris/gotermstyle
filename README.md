# gotermstyle

The house terminal style for the Go command-line tools — palette, section
header, help grammar, and the rules about when any of it is emitted at all.

Zero third-party dependencies, by this repo's own choice.

## Using it

```go
import "github.com/datapointchris/gotermstyle"
```

A `Screen` renders a help screen to one writer, with color resolved once
against that writer:

```go
s := gotermstyle.New(cmd.OutOrStdout())

s.HelpHeader("menu labs", "Small experiments worth revisiting.")
s.HelpUsage("menu labs <verb>")

s.HelpSection("Commands")
s.HelpRow("menu labs list", "", "List every lab")
s.HelpRow("menu labs show", "<id>", "Show one lab")

s.HelpEnd()
```

`HelpRow` buffers rather than writing. The flush sizes the description column
from the longest row in the section, so no call site passes a width or a color.
Use `HelpText` for prose between rows, so pending rows flush ahead of it rather
than landing under it.

For output that is not a help screen:

```go
p := gotermstyle.PaletteFor(os.Stdout)
fmt.Println(p.Cyan(name), p.Command(path))
```

`Clip(text, used)` shortens text to fit alongside `used` other columns. Pass the
width of the *uncolored* surrounding text and color the result afterwards, so a
clip can never land inside an escape sequence.

### Color and alignment together

Widths are measured on the uncolored string and the color applied after. Pad by
hand for the same reason: `text/tabwriter` counts an escape sequence as visible
width, so one colored cell shifts every column after it.

```go
fmt.Fprintf(w, "%s  %s\n", p.Cyan(pad(name, width)), address)  // wrong
fmt.Fprintf(w, "%s  %s\n", p.Cyan(name)+spaces(width-len(name)), address)
```

## Why this exists

Every Go CLI on the fleet renders monochrome help, and every Python and bash one
does not. Counting ANSI escapes in each tool's root help screen, through a pty:

| Stack | Tools | Escapes |
| --- | --- | --- |
| Go / cobra | icb, forge, todoui, toolbox, ifiles | **0 each** |
| Python / Typer + Rich | indy, relate, syncer, dectl | 26–45 |
| bash | theme, font, safekeep, menu, packages | 9–36 |

The gap is stack-shaped rather than tool-shaped, so it wants one answer rather
than five. Cobra renders help through `text/template`, which emits strings and
has no notion of a style it could apply or strip, so nothing arrives for free
the way `rich_help_panel` does in Typer.

## Why not a third-party library

Not because none is good enough. `lipgloss` is Rich-class and is a registered
exemplar here — it has tables, lists, layers and genuinely ANSI-aware width,
which is more than this package does.

The value being bought is narrower: a Go screen matching its Python and bash
neighbors on `PATH`, byte for byte, at zero dependency cost. The alternatives
considered and the reason each was rejected are in the planning doc.

The counter-example that settled it: Typer's Rich renderer emits 23–38 dim
sequences into every help screen on this fleet, none of them in any source file,
while `standards/cli-design.md` bans dim outright for being unreadable on half
the terminal themes in use. A framework that renders *for* you applies its
opinions where no source grep will find them.

## Relationship to the other two

| Language | Library | Lives in |
| --- | --- | --- |
| Go | `gotermstyle` | this repo |
| Python | `pytermstyle` | `~/tools/pytermstyle` |
| bash | `formatting.sh` + `colors.sh` | `~/dotfiles/configs/common/.local/shell/` |

The three emit the same escape codes for the same roles, and that is verified
against rendered bytes rather than against the source constants. A change to the
palette is a change to all three.

The API differs from the Python in one place. pytermstyle and `formatting.sh`
keep the buffered rows and the section counter in globals, which costs
pytermstyle a reset helper its tests must call. Go runs its tests in one process
and its programs concurrently, so that state lives on a `Screen` here. The
rendered output is identical.

## Constraints

- **No third-party dependencies.** This is the repo's choice rather than a
  fleet rule — `standards/repo-structure.md` retired the blanket ban, and what
  binds now is containment: a CLI frontend lives in its own package so a
  consumer importing only the core inherits nothing. The choice stands on its
  own here because everything this does is a string. Terminal detection is
  `os.Stat` plus `os.ModeCharDevice` rather than `golang.org/x/term`, and
  `Columns` reaches `TIOCGWINSZ` through stdlib `syscall`, so importing this
  downloads nothing and moves no Go floor.
- **No dim.** `standards/cli-design.md` § "Never use `[dim]`".
- **Color is emitted only where something will render it.** `NO_COLOR` is the
  user's preference and wins; `FORCE_COLOR` overrides the terminal detection.
- **Widths are measured in runes**, which is where this improves on the Python
  it ports. Wide runes and grapheme clusters are still assumed one column — the
  one thing `lipgloss` would buy, and a deliberate acceptance.

## Licence

MIT.
