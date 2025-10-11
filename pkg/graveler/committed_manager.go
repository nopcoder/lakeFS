package graveler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/treeverse/lakefs/pkg/logging"
)

type committedManager struct {
	metaRangeManagers map[StorageID]MetaRangeManager
	RangeManagers     map[StorageID]RangeManager
	params            *Params
}

func NewCommittedManager(m map[StorageID]MetaRangeManager, r map[StorageID]RangeManager, p Params) CommittedManager {
	return &committedManager{
		metaRangeManagers: m,
		RangeManagers:     r,
		params:            &p,
	}
}

func (c *committedManager) Exists(ctx context.Context, storageID StorageID, ns StorageNamespace, id MetaRangeID) (bool, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return false, err
	}
	return metaRangeManager.Exists(ctx, ns, id)
}

func (c *committedManager) Get(ctx context.Context, storageID StorageID, ns StorageNamespace, rangeID MetaRangeID, key Key) (*Value, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return nil, err
	}
	it, err := metaRangeManager.NewMetaRangeIterator(ctx, ns, rangeID)
	if err != nil {
		return nil, err
	}
	valIt := NewValueIterator(it)
	defer valIt.Close()
	valIt.SeekGE(key)
	// return the next value
	if !valIt.Next() {
		// error or not found
		if err := valIt.Err(); err != nil {
			return nil, err
		}
		return nil, ErrNotFound
	}
	// compare the key we found
	rec := valIt.Value()
	if !bytes.Equal(rec.Key, key) {
		return nil, ErrNotFound
	}
	return rec.Value, nil
}

func (c *committedManager) List(ctx context.Context, storageID StorageID, ns StorageNamespace, rangeID MetaRangeID) (ValueIterator, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return nil, err
	}
	it, err := metaRangeManager.NewMetaRangeIterator(ctx, ns, rangeID)
	if err != nil {
		return nil, err
	}
	return NewValueIterator(it), nil
}

func (c *committedManager) WriteRange(ctx context.Context, storageID StorageID, ns StorageNamespace, it ValueIterator) (*RangeInfo, error) {
	rangeManager, err := c.getRangeManager(storageID)
	if err != nil {
		return nil, err
	}
	writer, err := rangeManager.GetWriter(ctx, Namespace(ns), nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating range writer: %w", err)
	}
	writer.SetMetadata(MetadataTypeKey, MetadataRangesType)

	defer func() {
		if err := writer.Abort(); err != nil {
			logging.FromContext(ctx).WithError(err).Error("Aborting write to range")
		}
	}()

	for it.Next() {
		record := it.Value()
		// skip nil value (kv can hold value nil) and tombstones
		if record == nil || record.Value == nil {
			continue
		}
		v, err := MarshalValue(record.Value)
		if err != nil {
			return nil, err
		}

		if err := writer.WriteRecord(Record{Key: Key(record.Key), Value: v}); err != nil {
			return nil, fmt.Errorf("writing record: %w", err)
		}
		if writer.ShouldBreakAtKey(record.Key, c.params) {
			break
		}
	}
	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("getting value from iterator: %w", err)
	}

	info, err := writer.Close()
	if err != nil {
		return nil, fmt.Errorf("closing writer: %w", err)
	}

	return &RangeInfo{
		ID:                      RangeID(info.RangeID),
		MinKey:                  Key(info.First),
		MaxKey:                  Key(info.Last),
		Count:                   info.Count,
		EstimatedRangeSizeBytes: info.EstimatedRangeSizeBytes,
	}, nil
}

func (c *committedManager) WriteMetaRange(ctx context.Context, storageID StorageID, ns StorageNamespace, ranges []*RangeInfo) (*MetaRangeInfo, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return nil, err
	}
	writer := metaRangeManager.NewWriter(ctx, ns, nil)
	defer func() {
		if err := writer.Abort(); err != nil {
			logging.FromContext(ctx).WithError(err).Error("Aborting write to meta range")
		}
	}()

	sort.Slice(ranges, func(i, j int) bool {
		return bytes.Compare(ranges[i].MinKey, ranges[j].MinKey) < 0
	})

	for _, r := range ranges {
		if err := writer.WriteRange(Range{
			ID:            ID(r.ID),
			MinKey:        Key(r.MinKey),
			MaxKey:        Key(r.MaxKey),
			EstimatedSize: r.EstimatedRangeSizeBytes,
			Count:         int64(r.Count),
			Tombstone:     false,
		}); err != nil {
			logging.FromContext(ctx).WithError(err).Error("Aborting writing range to meta range")
			return nil, fmt.Errorf("writing range: %w", err)
		}
	}

	id, err := writer.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("closing metarange: %w", err)
	}

	return &MetaRangeInfo{
		ID: *id,
	}, nil
}

