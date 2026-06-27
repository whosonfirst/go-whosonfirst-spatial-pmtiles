package iterate

import (
	"context"
	"iter"
)

func init() {

	ctx := context.Background()
	err := RegisterIterator(ctx, "null", NewNullIterator)

	if err != nil {
		panic(err)
	}
}

// NullIterator implements the `Iterator` interface for appearing to crawl records but not doing anything.
type NullIterator struct {
	Iterator
}

// NewNullIterator() returns a new `NullIterator` instance configured by 'uri' in the form of:
//
//	null://
func NewNullIterator(ctx context.Context, uri string) (Iterator, error) {

	it := &NullIterator{}
	return it, nil
}

// Iterate() does nothing.
func (it *NullIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {
	return func(yield func(rec *Record, err error) bool) {}
}

// Seen() returns the total number of records processed so far.
func (it *NullIterator) Seen() int64 {
	return int64(0)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *NullIterator) IsIterator() bool {
	return false
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *NullIterator) Close() error {
	return nil
}
