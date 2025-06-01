package cmdconfig

import "github.com/spf13/cobra"

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage lctl configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	ConfigCmd.AddCommand(listCmd)
	ConfigCmd.AddCommand(getCmd)
	ConfigCmd.AddCommand(setCmd)
	ConfigCmd.AddCommand(initCmd)
	ConfigCmd.AddCommand(loginCmd)
	ConfigCmd.AddCommand(showPathCmd)
}