func (c *committedManager) WriteMetaRangeByIterator(ctx context.Context, storageID StorageID, ns StorageNamespace, it ValueIterator, metadata Metadata) (*MetaRangeID, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return nil, err
	}
	writer := metaRangeManager.NewWriter(ctx, ns, metadata)
	defer func() {
		if err := writer.Abort(); err != nil {
			logging.FromContext(ctx).WithError(err).Error("Aborting write to meta range")
		}
	}()

	for it.Next() {
		if err := writer.WriteRecord(*it.Value()); err != nil {
			return nil, fmt.Errorf("writing record: %w", err)
		}
	}
	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("getting value from iterator: %w", err)
	}
	id, err := writer.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("closing writer: %w", err)
	}

	return id, nil
}

func (c *committedManager) Diff(ctx context.Context, storageID StorageID, ns StorageNamespace, left, right MetaRangeID) (DiffIterator, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return nil, err
	}
	leftIt, err := metaRangeManager.NewMetaRangeIterator(ctx, ns, left)
	if err != nil {
		return nil, err
	}
	rightIt, err := metaRangeManager.NewMetaRangeIterator(ctx, ns, right)
	if err != nil {
		return nil, err
	}
	return NewDiffValueIterator(ctx, leftIt, rightIt), nil
}

func (c *committedManager) Import(ctx context.Context, storageID StorageID, ns StorageNamespace, destination, source MetaRangeID, prefixes []Prefix, _ ...SetOptionsFunc) (MetaRangeID, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return "", err
	}
	destIt, err := metaRangeManager.NewMetaRangeIterator(ctx, ns, destination)
	if err != nil {
		return "", fmt.Errorf("get destination iterator: %w", err)
	}
	destIt = NewSkipPrefixIterator(prefixes, destIt)
	defer destIt.Close()
	mctx := mergeContext{
		destIt:        destIt,
		strategy:      MergeStrategyNone,
		storageID:     storageID,
		ns:            ns,
		destinationID: destination,
		sourceID:      source,
		baseID:        "",
	}
	return c.merge(ctx, mctx)
}

func (c *committedManager) Merge(ctx context.Context, storageID StorageID, ns StorageNamespace, destination, source, base MetaRangeID, strategy MergeStrategy, opts ...SetOptionsFunc) (MetaRangeID, error) {
	options := NewSetOptions(opts)

	if source == base && !options.AllowEmpty && !options.Force {
		// no changes on source
		return "", ErrNoChanges
	}

	if destination == base {
		// changes introduced only on source
		return source, nil
	}

	mctx := mergeContext{
		strategy:      strategy,
		storageID:     storageID,
		ns:            ns,
		destinationID: destination,
		sourceID:      source,
		baseID:        base,
	}
	return c.merge(ctx, mctx)
}

type mergeContext struct {
	destIt        Iterator
	srcIt         Iterator
	baseIt        Iterator
	strategy      MergeStrategy
	storageID     StorageID
	ns            StorageNamespace
	destinationID MetaRangeID
	sourceID      MetaRangeID
	baseID        MetaRangeID
}

func (c *committedManager) merge(ctx context.Context, mctx mergeContext) (MetaRangeID, error) {
	metaRangeManager, err := c.getMetaRangeManager(mctx.storageID)
	if err != nil {
		return "", err
	}
	baseIt := mctx.baseIt
	if baseIt == nil {
		baseIt, err = metaRangeManager.NewMetaRangeIterator(ctx, mctx.ns, mctx.baseID)
		if err != nil {
			return "", fmt.Errorf("get base iterator: %w", err)
		}
		defer baseIt.Close()
	}

	destIt := mctx.destIt
	if destIt == nil {
		destIt, err = metaRangeManager.NewMetaRangeIterator(ctx, mctx.ns, mctx.destinationID)
		if err != nil {
			return "", fmt.Errorf("get destination iterator: %w", err)
		}
		defer destIt.Close()
	}

	srcIt := mctx.srcIt
	if srcIt == nil {
		srcIt, err = metaRangeManager.NewMetaRangeIterator(ctx, mctx.ns, mctx.sourceID)
		if err != nil {
			return "", fmt.Errorf("get source iterator: %w", err)
		}
		defer srcIt.Close()
	}

	mwWriter := metaRangeManager.NewWriter(ctx, mctx.ns, nil)
	defer func() {
		err = mwWriter.Abort()
		if err != nil {
			logging.FromContext(ctx).WithError(err).Error("Abort failed after Merge")
		}
	}()

	err = Merge(ctx, mwWriter, baseIt, srcIt, destIt, mctx.strategy)
	if err != nil {
		if !errors.Is(err, ErrUserVisible) {
			err = fmt.Errorf("merge ns=%s id=%s: %w", mctx.ns, mctx.destinationID, err)
		}
		return "", err
	}
	newID, err := mwWriter.Close(ctx)
	if newID == nil {
		return "", fmt.Errorf("close writer ns=%s id=%s: %w", mctx.ns, mctx.destinationID, err)
	}
	return *newID, err
}

