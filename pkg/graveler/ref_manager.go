package graveler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/treeverse/lakefs/pkg/batch"
	"github.com/treeverse/lakefs/pkg/cache"
	"github.com/treeverse/lakefs/pkg/config"
	"github.com/treeverse/lakefs/pkg/distributed"
	"github.com/treeverse/lakefs/pkg/httputil"
	"github.com/treeverse/lakefs/pkg/ident"
	"github.com/treeverse/lakefs/pkg/kv"
	"github.com/treeverse/lakefs/pkg/logging"
)

const (
	// commitIDStringLength string representation length of commit ID - based on hex representation of sha256
	commitIDStringLength = 64
	// ImportExpiryTime Expiry time to remove imports from ref-store
	ImportExpiryTime = 24 * time.Hour
)

type CacheConfig struct {
	Size   int
	Expiry time.Duration
	Jitter time.Duration
}

type Manager struct {
	kvStore         kv.Store
	kvStoreLimited  kv.Store
	addressProvider ident.AddressProvider
	batchExecutor   batch.Batcher
	repoCache       cache.Cache
	commitCache     cache.Cache
	maxBatchDelay   time.Duration
	branchOwnership *distributed.MostlyCorrectOwner
	storageConfig   config.StorageConfig
}

func branchFromProto(pb *BranchData) *Branch {
	var sealedTokens []StagingToken
	for _, st := range pb.SealedTokens {
		sealedTokens = append(sealedTokens, StagingToken(st))
	}
	branch := &Branch{
		CommitID:     CommitID(pb.CommitId),
		StagingToken: StagingToken(pb.StagingToken),
		SealedTokens: sealedTokens,
		Hidden:       pb.Hidden,
	}
	return branch
}

func protoFromBranch(branchID BranchID, b *Branch) *BranchData {
	var sealedTokens []string
	for _, st := range b.SealedTokens {
		sealedTokens = append(sealedTokens, st.String())
	}
	branch := &BranchData{
		Id:           branchID.String(),
		CommitId:     b.CommitID.String(),
		StagingToken: b.StagingToken.String(),
		SealedTokens: sealedTokens,
		Hidden:       b.Hidden,
	}
	return branch
}

// BranchApproximateOwnershipParams configures mostly-correct ownership of
// branches.  Branch correctness is safe _regardless_ of the values of these
// parameters.  They exist solely to reduce expensive operations when
// multiple concurrent updates race on the same branch.  Only one update can
// win a race.  Approximately correct ownership means others will generally
// back off and let that one update proceed.
type BranchApproximateOwnershipParams struct {
	// AcquireInterval is the interval at which to attempt to acquire
	// ownership of a branch.  It is a bound on the latency of the time
	// for one worker to acquire a branch when multiple operations race
	// on that branch.  Reducing it increases read load on the branch
	// ownership record when concurrent operations occur.
	AcquireInterval time.Duration
	// RefreshInterval the interval for which to assert ownership of a
	// branch.  It is a bound on the time to perform an operation on a
	// branch IF a previous worker crashed while owning that branch.  It
	// has no effect when there are no crashes.  Reducing it increases
	// write load on the branch ownership record when concurrent
	// operations occur.
	//
	// If zero or negative, ownership will not be asserted and branch
	// operations will race.  This is safe but can be slow.
	RefreshInterval time.Duration
}

type ManagerConfig struct {
	Executor                         batch.Batcher
	KVStore                          kv.Store
	KVStoreLimited                   kv.Store
	AddressProvider                  ident.AddressProvider
	RepositoryCacheConfig            CacheConfig
	CommitCacheConfig                CacheConfig
	MaxBatchDelay                    time.Duration
	BranchApproximateOwnershipParams BranchApproximateOwnershipParams
}

