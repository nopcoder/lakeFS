package shared

// contextKey is a custom type to avoid key collisions in context.Context.
type contextKey string

const (
	// APIClientContextKey is the context key for the lakeFS API client.
	APIClientContextKey = contextKey("apiClient")

	// FormatterContextKey is the context key for the output formatter.
	FormatterContextKey = contextKey("formatter")

	// OutputFormatContextKey is the context key for the raw output format string (e.g. "json", "text").
	OutputFormatContextKey = contextKey("outputFormat")

	// CLIContextContextKey is the context key for the loaded CLIContext struct.
	CLIContextContextKey = contextKey("cliContext")

	// CLIConfigContextKey is the context key for the loaded Configuration struct.
	CLIConfigContextKey = contextKey("cliConfig")
)
