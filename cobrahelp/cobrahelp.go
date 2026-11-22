// Package cobrahelp renders every help screen of a cobra command tree: what
// the command is, then every leaf beneath it, spelled from the screen's own
// command down to the leaf with its arguments, then its examples and flags. A
// command with children gets no row, even one that runs. Under NamespacesOnRoot
// the root screen gives each of its own commands one row instead, so a tree
// with hundreds of leaves does not list them all there.
//
// Cobra's own screen names only the next word down, so a command three levels
// down is three screens away, each ending in `[command]`. Listing the leaves
// of the whole subtree puts every leaf on the root and each namespace's share
// of them on that namespace. A row starts at the word after the screen's own
// command, as cobra's rows do, so a row's first word is always one of that
// command's children.
//
// The renderer replaces cobra's help and usage templates rather than swapping
// their command rows. Cobra's template gives every screen a Usage block, and
// puts the examples above the commands. On a namespace the block reads
// `<cmd> [command]`, and on a leaf it repeats what was typed. This screen has
// no Usage block. Its title line is the command with its arguments, and the
// commands come first.
//
// Everything a screen shows comes from the tree. A section is a group the
// command declares with AddGroup, in the order it adds them, and a command
// sits in the order the tree adds it. Commands no group claims follow under
// "Other commands".
package cobrahelp

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Install renders every screen in root's tree, help and usage alike.
//
// Every list prints in the order the tree adds it, which is the order a
// command is reached for, so Install turns off cobra's alphabetical sorting.
// That setting is cobra's and process-wide: every command tree in the process
// loses alphabetical sorting, not only root's.
func Install(root *cobra.Command, options ...Option) {
	var s settings
	for _, option := range options {
		option(&s)
	}
	cobra.EnableCommandSorting = false
	root.SetHelpFunc(func(cmd *cobra.Command, _ []string) { write(cmd.OutOrStdout(), cmd, s) })
	root.SetUsageFunc(func(cmd *cobra.Command) error {
		write(cmd.OutOrStderr(), cmd, s)
		return nil
	})
}

// Option changes how Install draws a tree's screens.
type Option func(*settings)

type settings struct{ namespacesOnRoot bool }

// NamespacesOnRoot draws the root screen with one row per command beneath it,
// a namespace as `<name> <command>` beside its own description, rather than
// every leaf in the tree. Every other screen still lists every leaf beneath
// it. It is for a tree whose leaves are too many to read on one screen.
func NamespacesOnRoot() Option {
	return func(s *settings) { s.namespacesOnRoot = true }
}

func write(out io.Writer, cmd *cobra.Command, s settings) {
	var b strings.Builder
	// UseLine appends " [flags]" whenever the command has a flag, --help
	// included. The flags have a section of their own, so the title drops it.
	fmt.Fprintf(&b, "%s — %s\n", strings.ReplaceAll(cmd.UseLine(), " [flags]", ""), cmd.Short)
	if long := strings.TrimSpace(cmd.Long); long != "" {
		fmt.Fprintf(&b, "\n%s\n", long)
	}
	rows := func(cmds []*cobra.Command) []row { return leafRows(cmd, cmds) }
	if s.namespacesOnRoot && !cmd.HasParent() {
		rows = namespaceRows
	}
	sections := commandSections(cmd, rows)
	flags := flagUsages(cmd)
	// Each section sizes its own column from the command lines that fit in
	// maxCommandColumn, so one long line does not push every description past
	// helpWidth. A longer line sits alone, with its description beneath.
	for _, s := range sections {
		width := 0
		for _, r := range s.rows {
			if n := utf8.RuneCountInString(r.line); n <= maxCommandColumn {
				width = max(width, n)
			}
		}
		indent := strings.Repeat(" ", width+4)
		fmt.Fprintf(&b, "\n%s:\n", s.title)
		for _, r := range s.rows {
			if utf8.RuneCountInString(r.line) > width {
				fmt.Fprintf(&b, "  %s\n%s", r.line, indent)
			} else {
				fmt.Fprintf(&b, "  %-*s  ", width, r.line)
			}
			fmt.Fprintf(&b, "%s\n", wrap(r.short, helpWidth-len(indent), indent))
		}
	}
	if cmd.HasExample() {
		fmt.Fprintf(&b, "\nExamples:\n%s\n", strings.TrimRight(cmd.Example, "\n"))
	}
	if flags != "" {
		fmt.Fprintf(&b, "\nFlags:\n%s\n", flags)
	}
	if len(sections) > 0 {
		fmt.Fprintf(&b, "\n`%s <command> --help` shows a command's flags and examples.\n", cmd.CommandPath())
	}
	_, _ = io.WriteString(out, b.String())
}

