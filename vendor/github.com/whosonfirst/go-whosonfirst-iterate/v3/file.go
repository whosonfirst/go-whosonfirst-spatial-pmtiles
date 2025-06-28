package iterate

import (
	"context"
	"fmt"
	"iter"
	"sync/atomic"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

func init() {
	ctx := context.Background()
	err := RegisterIterator(ctx, "file", NewFileIterator)

	if err != nil {
		panic(err)
	}
}

// FileIterator implements the `Iterator` interface for crawling individual file records.
type FileIterator struct {
	Iterator
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// iterating is a boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
	// seen is the count of documents that have been processed.
	seen int64
}

// NewFileIterator() returns a new `FileIterator` instance configured by 'uri' in the form of:
//
//	file://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewFileIterator(ctx context.Context, uri string) (Iterator, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	it := &FileIterator{
		filters:   f,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	return it, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *FileIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	return func(yield func(rec *Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		for _, uri := range uris {

			r, err := ReaderWithPath(ctx, uri)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to create reader for '%s', %w", uri, err)) {
					return
				}

				continue
			}

			atomic.AddInt64(&it.seen, 1)

			if it.filters != nil {

				ok, err := ApplyFilters(ctx, r, it.filters)

				if err != nil {

					r.Close()

					if !yield(nil, fmt.Errorf("Failed to apply filters for '%s', %w", uri, err)) {
						return
					}

					continue
				}

				if !ok {
					r.Close()
					continue
				}
			}

			rec := NewRecord(uri, r)

			if !yield(rec, nil) {
				return
			}
		}
	}

}

// Seen() returns the total number of records processed so far.
func (it *FileIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *FileIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *FileIterator) Close() error {
	return nil
}
