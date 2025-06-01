package cmdcontext

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
	"github.com/example/lctl/internal/output"
	"github.com/example/lctl/internal/shared"
	lakefsuri "github.com/treeverse/lakefs/pkg/uri" // Using alias to avoid conflict for Parse
)

const lctlLakeFSScheme = "lakefs" // Local definition

var setRepoCmd = &cobra.Command{
	Use:   "set-repo <repo_uri>",
	Short: "Set the current repository URI in CLI context (e.g., lakefs://my-repo)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURIStr := args[0]

		// Basic validation
		if !strings.HasPrefix(repoURIStr, lctlLakeFSScheme+"://") {
			return fmt.Errorf("invalid repository URI: must start with '%s://'", lctlLakeFSScheme)
		}
		// We still use lakefsuri.Parse to leverage its parsing logic
		parsedURI, err := lakefsuri.Parse(repoURIStr)
		if err != nil {
			return fmt.Errorf("invalid repository URI '%s': %w", repoURIStr, err)
		}
		if parsedURI.Repository == "" {
			return fmt.Errorf("invalid repository URI '%s': missing repository ID", repoURIStr)
		}
		// We only want to store the repo part, not ref or path from this command.
		// Construct a URI that only contains the repository.
		repoOnlyURI := fmt.Sprintf("%s://%s", lctlLakeFSScheme, parsedURI.Repository)


		cliCtx, err := config.LoadCLIContext()
		if err != nil {
			return fmt.Errorf("failed to load CLI context: %w", err)
		}

		cliCtx.CurrentRepoURI = repoOnlyURI

		err = config.SaveCLIContext(cliCtx)
		if err != nil {
			return fmt.Errorf("failed to save CLI context: %w", err)
		}

		fmt.Printf("Repository context set to: %s\n", cliCtx.CurrentRepoURI)

		// Print the full context
		formatterVal := cmd.Context().Value(shared.FormatterContextKey)
		if formatterVal == nil {
			return output.ErrFormatterNotLoaded
		}
		formatter := formatterVal.(output.Formatter)
		return formatter.Write(cliCtx)
	},
}
