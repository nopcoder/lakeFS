package graveler

import (
	"context"
)

// diffValuesIterator wraps a diffIterator in order to return only values
type diffValuesIterator struct {
	rangeDiffIter DiffIterator
}

func NewDiffValueIterator(ctx context.Context, left Iterator, right Iterator) DiffIterator {
	return &diffValuesIterator{
		rangeDiffIter: NewDiffIterator(ctx, left, right),
	}
}

func (d diffValuesIterator) Next() bool {
	for d.rangeDiffIter.Next() {
		if d.rangeDiffIter.Err() != nil {
			return false
		}
		val, _ := d.rangeDiffIter.Value()
		if val != nil {
			return true
		}
	}
	return false
}

func (d diffValuesIterator) SeekGE(id Key) {
	d.rangeDiffIter.SeekGE(id)
}

func (d diffValuesIterator) Value() *Diff {
	val, _ := d.rangeDiffIter.Value()
	return val
}

func (d diffValuesIterator) Err() error {
	return d.rangeDiffIter.Err()
}

func (d diffValuesIterator) Close() {
	d.rangeDiffIter.Close()
}
