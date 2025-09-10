package committed

import (
	"iter"
	"github.com/treeverse/lakefs/pkg/graveler"
)

type valueIterator struct {
	it Iterator
}

func (v *valueIterator) Next() bool {
	for v.it.Next() {
		if val, _ := v.it.Value(); val != nil {
			return true
		}
	}
	return false
}

func (v *valueIterator) SeekGE(id graveler.Key) {
	v.it.SeekGE(id)
}

func (v *valueIterator) Value() *graveler.ValueRecord {
	rec, _ := v.it.Value()
	return rec
}

func (v *valueIterator) Err() error {
	return v.it.Err()
}

func (v *valueIterator) Close() {
	v.it.Close()
}

func NewValueIterator(it Iterator) graveler.ValueIterator {
	return &valueIterator{
		it: it,
	}
}

// All returns a Go 1.23 iter.Seq that can be used in range-over-function loops.
// This provides a more idiomatic way to iterate over all value records.
//
// Example usage:
//   for value := range valueIterator.All() {
//       // process value
//   }
func (v *valueIterator) All() iter.Seq[*graveler.ValueRecord] {
	return graveler.IteratorToSeq(v)
}
