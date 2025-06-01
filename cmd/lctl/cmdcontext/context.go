package cmdcontext

import "github.com/spf13/cobra"

// ContextCmd represents the context command
var ContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage CLI context",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	ContextCmd.AddCommand(showCmd)
	ContextCmd.AddCommand(setRepoCmd)
	ContextCmd.AddCommand(setRefCmd)
	ContextCmd.AddCommand(clearCmd)
}