// row is one command a caller can run, and what it does.
type row struct{ line, short string }

type section struct {
	title string
	rows  []row
}

// commandSections is every command beneath cmd, under cmd's own groups where
// it has them. A namespace's commands sit under the group the namespace is in.
// A group's title is its own, less a trailing colon; the two this supplies,
// "Commands" and "Other commands", follow cobra's own heading.
func commandSections(cmd *cobra.Command, rows func([]*cobra.Command) []row) []section {
	children := typeable(cmd)
	if len(children) == 0 {
		return nil
	}
	groups := cmd.Groups()
	if len(groups) == 0 {
		return []section{{title: "Commands", rows: rows(children)}}
	}
	var sections []section
	for _, group := range groups {
		var in []*cobra.Command
		for _, child := range children {
			if child.GroupID == group.ID {
				in = append(in, child)
			}
		}
		if len(in) > 0 {
			sections = append(sections, section{title: strings.TrimSuffix(group.Title, ":"), rows: rows(in)})
		}
	}
	var rest []*cobra.Command
	for _, child := range children {
		if child.GroupID == "" {
			rest = append(rest, child)
		}
	}
	if len(rest) > 0 {
		sections = append(sections, section{title: "Other commands", rows: rows(rest)})
	}
	return sections
}

// typeable is cmd's children a caller can run. help is left out, since
// --help on any command is the same thing, and so is a hidden command.
func typeable(cmd *cobra.Command) []*cobra.Command {
	var children []*cobra.Command
	for _, child := range cmd.Commands() {
		if child.IsAvailableCommand() && child.Name() != "help" {
			children = append(children, child)
		}
	}
	return children
}

// leafRows is every leaf at or beneath cmds, each spelled from the word after
// screen's command down to the leaf, arguments included. A command with
// children gets no row of its own, even when it runs.
func leafRows(screen *cobra.Command, cmds []*cobra.Command) []row {
	var rows []row
	for _, cmd := range cmds {
		if children := typeable(cmd); len(children) > 0 {
			rows = append(rows, leafRows(screen, children)...)
			continue
		}
		line := strings.TrimPrefix(cmd.Parent().CommandPath()+" "+cmd.Use, screen.CommandPath()+" ")
		rows = append(rows, row{line: line, short: cmd.Short})
	}
	return rows
}

// namespaceRows is one row per command in cmds: a leaf with its arguments,
// and a namespace as `<name> <command>` beside its own description.
func namespaceRows(cmds []*cobra.Command) []row {
	var rows []row
	for _, cmd := range cmds {
		line := cmd.Use
		if len(typeable(cmd)) > 0 {
			line = cmd.Name() + " <command>"
		}
		rows = append(rows, row{line: line, short: cmd.Short})
	}
	return rows
}

// helpWidth is the column a command's or a flag's description wraps at, so a
// screen reads in a split pane.
const helpWidth = 80

// maxCommandColumn is the widest a section's command column grows, which
// leaves a description at least 36 columns beside it.
const maxCommandColumn = 40

// wrap breaks text into lines of at most width runes, each after the first
// starting with indent. A word longer than width keeps a line to itself.
func wrap(text string, width int, indent string) string {
	var b strings.Builder
	used := 0
	for i, word := range strings.Fields(text) {
		n := utf8.RuneCountInString(word)
		switch {
		case i == 0:
		case used+1+n > width:
			b.WriteString("\n" + indent)
			used = 0
		default:
			b.WriteByte(' ')
			used++
		}
		b.WriteString(word)
		used += n
	}
	return b.String()
}

// flagUsages is every flag cmd takes, its own and those it inherits, without
// --help, which every command has.
func flagUsages(cmd *cobra.Command) string {
	set := pflag.NewFlagSet(cmd.Name(), pflag.ContinueOnError)
	add := func(flag *pflag.Flag) {
		if flag.Name != "help" && !flag.Hidden && set.Lookup(flag.Name) == nil {
			set.AddFlag(flag)
		}
	}
	cmd.LocalFlags().VisitAll(add)
	cmd.InheritedFlags().VisitAll(add)
	return strings.TrimRight(set.FlagUsagesWrapped(helpWidth), "\n")
}
