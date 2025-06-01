package cmdrepl

import "github.com/spf13/cobra"

// REPLCmd represents the repl command
var REPLCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive REPL session for lctl",
	Long:  `The Read-Eval-Print Loop (REPL) allows for interactive command execution.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Define flags or specific REPL configurations
}
