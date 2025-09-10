package graveler

import (
	"iter"
)

// IteratorToSeq converts a traditional ValueIterator to a Go 1.23 iter.Seq[*ValueRecord].
// This allows using traditional iterators in range-over-function loops.
func IteratorToSeq(it ValueIterator) iter.Seq[*ValueRecord] {
	return func(yield func(*ValueRecord) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

// DiffIteratorToSeq converts a traditional DiffIterator to a Go 1.23 iter.Seq[*Diff].
// This allows using traditional diff iterators in range-over-function loops.
func DiffIteratorToSeq(it DiffIterator) iter.Seq[*Diff] {
	return func(yield func(*Diff) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

// BranchIteratorToSeq converts a traditional BranchIterator to a Go 1.23 iter.Seq[*BranchRecord].
// This allows using traditional branch iterators in range-over-function loops.
func BranchIteratorToSeq(it BranchIterator) iter.Seq[*BranchRecord] {
	return func(yield func(*BranchRecord) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

// TagIteratorToSeq converts a traditional TagIterator to a Go 1.23 iter.Seq[*TagRecord].
// This allows using traditional tag iterators in range-over-function loops.
func TagIteratorToSeq(it TagIterator) iter.Seq[*TagRecord] {
	return func(yield func(*TagRecord) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

// CommitIteratorToSeq converts a traditional CommitIterator to a Go 1.23 iter.Seq[*CommitRecord].
// This allows using traditional commit iterators in range-over-function loops.
func CommitIteratorToSeq(it CommitIterator) iter.Seq[*CommitRecord] {
	return func(yield func(*CommitRecord) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

// RepositoryIteratorToSeq converts a traditional RepositoryIterator to a Go 1.23 iter.Seq[*RepositoryRecord].
// This allows using traditional repository iterators in range-over-function loops.
func RepositoryIteratorToSeq(it RepositoryIterator) iter.Seq[*RepositoryRecord] {
	return func(yield func(*RepositoryRecord) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

// PullsIteratorToSeq converts a traditional PullsIterator to a Go 1.23 iter.Seq[*PullRequestRecord].
// This allows using traditional pulls iterators in range-over-function loops.
func PullsIteratorToSeq(it PullsIterator) iter.Seq[*PullRequestRecord] {
	return func(yield func(*PullRequestRecord) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}

// LinkAddressIteratorToSeq converts a traditional LinkAddressIterator to a Go 1.23 iter.Seq[*LinkAddressData].
// This allows using traditional link address iterators in range-over-function loops.
func LinkAddressIteratorToSeq(it LinkAddressIterator) iter.Seq[*LinkAddressData] {
	return func(yield func(*LinkAddressData) bool) {
		defer it.Close()
		for it.Next() {
			if !yield(it.Value()) {
				return
			}
		}
	}
}