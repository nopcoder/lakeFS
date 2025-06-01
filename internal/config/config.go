package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3" // For more precise saving if needed, viper handles most cases
)

const (
	configDirName  = ".lctl" // Assuming lctl-specific directory
	configFileName = "config.yaml"
)

// configFilePath holds the cached path to the config file
var configFilePath string

// GetConfigDirPath returns the absolute path to the lctl configuration directory.
func GetConfigDirPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}


// GetConfigFilePath resolves and returns the absolute path to the configuration file.
// It caches the result for subsequent calls.
func GetConfigFilePath(customPath string) (string, error) {
	if customPath != "" {
		return customPath, nil
	}
	if configFilePath != "" {
		return configFilePath, nil
	}

	dirPath, err := GetConfigDirPath()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dirPath, configFileName)
	configFilePath = path // Cache it
	return path, nil
}

// LoadConfiguration loads the CLI configuration.
// If customConfigPath is provided, it's used; otherwise, the default path is used.
// It starts with default values and overrides them with values from the config file if it exists.
func LoadConfiguration(customConfigPath string) (*Configuration, error) {
	resolvedPath, err := GetConfigFilePath(customConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to determine config path: %w", err)
	}

	cfg := DefaultConfiguration() // Start with defaults

	// Check if the file exists before trying to read with viper, to avoid error on first run.
	// Viper's ReadInConfig can also handle this, but explicit check gives more control.
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		// Config file not found. Return defaults.
		// It can be created upon first 'config set' or 'login'.
		return cfg, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat config file %s: %w", resolvedPath, err)
	}

	v := viper.New()
	v.SetConfigFile(resolvedPath)
	v.SetConfigType("yaml")

	// Attempt to read the config file
	if err := v.ReadInConfig(); err != nil {
		// Check if it's a "file not found" type of error, which is okay (use defaults)
		// For other errors (e.g., malformed YAML), return the error.
		// However, we already stat'd the file. If ReadInConfig fails now, it's likely parsing.
        return nil, fmt.Errorf("error reading config file %s: %w", resolvedPath, err)
	}

	// Unmarshal the config into our struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from %s: %w", resolvedPath, err)
	}

	return cfg, nil
}

// SaveConfiguration saves the given Configuration struct to the specified path or default.
// It ensures the configuration directory exists.
func SaveConfiguration(cfg *Configuration, customConfigPath string) error {
	resolvedPath, err := GetConfigFilePath(customConfigPath)
	if err != nil {
		return fmt.Errorf("failed to determine config path for saving: %w", err)
	}

	configDirPath := filepath.Dir(resolvedPath)
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDirPath, 0700); err != nil { // User read/write/execute
			return fmt.Errorf("failed to create config directory %s: %w", configDirPath, err)
		}
	}

	// Marshal the configuration to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to YAML: %w", err)
	}

	// Write the YAML data to the file
	if err := os.WriteFile(resolvedPath, data, 0600); err != nil { // User read/write
		return fmt.Errorf("failed to write config file %s: %w", resolvedPath, err)
	}

	return nil
}
