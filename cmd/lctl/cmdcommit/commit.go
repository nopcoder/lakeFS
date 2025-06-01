package cmdcommit

import "github.com/spf13/cobra"

// CommitCmd represents the commit command
var CommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a new commit in an lctl repository",
	Long:  `Records changes to the repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Define flags, e.g., for commit message
	CommitCmd.Flags().StringP("message", "m", "", "Commit message")
}
