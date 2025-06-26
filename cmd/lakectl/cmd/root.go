package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/go-openapi/swag"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/treeverse/lakefs/pkg/api/apigen"
	"github.com/treeverse/lakefs/pkg/api/apiutil"
	lakefsconfig "github.com/treeverse/lakefs/pkg/config"
	"github.com/treeverse/lakefs/pkg/git"
	giterror "github.com/treeverse/lakefs/pkg/git/errors"
	"github.com/treeverse/lakefs/pkg/local"
	"github.com/treeverse/lakefs/pkg/logging"
	"github.com/treeverse/lakefs/pkg/osinfo"
	"os/exec"
	"runtime"

	"github.com/treeverse/lakefs/pkg/uri"
	"github.com/treeverse/lakefs/pkg/version"
	"golang.org/x/term"
)

const (
	DefaultMaxIdleConnsPerHost = 100
	// version templates
	getLakeFSVersionErrorTemplate = `{{ "Failed getting lakeFS server version:" | red }} {{ . }}
`
	getLatestVersionErrorTemplate = `{{ "Failed getting latest lakectl version:" | red }} {{ . }}
`
	versionTemplate = `lakectl version: {{.LakectlVersion }}
{{- if .LakeFSVersion }}
lakeFS version: {{.LakeFSVersion}}
{{- end }}
{{- if .UpgradeURL }}{{ "\n" }}{{ end -}}
{{- if .LakectlLatestVersion }}
{{ "lakectl out of date!" | yellow }} (Available: {{ .LakectlLatestVersion }})
{{- end }}
{{- if .LakeFSLatestVersion }}
{{ "lakeFS out of date!" | yellow }} (Available: {{ .LakeFSLatestVersion }})
{{- end }}
{{- if .UpgradeURL }}
Get the latest release {{ .UpgradeURL|blue }}
{{- end }}
`
)

type RetriesCfg struct {
	Enabled         bool          `mapstructure:"enabled"`
	MaxAttempts     int           `mapstructure:"max_attempts"`      // MaxAttempts is the maximum number of attempts
	MinWaitInterval time.Duration `mapstructure:"min_wait_interval"` // MinWaitInterval is the minimum amount of time to wait between retries
	MaxWaitInterval time.Duration `mapstructure:"max_wait_interval"` // MaxWaitInterval is the maximum amount of time to wait between retries
}

