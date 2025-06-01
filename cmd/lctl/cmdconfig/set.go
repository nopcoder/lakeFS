package cmdconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
	// "github.com/example/lctl/internal/shared" // Not used
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value by key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		valueStr := args[1]

		// Get cfgFile path from root command's persistent flag, if set
		// This relies on cfgFile variable being accessible or passed down.
		// For simplicity, we assume RootCmd's PersistentPreRunE has already processed cfgFile global var.
		// A more robust way would be to get it from context if stored there, or re-evaluate flags.
		// For this version, we'll use the global cfgFile variable set by cobra.
		// This is not ideal as it makes this command depend on a global flag variable.
		// A better way would be for root command to store cfgFile path in context.

		// For now, we'll fetch the cfgFile path by accessing the flag directly from the root command
		// This is generally not recommended, but simpler than plumbing it through context for this specific case.
		cfgFilePath, err := cmd.Flags().GetString("config") // This gets the flag value for *this* command.
		if err != nil { // if not found on this command, try root.
			cfgFilePath, err = cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				// This means the --config flag itself is broken, which is unlikely.
				// If it's empty, config.LoadConfiguration will use the default path.
				cfgFilePath = ""
			}
		}


		// Load fresh config, update, then save
		loadedCfg, err := config.LoadConfiguration(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Use reflection to set the value
		v := reflect.ValueOf(loadedCfg).Elem() // Assuming loadedCfg is a pointer
		parts := strings.Split(key, ".")
		current := v
		found := true

		for i, part := range parts {
			if !current.IsValid() {
				found = false
				break
			}
			// Expect exact field names like "Server.EndpointURL"
			fieldName := part

			if current.Kind() == reflect.Ptr {
				current = current.Elem()
			}
			if current.Kind() != reflect.Struct {
				found = false;
				break;
			}

			current = current.FieldByName(fieldName)
			if !current.IsValid() {
				found = false
				break
			}
			// If it's not the last part, and it's a pointer to a struct, ensure it's not nil
			if i < len(parts)-1 && current.Kind() == reflect.Ptr && current.Type().Elem().Kind() == reflect.Struct {
				if current.IsNil() {
					current.Set(reflect.New(current.Type().Elem()))
				}
			}
		}

		if !found || !current.CanSet() {
			return fmt.Errorf("key '%s' not found, not settable, or path is invalid", key)
		}

		// Convert valueStr to the type of the field
		// This is a simplified conversion; robust solution needs type switches or libraries
		switch current.Kind() {
		case reflect.String:
			current.SetString(valueStr)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			iVal, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value '%s': %w", valueStr, err)
			}
			current.SetInt(iVal)
		case reflect.Bool:
			bVal, err := strconv.ParseBool(valueStr)
			if err != nil {
				return fmt.Errorf("invalid boolean value '%s': %w", valueStr, err)
			}
			current.SetBool(bVal)
		default:
			return fmt.Errorf("unsupported type for key '%s': %s", key, current.Kind())
		}

		err = config.SaveConfiguration(loadedCfg, cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Configuration key '%s' set successfully.\n", key)
		return nil
	},
}
