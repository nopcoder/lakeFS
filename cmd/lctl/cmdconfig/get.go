package cmdconfig

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/example/lctl/internal/config"
	"github.com/example/lctl/internal/output"
	"github.com/example/lctl/internal/shared"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		goCtx := cmd.Context()
		loadedCfgVal := goCtx.Value(shared.CLIConfigContextKey)
		if loadedCfgVal == nil {
			return config.ErrConfigNotLoaded
		}
		loadedCfg := loadedCfgVal.(*config.Configuration)

		formatterVal := goCtx.Value(shared.FormatterContextKey)
		if formatterVal == nil {
			return output.ErrFormatterNotLoaded
		}
		formatter := formatterVal.(output.Formatter)

		// Use reflection to get the value
		// This is a simplified example; a more robust solution might involve mapstructure tags
		// or a dedicated function in the config package.
		v := reflect.ValueOf(loadedCfg).Elem() // Assuming loadedCfg is a pointer
		var valueToPrint interface{}

		parts := strings.Split(key, ".")
		current := v
		found := true
		for _, part := range parts {
			if !current.IsValid() {
				found = false
				break
			}
			// Expect exact field names like "Server.EndpointURL"
			fieldName := part

			// Check if current is a pointer, and if so, get the element it points to.
			if current.Kind() == reflect.Ptr {
				current = current.Elem()
			}

			if current.Kind() != reflect.Struct {
				found = false
				break
			}
			current = current.FieldByName(fieldName)
			if !current.IsValid() {
				found = false
				break
			}
		}

		if !found || !current.IsValid() {
			return fmt.Errorf("key '%s' not found or not accessible", key)
		}
		valueToPrint = current.Interface()

		return formatter.Write(valueToPrint)
	},
}
