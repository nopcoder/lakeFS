# Go 1.23 Iterator Support for Graveler

This document describes the Go 1.23 iterator support added to the graveler package. The changes provide backward-compatible support for range-over-function syntax while maintaining all existing iterator functionality.

## Overview

Go 1.23 introduced range-over-function types (`iter.Seq[V]` and `iter.Seq2[K,V]`) that allow using custom iterators directly in `for range` loops. This implementation adds support for these patterns to all graveler iterators.

## Key Files Added

### `iterator_seq.go`
Contains adapter functions to convert traditional graveler iterators to Go 1.23 iterators:
- `IteratorToSeq(ValueIterator) iter.Seq[*ValueRecord]`
- `DiffIteratorToSeq(DiffIterator) iter.Seq[*Diff]`
- `BranchIteratorToSeq(BranchIterator) iter.Seq[*BranchRecord]`
- And similar functions for all iterator types

### `iterator_examples.go`
Comprehensive examples showing both traditional and new usage patterns:
- Side-by-side comparisons of old vs new syntax
- Iterator composition and pipeline examples
- Error handling patterns

### `iterator_seq_test.go` and `iterator_methods_test.go`
Tests demonstrating the new functionality and ensuring compatibility.

## Updated Iterator Types

The following iterator implementations now have `All()` or `AllWithRange()` methods:

### Core Iterators
- `CombinedIterator` - `All() iter.Seq[*ValueRecord]`
- `FilterTombstoneIterator` - `All() iter.Seq[*ValueRecord]`
- `emptyValueIterator` - `All() iter.Seq[*ValueRecord]`

### Diff Iterators
- `CombinedDiffIterator` - `All() iter.Seq[*Diff]`
- `JoinedDiffIterator` - `All() iter.Seq[*Diff]`
- `uncommittedDiffIterator` - `All() iter.Seq[*Diff]`

### Committed Package Iterators
- `iterator` - `AllWithRange() iter.Seq2[*ValueRecord, *Range]`
- `emptyIterator` - `AllWithRange() iter.Seq2[*ValueRecord, *Range]`
- `valueIterator` - `All() iter.Seq[*ValueRecord]`
- `UnmarshalIterator` - `All() iter.Seq[*ValueRecord]`
- `compareIterator` - `All() iter.Seq[*Diff]`
- `SkipPrefixIterator` - `AllWithRange() iter.Seq2[*ValueRecord, *Range]`

### Ref Package Iterators
- `BranchSimpleIterator` - `All() iter.Seq[*BranchRecord]`
- `BranchByCommitIterator` - `All() iter.Seq[*BranchRecord]`

## Usage Examples

### Traditional Pattern
```go
// Old way
defer iter.Close()
for iter.Next() {
    if err := iter.Err(); err != nil {
        return err
    }
    value := iter.Value()
    // process value
}
```

### New Go 1.23 Pattern
```go
// New way - automatic cleanup, no manual error checking needed in loop
for value := range iter.All() {
    // process value
}
```

### Combined Iterators
```go
// Combine multiple iterators
combined := graveler.NewCombinedIterator(iter1, iter2, iter3)
for value := range combined.All() {
    // process combined values in sorted order
}
```

### Diff Iterators
```go
// Iterate over diffs
for diff := range graveler.DiffIteratorToSeq(diffIter) {
    switch diff.Type {
    case graveler.DiffTypeAdded:
        // handle addition
    case graveler.DiffTypeRemoved:
        // handle removal
    }
}
```

### Range Iterators (committed package)
```go
// Iterate with range information
for value, rng := range iterator.AllWithRange() {
    // process value and its range
}
```

## Backward Compatibility

All existing code continues to work unchanged. The new methods are additive:
- Traditional `Next()`, `Value()`, `Err()`, `Close()` methods remain
- All existing iterator interfaces are unchanged
- Error handling patterns work as before
- Seeking with `SeekGE()` still available

## Benefits

1. **Cleaner Code**: Range-over-function is more readable than manual loops
2. **Automatic Cleanup**: Iterator cleanup happens automatically when exiting loops
3. **Early Exit**: Breaking from range loops properly cleans up iterators
4. **Composability**: iter.Seq works with other Go 1.23+ standard library functions
5. **Type Safety**: Strong typing maintained throughout

## Implementation Notes

- All `All()` methods use the existing iterator logic internally
- Cleanup (`Close()`) is called automatically via `defer` in the seq functions
- Early exit from range loops (via `break` or `return`) properly stops iteration
- The `yield` function return value controls whether to continue iteration

## Error Handling

For error handling, use traditional patterns since `iter.Seq` doesn't propagate errors:

```go
// For error handling, still use traditional pattern
defer iter.Close()
for iter.Next() {
    if err := iter.Err(); err != nil {
        return err
    }
    // process iter.Value()
}
```

Or use the error handling helper example in `iterator_examples.go`.