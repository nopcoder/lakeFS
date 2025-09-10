package graveler

import (
	"iter"
)

// Examples demonstrates how to use the new Go 1.23 iterator functionality
// with graveler iterators. These examples show both the old and new patterns.

// ExampleTraditionalIteratorPattern shows the traditional iterator usage
func ExampleTraditionalIteratorPattern(it ValueIterator) []*ValueRecord {
	defer it.Close()
	
	var results []*ValueRecord
	for it.Next() {
		if err := it.Err(); err != nil {
			// Handle error
			break
		}
		results = append(results, it.Value())
	}
	return results
}

// ExampleNewIteratorPattern shows the new Go 1.23 iterator usage
func ExampleNewIteratorPattern(it ValueIterator) []*ValueRecord {
	var results []*ValueRecord
	for value := range IteratorToSeq(it) {
		results = append(results, value)
	}
	return results
}

// ExampleCombinedIteratorUsage demonstrates using combined iterators with range-over-function
func ExampleCombinedIteratorUsage(iterA, iterB ValueIterator) []*ValueRecord {
	combined := NewCombinedIterator(iterA, iterB)
	
	var results []*ValueRecord
	for value := range combined.All() {
		results = append(results, value)
	}
	return results
}

// ExampleFilteredIteratorUsage demonstrates filtering tombstones with range-over-function
func ExampleFilteredIteratorUsage(it ValueIterator) []*ValueRecord {
	filtered := NewFilterTombstoneIterator(it)
	
	var results []*ValueRecord
	for value := range filtered.All() {
		// Only non-tombstone values will be yielded
		results = append(results, value)
	}
	return results
}

// ExampleDiffIteratorUsage demonstrates iterating over diffs with range-over-function
func ExampleDiffIteratorUsage(diffIter DiffIterator) []DiffType {
	var diffTypes []DiffType
	for diff := range DiffIteratorToSeq(diffIter) {
		diffTypes = append(diffTypes, diff.Type)
	}
	return diffTypes
}

// ExampleBranchIteratorUsage demonstrates iterating over branches with range-over-function
func ExampleBranchIteratorUsage(branchIter BranchIterator) []BranchID {
	var branchIDs []BranchID
	for branch := range BranchIteratorToSeq(branchIter) {
		branchIDs = append(branchIDs, branch.BranchID)
	}
	return branchIDs
}

// ExampleIteratorComposition demonstrates composing multiple iterators
func ExampleIteratorComposition(valueIters ...ValueIterator) iter.Seq[*ValueRecord] {
	// Combine multiple iterators into one
	combined := NewCombinedIterator(valueIters...)
	
	// Return the seq for further composition
	return combined.All()
}

// ExampleIteratorPipeline demonstrates a processing pipeline with iterators
func ExampleIteratorPipeline(it ValueIterator) iter.Seq[string] {
	// Create a pipeline: iterator -> filter tombstones -> extract keys
	filtered := NewFilterTombstoneIterator(it)
	
	return func(yield func(string) bool) {
		for value := range filtered.All() {
			// Convert key bytes to string and yield
			if !yield(string(value.Key)) {
				return
			}
		}
	}
}

// ExampleIteratorChaining demonstrates chaining multiple operations
func ExampleIteratorChaining(baseIter ValueIterator, maxItems int) []string {
	var results []string
	
	// Chain operations: base -> filter -> limit -> transform
	for value := range NewFilterTombstoneIterator(baseIter).All() {
		if len(results) >= maxItems {
			break
		}
		results = append(results, string(value.Key))
	}
	
	return results
}

// ExampleIteratorErrorHandling demonstrates error handling with new iterators
func ExampleIteratorErrorHandling(it ValueIterator) ([]*ValueRecord, error) {
	var results []*ValueRecord
	
	// Use the traditional iterator for error handling
	defer it.Close()
	for it.Next() {
		if err := it.Err(); err != nil {
			return nil, err
		}
		results = append(results, it.Value())
	}
	
	return results, it.Err()
}

// ExampleSeekableIterator demonstrates seeking with new iterator patterns
func ExampleSeekableIterator(it ValueIterator, startKey Key) []*ValueRecord {
	// Seek to start position
	it.SeekGE(startKey)
	
	// Use range-over-function from that position
	var results []*ValueRecord
	for value := range IteratorToSeq(it) {
		results = append(results, value)
	}
	
	return results
}