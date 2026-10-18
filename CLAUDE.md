# gotermstyle — Claude Code instructions

A Go library, not a CLI. No `main` package, nothing to install; it ships as git
tags, and Go command-line tools import it.

Read the README's "Why this exists" and "Why not a third-party library" before
writing any code here. The palette, the rejected alternatives, and the reason
each was rejected are settled there — this repo exists because of a decision,
not as a blank slate.

## Constraints that must not regress

- **Zero third-party dependencies — this repo's own choice.** The general rule
  for a Go library is *containment*: a CLI frontend lives in its own package so
  a consumer importing only the core inherits nothing. There is nothing to
  contain here — the repo is flat, has no subpackages, and `go.mod` declares no
  requirements at all, so containment is satisfied by having nothing to split
  off.

  The choice stands on its own anyway. This package exists so a Go screen
  looks the same as its Python and bash neighbors, and everything it needs
  to do that is a string. The bytes differ, and the README's "Relationship to
  the other two" says where. Terminal detection is `os.Stat` plus
  `os.ModeCharDevice` rather than `golang.org/x/term`, and `Columns` reaches
  `TIOCGWINSZ` through stdlib `syscall`, so importing this downloads no other
  module.

  What it gives up is wide-rune and grapheme-cluster width, which `lipgloss`
  handles and this does not. That is a deliberate acceptance, not an oversight.
  Reach for `lipgloss` in a repo that needs real layout rather than pretending
  this competes.
- **The `go.mod` floor is a statement about who may import this, not the lowest
  version the code needs.** Raising it excludes callers, because `go get` raises
  an importer's `go` line to match. So a standard-library advisory moves the
  `toolchain` line, never the floor. The floor also has a bottom: CI sets up the
  Go the `toolchain` line names, or the floor where there is none, under
  `GOTOOLCHAIN=local`, and golangci-lint refuses to run on a Go below its own
  minimum. A floor below that bottom with no `toolchain` line above it fails
  Lint while the code is fine.
- **The palette matches the Python and bash libraries.** The exact codes are
  pytermstyle's constants and the bash library's `colors.sh`; the README's
  "Relationship to the other two" says where each lives. Changing one without
  the other two is the failure this library exists to prevent — the point is
  that three screens on one `PATH` look like one product.
- **No dim, ever.** It is unreadable on many terminal themes. Use normal text,
  bold, or a color.
- **Color resolves once, and only where it will be rendered.** `NO_COLOR`
  outranks `FORCE_COLOR`: the first is the user's preference, the second only
  overrides the terminal detection.
- **Widths are measured on uncolored text.** An f-string-style padded field
  counts escape bytes and shoves one row out of alignment; the Python and bash
  versions both solve it by measuring plain and coloring after, and tests in
  both hold it. Port the tests too.

## Rules the hooks enforce

gofumpt formats and golangci-lint v2 lints, both as commit hooks and in CI;
`.pre-commit-config.yaml` and `.golangci.yml` configure them. The help grammar
this library renders is the one its Python and bash siblings render, and it
changes in all three or in none.

## Releases

- **No goreleaser.** A library has no binary to build, so go-semantic-release
  alone cuts the release.

`release.yml` cuts the tags. A tag is exactly what a consumer resolves, so
something has to cut it.

The tag is cut from `main` by `go-semantic-release`, after a build, vet and
`-race` test on ubuntu, macOS and windows. Windows is in that matrix because it
is the one platform this package cannot be exercised on locally: `Columns`
reaches `TIOCGWINSZ` through stdlib `syscall`, which windows does not have, so
the fallback sits behind a build constraint that nothing else proves.

`allow-initial-development-versions` holds the line at 0.x. Without it,
go-semantic-release majors on every change while the major is 0. For a library
a 1.0.0 is a compatibility promise, so staying on 0.x is deliberate here.

## Never write the breaking-change trailer in a commit message

The words `BREAKING CHANGE` — either number, colon or not, subject or body — cut a major release
once this repo is past 0.x, and a major from v1 on is an outage rather than a version. On 0.x,
`allow-initial-development-versions` cuts them as a minor. `commit-analyzer-cz` matches
them unanchored against the raw message and ORs the result with the configured major rules, so
`.semrelrc` cannot stop it and it majors even a `fix:` commit.

The module path carries no `/vN` suffix, so once a major exists `go get …@latest` cannot see it
and silently resolves the highest v1 instead. No command fails, so nothing tells a consumer it is
stranded on the old line. Recovery is not a rewrite — branch protection refuses one on `main`, and
the offending commit re-cuts the major on every push until a tag above it takes it out of range.

**The ban covers a commit that merely discusses the trailer.** One explaining this exact caveat cut
a fresh major on push. Name it some other way — "that marker" — and never quote it.

The release workflow reads this repo's own `.semrelrc`, and there is none, so no commit subject
maps to a major. Reaching 1.0 is dropping `allow-initial-development-versions`, and v1 needs no
path suffix. A deliberate v2 or later needs a `.semrelrc` mapping `chore(release-major)` to a
major, and the same release adds `/v2` to the module path.
