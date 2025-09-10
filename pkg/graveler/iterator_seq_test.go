package graveler_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treeverse/lakefs/pkg/graveler"
	"github.com/treeverse/lakefs/pkg/graveler/testutil"
)

func TestIteratorToSeq_ValueIterator(t *testing.T) {
	// Create test data
	testData := []graveler.ValueRecord{
		{
			Key: []byte("key1"),
			Value: &graveler.Value{
				Identity: []byte("id1"),
				Data:     []byte("data1"),
			},
		},
		{
			Key: []byte("key2"),
			Value: &graveler.Value{
				Identity: []byte("id2"),
				Data:     []byte("data2"),
			},
		},
		{
			Key: []byte("key3"),
			Value: &graveler.Value{
				Identity: []byte("id3"),
				Data:     []byte("data3"),
			},
		},
	}

	// Create iterator
	iter := testutil.NewValueIteratorFake(testData)

	// Convert to Go 1.23 iterator and collect values
	var collected []*graveler.ValueRecord
	for value := range graveler.IteratorToSeq(iter) {
		collected = append(collected, value)
	}

	// Verify results
	require.Len(t, collected, 3)
	require.Equal(t, []byte("key1"), collected[0].Key)
	require.Equal(t, []byte("key2"), collected[1].Key) 
	require.Equal(t, []byte("key3"), collected[2].Key)
	require.Equal(t, []byte("data1"), collected[0].Value.Data)
	require.Equal(t, []byte("data2"), collected[1].Value.Data)
	require.Equal(t, []byte("data3"), collected[2].Value.Data)
}

func TestIteratorToSeq_EmptyIterator(t *testing.T) {
	// Create empty iterator
	iter := testutil.NewValueIteratorFake([]graveler.ValueRecord{})

	// Convert to Go 1.23 iterator and collect values
	var collected []*graveler.ValueRecord
	for value := range graveler.IteratorToSeq(iter) {
		collected = append(collected, value)
	}

	// Verify no values collected
	require.Len(t, collected, 0)
}

func TestIteratorToSeq_EarlyExit(t *testing.T) {
	// Create test data with more items than we'll consume
	testData := []graveler.ValueRecord{
		{Key: []byte("key1"), Value: &graveler.Value{Identity: []byte("id1")}},
		{Key: []byte("key2"), Value: &graveler.Value{Identity: []byte("id2")}},
		{Key: []byte("key3"), Value: &graveler.Value{Identity: []byte("id3")}},
		{Key: []byte("key4"), Value: &graveler.Value{Identity: []byte("id4")}},
	}

	iter := testutil.NewValueIteratorFake(testData)

	// Convert to Go 1.23 iterator and exit early
	var collected []*graveler.ValueRecord
	for value := range graveler.IteratorToSeq(iter) {
		collected = append(collected, value)
		if len(collected) == 2 {
			break // Exit early after 2 items
		}
	}

	// Verify only 2 values collected
	require.Len(t, collected, 2)
	require.Equal(t, []byte("key1"), collected[0].Key)
	require.Equal(t, []byte("key2"), collected[1].Key)
}

func TestCombinedIterator_All(t *testing.T) {
	// Create test data for two iterators
	dataA := []graveler.ValueRecord{
		{Key: []byte("a1"), Value: &graveler.Value{Identity: []byte("idA1")}},
		{Key: []byte("a3"), Value: &graveler.Value{Identity: []byte("idA3")}},
	}
	dataB := []graveler.ValueRecord{
		{Key: []byte("a2"), Value: &graveler.Value{Identity: []byte("idB2")}},
		{Key: []byte("a4"), Value: &graveler.Value{Identity: []byte("idB4")}},
	}

	iterA := testutil.NewValueIteratorFake(dataA)
	iterB := testutil.NewValueIteratorFake(dataB)
	combined := graveler.NewCombinedIterator(iterA, iterB)

	// Test the All() method with range-over-function
	var collected []*graveler.ValueRecord
	for value := range combined.All() {
		collected = append(collected, value)
	}

	// Verify combined results are in sorted order
	require.Len(t, collected, 4)
	require.Equal(t, []byte("a1"), collected[0].Key)
	require.Equal(t, []byte("a2"), collected[1].Key)
	require.Equal(t, []byte("a3"), collected[2].Key)
	require.Equal(t, []byte("a4"), collected[3].Key)
}

func TestFilterTombstoneIterator_All(t *testing.T) {
	// Create test data with some tombstones
	testData := []graveler.ValueRecord{
		{Key: []byte("key1"), Value: &graveler.Value{Identity: []byte("id1"), Data: []byte("data1")}},
		{Key: []byte("key2"), Value: nil}, // Tombstone
		{Key: []byte("key3"), Value: &graveler.Value{Identity: []byte("id3"), Data: []byte("data3")}},
		{Key: []byte("key4"), Value: nil}, // Tombstone
	}

	baseIter := testutil.NewValueIteratorFake(testData)
	filterIter := graveler.NewFilterTombstoneIterator(baseIter)

	// Test the All() method - should filter out tombstones
	var collected []*graveler.ValueRecord
	for value := range filterIter.All() {
		collected = append(collected, value)
	}

	// Verify only non-tombstone values are collected
	require.Len(t, collected, 2)
	require.Equal(t, []byte("key1"), collected[0].Key)
	require.Equal(t, []byte("key3"), collected[1].Key)
	require.NotNil(t, collected[0].Value)
	require.NotNil(t, collected[1].Value)
}

func TestEmptyValueIterator_All(t *testing.T) {
	iter := graveler.NewEmptyValueIterator()

	// Test the All() method on empty iterator
	var collected []*graveler.ValueRecord
	for value := range iter.All() {
		collected = append(collected, value)
	}

	// Verify no values collected
	require.Len(t, collected, 0)
}

func TestDiffIteratorToSeq(t *testing.T) {
	// Create test diff data
	testDiffs := []graveler.Diff{
		{
			Type: graveler.DiffTypeAdded,
			Key:  []byte("added_key"),
			Value: &graveler.Value{
				Identity: []byte("added_id"),
				Data:     []byte("added_data"),
			},
		},
		{
			Type: graveler.DiffTypeRemoved,
			Key:  []byte("removed_key"),
			Value: &graveler.Value{
				Identity: []byte("removed_id"),
				Data:     []byte("removed_data"),
			},
			LeftIdentity: []byte("left_removed_id"),
		},
		{
			Type: graveler.DiffTypeChanged,
			Key:  []byte("changed_key"),
			Value: &graveler.Value{
				Identity: []byte("changed_id"),
				Data:     []byte("changed_data"),
			},
			LeftIdentity: []byte("left_changed_id"),
		},
	}

	iter := testutil.NewFakeDiffIterator(testDiffs)

	// Convert to Go 1.23 iterator and collect diffs
	var collected []*graveler.Diff
	for diff := range graveler.DiffIteratorToSeq(iter) {
		collected = append(collected, diff)
	}

	// Verify results
	require.Len(t, collected, 3)
	require.Equal(t, graveler.DiffTypeAdded, collected[0].Type)
	require.Equal(t, graveler.DiffTypeRemoved, collected[1].Type)
	require.Equal(t, graveler.DiffTypeChanged, collected[2].Type)
	require.Equal(t, []byte("added_key"), collected[0].Key)
	require.Equal(t, []byte("removed_key"), collected[1].Key)
	require.Equal(t, []byte("changed_key"), collected[2].Key)
}