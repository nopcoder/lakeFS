package cmdconfig

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
)

var (
	initEndpointURL    string
	initAccessKeyID    string
	initSecretKey      string
	initDefaultFormat  string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize lctl configuration with specified or default values",
	Long: `Creates a new configuration file.
If flags are provided, they are used. Otherwise, default values are written.
This will overwrite an existing configuration file if present.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfiguration() // Start with defaults

		// Override with flags if they were actually set (Cobra tracks this)
		if cmd.Flags().Changed("endpoint") {
			cfg.Server.EndpointURL = initEndpointURL
		}
		if cmd.Flags().Changed("access-key-id") {
			cfg.Credentials.AccessKeyID = initAccessKeyID
		}
		if cmd.Flags().Changed("secret-access-key") {
			cfg.Credentials.SecretAccessKey = initSecretKey
		}
		if cmd.Flags().Changed("default-format") {
			cfg.Output.DefaultFormat = initDefaultFormat
		}

		cfgFilePath, err := cmd.Flags().GetString("config") // Get global --config flag
		if err != nil { // Should not happen if flag is defined on root
			cfgFilePath = ""
		}


		err = config.SaveConfiguration(cfg, cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to save initial configuration: %w", err)
		}

		fmt.Println("lctl configuration initialized successfully.")
		if cfgFilePath != "" {
			fmt.Printf("Configuration saved to: %s\n", cfgFilePath)
		} else {
			defaultPath, _ := config.GetConfigFilePath("")
			fmt.Printf("Configuration saved to default location: %s\n", defaultPath)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initEndpointURL, "endpoint", "", "Server endpoint URL")
	initCmd.Flags().StringVar(&initAccessKeyID, "access-key-id", "", "Access Key ID")
	initCmd.Flags().StringVar(&initSecretKey, "secret-access-key", "", "Secret Access Key")
	initCmd.Flags().StringVar(&initDefaultFormat, "default-format", "", "Default output format (e.g., text, json, yaml)")
}
