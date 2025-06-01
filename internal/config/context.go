package config

import (
	"fmt"
	"os"
	"path/filepath"

	// go-homedir is already a dependency via config.go
	// "gopkg.in/yaml.v3" is already a dependency via config.go
	"gopkg.in/yaml.v3"
)

const (
	contextFileName = "context.yaml"
)

// contextFilePath holds the cached path to the context file
var contextFilePath string

// GetContextFilePath resolves and returns the absolute path to the CLI context file.
// It caches the result for subsequent calls.
func GetContextFilePath() (string, error) {
	if contextFilePath != "" {
		return contextFilePath, nil
	}
	dirPath, err := GetConfigDirPath() // Uses GetConfigDirPath from config.go
	if err != nil {
		return "", fmt.Errorf("failed to determine context directory path: %w", err)
	}
	path := filepath.Join(dirPath, contextFileName)
	contextFilePath = path // Cache it
	return path, nil
}

// LoadCLIContext loads the CLI context from the default location.
// If the context file does not exist, it returns a default (empty) CLIContext.
func LoadCLIContext() (*CLIContext, error) {
	resolvedPath, err := GetContextFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine context file path: %w", err)
	}

	ctx := DefaultCLIContext() // Start with a default empty context

	// Check if the file exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		// Context file not found, return default context. Not an error.
		return ctx, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat context file %s: %w", resolvedPath, err)
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read context file %s: %w", resolvedPath, err)
	}

	if err := yaml.Unmarshal(data, ctx); err != nil {
		// If file is empty, Unmarshal might return EOF or similar, treat as default.
		// For now, any unmarshal error is considered problematic if file is not empty.
		if len(data) == 0 {
			return DefaultCLIContext(), nil
		}
		return nil, fmt.Errorf("failed to unmarshal context file %s: %w", resolvedPath, err)
	}
	return ctx, nil
}

// SaveCLIContext saves the given CLIContext struct to the default location.
// It ensures the configuration directory exists.
func SaveCLIContext(ctx *CLIContext) error {
	resolvedPath, err := GetContextFilePath()
	if err != nil {
		return fmt.Errorf("failed to determine context file path for saving: %w", err)
	}

	// Ensure the directory exists (GetConfigDirPath from config.go handles this,
	// but SaveConfiguration also creates it, so it should exist if config was saved)
	configDirPath := filepath.Dir(resolvedPath)
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDirPath, 0700); err != nil { // User read/write/execute
			return fmt.Errorf("failed to create context directory %s: %w", configDirPath, err)
		}
	}

	data, err := yaml.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal CLI context to YAML: %w", err)
	}

	if err := os.WriteFile(resolvedPath, data, 0600); err != nil { // User read/write
		return fmt.Errorf("failed to write context file %s: %w", resolvedPath, err)
	}
	return nil
}
