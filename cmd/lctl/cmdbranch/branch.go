package cmdbranch

import "github.com/spf13/cobra"

// BranchCmd represents the branch command
var BranchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Manage branches in an lctl repository",
	Long:  `Commands for creating, listing, deleting, and switching branches within a repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Define flags and subcommands here
}
