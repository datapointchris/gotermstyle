package cobrahelp

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func tree() *cobra.Command {
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
	Install(root)
	return root
}

// A screen is pinned whole, because its layout is the contract every tool
// that installs it shares.
func TestEveryScreenListsLeavesUnderTheirGroupsAndWrapsAtEighty(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"--help"}, `lab — Small experiments

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
      --json   Emit JSON

` + "`lab <command> --help`" + ` shows a command's flags and examples.
`},
		{[]string{"trips", "list", "--help"}, `lab trips list — List every trip

Usage:
  lab trips list [flags]

Flags:
      --json   Emit JSON
`},
	}
	for _, c := range cases {
		root := tree()
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
