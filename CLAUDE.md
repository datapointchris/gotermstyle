# gocore — Claude Code instructions

A Go library, not a CLI. No `main` package, nothing to install; it ships as git
tags and is consumed by the Go tools in `~/tools/` and the `cli/` modules inside
the webapps.

Read `.planning/design.md` before writing any code here. The palette, the
rejected alternatives, and the reason each was rejected are settled there — this
repo exists because of a decision, not as a blank slate.

## Constraints that must not regress

- **Zero third-party dependencies.** The same constraint `goselfupdate` holds
  and for the same reason: a consumer importing this must download nothing and
  must not inherit a Go floor. Terminal detection is `os.Stat` plus
  `os.ModeCharDevice` (the idiom `goselfupdate/autoupdate` already uses), never
  `golang.org/x/term`.
- **The `go.mod` floor matches the fleet, not the lowest version the code
  needs.** A library would ideally floor lower, but the generated CI reads the
  floor from `go.mod`, pins `GOTOOLCHAIN=local`, and then installs golangci-lint,
  which has its own minimum — so a lower floor fails Lint on a repo whose code is
  fine. This was scaffolded at 1.23 and failed exactly that way before matching
  `goselfupdate`. See `~/dev/standards/go.md` § "Go version floor".
- **The palette matches `~/dotfiles`.** The exact codes are in
  `appcore/formatting.py` and `configs/common/.local/shell/colors.sh`. Changing
  one without the other two is the failure this library exists to prevent — the
  point is that three screens on one `PATH` look like one product.
- **No dim, ever.** `~/dev/standards/cli-design.md` § "Never use `[dim]`".
- **Colour resolves once, and only where it will be rendered.** `NO_COLOR`
  outranks `FORCE_COLOR`: the first is the user's preference, the second only
  overrides the terminal detection.
- **Widths are measured on uncoloured text.** An f-string-style padded field
  counts escape bytes and shoves one row out of alignment; the Python and bash
  versions both solve it by measuring plain and colouring after, and there are
  tests in `~/dotfiles` holding it. Port the tests too.

## Rules that live elsewhere

`~/dev/standards/go.md` for layout, gofumpt, golangci-lint v2, and the module
rules. `~/dev/standards/cli-design.md` for what a help screen must contain —
this library renders that grammar and does not redefine it.

## Sanctioned exceptions

- **No goreleaser and no `release.yml`.** A library with no binary has nothing
  to release; it is consumed by tag, like `goselfupdate`.