func NewRefManager(cfg ManagerConfig, storageCfg config.StorageConfig) *RefManager {
	var branchOwnership *distributed.MostlyCorrectOwner
	if cfg.BranchApproximateOwnershipParams.RefreshInterval > 0 {
		log := logging.ContextUnavailable().WithField("component", "RefManager approximate branch ownership")
		branchOwnership = distributed.NewMostlyCorrectOwner(
			log,
			cfg.KVStore,
			"run-refs/approximate-branch-owner",
			cfg.BranchApproximateOwnershipParams.AcquireInterval,
			cfg.BranchApproximateOwnershipParams.RefreshInterval,
		)
		log.Info("Initialized")
	}

	return &RefManager{
		kvStore:         cfg.KVStore,
		kvStoreLimited:  cfg.KVStoreLimited,
		addressProvider: cfg.AddressProvider,
		batchExecutor:   cfg.Executor,
		repoCache:       newCache(cfg.RepositoryCacheConfig),
		commitCache:     newCache(cfg.CommitCacheConfig),
		maxBatchDelay:   cfg.MaxBatchDelay,
		branchOwnership: branchOwnership,
		storageConfig:   storageCfg,
	}
}

func (m *RefManager) getRepository(ctx context.Context, repositoryID RepositoryID) (*RepositoryRecord, error) {
	data := RepositoryData{}
	_, err := kv.GetMsg(ctx, m.kvStore, RepositoriesPartition(), []byte(RepoPath(repositoryID)), &data)
	if err != nil {
		if errors.Is(err, kv.ErrNotFound) {
			err = ErrRepositoryNotFound
		}
		return nil, err
	}
	repo := RepoFromProto(&data)
	repo.StorageID = StorageID(config.GetActualStorageID(m.storageConfig, repo.StorageID.String()))

	return repo, nil
}

func (m *RefManager) getRepositoryBatch(ctx context.Context, repositoryID RepositoryID) (*RepositoryRecord, error) {
	key := fmt.Sprintf("GetRepository:%s", repositoryID)
	repository, err := m.batchExecutor.BatchFor(ctx, key, m.maxBatchDelay, batch.ExecuterFunc(func() (interface{}, error) {
		return m.getRepository(context.Background(), repositoryID)
	}))
	if err != nil {
		return nil, err
	}
	return repository.(*RepositoryRecord), nil
}

func (m *RefManager) GetRepository(ctx context.Context, repositoryID RepositoryID) (*RepositoryRecord, error) {
	rec, err := m.repoCache.GetOrSet(repositoryID, func() (interface{}, error) {
		repo, err := m.getRepositoryBatch(ctx, repositoryID)
		if err != nil {
			return nil, err
		}

		switch repo.State {
		case RepositoryState_ACTIVE:
			return repo, nil
		case RepositoryState_IN_DELETION:
			return nil, ErrRepositoryInDeletion
		default:
			return nil, fmt.Errorf("invalid repository state (%d) rec: %w", repo.State, ErrInvalid)
		}
	})
	if err != nil {
		return nil, err
	}
	return rec.(*RepositoryRecord), nil
}

func (m *RefManager) createBareRepository(ctx context.Context, repositoryID RepositoryID, repository Repository) (*RepositoryRecord, error) {
	repoRecord := &RepositoryRecord{
		RepositoryID: repositoryID,
		Repository:   &repository,
	}
	repo := ProtoFromRepo(repoRecord)

	err := kv.SetMsgIf(ctx, m.kvStore, RepositoriesPartition(), []byte(RepoPath(repositoryID)), repo, nil)
	if err != nil {
		if errors.Is(err, kv.ErrPredicateFailed) {
			err = ErrNotUnique
		}
		return nil, err
	}

	return repoRecord, nil
}

