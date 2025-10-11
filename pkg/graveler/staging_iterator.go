package graveler

import (
	"context"

	"github.com/treeverse/lakefs/pkg/kv"
)

type StagingIterator struct {
	ctx   context.Context
	itr   *kv.PartitionIterator
	entry *ValueRecord
	err   error
}

// NewStagingIterator initiates the staging iterator with a batchSize
func NewStagingIterator(ctx context.Context, kvStore kv.Store, st StagingToken, batchSize int) *StagingIterator {
	itr := kv.NewPartitionIterator(ctx, kvStore, (&StagedEntryData{}).ProtoReflect().Type(), StagingTokenPartition(st), batchSize)
	return &StagingIterator{
		ctx: ctx,
		itr: itr,
	}
}

func (s *StagingIterator) Next() bool {
	if s.Err() != nil {
		return false
	}
	if !s.itr.Next() {
		s.entry = nil
		return false
	}
	entry := s.itr.Entry()
	if entry == nil {
		s.err = ErrInvalid
		return false
	}
	key := entry.Value.(*StagedEntryData).Key
	value := StagedEntryFromProto(entry.Value.(*StagedEntryData))
	s.entry = &ValueRecord{
		Key:   key,
		Value: value,
	}
	return true
}

func (s *StagingIterator) SeekGE(key Key) {
	s.itr.SeekGE(key)
}

func (s *StagingIterator) Value() *ValueRecord {
	if s.Err() != nil {
		return nil
	}
	// Tombstone handling
	if s.entry != nil && s.entry.Value != nil && s.entry.Identity == nil {
		s.entry.Value = nil
	}
	return s.entry
}

func (s *StagingIterator) Err() error {
	if s.err == nil {
		return s.itr.Err()
	}
	return s.err
}

func (s *StagingIterator) Close() {
	s.itr.Close()
}
