package graveler_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treeverse/lakefs/pkg/graveler"
)

// TestNewIteratorMethods verifies that the new All() methods exist and can be called
func TestNewIteratorMethods(t *testing.T) {
	t.Run("EmptyValueIterator_All", func(t *testing.T) {
		iter := graveler.NewEmptyValueIterator()
		
		// Verify All() method exists and returns iter.Seq
		seq := iter.All()
		require.NotNil(t, seq)
		
		// Test that range-over-function works
		count := 0
		for _ = range seq {
			count++
		}
		require.Equal(t, 0, count, "empty iterator should yield no values")
	})

	t.Run("CombinedIterator_All", func(t *testing.T) {
		// Create empty iterators for testing
		iter1 := graveler.NewEmptyValueIterator()
		iter2 := graveler.NewEmptyValueIterator()
		combined := graveler.NewCombinedIterator(iter1, iter2)
		
		// Verify All() method exists
		seq := combined.All()
		require.NotNil(t, seq)
		
		// Test that range-over-function works
		count := 0
		for _ = range seq {
			count++
		}
		require.Equal(t, 0, count, "combined empty iterators should yield no values")
	})

	t.Run("FilterTombstoneIterator_All", func(t *testing.T) {
		// Create empty iterator for testing
		baseIter := graveler.NewEmptyValueIterator()
		filtered := graveler.NewFilterTombstoneIterator(baseIter)
		
		// Verify All() method exists
		seq := filtered.All()
		require.NotNil(t, seq)
		
		// Test that range-over-function works
		count := 0
		for _ = range seq {
			count++
		}
		require.Equal(t, 0, count, "filtered empty iterator should yield no values")
	})
}

// TestAdapterFunctions verifies that the adapter functions exist and work
func TestAdapterFunctions(t *testing.T) {
	t.Run("IteratorToSeq", func(t *testing.T) {
		iter := graveler.NewEmptyValueIterator()
		seq := graveler.IteratorToSeq(iter)
		require.NotNil(t, seq)
		
		// Verify it works in range loop
		for _ = range seq {
			t.Fatal("empty iterator should not yield any values")
		}
	})
}