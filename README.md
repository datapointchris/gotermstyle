# gotermstyle

The house terminal style for the Go command-line tools — palette, section
header, help grammar, and the rules about when any of it is emitted at all.

Zero third-party dependencies.

> **Status: scaffolded, not implemented.** The design and the reasoning behind
> it live in `.planning/design.md` (local only, not in git).

## Why this exists

Every Go CLI on the fleet renders monochrome help, and every Python and bash one
does not. Counting ANSI escapes in each tool's root help screen, through a pty:

| Stack | Tools | Escapes |
| --- | --- | --- |
| Go / cobra | icb, forge, todoui, toolbox, ifiles | **0 each** |
| Python / Typer + Rich | indy, relate, syncer, dectl | 26–45 |
| bash + appcore | theme, font, safekeep, menu, packages | 9–36 |

The gap is stack-shaped rather than tool-shaped, so it wants one answer rather
than five. Cobra renders help through `text/template`, which emits strings and
has no notion of a style it could apply or strip, so nothing arrives for free
the way `rich_help_panel` does in Typer.

## Why not a third-party library

The value being bought is that a Go screen matches its Python and bash
neighbours on `PATH`. Every styling library brings its own aesthetic instead,
which is the one thing that cannot be adopted here. The alternatives considered
and the specific reason each was rejected are in the planning doc.

The counter-example that settled it: Typer's Rich renderer emits 23–38 dim
sequences into every help screen on this fleet, none of them in any source file,
while `standards/cli-design.md` bans dim outright for being unreadable on
half the terminal themes in use. Adopting a library means adopting its opinions,
including the ones that contradict your own, invisibly to any source grep.

## Relationship to the other two

| Language | Library | Lives in |
| --- | --- | --- |
| Go | `gotermstyle` | this repo |
| Python | `appcore.formatting` | `~/dotfiles/appcore/` |
| bash | `formatting.sh` + `colors.sh` | `~/dotfiles/configs/common/.local/shell/` |

The three emit the same escape codes for the same roles. A change to the palette
is a change to all three.

## Constraints

- **No third-party dependencies**, matching `goselfupdate`. Terminal detection
  is `os.Stat` plus `os.ModeCharDevice`, not `golang.org/x/term`, so importing
  this costs a consumer nothing and moves no Go floor.
- **No dim.** `standards/cli-design.md` § "Never use `[dim]`".
- **Colour is emitted only where something will render it.** `NO_COLOR` is the
  user's preference and wins; `FORCE_COLOR` overrides the terminal detection.

## Licence

MIT.
