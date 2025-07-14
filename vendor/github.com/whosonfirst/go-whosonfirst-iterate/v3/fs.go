package iterate

import (
	"context"
	"fmt"
	"io/fs"
	"iter"
	"log/slog"
	"sync/atomic"

	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

// FSIterator implements the `Iterator` interface for crawling records in a fs.
type FSIterator struct {
	Iterator
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// The fs.FS filesystem to iterate through.
	fs fs.FS
	// seen is the count of documents that have been processed.
	seen int64
	// iterating is a boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewFSIterator() returns a new `FSIterator` instance (wrapped by the `concurrentIterator` implementation) configured by 'uri' for iterating over 'fs'.
// Where 'uri' takes the form of:
//
//	fs://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
// * `?processes=` An optional number assigning the maximum number of database rows that will be processed simultaneously. (Default is defined by `runtime.NumCPU()`.)
func NewFSIterator(ctx context.Context, uri string, iterator_fs fs.FS) (Iterator, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	it := &FSIterator{
		fs:        iterator_fs,
		filters:   f,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	return NewConcurrentIterator(ctx, uri, it)
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *FSIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	return func(yield func(rec *Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		var walk_func func(path string, d fs.DirEntry, err error) error

		walk_func = func(path string, d fs.DirEntry, err error) error {

			if err != nil {
				return fmt.Errorf("Failed to walk %s, %w", path, err)
			}

			if d.IsDir() {
				return nil
			}

			r, err := it.fs.Open(path)

			if err != nil {

				if !yield(nil, fmt.Errorf("Failed to open %s for reading, %w", path, err)) {
					return err
				}

				return nil
			}

			rsc, err := ioutil.NewReadSeekCloser(r)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to create ReadSeekCloser for %s, %w", path, err)) {
					return err
				}

				return nil
			}

			if it.filters != nil {

				ok, err := ApplyFilters(ctx, rsc, it.filters)

				if err != nil {
					rsc.Close()
					if !yield(nil, fmt.Errorf("Failed to apply filters for '%s', %w", path, err)) {
						return err
					}

					return nil
				}

				if !ok {
					rsc.Close()
					return nil
				}
			}

			rec := NewRecord(path, rsc)
			yield(rec, nil)

			return nil
		}

		for _, uri := range uris {

			logger := slog.Default()
			logger = logger.With("uri", uri)

			err := fs.WalkDir(it.fs, uri, walk_func)

			if err != nil {
				logger.Error("Failed to walk filesystem", "error", err)
				return
			}
		}
	}

	return nil
}

// Seen() returns the total number of records processed so far.
func (it *FSIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *FSIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *FSIterator) Close() error {
	return nil
}
