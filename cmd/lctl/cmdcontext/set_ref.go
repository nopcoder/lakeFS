package cmdcontext

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
	"github.com/example/lctl/internal/output"
	"github.com/example/lctl/internal/shared"
)

var setRefCmd = &cobra.Command{
	Use:   "set-ref <ref_name>",
	Short: "Set the current reference (branch/tag/commit) in CLI context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		refName := args[0]

		cliCtx, err := config.LoadCLIContext()
		if err != nil {
			return fmt.Errorf("failed to load CLI context: %w", err)
		}

		cliCtx.CurrentRef = refName

		err = config.SaveCLIContext(cliCtx)
		if err != nil {
			return fmt.Errorf("failed to save CLI context: %w", err)
		}

		fmt.Printf("Reference context set to: %s\n", cliCtx.CurrentRef)

		formatterVal := cmd.Context().Value(shared.FormatterContextKey)
		if formatterVal == nil {
			return output.ErrFormatterNotLoaded
		}
		formatter := formatterVal.(output.Formatter)
		return formatter.Write(cliCtx)
	},
}
