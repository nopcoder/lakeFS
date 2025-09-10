package graveler_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treeverse/lakefs/pkg/graveler"
	"github.com/treeverse/lakefs/pkg/graveler/testutil"
)

// TestCompleteWorkflow demonstrates a complete workflow using both old and new iterator patterns
func TestCompleteWorkflow(t *testing.T) {
	// Create test data
	data1 := []graveler.ValueRecord{
		{Key: []byte("file1.txt"), Value: &graveler.Value{Identity: []byte("hash1"), Data: []byte("content1")}},
		{Key: []byte("file3.txt"), Value: &graveler.Value{Identity: []byte("hash3"), Data: []byte("content3")}},
		{Key: []byte("file5.txt"), Value: &graveler.Value{Identity: []byte("hash5"), Data: []byte("content5")}},
	}
	
	data2 := []graveler.ValueRecord{
		{Key: []byte("file2.txt"), Value: &graveler.Value{Identity: []byte("hash2"), Data: []byte("content2")}},
		{Key: []byte("file4.txt"), Value: &graveler.Value{Identity: []byte("hash4"), Data: []byte("content4")}},
		{Key: []byte("file6.txt"), Value: nil}, // Tombstone
	}

	t.Run("TraditionalPattern", func(t *testing.T) {
		iter1 := testutil.NewValueIteratorFake(data1)
		iter2 := testutil.NewValueIteratorFake(data2)
		combined := graveler.NewCombinedIterator(iter1, iter2)
		filtered := graveler.NewFilterTombstoneIterator(combined)

		// Traditional approach
		var files []string
		defer filtered.Close()
		for filtered.Next() {
			if err := filtered.Err(); err != nil {
				t.Fatal(err)
			}
			value := filtered.Value()
			files = append(files, string(value.Key))
		}

		// Verify results
		expected := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}
		require.Equal(t, expected, files)
	})

	t.Run("NewPatternWithRangeOverFunction", func(t *testing.T) {
		iter1 := testutil.NewValueIteratorFake(data1)
		iter2 := testutil.NewValueIteratorFake(data2)
		combined := graveler.NewCombinedIterator(iter1, iter2)
		filtered := graveler.NewFilterTombstoneIterator(combined)

		// New Go 1.23 approach
		var files []string
		for value := range filtered.All() {
			files = append(files, string(value.Key))
		}

		// Verify results
		expected := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}
		require.Equal(t, expected, files)
	})

	t.Run("FunctionalPipeline", func(t *testing.T) {
		iter1 := testutil.NewValueIteratorFake(data1)
		iter2 := testutil.NewValueIteratorFake(data2)
		
		// Create a functional pipeline
		seq := graveler.NewCombinedIterator(iter1, iter2).All()
		
		var totalSize int
		var fileCount int
		for value := range seq {
			if value.Value != nil { // Skip tombstones
				totalSize += len(value.Value.Data)
				fileCount++
			}
		}

		// Verify pipeline results
		require.Equal(t, 5, fileCount) // 5 non-tombstone files
		require.Equal(t, 40, totalSize) // 8 bytes per content * 5 files
	})
}

// TestIteratorComposition demonstrates composing iterators in various ways
func TestIteratorComposition(t *testing.T) {
	// Create test data with different scenarios
	regularFiles := []graveler.ValueRecord{
		{Key: []byte("regular1.txt"), Value: &graveler.Value{Identity: []byte("r1")}},
		{Key: []byte("regular2.txt"), Value: &graveler.Value{Identity: []byte("r2")}},
	}
	
	deletedFiles := []graveler.ValueRecord{
		{Key: []byte("deleted1.txt"), Value: nil}, // Tombstone
		{Key: []byte("deleted2.txt"), Value: nil}, // Tombstone
	}
	
	hiddenFiles := []graveler.ValueRecord{
		{Key: []byte(".hidden1"), Value: &graveler.Value{Identity: []byte("h1")}},
		{Key: []byte(".hidden2"), Value: &graveler.Value{Identity: []byte("h2")}},
	}

	t.Run("CompositeIteratorProcessing", func(t *testing.T) {
		// Combine all iterators
		iter1 := testutil.NewValueIteratorFake(regularFiles)
		iter2 := testutil.NewValueIteratorFake(deletedFiles)
		iter3 := testutil.NewValueIteratorFake(hiddenFiles)
		
		// Compose: combine -> filter tombstones -> process
		all := graveler.NewCombinedIterator(iter1, iter2, iter3)
		filtered := graveler.NewFilterTombstoneIterator(all)
		
		// Process with new pattern
		regularCount := 0
		hiddenCount := 0
		for value := range filtered.All() {
			filename := string(value.Key)
			if filename[0] == '.' {
				hiddenCount++
			} else {
				regularCount++
			}
		}
		
		require.Equal(t, 2, regularCount)
		require.Equal(t, 2, hiddenCount)
	})
}

// TestSeekingWithNewIterators demonstrates seeking capabilities with new patterns
func TestSeekingWithNewIterators(t *testing.T) {
	// Create sorted test data
	data := []graveler.ValueRecord{
		{Key: []byte("a/file1.txt"), Value: &graveler.Value{Identity: []byte("a1")}},
		{Key: []byte("a/file2.txt"), Value: &graveler.Value{Identity: []byte("a2")}},
		{Key: []byte("b/file1.txt"), Value: &graveler.Value{Identity: []byte("b1")}},
		{Key: []byte("b/file2.txt"), Value: &graveler.Value{Identity: []byte("b2")}},
		{Key: []byte("c/file1.txt"), Value: &graveler.Value{Identity: []byte("c1")}},
	}

	t.Run("SeekAndIterate", func(t *testing.T) {
		iter := testutil.NewValueIteratorFake(data)
		
		// Seek to files in 'b' directory
		iter.SeekGE([]byte("b/"))
		
		// Collect files starting from 'b' using new pattern
		var files []string
		for value := range graveler.IteratorToSeq(iter) {
			filename := string(value.Key)
			// Stop when we move past 'b/' directory
			if filename >= "c/" {
				break
			}
			files = append(files, filename)
		}
		
		expected := []string{"b/file1.txt", "b/file2.txt"}
		require.Equal(t, expected, files)
	})
}

// BenchmarkIteratorPatterns compares performance of old vs new patterns
func BenchmarkIteratorPatterns(b *testing.B) {
	// Create large test dataset
	var data []graveler.ValueRecord
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("file%04d.txt", i)
		data = append(data, graveler.ValueRecord{
			Key:   []byte(key),
			Value: &graveler.Value{Identity: []byte(fmt.Sprintf("id%d", i))},
		})
	}

	b.Run("TraditionalPattern", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := testutil.NewValueIteratorFake(data)
			count := 0
			defer iter.Close()
			for iter.Next() {
				_ = iter.Value()
				count++
			}
		}
	})

	b.Run("NewRangePattern", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := testutil.NewValueIteratorFake(data)
			count := 0
			for _ = range graveler.IteratorToSeq(iter) {
				count++
			}
		}
	})
}