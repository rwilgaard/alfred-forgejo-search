package cmd

import (
	"github.com/spf13/cobra"
)

// recordCmd counts a use for ranking when ⏎ opens a repository in the
// browser, a branch that runs no other Go code.
var recordCmd = &cobra.Command{
	Use:    "record <repo>",
	Short:  "record a repository use for ranking",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(_ *cobra.Command, args []string) error {
		return wf.Usage.Record(args[0])
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)
}