// Configuration is the user-visible configuration structure in Golang form.
// When editing, make sure *all* fields have a `mapstructure:"..."` tag, to simplify future refactoring.
type Configuration struct {
	Credentials struct {
		AccessKeyID     lakefsconfig.OnlyString `mapstructure:"access_key_id"`
		SecretAccessKey lakefsconfig.OnlyString `mapstructure:"secret_access_key"`
		Provider        struct {
			Type   lakefsconfig.OnlyString `mapstructure:"type"`
			AWSIAM struct {
				TokenTTL            time.Duration      `mapstructure:"token_ttl_seconds"`
				URLPresignTTL       time.Duration      `mapstructure:"url_presign_ttl_seconds"`
				RefreshInterval     time.Duration      `mapstructure:"refresh_interval"`
				TokenRequestHeaders *map[string]string `mapstructure:"token_request_headers"`
			} `mapstructure:"aws_iam"`
		} `mapstructure:"provider"`
	} `mapstructure:"credentials"`
	Network struct {
		HTTP2 struct {
			Enabled bool `mapstructure:"enabled"`
		} `mapstructure:"http2"`
	} `mapstructure:"network"`
	Server struct {
		EndpointURL lakefsconfig.OnlyString `mapstructure:"endpoint_url"`
		Retries     RetriesCfg              `mapstructure:"retries"`
	} `mapstructure:"server"`
	Options struct {
		Parallelism int `mapstructure:"parallelism"`
	} `mapstructure:"options"`
	Metastore struct {
		Type lakefsconfig.OnlyString `mapstructure:"type"`
		Hive struct {
			URI           lakefsconfig.OnlyString `mapstructure:"uri"`
			DBLocationURI lakefsconfig.OnlyString `mapstructure:"db_location_uri"`
		} `mapstructure:"hive"`
		Glue struct {
			// TODO(ariels): Refactor credentials to share with server side.
			Profile         lakefsconfig.OnlyString `mapstructure:"profile"`
			CredentialsFile lakefsconfig.OnlyString `mapstructure:"credentials_file"`
			DBLocationURI   lakefsconfig.OnlyString `mapstructure:"db_location_uri"`
			Credentials     *struct {
				AccessKeyID     lakefsconfig.OnlyString `mapstructure:"access_key_id"`
				AccessSecretKey lakefsconfig.OnlyString `mapstructure:"access_secret_key"`
				SessionToken    lakefsconfig.OnlyString `mapstructure:"session_token"`
			} `mapstructure:"credentials"`

			Region    lakefsconfig.OnlyString `mapstructure:"region"`
			CatalogID lakefsconfig.OnlyString `mapstructure:"catalog_id"`
		} `mapstructure:"glue"`
		// setting FixSparkPlaceholder to true will change spark placeholder with the actual location. for more information see https://github.com/treeverse/lakeFS/issues/2213
		FixSparkPlaceholder bool `mapstructure:"fix_spark_placeholder"`
	}
	Local struct {
		// SkipNonRegularFiles - By default lakectl local fails if local directory contains a symbolic link. When set, lakectl will ignore the symbolic links instead.
		SkipNonRegularFiles bool `mapstructure:"skip_non_regular_files"`
	} `mapstructure:"local"`
	// Experimental - Use caution when enabling experimental features. It should only be used after consulting with the lakeFS team!
	Experimental struct {
		Local struct {
			POSIXPerm struct {
				Enabled    bool `mapstructure:"enabled"`
				IncludeUID bool `mapstructure:"include_uid"`
				IncludeGID bool `mapstructure:"include_gid"`
			} `mapstructure:"posix_permissions"`
		} `mapstructure:"local"`
	} `mapstructure:"experimental"`
}

type versionInfo struct {
	LakectlVersion       string
	LakeFSVersion        string
	LakectlLatestVersion string
	LakeFSLatestVersion  string
	UpgradeURL           string
}

var (
	cfgFile string
	cfgErr  error
	cfg     *Configuration

	// baseURI default value is set by the environment variable LAKECTL_BASE_URI and
	// override by flag 'base-url'. The baseURI is used as a prefix when we parse lakefs address (repo, ref or path).
	// The prefix is used only when the address we parse is not a full address (starts with 'lakefs://' scheme).
	// Examples:
	//   `--base-uri lakefs:// repo1` will resolve to repository `lakefs://repo1`
	//   `--base-uri lakefs://repo1 /main/file.md` will resolve to path `lakefs://repo1/main/file.md`
	baseURI string

	// logLevel logging level (default is off)
	logLevel string
	// logFormat logging output format
	logFormat string
	// logOutputs logging outputs
	logOutputs []string

	// noColorRequested is set to true when the user requests no color output
	noColorRequested = false

	// verboseMode is set to true when the user requests verbose output
	verboseMode = false
)

const (
	recursiveFlagName     = "recursive"
	recursiveFlagShort    = "r"
	storageIDFlagName     = "storage-id"
	presignFlagName       = "pre-sign"
	parallelismFlagName   = "parallelism"
	noProgressBarFlagName = "no-progress"

	defaultParallelism = 25
	defaultSyncPresign = true
	defaultNoProgress  = false

	paginationPrefixFlagName = "prefix"
	paginationAfterFlagName  = "after"
	paginationAmountFlagName = "amount"

	myRepoExample   = "lakefs://my-repo"
	myBucketExample = "s3://my-bucket"
	myBranchExample = "my-branch"
	myRunIDExample  = "20230719152411arS0z6I"
	myDigestExample = "600dc0ffee"

	commitMsgFlagName     = "message"
	allowEmptyMsgFlagName = "allow-empty-message"
	fmtErrEmptyMsg        = `commit with no message without specifying the "--allow-empty-message" flag`
	metaFlagName          = "meta"

	defaultHTTP2Enabled     = true
	defaultMaxAttempts      = 4
	defaultMaxRetryInterval = 30 * time.Second
	defaultMinRetryInterval = 200 * time.Millisecond
)

