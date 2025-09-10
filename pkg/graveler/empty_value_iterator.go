package graveler

import (
	"iter"
)

func NewEmptyValueIterator() ValueIterator {
	return &emptyValueIterator{}
}

type emptyValueIterator struct{}

func (e *emptyValueIterator) SeekGE(_ Key) {
}

func (e *emptyValueIterator) Value() *ValueRecord {
	return nil
}

func (e *emptyValueIterator) Err() error {
	return nil
}

func (e *emptyValueIterator) Close() {
}

func (e *emptyValueIterator) Next() bool {
	return false
}

// All returns a Go 1.23 iter.Seq that can be used in range-over-function loops.
// For empty iterators, this returns an empty sequence.
//
// Example usage:
//   for value := range iterator.All() {
//       // never executed for empty iterator
//   }
func (e *emptyValueIterator) All() iter.Seq[*ValueRecord] {
	return IteratorToSeq(e)
}
