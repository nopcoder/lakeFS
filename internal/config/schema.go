package config

// Credentials holds the API access credentials.
type Credentials struct {
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key"`
}

// ServerConfig holds the server connection details.
type ServerConfig struct {
	EndpointURL string `mapstructure:"endpoint_url" yaml:"endpoint_url"`
}

// OutputConfig holds output related configurations.
type OutputConfig struct {
	DefaultFormat string `mapstructure:"default_format" yaml:"default_format"` // e.g., "text", "json"
}

// Configuration is the main structure for ~/.lctl/config.yaml
type Configuration struct {
	Credentials  Credentials  `mapstructure:"credentials" yaml:"credentials"`
	Server       ServerConfig `mapstructure:"server" yaml:"server"`
	Output       OutputConfig `mapstructure:"output" yaml:"output"`
}

// DefaultConfiguration returns a new Configuration with default values.
func DefaultConfiguration() *Configuration {
	return &Configuration{
		Credentials: Credentials{},
		Server: ServerConfig{
			EndpointURL: "http://localhost:8000", // Default lakeFS endpoint
		},
		Output: OutputConfig{
			DefaultFormat: "text",
		},
	}
}

// CLIContext holds the current operational context for the CLI.
type CLIContext struct {
	CurrentRepoURI string `yaml:"current_repo_uri,omitempty"` // e.g., "lakefs://my-repo"
	CurrentRef     string `yaml:"current_ref,omitempty"`      // e.g., "main", "dev", commit ID
}

// DefaultCLIContext returns an empty CLI context.
func DefaultCLIContext() *CLIContext {
    return &CLIContext{}
}
