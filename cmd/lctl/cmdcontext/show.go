package cmdcontext

import (
	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config" // For config.CLIContext type
	"github.com/example/lctl/internal/output"
	"github.com/example/lctl/internal/shared"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current lctl CLI context (repository and reference)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// CLIContext is loaded by RootCmd's PersistentPreRunE
		goCtx := cmd.Context()
		loadedCLIContextVal := goCtx.Value(shared.CLIContextContextKey)
		if loadedCLIContextVal == nil {
			return config.ErrCLIContextNotLoaded // Using error from config package
		}
		loadedCLIContext := loadedCLIContextVal.(*config.CLIContext)

		formatterVal := goCtx.Value(shared.FormatterContextKey)
		if formatterVal == nil {
			return output.ErrFormatterNotLoaded // Using error from output package
		}
		formatter := formatterVal.(output.Formatter)

		return formatter.Write(loadedCLIContext)
	},
}
// func init() {
//     ContextCmd.AddCommand(showCmd)
// }