func (m *RefManager) CreateRepository(ctx context.Context, repositoryID RepositoryID, repository Repository) (*RepositoryRecord, error) {
	firstCommit := NewCommit()
	firstCommit.Message = FirstCommitMsg
	firstCommit.Generation = 1

	repo := &RepositoryRecord{
		RepositoryID: repositoryID,
		Repository:   &repository,
	}
	// If branch creation fails - this commit will become dangling. This is a known issue that can be resolved via garbage collection
	commitID, err := m.addCommit(ctx, RepoPartition(repo), firstCommit)
	if err != nil {
		return nil, err
	}

	branch := Branch{
		CommitID:     commitID,
		StagingToken: GenerateStagingToken(repositoryID, repository.DefaultBranchID),
		SealedTokens: nil,
	}
	err = m.createBranch(ctx, RepoPartition(repo), repository.DefaultBranchID, branch)
	if err != nil {
		return nil, err
	}
	_, err = m.createBareRepository(ctx, repositoryID, repository)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (m *RefManager) CreateBareRepository(ctx context.Context, repositoryID RepositoryID, repository Repository) (*RepositoryRecord, error) {
	return m.createBareRepository(ctx, repositoryID, repository)
}

func (m *RefManager) ListRepositories(ctx context.Context) (RepositoryIterator, error) {
	return NewRepositoryIterator(ctx, m.kvStore, m.storageConfig)
}

func (m *RefManager) updateRepoState(ctx context.Context, repo *RepositoryRecord, state RepositoryState) error {
	repo.State = state
	return kv.SetMsg(ctx, m.kvStore, RepositoriesPartition(), []byte(RepoPath(repo.RepositoryID)), ProtoFromRepo(repo))
}

func (m *RefManager) deleteRepositoryBranches(ctx context.Context, repository *RepositoryRecord) error {
	itr, err := m.ListBranches(ctx, repository, ListOptions{ShowHidden: true})
	if err != nil {
		return err
	}
	defer itr.Close()
	var wg multierror.Group
	for itr.Next() {
		b := itr.Value()
		wg.Go(func() error {
			return m.DeleteBranch(ctx, repository, b.BranchID)
		})
	}
	return wg.Wait().ErrorOrNil()
}

func (m *RefManager) deleteRepositoryTags(ctx context.Context, repository *RepositoryRecord) error {
	itr, err := m.ListTags(ctx, repository)
	if err != nil {
		return err
	}
	defer itr.Close()
	var wg multierror.Group
	for itr.Next() {
		tag := itr.Value()
		wg.Go(func() error {
			return m.DeleteTag(ctx, repository, tag.TagID)
		})
	}
	return wg.Wait().ErrorOrNil()
}

func (m *RefManager) deleteRepositoryCommits(ctx context.Context, repository *RepositoryRecord) error {
	itr, err := m.ListCommits(ctx, repository)
	if err != nil {
		return err
	}
	defer itr.Close()
	var wg multierror.Group
	for itr.Next() {
		commit := itr.Value()
		wg.Go(func() error {
			return m.RemoveCommit(ctx, repository, commit.CommitID)
		})
	}
	return wg.Wait().ErrorOrNil()
}

func (m *RefManager) deleteRepositoryMetadata(ctx context.Context, repository *RepositoryRecord) error {
	return m.kvStore.Delete(ctx, []byte(RepoPartition(repository)), []byte(RepoMetadataPath()))
}

func (m *RefManager) deleteRepository(ctx context.Context, repo *RepositoryRecord) error {
	// ctx := context.Background() TODO (niro): When running this async create a new context and remove ctx from signature
	var wg multierror.Group
	wg.Go(func() error {
		return m.deleteRepositoryBranches(ctx, repo)
	})
	wg.Go(func() error {
		return m.deleteRepositoryTags(ctx, repo)
	})
	wg.Go(func() error {
		return m.deleteRepositoryCommits(ctx, repo)
	})
	wg.Go(func() error {
		return m.deleteRepositoryMetadata(ctx, repo)
	})

	if err := wg.Wait().ErrorOrNil(); err != nil {
		return err
	}

	// Finally delete the repository record itself
	return m.kvStore.Delete(ctx, []byte(RepositoriesPartition()), []byte(RepoPath(repo.RepositoryID)))
}

func (m *RefManager) DeleteRepository(ctx context.Context, repositoryID RepositoryID, opts ...SetOptionsFunc) error {
	repo, err := m.getRepository(ctx, repositoryID)
	if err != nil {
		return err
	}

	options := &SetOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if repo.ReadOnly && !options.Force {
		return ErrReadOnlyRepository
	}

	// Set repository state to deleted and then perform background delete.
	if repo.State != RepositoryState_IN_DELETION {
		err = m.updateRepoState(ctx, repo, RepositoryState_IN_DELETION)
		if err != nil {
			return err
		}
	}

	// TODO(niro): This should be a background delete process
	return m.deleteRepository(ctx, repo)
}

func (m *RefManager) getRepositoryMetadata(ctx context.Context, repo *RepositoryRecord) (RepositoryMetadata, kv.Predicate, error) {
	data := RepoMetadata{}
	pred, err := kv.GetMsg(ctx, m.kvStore, RepoPartition(repo), []byte(RepoMetadataPath()), &data)
	if err != nil {
		return nil, nil, err
	}
	return RepoMetadataFromProto(&data), pred, nil
}

func (m *RefManager) GetRepositoryMetadata(ctx context.Context, repositoryID RepositoryID) (RepositoryMetadata, error) {
	repo, err := m.getRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	metadata, _, err := m.getRepositoryMetadata(ctx, repo)
	if errors.Is(err, kv.ErrNotFound) { // Return nil map if not exists
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (m *RefManager) SetRepositoryMetadata(ctx context.Context, repo *RepositoryRecord, updateFunc RepoMetadataUpdateFunc) error {
	metadata, pred, err := m.getRepositoryMetadata(ctx, repo)
	if errors.Is(err, kv.ErrNotFound) { // Create new metadata map and set predicate to nil for setIf not exists
		metadata = RepositoryMetadata{}
		pred = nil
	} else if err != nil {
		return err
	}

	newMetadata, err := updateFunc(metadata)
	// return on error or nothing to update
	if err != nil || newMetadata == nil {
		return err
	}
	return kv.SetMsgIf(ctx, m.kvStore, RepoPartition(repo), []byte(RepoMetadataPath()), ProtoFromRepositoryMetadata(newMetadata), pred)
}

func (m *RefManager) ParseRef(ref Ref) (RawRef, error) {
	return ParseRef(ref)
}

func (m *RefManager) ResolveRawRef(ctx context.Context, repository *RepositoryRecord, raw RawRef) (*ResolvedRef, error) {
	return ResolveRawRef(ctx, m, m.addressProvider, repository, raw)
}

func (m *RefManager) getBranchWithPredicate(ctx context.Context, repository *RepositoryRecord, branchID BranchID) (*Branch, kv.Predicate, error) {
	key := fmt.Sprintf("GetBranch:%s:%s", repository.RepositoryID, branchID)
	type branchPred struct {
		*Branch
		kv.Predicate
	}
	result, err := m.batchExecutor.BatchFor(ctx, key, m.maxBatchDelay, batch.ExecuterFunc(func() (interface{}, error) {
		key := BranchPath(branchID)
		data := BranchData{}
		pred, err := kv.GetMsg(context.Background(), m.kvStore, RepoPartition(repository), []byte(key), &data)
		if err != nil {
			return nil, err
		}
		return &branchPred{Branch: branchFromProto(&data), Predicate: pred}, nil
	}))
	if errors.Is(err, kv.ErrNotFound) {
		err = ErrBranchNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	branchWithPred := result.(*branchPred)
	return branchWithPred.Branch, branchWithPred.Predicate, nil
}

func (m *RefManager) GetBranch(ctx context.Context, repository *RepositoryRecord, branchID BranchID) (*Branch, error) {
	branch, _, err := m.getBranchWithPredicate(ctx, repository, branchID)
	return branch, err
}

func (m *RefManager) createBranch(ctx context.Context, repositoryPartition string, branchID BranchID, branch Branch) error {
	err := kv.SetMsgIf(ctx, m.kvStore, repositoryPartition, []byte(BranchPath(branchID)), protoFromBranch(branchID, &branch), nil)
	if errors.Is(err, kv.ErrPredicateFailed) {
		err = ErrBranchExists
	}
	return err
}

func (m *RefManager) CreateBranch(ctx context.Context, repository *RepositoryRecord, branchID BranchID, branch Branch) error {
	return m.createBranch(ctx, RepoPartition(repository), branchID, branch)
}

func (m *RefManager) SetBranch(ctx context.Context, repository *RepositoryRecord, branchID BranchID, branch Branch) error {
	return kv.SetMsg(ctx, m.kvStore, RepoPartition(repository), []byte(BranchPath(branchID)), protoFromBranch(branchID, &branch))
}

func (m *RefManager) BranchUpdate(ctx context.Context, repository *RepositoryRecord, branchID BranchID, f BranchUpdateFunc) error {
	// TODO(ariels): Get request ID in a nicer way.
	requestIDPtr := httputil.RequestIDFromContext(ctx)
	// Grab ownership if configured.  Also check we actually have a
	// request-ID on the request.  (lakeFS middleware should *always*
	// place a request ID anyways.)
	if m.branchOwnership != nil && requestIDPtr != nil {
		requestID := *requestIDPtr
		release, err := m.branchOwnership.Own(ctx, requestID, string(branchID))
		if err != nil {
			logging.FromContext(ctx).
				WithError(err).
				Warn("Failed to get ownership on branch; continuing but may be slow")
		} else {
			defer release()
		}
	}
	b, pred, err := m.getBranchWithPredicate(ctx, repository, branchID)
	if err != nil {
		return err
	}

	// clone the branch information to avoid mutating the shared result returned by batch executor
	b = b.Clone()

	newBranch, err := f(b)
	// return on error or nothing to update
	if err != nil || newBranch == nil {
		return err
	}
	return kv.SetMsgIf(ctx, m.kvStore, RepoPartition(repository), []byte(BranchPath(branchID)), protoFromBranch(branchID, newBranch), pred)
}

func (m *RefManager) DeleteBranch(ctx context.Context, repository *RepositoryRecord, branchID BranchID) error {
	_, err := m.GetBranch(ctx, repository, branchID)
	if err != nil {
		return err
	}
	return m.kvStore.Delete(ctx, []byte(RepoPartition(repository)), []byte(BranchPath(branchID)))
}

func (m *RefManager) ListBranches(ctx context.Context, repository *RepositoryRecord, opts ListOptions) (BranchIterator, error) {
	return NewBranchSimpleIterator(ctx, m.kvStore, repository, opts)
}

func (m *RefManager) GCBranchIterator(ctx context.Context, repository *RepositoryRecord) (BranchIterator, error) {
	return NewBranchByCommitIterator(ctx, m.kvStore, repository, ListOptions{ShowHidden: true})
}

func (m *RefManager) GetTag(ctx context.Context, repository *RepositoryRecord, tagID TagID) (*CommitID, error) {
	key := fmt.Sprintf("GetTag:%s:%s", repository.RepositoryID, tagID)
	commitID, err := m.batchExecutor.BatchFor(ctx, key, m.maxBatchDelay, batch.ExecuterFunc(func() (interface{}, error) {
		tagKey := TagPath(tagID)
		t := TagData{}
		_, err := kv.GetMsg(context.Background(), m.kvStore, RepoPartition(repository), []byte(tagKey), &t)
		if err != nil {
			return nil, err
		}
		commitID := CommitID(t.CommitId)
		return &commitID, nil
	}))
	if errors.Is(err, kv.ErrNotFound) {
		err = ErrTagNotFound
	}
	if err != nil {
		return nil, err
	}
	return commitID.(*CommitID), nil
}

func (m *RefManager) CreateTag(ctx context.Context, repository *RepositoryRecord, tagID TagID, commitID CommitID) error {
	t := &TagData{
		Id:       tagID.String(),
		CommitId: commitID.String(),
	}
	tagKey := TagPath(tagID)
	err := kv.SetMsgIf(ctx, m.kvStore, RepoPartition(repository), []byte(tagKey), t, nil)
	if err != nil {
		if errors.Is(err, kv.ErrPredicateFailed) {
			err = ErrTagAlreadyExists
		}
		return err
	}
	return nil
}

func (m *RefManager) DeleteTag(ctx context.Context, repository *RepositoryRecord, tagID TagID) error {
	tagKey := TagPath(tagID)
	// TODO (issue 3640) align with delete tag DB - return ErrNotFound when tag does not exist
	return m.kvStore.Delete(ctx, []byte(RepoPartition(repository)), []byte(tagKey))
}

func (m *RefManager) ListTags(ctx context.Context, repository *RepositoryRecord) (TagIterator, error) {
	return NewTagIterator(ctx, m.kvStore, repository)
}

func (m *RefManager) GetCommitByPrefix(ctx context.Context, repository *RepositoryRecord, prefix CommitID) (*Commit, error) {
	// optimize by get if prefix is not a prefix, but a full length commit id
	if len(prefix) == commitIDStringLength {
		return m.GetCommit(ctx, repository, prefix)
	}
	key := fmt.Sprintf("GetCommitByPrefix:%s:%s", repository.RepositoryID, prefix)
	commit, err := m.batchExecutor.BatchFor(ctx, key, m.maxBatchDelay, batch.ExecuterFunc(func() (interface{}, error) {
		it, err := NewOrderedCommitIterator(context.Background(), m.kvStore, repository, false)
		if err != nil {
			return nil, err
		}
		defer it.Close()
		it.SeekGE(prefix)
		var commit *Commit
		for it.Next() {
			c := it.Value()
			if strings.HasPrefix(string(c.CommitID), string(prefix)) {
				if commit != nil {
					return nil, ErrCommitNotFound // more than 1 commit starts with the ID prefix
				}
				commit = c.Commit
			} else {
				break
			}
		}
		if err := it.Err(); err != nil {
			return nil, err
		}
		if commit == nil {
			return nil, ErrCommitNotFound
		}
		return commit, nil
	}))
	if err != nil {
		return nil, err
	}
	return commit.(*Commit), nil
}

func (m *RefManager) GetCommit(ctx context.Context, repository *RepositoryRecord, commitID CommitID) (*Commit, error) {
	key := fmt.Sprintf("%s:%s", repository.RepositoryID, commitID)
	v, err := m.commitCache.GetOrSet(key, func() (v interface{}, err error) {
		return m.getCommitBatch(ctx, repository, commitID)
	})
	if err != nil {
		return nil, err
	}
	return v.(*Commit), nil
}

func (m *RefManager) getCommitBatch(ctx context.Context, repository *RepositoryRecord, commitID CommitID) (*Commit, error) {
	key := fmt.Sprintf("GetCommit:%s:%s", repository.RepositoryID, commitID)
	commit, err := m.batchExecutor.BatchFor(ctx, key, m.maxBatchDelay, batch.ExecuterFunc(func() (interface{}, error) {
		return m.getCommit(context.Background(), commitID, repository)
	}))
	if err != nil {
		return nil, err
	}
	return commit.(*Commit), nil
}

func (m *RefManager) getCommit(ctx context.Context, commitID CommitID, repository *RepositoryRecord) (interface{}, error) {
	commitKey := CommitPath(commitID)
	c := CommitData{}
	_, err := kv.GetMsg(ctx, m.kvStore, RepoPartition(repository), []byte(commitKey), &c)
	if errors.Is(err, kv.ErrNotFound) {
		err = ErrCommitNotFound
	}
	if err != nil {
		return nil, err
	}
	return CommitFromProto(&c), nil
}

func (m *RefManager) addCommit(ctx context.Context, repoPartition string, commit Commit) (CommitID, error) {
	commitID := m.addressProvider.ContentAddress(commit)
	c := ProtoFromCommit(CommitID(commitID), &commit)
	commitKey := CommitPath(CommitID(commitID))
	err := kv.SetMsgIf(ctx, m.kvStore, repoPartition, []byte(commitKey), c, nil)
	// commits are written based on their content hash, if we insert the same ID again,
	// it will necessarily have the same attributes as the existing one, so if a commit already exists doesn't return an error
	if err != nil && !errors.Is(err, kv.ErrPredicateFailed) {
		return "", err
	}
	return CommitID(commitID), nil
}

func (m *RefManager) AddCommit(ctx context.Context, repository *RepositoryRecord, commit Commit) (CommitID, error) {
	return m.addCommit(ctx, RepoPartition(repository), commit)
}

func (m *RefManager) CreateCommitRecord(ctx context.Context, repository *RepositoryRecord, commitID CommitID, commit Commit) error {
	if m.addressProvider.ContentAddress(commit) != commitID.String() {
		return ErrInvalidCommitID
	}
	c := ProtoFromCommit(commitID, &commit)
	commitKey := CommitPath(commitID)
	err := kv.SetMsgIf(ctx, m.kvStore, RepoPartition(repository), []byte(commitKey), c, nil)
	if errors.Is(err, kv.ErrPredicateFailed) {
		return ErrCommitAlreadyExists
	} else if err != nil {
		return err
	}
	return nil
}

func (m *RefManager) RemoveCommit(ctx context.Context, repository *RepositoryRecord, commitID CommitID) error {
	commitKey := CommitPath(commitID)
	return m.kvStore.Delete(ctx, []byte(RepoPartition(repository)), []byte(commitKey))
}

func (m *RefManager) FindMergeBase(ctx context.Context, repository *RepositoryRecord, commitIDs ...CommitID) (*Commit, error) {
	const allowedCommitsToCompare = 2
	if len(commitIDs) != allowedCommitsToCompare {
		return nil, ErrInvalidMergeBase
	}
	return FindMergeBase(ctx, m, repository, commitIDs[0], commitIDs[1])
}

func (m *RefManager) Log(ctx context.Context, repository *RepositoryRecord, from CommitID, firstParent bool, since *time.Time) (CommitIterator, error) {
	return NewCommitIterator(ctx, &CommitIteratorConfig{
		repository:  repository,
		start:       from,
		firstParent: firstParent,
		since:       since,
		manager:     m,
	}), nil
}

func (m *RefManager) ListCommits(ctx context.Context, repository *RepositoryRecord) (CommitIterator, error) {
	return NewOrderedCommitIterator(ctx, m.kvStore, repository, false)
}

func (m *RefManager) GCCommitIterator(ctx context.Context, repository *RepositoryRecord) (CommitIterator, error) {
	return NewOrderedCommitIterator(ctx, m.kvStore, repository, true)
}

func newCache(cfg CacheConfig) cache.Cache {
	if cfg.Size == 0 {
		return cache.NoCache
	}
	return cache.NewCache(cfg.Size, cfg.Expiry, cache.NewJitterFn(cfg.Jitter))
}

func (m *RefManager) DeleteExpiredImports(ctx context.Context, repository *RepositoryRecord) error {
	expiry := time.Now().Add(-ImportExpiryTime)
	repoPartition := RepoPartition(repository)
	key := []byte(ImportsPath(""))
	options := kv.IteratorOptionsFrom([]byte(""))
	itr, err := kv.NewPrimaryIterator(ctx, m.kvStoreLimited, (&ImportStatusData{}).ProtoReflect().Type(), repoPartition, key, options)
	if err != nil {
		return fmt.Errorf("failed to get imports iterator from store: %w", err)
	}
	defer itr.Close()

	var errs multierror.Error
	for itr.Next() {
		entry := itr.Entry()
		status, ok := entry.Value.(*ImportStatusData)
		if !ok {
			return fmt.Errorf("invalid protobuf type %s: %w", entry.Value.ProtoReflect().Type().Descriptor().FullName(), ErrReadingFromStore)
		}
		if status.UpdatedAt.AsTime().Before(expiry) {
			if !status.Completed && status.Error == "" {
				logging.FromContext(ctx).WithFields(logging.Fields{"import_id": status.Id}).Warning("removing stale import")
			}
			err = m.kvStoreLimited.Delete(ctx, []byte(repoPartition), entry.Key)
			if err != nil {
				errs.Errors = append(errs.Errors, fmt.Errorf("delete failed for import ID %s: %w", status.Id, err))
			}
		}
	}
	return errs.ErrorOrNil()
}

// Pull Requests logic
// TODO (niro): In the future we would probably like to move all the PR logic into a dedicated service similar to actions.
// TODO (niro): For now we put all the logic here under a single block

const (
	// pullRequestsPrefix used for repo context listing
	pullRequestsPrefix = "pulls"
	// PullsPartitionKey used for lookup per source-dest (future)
	PullsPartitionKey = "pulls"
	reposPrefix       = "repos"
)

func PullRequestPath(pullID PullRequestID) string {
	return kv.FormatPath(pullRequestsPrefix, pullID.String())
}

func basePullsPath(repoID string) string {
	return kv.FormatPath(reposPrefix, repoID)
}

func PullBySrcDstPath(repository *RepositoryRecord, srcBranch, dstBranch string) string {
	return kv.FormatPath(basePullsPath(RepoPartition(repository)), srcBranch, dstBranch)
}

func (m *RefManager) getPullWithPredicate(ctx context.Context, repository *RepositoryRecord, pullID PullRequestID) (*PullRequest, kv.Predicate, error) {
	type pullWithPred struct {
		*PullRequestRecord
		kv.Predicate
	}
	key := fmt.Sprintf("GetPullRequest:%s:%s", repository.RepositoryID, pullID)
	result, err := m.batchExecutor.BatchFor(ctx, key, m.maxBatchDelay, batch.ExecuterFunc(func() (interface{}, error) {
		pullKey := PullRequestPath(pullID)
		data := PullRequestData{}
		pred, err := kv.GetMsg(context.Background(), m.kvStore, RepoPartition(repository), []byte(pullKey), &data)
		if err != nil {
			return nil, err
		}
		return &pullWithPred{PullRequestFromProto(&data), pred}, nil
	}))
	if errors.Is(err, kv.ErrNotFound) {
		err = ErrPullRequestNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	p := result.(*pullWithPred)
	return &p.PullRequest, p.Predicate, nil
}

func (m *RefManager) GetPullRequest(ctx context.Context, repository *RepositoryRecord, pullID PullRequestID) (*PullRequest, error) {
	pull, _, err := m.getPullWithPredicate(ctx, repository, pullID)
	return pull, err
}

func (m *RefManager) ListPullRequests(ctx context.Context, repository *RepositoryRecord) (PullsIterator, error) {
	return NewPullsIterator(ctx, m.kvStore, repository)
}

func (m *RefManager) CreatePullRequest(ctx context.Context, repository *RepositoryRecord, pullRequestID PullRequestID, pullRequest *PullRequest) error {
	// Save secondary index by source - dest. For now, we override the value. In the future we should allow only single src-dest to exist
	secondaryKey := []byte(PullBySrcDstPath(repository, pullRequest.Source, pullRequest.Destination))
	err := kv.SetMsg(ctx, m.kvStore, PullsPartitionKey, secondaryKey, &kv.SecondaryIndex{PrimaryKey: []byte(pullRequestID.String())})
	if err != nil {
		return fmt.Errorf("save secondary index by src-dest (key %s): %w", secondaryKey, err)
	}

	// Save primary
	err = kv.SetMsgIf(ctx, m.kvStore, RepoPartition(repository), []byte(PullRequestPath(pullRequestID)), ProtoFromPullRequest(pullRequestID, pullRequest), nil)
	if errors.Is(err, kv.ErrPredicateFailed) {
		err = ErrPullRequestExists
	}
	return err
}

func (m *RefManager) DeletePullRequest(ctx context.Context, repository *RepositoryRecord, pullRequestID PullRequestID) error {
	pr, _, err := m.getPullWithPredicate(ctx, repository, pullRequestID)
	if err != nil {
		if errors.Is(err, ErrPullRequestNotFound) { // Ignore if not exists
			return nil
		}
		return err
	}

	// Delete secondary key
	secondaryKey := []byte(PullBySrcDstPath(repository, pr.Source, pr.Destination))
	if err = m.kvStore.Delete(ctx, []byte(PullsPartitionKey), secondaryKey); err != nil {
		return fmt.Errorf("delete secondary index by src-dest (key %s): %w", secondaryKey, err)
	}

	// Delete primary key
	pullKey := PullRequestPath(pullRequestID)
	return m.kvStore.Delete(ctx, []byte(RepoPartition(repository)), []byte(pullKey))
}

func (m *RefManager) UpdatePullRequest(ctx context.Context, repository *RepositoryRecord, pullRequestID PullRequestID, f PullUpdateFunc) error {
	b, pred, err := m.getPullWithPredicate(ctx, repository, pullRequestID)
	if err != nil {
		return err
	}
	newPull, err := f(b)
	// return on error or nothing to update
	if err != nil || newPull == nil {
		return err
	}
	return kv.SetMsgIf(ctx, m.kvStore, RepoPartition(repository), []byte(PullRequestPath(pullRequestID)), ProtoFromPullRequest(pullRequestID, newPull), pred)
}
