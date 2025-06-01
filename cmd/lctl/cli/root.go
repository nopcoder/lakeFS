package cli

import (
	"context" // Go context
	"fmt"
	"os"

	"github.com/spf13/cobra"

	apiclient "github.com/example/lctl/internal/client"
	"github.com/example/lctl/internal/config"
	"github.com/example/lctl/internal/output"
	"github.com/example/lctl/internal/shared"
    // "github.com/treeverse/lakefs/pkg/api/apigen" // No longer directly used here

    // Blank imports for command registration
    // _ "github.com/example/lctl/cmd/lctl/cmdbranch" // No longer needed if added in init
    // _ "github.com/example/lctl/cmd/lctl/cmdcommit"
    // _ "github.com/example/lctl/cmd/lctl/cmdconfig"
    // _ "github.com/example/lctl/cmd/lctl/cmdcontext"
    // _ "github.com/example/lctl/cmd/lctl/cmdfs"
    // _ "github.com/example/lctl/cmd/lctl/cmdrepl"
    // _ "github.com/example/lctl/cmd/lctl/cmdrepo"

	// Imports for command packages (needed for init())
	"github.com/example/lctl/cmd/lctl/cmdbranch"
	"github.com/example/lctl/cmd/lctl/cmdcommit"
	"github.com/example/lctl/cmd/lctl/cmdconfig"
	"github.com/example/lctl/cmd/lctl/cmdcontext"
	"github.com/example/lctl/cmd/lctl/cmdfs"
	"github.com/example/lctl/cmd/lctl/cmdrepl"
	"github.com/example/lctl/cmd/lctl/cmdrepo"
)

var (
	cfgFile        string
	outputFormat   string
	endpointURL    string
	accessKeyID    string
	secretAccessKey string
)

var RootCmd = &cobra.Command{
	Use:   "lctl",
	Short: "lctl is an alternative CLI for lakeFS",
	Long: `lctl provides a command-line interface to interact with a lakeFS server,
offering features for repository management, data versioning, and more.
It supports multiple output formats and context-aware operations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		goCtx := cmd.Context()
		if goCtx == nil {
			goCtx = context.Background()
		}

		loadedCfg, err := config.LoadConfiguration(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if endpointURL != "" { loadedCfg.Server.EndpointURL = endpointURL }
		if accessKeyID != "" { loadedCfg.Credentials.AccessKeyID = accessKeyID }
		if secretAccessKey != "" { loadedCfg.Credentials.SecretAccessKey = secretAccessKey }

        currentOutputFormat := loadedCfg.Output.DefaultFormat
        if outputFormat != "" { // Global flag overrides config default
			currentOutputFormat = outputFormat
		}

		goCtx = context.WithValue(goCtx, shared.CLIConfigContextKey, loadedCfg)

		var loadedCLIContext *config.CLIContext
		if c := goCtx.Value(shared.CLIContextContextKey); c != nil {
			loadedCLIContext = c.(*config.CLIContext)
		} else {
			loadedCLIContext, err = config.LoadCLIContext()
			if err != nil { return fmt.Errorf("failed to load CLI context: %w", err) }
		}
		goCtx = context.WithValue(goCtx, shared.CLIContextContextKey, loadedCLIContext)

		var apiClientInstance apiclient.LctlAPIClient // Use the interface type
		if c := goCtx.Value(shared.APIClientContextKey); c != nil {
			apiClientInstance = c.(apiclient.LctlAPIClient)
		} else {
			if !isOfflineCommand(cmd) {
				// var err error // err already declared by loadedCfg
				apiClientInstance, err = apiclient.NewClient(loadedCfg) // client.NewClient now returns LctlAPIClient
				if err != nil { return fmt.Errorf("failed to initialize API client: %w", err) }
			}
		}
		goCtx = context.WithValue(goCtx, shared.APIClientContextKey, apiClientInstance)

		var formatterInstance output.Formatter
		if f := goCtx.Value(shared.FormatterContextKey); f != nil {
			formatterInstance = f.(output.Formatter)
		} else {
			formatterInstance, err = output.NewFormatter(currentOutputFormat, "", cmd.OutOrStdout())
			if err != nil { return fmt.Errorf("failed to initialize output formatter: %w", err) }
		}
		goCtx = context.WithValue(goCtx, shared.FormatterContextKey, formatterInstance)
        goCtx = context.WithValue(goCtx, shared.OutputFormatContextKey, currentOutputFormat)

		cmd.SetContext(goCtx)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
        cmd.Help()
    },
}

func isOfflineCommand(cmd *cobra.Command) bool {
	// This function might be called by PersistentPreRunE if not fully commented out,
	// so keeping a minimal version or ensuring it's not called.
	// For the super-minimal test, PersistentPreRunE is entirely replaced.
	if cmd == nil { return false }
	switch cmd.Name() {
	case "config", "login", "context", "help", "version", "completion", "repl", "lctl":
		return true
	}
	if cmd.HasParent() { return isOfflineCommand(cmd.Parent()) }
	return false
}

func Execute(ctx context.Context) {
	if err := RootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file path (default is ~/.lctl/config.yaml)")
	RootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (text, json, yaml, table). Overrides config default.")
	RootCmd.PersistentFlags().StringVar(&endpointURL, "endpoint-url", "", "lakeFS server endpoint URL (overrides config)")
	RootCmd.PersistentFlags().StringVar(&accessKeyID, "access-key-id", "", "lakeFS access key ID (overrides config)")
	RootCmd.PersistentFlags().StringVar(&secretAccessKey, "secret-access-key", "", "lakeFS secret access key (overrides config)")

	// Register top-level commands
	RootCmd.AddCommand(cmdconfig.ConfigCmd)
	RootCmd.AddCommand(cmdcontext.ContextCmd)
	RootCmd.AddCommand(cmdrepo.RepoCmd)
	RootCmd.AddCommand(cmdbranch.BranchCmd)
	RootCmd.AddCommand(cmdfs.FSCmd)
	RootCmd.AddCommand(cmdcommit.CommitCmd)
	RootCmd.AddCommand(cmdrepl.REPLCmd)
}
