package cmdconfig

import (
	"github.com/spf13/cobra"
	// Assuming shared.CLIConfigContextKey and shared.FormatterContextKey are defined and used in root's PersistentPreRunE
	"github.com/example/lctl/internal/config" // For config.Configuration type
	"github.com/example/lctl/internal/output"
	"github.com/example/lctl/internal/shared"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all current lctl configuration settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Configuration is loaded by RootCmd's PersistentPreRunE and put into context
		goCtx := cmd.Context()
		loadedCfgVal := goCtx.Value(shared.CLIConfigContextKey)
		if loadedCfgVal == nil {
			return config.ErrConfigNotLoaded // Or a more specific error
		}
		loadedCfg := loadedCfgVal.(*config.Configuration)

		formatterVal := goCtx.Value(shared.FormatterContextKey)
		if formatterVal == nil {
			return output.ErrFormatterNotLoaded // Or a more specific error
		}
		formatter := formatterVal.(output.Formatter)

		return formatter.Write(loadedCfg)
	},
}

// init() function for listCmd to add itself to ConfigCmd is handled in config.go
// func init() {
//     ConfigCmd.AddCommand(listCmd)
// }
