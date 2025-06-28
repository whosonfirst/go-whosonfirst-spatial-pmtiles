package iterate

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

func init() {
	ctx := context.Background()
	err := RegisterIterator(ctx, "directory", NewDirectoryIterator)

	if err != nil {
		panic(err)
	}
}

// DirectoryIterator implements the `Iterator` interface for crawling records in a directory.
type DirectoryIterator struct {
	Iterator
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// seen is the count of documents that have been processed.
	seen int64
	// iterating is a boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewDirectoryIterator() returns a new `DirectoryIterator` instance configured by 'uri' in the form of:
//
//	directory://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewDirectoryIterator(ctx context.Context, uri string) (Iterator, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	it := &DirectoryIterator{
		filters:   f,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	return it, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *DirectoryIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	return func(yield func(rec *Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		for _, uri := range uris {

			logger := slog.Default()
			logger = logger.With("uri", uri)

			abs_path, err := filepath.Abs(uri)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to derive absolute path for '%s', %w", uri, err)) {
					return
				}

				continue
			}

			logger = logger.With("path", abs_path)
			root, err := os.OpenRoot(abs_path)

			if err != nil {
				logger.Error("Failed to open root", "error", err)
				if !yield(nil, fmt.Errorf("Failed to open root for '%s', %w", abs_path, err)) {
					return
				}

				continue
			}

			root_fs := root.FS()

			err = fs.WalkDir(root_fs, ".", func(path string, d fs.DirEntry, err error) error {

				select {
				case <-ctx.Done():
					return nil
				default:
					// pass
				}

				if err != nil {

					if !yield(nil, err) {
						return err
					}

					return nil
				}

				if d.IsDir() {
					return nil
				}

				atomic.AddInt64(&it.seen, 1)

				r, err := root.Open(path)

				if err != nil {

					r.Close()

					if !yield(nil, fmt.Errorf("Failed to create reader for '%s', %w", abs_path, err)) {
						return err
					}

					return nil
				}

				if it.filters != nil {

					ok, err := ApplyFilters(ctx, r, it.filters)

					if err != nil {
						r.Close()
						if !yield(nil, fmt.Errorf("Failed to apply filters for '%s', %w", path, err)) {
							return err
						}

						return nil
					}

					if !ok {
						r.Close()
						return nil
					}
				}

				rec := NewRecord(path, r)

				if !yield(rec, nil) {
					return io.EOF
				}

				return nil
			})

			if err != nil && err != io.EOF {
				logger.Error("Failed to walk dir", "error", err)
			}
		}
	}
}

// Seen() returns the total number of records processed so far.
func (it *DirectoryIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *DirectoryIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *DirectoryIterator) Close() error {
	return nil
}