func withRecursiveFlag(cmd *cobra.Command, usage string) {
	cmd.Flags().BoolP(recursiveFlagName, recursiveFlagShort, false, usage)
}

func withStorageID(cmd *cobra.Command) {
	cmd.Flags().String(storageIDFlagName, "", "")
	if err := cmd.Flags().MarkHidden(storageIDFlagName); err != nil {
		DieErr(err)
	}
}

func withParallelismFlag(cmd *cobra.Command) {
	cmd.Flags().IntP(parallelismFlagName, "p", defaultParallelism,
		"Max concurrent operations to perform")
}

func withPresignFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(presignFlagName, defaultSyncPresign,
		"Use pre-signed URLs when downloading/uploading data (recommended)")
}

func withNoProgress(cmd *cobra.Command) {
	cmd.Flags().Bool(noProgressBarFlagName, defaultNoProgress,
		"Disable progress bar animation for IO operations")
}

func withSyncFlags(cmd *cobra.Command) {
	withParallelismFlag(cmd)
	withPresignFlag(cmd)
	withNoProgress(cmd)
}

func getStorageConfigOrDie(ctx context.Context, client *apigen.ClientWithResponses, repositoryID string) *apigen.StorageConfig {
	confResp, err := client.GetConfigWithResponse(ctx)
	DieOnErrorOrUnexpectedStatusCode(confResp, err, http.StatusOK)
	if confResp.JSON200 == nil {
		Die("Bad response from server for GetConfig", 1)
	}

	storageConfigList := confResp.JSON200.StorageConfigList
	if storageConfigList != nil && len(*storageConfigList) > 1 {
		repoResp, errRepo := client.GetRepositoryWithResponse(ctx, repositoryID)
		DieOnErrorOrUnexpectedStatusCode(repoResp, errRepo, http.StatusOK)
		if repoResp.JSON200 == nil {
			Die("Bad response from server for GetRepository", 1)
		}
		storageID := repoResp.JSON200.StorageId

		// find the storage config for the repository
		for _, storageConfig := range *storageConfigList {
			if swag.StringValue(storageConfig.BlockstoreId) == swag.StringValue(storageID) {
				return &storageConfig
			}
		}

		Die("Storage config not found for repo "+repositoryID, 1)
	}

	storageConfig := confResp.JSON200.StorageConfig
	if storageConfig == nil {
		Die("Bad response from server for GetConfig", 1)
	}
	return storageConfig
}

type PresignMode struct {
	Enabled   bool
	Multipart bool
}

func getServerPreSignMode(ctx context.Context, client *apigen.ClientWithResponses, repositoryID string) PresignMode {
	storageConfig := getStorageConfigOrDie(ctx, client, repositoryID)
	return PresignMode{
		Enabled:   storageConfig.PreSignSupport,
		Multipart: swag.BoolValue(storageConfig.PreSignMultipartUpload),
	}
}

func getPresignMode(cmd *cobra.Command, client *apigen.ClientWithResponses, repositoryID string) PresignMode {
	// use flags if set
	presignFlag := cmd.Flags().Lookup(presignFlagName)
	var presignMode PresignMode
	if presignFlag.Changed {
		presignMode.Enabled = Must(cmd.Flags().GetBool(presignFlagName))
	}
	// fetch server config if needed
	// if presign flag is not set, use server config
	// if presign flag is set, check if server supports multipart upload
	if !presignFlag.Changed || presignMode.Enabled {
		presignMode = getServerPreSignMode(cmd.Context(), client, repositoryID)
	}
	return presignMode
}

func getNoProgressMode(cmd *cobra.Command) bool {
	// Disable progress bar if stdout is not tty
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return true
	}
	return Must(cmd.Flags().GetBool(noProgressBarFlagName))
}

