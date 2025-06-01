package cmdfs

import "github.com/spf13/cobra"

// FSCmd represents the fs command
var FSCmd = &cobra.Command{
	Use:   "fs",
	Short: "Interact with the lctl file system",
	Long:  `Commands for listing files (ls), uploading (put), downloading (get), and other file system operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Define flags and subcommands like ls, get, put
}
