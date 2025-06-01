package cmdrepo

import "github.com/spf13/cobra"

// RepoCmd represents the repo command
var RepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage lctl repositories",
	Long:  `Commands for creating, listing, inspecting, and deleting lctl repositories.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Define flags and subcommands here
	// e.g., repo create, repo list, repo delete
}
