package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root cobra command for the linear CLI.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "linear",
		Short:         "Linear CLI - manage your Linear workspace from the terminal",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		// show help when no subcommand is given
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.PersistentFlags().Bool("json", false, "output as JSON")
	root.AddCommand(newAuthCommand())
	root.AddCommand(newIssueCommand())
	root.AddCommand(newMeCommand())
	root.AddCommand(newTeamCommand())
	root.AddCommand(newUserCommand())

	return root
}
