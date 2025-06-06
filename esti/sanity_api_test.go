package esti

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treeverse/lakefs/pkg/api/apigen"
	"github.com/treeverse/lakefs/pkg/api/apiutil"
)

func TestSanityAPI(t *testing.T) {
	ctx, log, repo := setupTest(t)

	log.Debug("list entries")
	entries := ListRepositoryObjects(ctx, t, repo, mainBranch, client)
	require.Len(t, entries, 0, "expected no entries")

	log.Debug("upload some files")
	const numOfFiles = 5
	paths := make([]string, numOfFiles)
	contents := make([]string, numOfFiles)
	for i := 0; i < numOfFiles; i++ {
		paths[i] = fmt.Sprintf("file%d", i)
		_, contents[i] = UploadFileRandomData(ctx, t, repo, mainBranch, paths[i], nil)
	}

	log.Debug("verify upload content")
	for i, p := range paths {
		resp, err := client.GetObjectWithResponse(ctx, repo, mainBranch, &apigen.GetObjectParams{Path: p})
		require.NoError(t, err, "get object for", p)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		content := string(resp.Body)
		require.Equal(t, contents[i], content, "content should be the same", p)
	}

	log.Debug("list uncommitted files")
	entries = ListRepositoryObjects(ctx, t, repo, mainBranch, client)
	require.Len(t, entries, numOfFiles, "repository should have files")

	log.Debug("commit changes")
	commitResp, err := client.CommitWithResponse(ctx, repo, mainBranch, &apigen.CommitParams{}, apigen.CommitJSONRequestBody{
		Message: "first commit",
	})
	require.NoError(t, err, "initial commit")
	require.Equal(t, http.StatusCreated, commitResp.StatusCode())

	log.Debug("list files on main")
	entries = ListRepositoryObjects(ctx, t, repo, mainBranch, client)
	require.Len(t, entries, numOfFiles, "repository should have files")

	log.Debug("create 'branch1' based on 'main'")
	createBranchResp, err := client.CreateBranchWithResponse(ctx, repo, apigen.CreateBranchJSONRequestBody{
		Name:   "branch1",
		Source: mainBranch,
	})
	require.NoError(t, err, "failed to create branch1 from main")
	require.Equal(t, http.StatusCreated, createBranchResp.StatusCode())
	branchRef := string(createBranchResp.Body)
	require.NotEmpty(t, branchRef, "reference to new branch")

	log.Debug("list branches")
	branchesResp, err := client.ListBranchesWithResponse(ctx, repo, &apigen.ListBranchesParams{})
	require.NoError(t, err, "list branches")
	require.Equal(t, http.StatusOK, branchesResp.StatusCode())

	payload := branchesResp.JSON200
	var branches []string
	for _, ref := range payload.Results {
		branch := ref.Id
		commitID := ref.CommitId
		require.NotEmpty(t, commitID, "branch should have commit ID")
		require.NotEqual(t, branch, commitID, "commit ID should not be the branch name")
		// collect the branch names
		branches = append(branches, branch)
	}
	require.ElementsMatch(t, branches, []string{mainBranch, "branch1"},
		"match existing branches")

	log.Debug("branch1 - change file0")
	_, _ = UploadFileRandomData(ctx, t, repo, "branch1", "file0", nil)

	log.Debug("branch1 - delete file1")
	deleteResp, err := client.DeleteObjectWithResponse(ctx, repo, "branch1", &apigen.DeleteObjectParams{Path: "file1"})
	require.NoError(t, err, "delete object")
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode())

	log.Debug("branch1 - add fileX")
	_, _ = UploadFileRandomData(ctx, t, repo, "branch1", "fileX", nil)

	log.Debug("main - list files")
	mainObjects := ListRepositoryObjects(ctx, t, repo, "main", client)
	mainPaths := make([]string, len(mainObjects))
	for i, obj := range mainObjects {
		mainPaths[i] = obj.Path
	}
	require.EqualValues(t, mainPaths, paths)

	log.Debug("branch1 - list objects")
	branch1Objects := ListRepositoryObjects(ctx, t, repo, "branch1", client)
	for i := range branch1Objects {
		mainPaths[i] = branch1Objects[i].Path
	}
	pathsBranch1 := make([]string, len(paths))
	copy(pathsBranch1, paths)
	pathsBranch1 = append(append(paths[:1], paths[2:]...), "fileX")
	require.EqualValues(t, pathsBranch1, mainPaths)

	log.Debug("branch1 - diff changes with main")
	diffResp, err := client.DiffRefsWithResponse(ctx, repo, mainBranch, "branch1", &apigen.DiffRefsParams{})
	require.NoError(t, err, "diff between branch1 and main")
	require.Equal(t, http.StatusOK, diffResp.StatusCode())
	require.Len(t, diffResp.JSON200.Results, 0, "no changes should be found as we didn't commit anything")

	log.Debug("branch1 - commit changes")
	commitResp, err = client.CommitWithResponse(ctx, repo, "branch1", &apigen.CommitParams{}, apigen.CommitJSONRequestBody{
		Message: "3 changes",
	})
	require.NoError(t, err, "commit 3 changes")
	require.Equal(t, http.StatusCreated, commitResp.StatusCode())

	log.Debug("branch1 - diff changes with main")
	diffResp, err = client.DiffRefsWithResponse(ctx, repo, mainBranch, "branch1", &apigen.DiffRefsParams{
		Amount: apiutil.Ptr(apigen.PaginationAmount(-1)),
	})
	require.NoError(t, err, "diff between branch1 and main")
	require.Equal(t, http.StatusOK, diffResp.StatusCode())
	size := int64(randomDataContentLength)
	require.ElementsMatch(t, diffResp.JSON200.Results, []apigen.Diff{
		{Path: "file0", PathType: "object", Type: "changed", SizeBytes: &size},
		{Path: "file1", PathType: "object", Type: "removed", SizeBytes: &size},
		{Path: "fileX", PathType: "object", Type: "added", SizeBytes: &size},
	})

	log.Debug("branch1 - merge changes to main")
	mergeResp, err := client.MergeIntoBranchWithResponse(ctx, repo, "branch1", mainBranch, apigen.MergeIntoBranchJSONRequestBody{})
	require.NoError(t, err, "merge branch1 to main")
	require.Equal(t, http.StatusOK, mergeResp.StatusCode())
	require.NotEmpty(t, mergeResp.JSON200.Reference, "merge should return a commit reference")

	log.Debug("branch1 - diff after merge")
	diffResp, err = client.DiffRefsWithResponse(ctx, repo, mainBranch, "branch1", &apigen.DiffRefsParams{})
	require.NoError(t, err, "diff between branch1 and main")
	require.Equal(t, http.StatusOK, diffResp.StatusCode())
	require.Len(t, diffResp.JSON200.Results, 0, "no diff between branch1 and main")

	log.Debug("main - diff with branch1")
	diffResp, err = client.DiffRefsWithResponse(ctx, repo, "branch1", mainBranch, &apigen.DiffRefsParams{})
	require.NoError(t, err, "diff between main and branch1")
	require.Equal(t, http.StatusOK, diffResp.StatusCode())
	require.Len(t, diffResp.JSON200.Results, 0, "no diff between main and branch1")

	log.Debug("delete test repository")
	deleteRepoResp, err := client.DeleteRepositoryWithResponse(ctx, repo, &apigen.DeleteRepositoryParams{})
	require.NoError(t, err, "failed to delete repository")
	require.Equal(t, http.StatusNoContent, deleteRepoResp.StatusCode())
}
