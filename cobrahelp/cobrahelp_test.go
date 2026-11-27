package cobrahelp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func tree(options ...Option) *cobra.Command {
	run := func(*cobra.Command, []string) {}
	root := &cobra.Command{Use: "lab", Short: "Small experiments", Long: "Every experiment is a trip with stops."}
	root.AddGroup(&cobra.Group{ID: "plan", Title: "Planning commands:"})
	trips := &cobra.Command{Use: "trips", Short: "Trips", GroupID: "plan"}
	trips.AddCommand(
		&cobra.Command{Use: "list", Short: "List every trip", Run: run},
		&cobra.Command{
			Use:   "show <id>",
			Short: "Show one trip with every stop it makes, the days between them, and what each stop costs",
			Run:   run,
		},
		&cobra.Command{Use: "secret", Short: "Hidden", Hidden: true, Run: run},
	)
	root.AddCommand(
		trips,
		&cobra.Command{Use: "import <file> [--replace-every-stop-already-recorded]", Short: "Import trips", GroupID: "plan", Run: run},
		&cobra.Command{Use: "version", Short: "Print the version", Run: run},
	)
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().Bool("json", false, "Emit JSON")
	Install(root, options...)
	return root
}

// A screen is pinned whole, because its layout is the contract every tool
// that installs it shares. The flag lines are the exception. pflag draws
// them, so each screen expects whatever pflag draws for the one flag the tree
// declares, and a change to pflag's layout leaves this test passing.
func TestEveryScreenListsLeavesUnderTheirGroupsAndWrapsAtEighty(t *testing.T) {
	declared := pflag.NewFlagSet("lab", pflag.ContinueOnError)
	declared.Bool("json", false, "Emit JSON")
	flags := strings.TrimRight(declared.FlagUsagesWrapped(80), "\n")
	cases := []struct {
		args    []string
		options []Option
		want    string
	}{
		{[]string{"--help"}, nil, `lab — Small experiments

Every experiment is a trip with stops.

Planning commands:
  trips list       List every trip
  trips show <id>  Show one trip with every stop it makes, the days between
                   them, and what each stop costs
  import <file> [--replace-every-stop-already-recorded]
                   Import trips

Other commands:
  version  Print the version

Flags:
` + flags + `

` + "`lab <command> --help`" + ` shows a command's flags and examples.
`},
		{[]string{"import", "--help"}, nil, `lab import <file> [--replace-every-stop-already-recorded] — Import trips

Flags:
` + flags + `
`},
		{[]string{"--help"}, []Option{NamespacesOnRoot()}, `lab — Small experiments

Every experiment is a trip with stops.

Planning commands:
  trips <command>  Trips
  import <file> [--replace-every-stop-already-recorded]
                   Import trips

Other commands:
  version  Print the version

Flags:
` + flags + `

` + "`lab <command> --help`" + ` shows a command's flags and examples.
`},
		{[]string{"trips", "--help"}, []Option{NamespacesOnRoot()}, `lab trips — Trips

Commands:
  list       List every trip
  show <id>  Show one trip with every stop it makes, the days between them, and
             what each stop costs

Flags:
` + flags + `

` + "`lab trips <command> --help`" + ` shows a command's flags and examples.
`},
	}
	for _, c := range cases {
		root := tree(c.options...)
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs(c.args)
		if err := root.Execute(); err != nil {
			t.Fatalf("%v: %v", c.args, err)
		}
		if got := out.String(); got != c.want {
			t.Errorf("%v rendered\n%s\nwant\n%s", c.args, got, c.want)
		}
	}
}
