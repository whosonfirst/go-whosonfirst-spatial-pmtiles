package iterate

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"path/filepath"
	"sync/atomic"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

func init() {
	ctx := context.Background()
	err := RegisterIterator(ctx, "filelist", NewFileListIterator)
	if err != nil {
		panic(err)
	}
}

// FileListIterator implements the `Iterator` interface for crawling records listed in a "file list" (a plain text newline-delimted list of files).
type FileListIterator struct {
	Iterator
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// seen is the count of documents that have been processed.
	seen int64
	// iterating is a boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewFileListIterator() returns a new `FileListIterator` instance configured by 'uri' in the form of:
//
//	file://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewFileListIterator(ctx context.Context, uri string) (Iterator, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	it := &FileListIterator{
		filters:   f,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	return it, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *FileListIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	return func(yield func(rec *Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		for _, uri := range uris {

			abs_path, err := filepath.Abs(uri)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to derive absolute path for '%s', %w", uri, err)) {
					return
				}

				continue
			}

			r, err := ReaderWithPath(ctx, abs_path)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to create reader for '%s', %w", abs_path, err)) {
					return
				}

				continue
			}

			defer r.Close()

			scanner := bufio.NewScanner(r)

			for scanner.Scan() {

				select {
				case <-ctx.Done():
					break
				default:
					// pass
				}

				path := scanner.Text()

				r2, err := ReaderWithPath(ctx, path)

				if err != nil {
					if !yield(nil, fmt.Errorf("Failed to create reader for '%s', %w", path, err)) {
						return
					}

					continue
				}

				atomic.AddInt64(&it.seen, 1)

				if it.filters != nil {

					ok, err := ApplyFilters(ctx, r2, it.filters)

					if err != nil {

						r2.Close()

						if !yield(nil, fmt.Errorf("Failed to apply filters for '%s', %w", path, err)) {
							return
						}

						continue
					}

					if !ok {
						r2.Close()
						continue
					}
				}

				rec := NewRecord(path, r2)

				if !yield(rec, nil) {
					break
				}
			}

			err = scanner.Err()

			if err != nil {
				yield(nil, err)
			}
		}
	}
}

// Seen() returns the total number of records processed so far.
func (it *FileListIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *FileListIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *FileListIterator) Close() error {
	return nil
}
