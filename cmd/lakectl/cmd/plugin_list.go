package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/treeverse/lakefs/pkg/logging"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const pluginListTemplate = `The following lakectl-compatible plugins are available:

{{ range .Plugins }}
  {{ .Path }}
{{- if .Warning }}
    - warning: {{ .Warning }}
{{- end }}
{{ end }}
{{- if .Errors }}
{{ "\n" }}Errors:
{{- range .Errors }}
  - {{ . }}
{{- end }}
{{- end }}
`

type pluginInfo struct {
	Name    string
	Path    string
	Warning string
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available lakectl plugins",
	Long:  `Scans the PATH for executables named "lakectl-*" and lists the detected plugins.`,
	Run: func(cmd *cobra.Command, args []string) {
		// pluginPaths := make(map[string][]string) // This was used in an earlier iteration
		var errorsList []string

		pathEnv := os.Getenv("PATH")
		pathDirs := filepath.SplitList(pathEnv)

		processedPaths := make(map[string]bool) // To avoid processing the same file path multiple times

		for _, dir := range pathDirs {
			if dir == "" {
				// PATH may contain empty strings, which means current directory.
				// For security reasons and consistency with kubectl, we might want to skip this,
				// or ensure it's handled carefully. Kubectl seems to include it.
				// For now, let's resolve it to an absolute path.
				cwd, err := os.Getwd()
				if err != nil {
					errorsList = append(errorsList, fmt.Sprintf("error getting current working directory: %v", err))
					continue
				}
				dir = cwd
			}

			files, err := os.ReadDir(dir)
			if err != nil {
				// Silently ignore errors from non-existent or unreadable directories in PATH
				continue
			}

			for _, file := range files {
				fileName := file.Name()
				if strings.HasPrefix(fileName, "lakectl-") {
					pluginName := strings.TrimPrefix(fileName, "lakectl-")
					if pluginName == "" { // Should not happen if filename is "lakectl-"
						continue
					}
					// Cobra commands can't have dashes, but plugin names can represent them with underscores.
					// For listing, we show the actual filename.
					// pluginName = strings.ReplaceAll(pluginName, "-", "_")

					fullPath := filepath.Join(dir, fileName)
					absPath, err := filepath.Abs(fullPath)
					if err != nil {
						errorsList = append(errorsList, fmt.Sprintf("error getting absolute path for %s: %v", fullPath, err))
						continue
					}

					if processedPaths[absPath] {
						continue // Already processed this exact file path (e.g. via symlink or duplicate PATH entry)
					}
					processedPaths[absPath] = true

					info, err := file.Info()
					if err != nil {
						warnings = append(warnings, fmt.Sprintf("%s: error getting file info: %v", fullPath, err))
						pluginPaths[pluginName] = append(pluginPaths[pluginName], fullPath)
						continue
					}

					isExecutable := false
					if info.IsDir() {
						// Skip directories
						continue
					} else {
						// Check execute permissions
						if runtime.GOOS == "windows" {
							// On Windows, any .exe, .com, .bat, .cmd file found by LookPath is considered executable.
							// Here, we just check if it's a common executable extension.
							// A more robust check might involve trying to use exec.LookPath on the full path.
							ext := strings.ToLower(filepath.Ext(fileName))
							if ext == ".exe" || ext == ".com" || ext == ".bat" || ext == ".cmd" {
								isExecutable = true
							}
						} else {
							if info.Mode().Perm()&0111 != 0 { // Check for user, group or other execute bit
								isExecutable = true
							}
						}
					}

					if !isExecutable {
						warnings = append(warnings, fmt.Sprintf("%s identified as a kubectl plugin, but it is not executable", fullPath))
					}
					// Store all found, executable or not, to handle overshadowing warnings correctly.
					pluginPaths[pluginName] = append(pluginPaths[pluginName], fullPath)
				}
			}
		}

		var displayPlugins []pluginInfo
		sortedPluginNames := maps.Keys(pluginPaths)
		slices.Sort(sortedPluginNames)

		for _, name := range sortedPluginNames {
			paths := pluginPaths[name]
			if len(paths) == 0 {
				continue
			}
			// The first one in the list for a given name is the one that would be executed,
			// assuming PATH order is preserved by SplitList and our iteration.
			// For simplicity, we'll just list them. Overshadowing logic is more complex.

			// Determine effective plugin and overshadowing
			// We need to consider the original PATH order for overshadowing.
			// This simple loop doesn't fully respect that yet for choosing the "active" one.
			// A more accurate way: iterate dirs, then files. The first hit for a plugin name is active.

			effectivePath := paths[0] // Placeholder: first one found based on sorted names
			isEffectivePathExecutable := true // Assume true, warning added earlier if not

			// Re-check executability for the effective path to avoid redundant warning storage
			info, err := os.Stat(effectivePath)
			if err != nil {
				// error already logged
			} else {
				if runtime.GOOS == "windows" {
					ext := strings.ToLower(filepath.Ext(effectivePath))
					isEffectivePathExecutable = (ext == ".exe" || ext == ".com" || ext == ".bat" || ext == ".cmd")
				} else {
					isEffectivePathExecutable = (info.Mode().Perm()&0111 != 0)
				}
			}


			pi := pluginInfo{Name: name, Path: effectivePath}
			if !isEffectivePathExecutable {
				// This warning is now redundant if already added to the global 'warnings' list
				// but good to associate directly if we refine display.
				// Let's ensure warnings are associated correctly.
			}
			displayPlugins = append(displayPlugins, pi)


			// Simplified overshadowing: if more than one path for the same plugin name,
			// all but the first (based on current paths sorting) are overshadowed.
			// For a more accurate overshadowing, we need to respect original PATH order.
			// This current loop iterates sorted plugin names, then their paths as found.
			// A better way for overshadowing:
			// 1. Iterate PATH directories.
			// 2. For each dir, find lakectl-* files.
			// 3. Maintain a map of pluginName -> firstFoundPath.
			// 4. If a subsequent file has the same pluginName, it's overshadowed by firstFoundPath.

			// For now, just list all found paths per plugin name and let user infer from PATH order.
			// Or, pick the first one from `paths` as primary and list others as overshadowed.
		}

		// More accurate listing and overshadowing:
		// Iterate through pathDirs again, this time to build the final list with correct overshadowing
		finalPluginsOutput := []pluginInfo{}
		foundEffectivePlugins := make(map[string]string) // plugin name -> path of chosen executable

		for _, dir := range pathDirs {
			if dir == "" {
				cwd, _ := os.Getwd()
				dir = cwd
			}
			files, err := os.ReadDir(dir)
			if err != nil { continue }

			for _, file := range files {
				fileName := file.Name()
				if strings.HasPrefix(fileName, "lakectl-") {
					pluginName := strings.TrimPrefix(fileName, "lakectl-")
					if pluginName == "" { continue }

					fullPath := filepath.Join(dir, fileName)
					absPath, _ := filepath.Abs(fullPath)

					fileInfo, statErr := file.Info()
					if statErr != nil {
						// This path will be included with a warning
						finalPluginsOutput = append(finalPluginsOutput, pluginInfo{Path: fullPath, Warning: fmt.Sprintf("error getting file info: %v", statErr)})
						continue
					}
					if fileInfo.IsDir() {
						continue
					}

					isExec := false
					if runtime.GOOS == "windows" {
						ext := strings.ToLower(filepath.Ext(fileName))
						isExec = (ext == ".exe" || ext == ".com" || ext == ".bat" || ext == ".cmd")
					} else {
						isExec = (fileInfo.Mode().Perm()&0111 != 0)
					}

					currentPI := pluginInfo{Name: pluginName, Path: absPath}
					if !isExec {
						currentPI.Warning = "identified as a lakectl plugin, but it is not executable"
					}

					if existingPath, ok := foundEffectivePlugins[pluginName]; ok {
						// Already found an effective plugin for this name. This one is overshadowed.
						if currentPI.Warning != "" { // It has its own issue (e.g. not executable)
							currentPI.Warning = fmt.Sprintf("%s (also overshadowed by %s)", currentPI.Warning, existingPath)
						} else {
							currentPI.Warning = fmt.Sprintf("overshadowed by %s", existingPath)
						}
					} else if isExec {
						// This is the first executable plugin found for this name.
						foundEffectivePlugins[pluginName] = absPath
					}
					// Add all found plugins (executable or not, overshadowed or not) to the list for display
					finalPluginsOutput = append(finalPluginsOutput, currentPI)
				}
			}
		}


		// Filter out duplicates from finalPluginsOutput (e.g. if PATH has duplicate dirs)
		// and ensure warnings are consolidated if a file had multiple issues (though less likely with current logic)
		uniqueFinalPlugins := []pluginInfo{}
		seenPaths := make(map[string]bool)
		for _, p := range finalPluginsOutput {
			if !seenPaths[p.Path] {
				uniqueFinalPlugins = append(uniqueFinalPlugins, p)
				seenPaths[p.Path] = true
			}
		}

		// Sort final list by path for consistent output
		slices.SortFunc(uniqueFinalPlugins, func(a, b pluginInfo) int {
			return strings.Compare(a.Path, b.Path)
		})

		// Use the standard Write function for template execution
		Write(pluginListTemplate, struct {
			Plugins []pluginInfo
			Errors  []string
		}{
			Plugins: uniqueFinalPlugins,
			Errors:  errorsList,
		})

		// Collect all distinct warnings from the plugins
		var distinctWarnings []string
		warningTexts := make(map[string]bool)
		for _, p := range uniqueFinalPlugins {
			if p.Warning != "" {
				if !warningTexts[p.Warning] {
					distinctWarnings = append(distinctWarnings, p.Warning)
					warningTexts[p.Warning] = true
				}
			}
		}

		if len(distinctWarnings) > 0 || len(errorsList) > 0 {
			// Print errors and warnings to stderr, but exit 0 as per kubectl behavior unless a fatal error occurred earlier.
			// DieErr would have already exited for fatal errors.
			if len(errorsList) > 0 {
				fmt.Fprintln(os.Stderr, "\nErrors encountered while listing plugins:")
				for _, errMsg := range errorsList {
					fmt.Fprintln(os.Stderr, "  -", errMsg)
				}
			}
			if len(distinctWarnings) > 0 {
				// This message format is closer to kubectl's output for warnings.
				// Kubectl prints "error: X plugin warnings were found" for its own warnings,
				// and individual warnings under each plugin. Our template does the latter.
				// A summary message to stderr is good.
				fmt.Fprintf(os.Stderr, "\nWarning: %d plugin issue(s) found.\n", len(distinctWarnings))
			}
		}
	},
}

//nolint:gochecknoinits
func init() {
	pluginCmd.AddCommand(pluginListCmd)
}
