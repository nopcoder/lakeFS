package graveler

import "github.com/cockroachdb/pebble/sstable"

// SSTableIterator returns ordered iteration of the SSTable entries
type SSTableIterator struct {
	it sstable.Iterator

	currKey   *sstable.InternalKey
	currValue []byte

	postSeek bool
	err      error
	derefer  func() error
}

func NewSSTableIterator(it sstable.Iterator, derefer func() error) *SSTableIterator {
	iter := &SSTableIterator{
		it:      it,
		derefer: derefer,
	}

	return iter
}

func (iter *SSTableIterator) SeekGE(lookup Key) {
	key, value := iter.it.SeekGE(lookup, sstable.SeekGEFlags(0))
	val, err := retrieveValue(value)
	iter.currKey = key
	iter.currValue = val
	iter.err = err
	iter.postSeek = true
}

func (iter *SSTableIterator) Next() bool {
	if !iter.postSeek {
		key, value := iter.it.Next()
		iter.currKey = key
		val, err := retrieveValue(value)
		iter.err = err
		iter.currValue = val
	}
	iter.postSeek = false

	if iter.currKey == nil && iter.currValue == nil {
		return false
	}

	return true
}

func (iter *SSTableIterator) Value() *ValueRecord {
	if iter.currKey == nil || iter.err != nil || iter.postSeek {
		return nil
	}
	v, err := UnmarshalValue(iter.currValue)
	if err != nil {
		iter.err = err
		return nil
	}
	return &ValueRecord{
		Key:   iter.currKey.UserKey,
		Value: v,
	}
}

func (iter *SSTableIterator) Err() error {
	return iter.err
}

func (iter *SSTableIterator) Close() {
	if iter.it == nil {
		return
	}
	err := iter.it.Close()
	iter.updateOnNilErr(err)

	err = iter.derefer()
	iter.updateOnNilErr(err)

	iter.it = nil
}

func (iter *SSTableIterator) updateOnNilErr(err error) {
	if iter.err == nil {
		// avoid overriding earlier errors
		iter.err = err
	}
}