func (c *committedManager) Commit(ctx context.Context, storageID StorageID, ns StorageNamespace, baseMetaRangeID MetaRangeID, changes ValueIterator, allowEmpty bool, _ ...SetOptionsFunc) (MetaRangeID, DiffSummary, error) {
	summary := DiffSummary{
		Count: map[DiffType]int{},
	}
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return "", summary, err
	}
	mwWriter := metaRangeManager.NewWriter(ctx, ns, nil)
	defer func() {
		err := mwWriter.Abort()
		if err != nil {
			logging.FromContext(ctx).WithError(err).Error("Abort failed after Commit")
		}
	}()
	metaRangeIterator, err := metaRangeManager.NewMetaRangeIterator(ctx, ns, baseMetaRangeID)
	if err != nil {
		return "", summary, fmt.Errorf("get metarange ns=%s id=%s: %w", ns, baseMetaRangeID, err)
	}
	defer metaRangeIterator.Close()
	summary, err = Commit(ctx, mwWriter, metaRangeIterator, changes, &CommitOptions{AllowEmpty: allowEmpty})
	if err != nil {
		if !errors.Is(err, ErrUserVisible) {
			err = fmt.Errorf("commit ns=%s id=%s: %w", ns, baseMetaRangeID, err)
		}
		return "", summary, err
	}
	newID, err := mwWriter.Close(ctx)
	if newID == nil {
		return "", summary, fmt.Errorf("close writer ns=%s metarange id=%s: %w", ns, baseMetaRangeID, err)
	}
	return *newID, summary, err
}

func (c *committedManager) Compare(ctx context.Context, storageID StorageID, ns StorageNamespace, destination, source, base MetaRangeID) (DiffIterator, error) {
	diffIt, err := c.Diff(ctx, storageID, ns, destination, source)
	if err != nil {
		return nil, fmt.Errorf("diff: %w", err)
	}
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return nil, err
	}
	baseIt, err := metaRangeManager.NewMetaRangeIterator(ctx, ns, base)
	if err != nil {
		diffIt.Close()
		return nil, fmt.Errorf("get base iterator: %w", err)
	}
	return NewCompareValueIterator(ctx, NewDiffIteratorWrapper(diffIt), baseIt), nil
}

func (c *committedManager) GetMetaRange(ctx context.Context, storageID StorageID, id MetaRangeID) (MetaRangeAddress, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return "", err
	}
	uri, err := metaRangeManager.GetMetaRangeURI(ctx, id)
	return MetaRangeAddress(uri), err
}

func (c *committedManager) GetRange(ctx context.Context, storageID StorageID, id RangeID) (RangeAddress, error) {
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return "", err
	}
	uri, err := metaRangeManager.GetRangeURI(ctx, id)
	return RangeAddress(uri), err
}

func (c *committedManager) GetRangeIDByKey(ctx context.Context, storageID StorageID, ns StorageNamespace, id MetaRangeID, key Key) (RangeID, error) {
	if id == "" {
		return "", ErrNotFound
	}
	metaRangeManager, err := c.getMetaRangeManager(storageID)
	if err != nil {
		return "", err
	}
	r, err := metaRangeManager.GetRangeByKey(ctx, ns, id, key)
	if err != nil {
		return "", fmt.Errorf("get range for key: %w", err)
	}
	return RangeID(r.ID), nil
}

func (c *committedManager) getRangeManager(id StorageID) (RangeManager, error) {
	rm, exists := c.RangeManagers[id]
	if exists {
		return rm, nil
	} else {
		return nil, fmt.Errorf("RangeManager not found for storage ID %s: %w", id, ErrInvalidStorageID)
	}
}

func (c *committedManager) getMetaRangeManager(id StorageID) (MetaRangeManager, error) {
	rm, exists := c.metaRangeManagers[id]
	if exists {
		return rm, nil
	} else {
		return nil, fmt.Errorf("MetaRangeManager not found for storage ID %s: %w", id, ErrInvalidStorageID)
	}
}
