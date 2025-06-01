package cmdcontext

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
	"github.com/example/lctl/internal/output"
	"github.com/example/lctl/internal/shared"
)

var (
	clearRepo bool
	clearRef  bool
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear parts of the current lctl CLI context (repository, reference, or both)",
	Long: `Clears repository URI, reference, or both from the CLI context.
If no flags are specified, both repository and reference are cleared.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cliCtx, err := config.LoadCLIContext()
		if err != nil {
			return fmt.Errorf("failed to load CLI context: %w", err)
		}

		clearedSomething := false
		// If no flags are set, clear both. Otherwise, only clear specified.
		clearAll := !clearRepo && !clearRef

		if clearRepo || clearAll {
			if cliCtx.CurrentRepoURI != "" {
				cliCtx.CurrentRepoURI = ""
				clearedSomething = true
				fmt.Println("Repository context cleared.")
			}
		}
		if clearRef || clearAll {
			if cliCtx.CurrentRef != "" {
				cliCtx.CurrentRef = ""
				clearedSomething = true
				fmt.Println("Reference context cleared.")
			}
		}

		if !clearedSomething && !clearAll { // clearAll implies intent even if nothing was set
		    fmt.Println("No context parts specified to clear, or already empty.")
        }


		err = config.SaveCLIContext(cliCtx)
		if err != nil {
			return fmt.Errorf("failed to save CLI context: %w", err)
		}

		if clearedSomething || clearAll {
			fmt.Println("Current context after clearing:")
			formatterVal := cmd.Context().Value(shared.FormatterContextKey)
			if formatterVal == nil {
				return output.ErrFormatterNotLoaded
			}
			formatter := formatterVal.(output.Formatter)
			return formatter.Write(cliCtx)
		}
		return nil
	},
}

func init() {
	clearCmd.Flags().BoolVar(&clearRepo, "repo", false, "Clear only the repository URI from context")
	clearCmd.Flags().BoolVar(&clearRef, "ref", false, "Clear only the reference (branch/tag/commit) from context")
	// Can add --all flag later if needed, for now, no flags means all.
}
