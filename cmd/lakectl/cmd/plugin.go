package cmd

import (
	"github.com/spf13/cobra"
)

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage lakectl plugins",
	Long: `Provides utilities for managing lakectl plugins.

Plugins are standalone executable files that extend lakectl's functionality.
lakectl discovers plugins by looking for executables in your system's PATH
that are named with the prefix "lakectl-".

For example, an executable named "lakectl-myfeature" can be invoked as
"lakectl myfeature [args...]".

Plugin Naming:
  - The executable must start with "lakectl-".
  - The part after "lakectl-" becomes the command name users type.
    (e.g., "lakectl-foo" -> "lakectl foo")
  - Dashes in the plugin name after "lakectl-" (e.g., "lakectl-foo-bar")
    will be part of the command users type ("lakectl foo-bar").
    (Note: This differs slightly from kubectl where dashes create subcommands.
    For lakectl, "lakectl-foo-bar" is a single plugin "foo-bar", not "foo bar".)
  - To include dashes in the invoked command name itself (e.g. "lakectl my-command"),
    use an underscore in the executable filename after the initial "lakectl-" prefix
    (e.g., "lakectl-my_command"). This will be callable as "lakectl my-command".

Installation:
  - Place your "lakectl-..." executable file in a directory listed in your PATH.
  - Ensure the file has execute permissions.

Execution:
  - When you run "lakectl some-plugin arg1 --flag", lakectl searches for
    "lakectl-some-plugin" in PATH.
  - If found and executable, it runs the plugin, passing "arg1 --flag" as arguments.
  - The plugin inherits environment variables from lakectl.
  - Standard output, standard error, and the exit code of the plugin are propagated.
  - Built-in lakectl commands always take precedence over plugins.

Use "lakectl plugin list" to see all detected plugins and any warnings.
`,
	// Run: func(cmd *cobra.Command, args []string) {}, // No action for 'lakectl plugin' itself
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(pluginCmd)
}
