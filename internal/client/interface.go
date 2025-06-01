package client

import (
	"context"
	// Use basic types or define simple structs here, avoid direct apigen imports
)

// RepositoryInfo holds basic information about a repository.
type RepositoryInfo struct {
	ID               string
	DefaultBranch    string
	StorageNamespace string
	CreationDate     int64 // Unix timestamp
	// Add more fields as needed, e.g., from apigen.Repository
}

// LctlAPIClient defines the interface for lctl's interactions with the lakeFS API.
type LctlAPIClient interface {
	ListRepositories(ctx context.Context, after string, amount int) ([]RepositoryInfo, string, error) // Returns repos, next_after_token, error
	// Add other methods that lctl commands will need, e.g.:
	// GetRepository(ctx context.Context, repositoryID string) (RepositoryInfo, error)
	// CreateBranch(ctx context.Context, repoID string, branchName string, sourceRef string) (string, error) // Returns commit_id, error
	// ListObjects(ctx context.Context, repoID string, refID string, path string, recursive bool, after string, amount int) ([]ObjectStatsInfo, string, error)
	// Commit(ctx context.Context, repoID string, branchID string, message string, metadata map[string]string) (string, error) // Returns commit_id, error
}
