package cmdconfig

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Interactively configure server endpoint and credentials",
	Long:  `Prompts for server endpoint URL, access key ID, and secret access key, then saves them to the configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFilePath, err := cmd.Flags().GetString("config")
		if err != nil {
			cfgFilePath = ""
		}

		loadedCfg, err := config.LoadConfiguration(cfgFilePath)
		if err != nil {
			// If config is malformed, start fresh but warn user?
			// For now, let's proceed with defaults if loading fails badly,
			// as LoadConfiguration returns defaults on "not found".
			// If it's a parse error, LoadConfiguration returns the error, so this RunE won't be hit.
			// This means loadedCfg should always be a valid struct here (either from file or defaults).
			fmt.Fprintf(os.Stderr, "Warning: could not load existing configuration, starting with defaults: %v\n", err)
			loadedCfg = config.DefaultConfiguration()
		}


		// Prompt for Endpoint URL
		promptEndpoint := promptui.Prompt{
			Label:   "Server Endpoint URL",
			Default: loadedCfg.Server.EndpointURL,
		}
		endpointURL, err := promptEndpoint.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		loadedCfg.Server.EndpointURL = endpointURL

		// Prompt for Access Key ID
		promptAccessKey := promptui.Prompt{
			Label:   "Access Key ID",
			Default: loadedCfg.Credentials.AccessKeyID,
		}
		accessKeyID, err := promptAccessKey.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		loadedCfg.Credentials.AccessKeyID = accessKeyID

		// Prompt for Secret Access Key
		promptSecretKey := promptui.Prompt{
			Label: "Secret Access Key",
			Mask:  '*',
			// No default for secret for security, user must re-enter
		}
		secretAccessKey, err := promptSecretKey.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		loadedCfg.Credentials.SecretAccessKey = secretAccessKey

		err = config.SaveConfiguration(loadedCfg, cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println("Configuration updated successfully.")
		if cfgFilePath != "" {
			fmt.Printf("Configuration saved to: %s\n", cfgFilePath)
		} else {
			defaultPath, _ := config.GetConfigFilePath("")
			fmt.Printf("Configuration saved to default location: %s\n", defaultPath)
		}
		return nil
	},
}