func getSyncFlags(cmd *cobra.Command, client *apigen.ClientWithResponses, repositoryID string) local.SyncFlags {
	parallelism := Must(cmd.Flags().GetInt(parallelismFlagName))
	if parallelism < 1 {
		DieFmt("Invalid value for parallelism (%d), minimum is 1.\n", parallelism)
	}
	changed := cmd.Flags().Changed(parallelismFlagName)
	if viper.IsSet("options.parallelism") && !changed {
		parallelism = cfg.Options.Parallelism
	}

	presignMode := getPresignMode(cmd, client, repositoryID)
	return local.SyncFlags{
		Parallelism:      parallelism,
		Presign:          presignMode.Enabled,
		PresignMultipart: presignMode.Multipart,
		NoProgress:       getNoProgressMode(cmd),
	}
}

// getSyncArgs parses arguments to extract a remote URI and deduces the local path.
// If the local path isn't provided and considerGitRoot is true, it uses the git repository root.
func getSyncArgs(args []string, requireRemote bool, considerGitRoot bool) (remote *uri.URI, localPath string) {
	idx := 0
	if requireRemote {
		remote = MustParsePathURI("path URI", args[0])
		idx += 1
	}

	if len(args) > idx {
		expanded := Must(homedir.Expand(args[idx]))
		localPath = Must(filepath.Abs(expanded))
		return
	}

	localPath = Must(filepath.Abs("."))
	if considerGitRoot {
		gitRoot, err := git.GetRepositoryPath(localPath)
		if err == nil {
			localPath = gitRoot
		} else if !(errors.Is(err, giterror.ErrNotARepository) || errors.Is(err, giterror.ErrNoGit)) { // allow support in environments with no git
			DieErr(err)
		}
	}
	return
}

func getPaginationFlags(cmd *cobra.Command) (prefix string, after string, amount int) {
	prefix = Must(cmd.Flags().GetString(paginationPrefixFlagName))
	after = Must(cmd.Flags().GetString(paginationAfterFlagName))
	amount = Must(cmd.Flags().GetInt(paginationAmountFlagName))

	return
}

type PaginationOptions func(*cobra.Command)

func withoutPrefix(cmd *cobra.Command) {
	if err := cmd.Flags().MarkHidden(paginationPrefixFlagName); err != nil {
		DieErr(err)
	}
}

func withPaginationFlags(cmd *cobra.Command, options ...PaginationOptions) {
	cmd.Flags().SortFlags = false
	cmd.Flags().Int(paginationAmountFlagName, defaultAmountArgumentValue, "how many results to return")
	cmd.Flags().String(paginationAfterFlagName, "", "show results after this value (used for pagination)")
	cmd.Flags().String(paginationPrefixFlagName, "", "filter results by prefix (used for pagination)")

	for _, option := range options {
		option(cmd)
	}
}

func withMessageFlags(cmd *cobra.Command, allowEmpty bool) {
	cmd.Flags().StringP(commitMsgFlagName, "m", "", "commit message")
	cmd.Flags().Bool(allowEmptyMsgFlagName, allowEmpty, "allow an empty commit message")
}

func withMetadataFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice(metaFlagName, []string{}, "key value pair in the form of key=value")
}

func withCommitFlags(cmd *cobra.Command, allowEmptyMessage bool) {
	withMessageFlags(cmd, allowEmptyMessage)
	withMetadataFlag(cmd)
}

func getCommitFlags(cmd *cobra.Command) (string, map[string]string) {
	message := Must(cmd.Flags().GetString(commitMsgFlagName))
	emptyMessageBool := Must(cmd.Flags().GetBool(allowEmptyMsgFlagName))
	if strings.TrimSpace(message) == "" && !emptyMessageBool {
		DieFmt(fmtErrEmptyMsg)
	}

	kvPairs, err := getKV(cmd, metaFlagName)
	if err != nil {
		DieErr(err)
	}

	return message, kvPairs
}

