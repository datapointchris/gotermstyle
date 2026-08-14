# gotermstyle — Claude Code instructions

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
  `goselfupdate`. See `standards/go.md` § "Go version floor".
- **The palette matches `~/dotfiles`.** The exact codes are in
  `appcore/formatting.py` and `configs/common/.local/shell/colors.sh`. Changing
  one without the other two is the failure this library exists to prevent — the
  point is that three screens on one `PATH` look like one product.
- **No dim, ever.** `standards/cli-design.md` § "Never use `[dim]`".
- **Colour resolves once, and only where it will be rendered.** `NO_COLOR`
  outranks `FORCE_COLOR`: the first is the user's preference, the second only
  overrides the terminal detection.
- **Widths are measured on uncoloured text.** An f-string-style padded field
  counts escape bytes and shoves one row out of alignment; the Python and bash
  versions both solve it by measuring plain and colouring after, and there are
  tests in `~/dotfiles` holding it. Port the tests too.

## Rules that live elsewhere

`standards/go.md` for layout, gofumpt, golangci-lint v2, and the module
rules. `standards/cli-design.md` for what a help screen must contain —
this library renders that grammar and does not redefine it.

## Sanctioned exceptions

- **No goreleaser.** There is no binary to build, upload or install. A consumer
  resolves this by module path and tag.

`release.yml` is not an exception and the repo has one. A tag is exactly what a
consumer resolves, so something has to cut it — `goselfupdate` and
`pytermstyle` both cut theirs the same way, and gotermstyle went unreleasable
for a week while this file said the opposite.

The tag is cut from `main` by `go-semantic-release`, after a build, vet and
`-race` test on ubuntu, macOS and windows. Windows is in that matrix because it
is the one platform this package cannot be exercised on locally: `Columns`
reaches `TIOCGWINSZ` through stdlib `syscall`, which windows does not have, so
the fallback sits behind a build tag that nothing else proves.

`allow-initial-development-versions` holds the line at 0.x. Dropping it is how
`bbkt`, `sess` and `bashselfupdate` each shipped 1.0.0 as their first automated
release; for a library that is a compatibility promise, so it stays deliberate.

## Never write the breaking-change trailer in a commit message

The words `BREAKING CHANGE` — either number, colon or not, subject or body — cut a major release
here, and a major on this repo is an outage rather than a version. `commit-analyzer-cz` matches
them unanchored against the raw message and ORs the result with the configured major rules, so
`.semrelrc` cannot stop it and it majors even a `fix:` commit.

The module path carries no `/vN` suffix, so once a major exists `go install …@latest` cannot see it
and silently resolves the highest v1 instead — `dotfiles check` reports the tool stale forever
while `apply` exits 0 having installed nothing. Every already-installed binary is stranded too:
`goselfupdate` refuses a lower version and reports "already up to date". Recovery is a reinstall on
each machine, and it is not a rewrite — branch protection refuses one on `main`, and the offending
commit re-cuts the major on every push until a tag above it takes it out of range.

**The ban covers a commit that merely discusses the trailer.** One explaining this exact caveat cut
a fresh major on push. Name it some other way — "that marker" — and never quote it.

Deliberate majors use `chore(release-major)`, the one subject `.semrelrc` leaves as a major. Full
reasoning and the reset procedure: `standards/release.md` § "Never write the breaking-change
trailer in a Go repo's commit message".
