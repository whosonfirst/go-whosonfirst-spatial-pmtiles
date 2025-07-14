package iterate

import (
	"context"
	"fmt"
	"iter"
	"path/filepath"
)

func init() {
	ctx := context.Background()
	err := RegisterIterator(ctx, "repo", NewRepoIterator)

	if err != nil {
		panic(err)
	}
}

// RepoIterator implements the `Iterator` interface for crawling records in a Who's On First style data directory.
type RepoIterator struct {
	Iterator
	// iterator is the underlying `DirectoryIterator` instance for crawling records.
	iterator Iterator
}

// NewDirectoryIterator() returns a new `RepoIterator` instance configured by 'uri' in the form of:
//
//	repo://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewRepoIterator(ctx context.Context, uri string) (Iterator, error) {

	directory_it, err := NewDirectoryIterator(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new directory iterator, %w", err)
	}

	it := &RepoIterator{
		iterator: directory_it,
	}

	return it, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *RepoIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	data_uris := make([]string, len(uris))

	for idx, path := range uris {

		abs_path, err := filepath.Abs(path)

		if err != nil {

			return func(yield func(rec *Record, err error) bool) {
				yield(nil, err)
			}
		}

		data_path := filepath.Join(abs_path, "data")
		data_uris[idx] = data_path
	}

	return it.iterator.Iterate(ctx, data_uris...)
}

// Seen() returns the total number of records processed so far.
func (it *RepoIterator) Seen() int64 {
	return it.iterator.Seen()
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *RepoIterator) IsIterating() bool {
	return it.iterator.IsIterating()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *RepoIterator) Close() error {
	return it.iterator.Close()
}
