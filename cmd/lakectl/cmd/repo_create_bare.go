package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/treeverse/lakefs/pkg/api/apigen"
)

// repoCreateBareCmd represents the create repo command
var repoCreateBareCmd = &cobra.Command{
	Use:               "create-bare <repository URI> <storage namespace>",
	Short:             "Create a new repository with no initial branch or commit",
	Example:           "lakectl create-bare " + myRepoExample + " " + myBucketExample,
	Hidden:            true,
	Args:              cobra.ExactArgs(repoCreateCmdArgs),
	ValidArgsFunction: ValidArgsRepository,
	Run: func(cmd *cobra.Command, args []string) {
		clt := getClient()
		u := MustParseRepoURI("repository URI", args[0])
		fmt.Println("Repository:", u)

		defaultBranch := Must(cmd.Flags().GetString(defaultBranchFlagName))
		storageID := Must(cmd.Flags().GetString(storageIDFlagName))

		bareRepo := true

		resp, err := clt.CreateRepositoryWithResponse(cmd.Context(), &apigen.CreateRepositoryParams{
			Bare: &bareRepo,
		}, apigen.CreateRepositoryJSONRequestBody{
			DefaultBranch:    &defaultBranch,
			Name:             u.Repository,
			StorageId:        &storageID,
			StorageNamespace: args[1],
		})
		DieOnErrorOrUnexpectedStatusCode(resp, err, http.StatusCreated)
		if resp.JSON201 == nil {
			Die("Bad response from server", 1)
		}
		repo := resp.JSON201
		fmt.Printf("Repository '%s' created:\nstorage namespace: %s\ndefault branch: %s\ntimestamp: %d\n", repo.Id, repo.StorageNamespace, repo.DefaultBranch, repo.CreationDate)
	},
}

//nolint:gochecknoinits
func init() {
	repoCreateBareCmd.Flags().StringP(defaultBranchFlagName, "d", defaultBranchFlagValue, "the default branch name of this repository (will not be created)")
	withStorageID(repoCreateBareCmd)

	repoCmd.AddCommand(repoCreateBareCmd)
}