func getKV(cmd *cobra.Command, name string) (map[string]string, error) {
	kvList, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		return nil, err
	}

	kv := make(map[string]string)
	for _, pair := range kvList {
		key, value, found := strings.Cut(pair, "=")
		if !found {
			return nil, errInvalidKeyValueFormat
		}
		kv[key] = value
	}
	return kv, nil
}

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:   "lakectl",
	Short: "A cli tool to explore manage and work with lakeFS",
	Long: `lakectl is a CLI tool allowing exploration and manipulation of a lakeFS environment.

It can be extended with plugins; see 'lakectl plugin --help' for more information.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		preRunCmd(cmd)
		sendStats(cmd, "")
	},
	Run: func(cmd *cobra.Command, args []string) {
		if !Must(cmd.Flags().GetBool("version")) {
			if err := cmd.Help(); err != nil {
				WriteIfVerbose("failed showing help {{ . }}", err)
			}
			return
		}

		info := versionInfo{LakectlVersion: version.Version}

		// get lakeFS server version

		client := getClient()

		resp, err := client.GetConfigWithResponse(cmd.Context())
		if err != nil {
			WriteIfVerbose(getLakeFSVersionErrorTemplate, err)
		} else if resp.JSON200 == nil {
			WriteIfVerbose(getLakeFSVersionErrorTemplate, resp.Status())
		} else {
			lakefsVersion := resp.JSON200
			info.LakeFSVersion = swag.StringValue(lakefsVersion.VersionConfig.Version)
			if swag.BoolValue(lakefsVersion.VersionConfig.UpgradeRecommended) {
				info.LakeFSLatestVersion = swag.StringValue(lakefsVersion.VersionConfig.LatestVersion)
			}
			upgradeURL := swag.StringValue(lakefsVersion.VersionConfig.UpgradeUrl)
			if upgradeURL != "" {
				info.UpgradeURL = upgradeURL
			}
		}
		// get lakectl latest version
		ghReleases := version.NewGithubReleases(version.GithubRepoOwner, version.GithubRepoName)
		latestVer, err := ghReleases.FetchLatestVersion()
		if err != nil {
			WriteIfVerbose(getLatestVersionErrorTemplate, err)
		} else {
			latest, err := version.CheckLatestVersion(latestVer)
			if err != nil {
				WriteIfVerbose("failed parsing {{ . }}", err)
			} else if latest.Outdated {
				info.LakectlLatestVersion = latest.LatestVersion
				if info.UpgradeURL == "" {
					info.UpgradeURL = version.DefaultReleasesURL
				}
			}
		}

		Write(versionTemplate, info)
	},
}

var excludeStatsCmds = []string{
	"doctor",
	"config",
}

func preRunCmd(cmd *cobra.Command) {
	logging.SetLevel(logLevel)
	logging.SetOutputFormat(logFormat)
	err := logging.SetOutputs(logOutputs, 0, 0)
	if err != nil {
		DieFmt("Failed to setup logging: %s", err)
	}
	if noColorRequested {
		DisableColors()
	}
	if cmd == configCmd {
		return
	}

	if cfgFile != "" && cfgErr != nil {
		DieFmt("error reading configuration file: %v", cfgErr)
	}

	logging.ContextUnavailable().
		WithField("file", viper.ConfigFileUsed()).
		Debug("loaded configuration from file")
	err = viper.UnmarshalExact(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			lakefsconfig.DecodeOnlyString,
			mapstructure.StringToTimeDurationHookFunc(),
			lakefsconfig.DecodeStringToMap(),
		)))
	if err != nil {
		DieFmt("error unmarshal configuration: %v", err)
	}
}

func sendStats(cmd *cobra.Command, cmdSuffix string) {
	if version.IsVersionUnreleased() || !cmd.HasParent() { // Don't send statistics for root command
		return
	}
	var cmdName string
	for curr := cmd; curr.HasParent(); curr = curr.Parent() {
		if cmdName != "" {
			cmdName = curr.Name() + "_" + cmdName
		} else {
			cmdName = curr.Name()
		}
	}
	if cmdSuffix != "" {
		cmdName = cmdName + "_" + cmdSuffix
	}
	if !slices.Contains(excludeStatsCmds, cmdName) { // Skip excluded commands
		resp, err := getClient().PostStatsEventsWithResponse(cmd.Context(), apigen.PostStatsEventsJSONRequestBody{
			Events: []apigen.StatsEvent{
				{
					Class: "lakectl",
					Name:  cmdName,
					Count: 1,
				},
			},
		})

		var errStr string
		if err != nil {
			errStr = err.Error()
		} else if resp.StatusCode() != http.StatusNoContent {
			errStr = resp.Status()
		}
		if errStr != "" {
			logging.ContextUnavailable().Debugf("Warning: failed sending statistics: %s\n", errStr)
		}
	}
}

func getHTTPClient() *http.Client {
	// Override MaxIdleConnsPerHost to allow highly concurrent access to our API client.
	// This is done to avoid accumulating many sockets in `TIME_WAIT` status that were closed
	// only to be immediately reopened.
	// see: https://stackoverflow.com/a/39834253
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if !cfg.Network.HTTP2.Enabled {
		transport.ForceAttemptHTTP2 = false
		transport.TLSClientConfig.NextProtos = []string{}
	}
	transport.MaxIdleConnsPerHost = DefaultMaxIdleConnsPerHost
	if !cfg.Server.Retries.Enabled {
		return &http.Client{Transport: transport}
	}
	return NewRetryClient(cfg.Server.Retries, transport)
}

func getClient() *apigen.ClientWithResponses {
	httpClient := getHTTPClient()

	accessKeyID := cfg.Credentials.AccessKeyID
	secretAccessKey := cfg.Credentials.SecretAccessKey
	basicAuthProvider, err := securityprovider.NewSecurityProviderBasicAuth(string(accessKeyID), string(secretAccessKey))
	if err != nil {
		DieErr(err)
	}

	serverEndpoint, err := apiutil.NormalizeLakeFSEndpoint(cfg.Server.EndpointURL.String())
	if err != nil {
		DieErr(err)
	}

	oss := osinfo.GetOSInfo()
	client, err := apigen.NewClientWithResponses(
		serverEndpoint,
		apigen.WithHTTPClient(httpClient),
		apigen.WithRequestEditorFn(basicAuthProvider.Intercept),
		apigen.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			// This UA string structure is agreed upon
			// Please consider that when making changes
			req.Header.Set("User-Agent", fmt.Sprintf("lakectl/%s/%s/%s/%s", version.Version, oss.OS, oss.Version, oss.Platform))
			return nil
		}),
	)
	if err != nil {
		Die(fmt.Sprintf("could not initialize API client: %s", err), 1)
	}
	return client
}

// isUnknownCommandError checks if the error from ExecuteC is an unknown command error.
// Cobra doesn't expose a specific error type for this, so we check the message.
func isUnknownCommandError(err error) bool {
	if err == nil {
		return false
	}
	// This is based on the error message format from cobra.findSuggestions
	// It might be brittle if Cobra changes its error messages.
	return strings.HasPrefix(err.Error(), "unknown command")
}

// handlePluginCommand attempts to find and execute a lakectl plugin.
// It returns true if a plugin was found and executed (or an attempt was made),
// and false otherwise.
func handlePluginCommand(cmd *cobra.Command, args []string) bool {
	if len(args) == 0 {
		return false // No command to interpret as a plugin
	}

	pluginCmdName := args[0]
	pluginExecName := "lakectl-" + pluginCmdName

	// Find the plugin in PATH
	pluginPath, err := exec.LookPath(pluginExecName)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			// Plugin not found in PATH, not necessarily an error here,
			// could be a genuine unknown command for lakectl itself.
			return false
		}
		// Other error from LookPath (e.g., permission issues)
		DieErr(fmt.Errorf("error looking up plugin %s: %w", pluginExecName, err))
	}

	// Check file stats
	info, err := os.Stat(pluginPath)
	if err != nil {
		// This might happen if the file was removed between LookPath and Stat, or permissions changed.
		DieErr(fmt.Errorf("error getting info for plugin %s: %w", pluginPath, err))
	}

	if info.IsDir() {
		return false // It's a directory, not a plugin file, let Cobra handle.
	}

	// Check if executable (LookPath does this for non-Windows, but explicit check is good)
	// On Windows, LookPath checks if it's an executable type, not necessarily permissions.
	// On Unix, os.FileMode.Perm() & 0111 checks execute bit for user, group, or other.
	if runtime.GOOS != "windows" && (info.Mode().Perm()&0111 == 0) {
		DieErr(fmt.Errorf("plugin '%s' found at '%s' but is not executable", pluginCmdName, pluginPath))
	}
	// For Windows, LookPath finding it and it not being a directory is usually enough.

	// Prepare and execute the plugin command
	pluginArgs := args[1:]
	externalCmd := exec.Command(pluginPath, pluginArgs...)
	externalCmd.Stdout = os.Stdout
	externalCmd.Stderr = os.Stderr
	externalCmd.Stdin = os.Stdin
	externalCmd.Env = os.Environ() // Pass parent environment

	// Run the command
	if err := externalCmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// The command ran and exited with a non-zero status.
			// We want lakectl to exit with the same status code.
			os.Exit(exitError.ExitCode())
		} else {
			// Other errors (e.g., failed to start, I/O errors not related to exit status)
			DieErr(fmt.Errorf("failed to run plugin %s: %w", pluginExecName, err))
		}
	}
	return true // Plugin was handled
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	executedCmd, err := rootCmd.ExecuteC()
	if err != nil {
		// Check if it's an unknown command error and try to handle it as a plugin
		if isUnknownCommandError(err) {
			// executedCmd.Context() is nil for unknown commands, so we need to find the args from os.Args
			// os.Args[0] is "lakectl", os.Args[1] would be the potential plugin command.
			// We need to filter out flags that Cobra might have parsed before realizing it's an unknown command.
			// However, for simplicity in the first pass, we'll use executedCmd.Flags().Args()
			// if available, or fallback to os.Args, being mindful that Cobra's parsing
			// When ExecuteC returns an unknown command error, 'executedCmd' is the command
			// that was being processed (often rootCmd itself if the first arg is unknown).
			// Its Flags().Args() should contain the sequence of non-flag arguments
			// that Cobra couldn't match to a known command.
			// For example, if `lakectl myplugin --myflag val arg1` is called,
			// and `myplugin` is not built-in, then executedCmd.Flags().Args()
			// should be `["myplugin", "--myflag", "val", "arg1"]` if --myflag is not a global flag,
			// or `["myplugin", "val", "arg1"]` if --myflag was consumed by the plugin.
			// More accurately, it will be all args that were not parsed as flags by Cobra for the executedCmd.
			pluginCandidateArgs := executedCmd.Flags().Args()

			// If executedCmd.Flags().Args() is empty, we might inspect os.Args directly.
			// This could happen if the input was `lakectl unknown` and `unknown` was not parsed as an arg.
			// However, `handlePluginCommand` itself checks if `pluginCandidateArgs` is empty.
			// A more direct way to get what the user typed after `lakectl` and any global flags
			// would be to find the command path in os.Args.
			// For `lakectl --global-flag x unknown-cmd --plugin-flag y`:
			// os.Args = ["lakectl", "--global-flag", "x", "unknown-cmd", "--plugin-flag", "y"]
			// We need to pass ["unknown-cmd", "--plugin-flag", "y"] to handlePluginCommand.

			// Let's find the actual command args passed by the user, after stripping lakectl itself
			// and any flags that were processed by Cobra for the rootCmd.
			// The `executedCmd.Flags().Args()` should give us the arguments that were not recognized
			// by the command `executedCmd` (which is `rootCmd` in this case of top-level unknown command).
			// So, if the user typed `lakectl --config my.cfg plugin-name --plugin-arg`,
			// `rootCmd.ExecuteC()` would parse `--config my.cfg`, and then `rootCmd.Flags().Args()`
			// should yield `["plugin-name", "--plugin-arg"]`.

			if !handlePluginCommand(executedCmd, pluginCandidateArgs) {
				// Not handled as a plugin, or plugin not found, so propagate original error.
				// We also want to show help for the command that was attempted, if possible.
				_ = executedCmd.Help() // Show help for the command that failed (e.g. rootCmd help)
				DieErr(err)            // Then exit with the original error.
			}
			// If handlePluginCommand returned true, it means the plugin was executed (or attempted)
			// and it handled os.Exit itself or completed successfully.
			return
		}
		// For other errors, DieErr as before
		DieErr(err)
	}
}

// For testing purposes only
func ResetConfigForTesting() {
	cfgFile = "" // Prevent loading from default user config file during tests
	cfg = new(Configuration) // Create a new instance for cfg

	viper.Reset() // Reset all viper settings

	// Re-initialize viper with default values by calling initConfig directly.
	// initConfig sets defaults and then tries to read a config file (which won't be found if cfgFile is empty).
	initConfig()

	// After initConfig, viper has the defaults (and potentially env vars). Unmarshal into our cfg.
	err := viper.UnmarshalExact(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			lakefsconfig.DecodeOnlyString,
			mapstructure.StringToTimeDurationHookFunc(),
			lakefsconfig.DecodeStringToMap(),
		)))
	if err != nil {
		// This is a fatal error in test setup, so panic.
		panic(fmt.Sprintf("ResetConfigForTesting: failed to unmarshal config into cfg: %v. Viper keys: %v", err, viper.AllKeys()))
	}
}

// BuildRootCmd returns the root command for lakectl.
// For testing, this ensures we get the command structure.
func BuildRootCmd() *cobra.Command {
	// Since subcommands are added to the global rootCmd in their init() functions,
	// and flags are also defined on this global, returning it is appropriate.
	// Test execution will call SetArgs on the returned command.
	return rootCmd
}


//nolint:gochecknoinits
func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.lakectl.yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColorRequested, "no-color", getEnvNoColor(), "don't use fancy output colors (default value can be set by NO_COLOR environment variable)")
	rootCmd.PersistentFlags().StringVarP(&baseURI, "base-uri", "", os.Getenv("LAKECTL_BASE_URI"), "base URI used for lakeFS address parse")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", "none", "set logging level")
	rootCmd.PersistentFlags().StringVarP(&logFormat, "log-format", "", "", "set logging output format")
	rootCmd.PersistentFlags().StringSliceVarP(&logOutputs, "log-output", "", []string{}, "set logging output(s)")
	rootCmd.PersistentFlags().BoolVar(&verboseMode, "verbose", false, "run in verbose mode")
	rootCmd.Flags().BoolP("version", "v", false, "version for lakectl")
}

func getEnvNoColor() bool {
	v := os.Getenv("NO_COLOR")
	return v != "" && v != "0"
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else if envCfgFile, _ := os.LookupEnv("LAKECTL_CONFIG_FILE"); envCfgFile != "" {
		// Use config file from the env variable.
		viper.SetConfigFile(envCfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			DieErr(err)
		}

		// Search config in home directory
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".lakectl")
	}
	viper.SetEnvPrefix("LAKECTL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // support nested config
	viper.AutomaticEnv()                                   // read in environment variables that match

	// Inform viper of all expected fields.
	// Otherwise, it fails to deserialize from the environment.
	var conf Configuration
	keys := lakefsconfig.GetStructKeys(reflect.TypeOf(conf), "mapstructure", "squash")
	for _, key := range keys {
		viper.SetDefault(key, nil)
	}

	// set defaults
	viper.SetDefault("metastore.hive.db_location_uri", "file:/user/hive/warehouse/")
	viper.SetDefault("server.endpoint_url", "http://127.0.0.1:8000")
	viper.SetDefault("server.retries.enabled", true)
	viper.SetDefault("server.retries.max_attempts", defaultMaxAttempts)
	viper.SetDefault("network.http2.enabled", defaultHTTP2Enabled)
	viper.SetDefault("server.retries.max_wait_interval", defaultMaxRetryInterval)
	viper.SetDefault("server.retries.min_wait_interval", defaultMinRetryInterval)
	viper.SetDefault("experimental.local.posix_permissions.enabled", false)
	viper.SetDefault("local.skip_non_regular_files", false)
	cfgErr = viper.ReadInConfig()
}
