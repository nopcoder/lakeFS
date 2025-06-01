package cmdconfig

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
)

var showPathCmd = &cobra.Command{
	Use:   "show-path",
	Short: "Show the path to the lctl configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFilePathFlag, err := cmd.Flags().GetString("config")
		if err != nil { // Should not happen if flag is defined on root
			cfgFilePathFlag = ""
		}

		path, err := config.GetConfigFilePath(cfgFilePathFlag)
		if err != nil {
			return fmt.Errorf("could not determine config file path: %w", err)
		}
		fmt.Println(path)
		return nil
	},
}
