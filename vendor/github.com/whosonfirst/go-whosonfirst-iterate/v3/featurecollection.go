package iterate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"sync/atomic"

	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

func init() {
	ctx := context.Background()
	err := RegisterIterator(ctx, "featurecollection", NewFeatureCollectionIterator)

	if err != nil {
		panic(err)
	}
}

// FeatureCollectionIterator implements the `Iterator` interface for crawling features in a GeoJSON FeatureCollection record.
type FeatureCollectionIterator struct {
	Iterator
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// seen is the count of documents that have been processed.
	seen int64
	// iterating is a boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewFeatureCollectionIterator() returns a new `FeatureCollectionIterator` instance configured by 'uri' in the form of:
//
//	featurecollection://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewFeatureCollectionIterator(ctx context.Context, uri string) (Iterator, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	i := &FeatureCollectionIterator{
		filters:   f,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	return i, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *FeatureCollectionIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

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

			defer r.Close()

			body, err := io.ReadAll(r)

			if err != nil {

				if !yield(nil, fmt.Errorf("Failed to read body for '%s', %w", uri, err)) {
					return
				}

				continue
			}

			type FC struct {
				Type     string
				Features []interface{}
			}

			var collection FC

			err = json.Unmarshal(body, &collection)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to unmarshal '%s' as a feature collection, %w", uri, err)) {
					return
				}

				continue
			}

			for i, f := range collection.Features {

				select {
				case <-ctx.Done():
					break
				default:
					// pass
				}

				path := fmt.Sprintf("%s#%d", uri, i)

				feature, err := json.Marshal(f)

				if err != nil {
					if !yield(nil, fmt.Errorf("Failed to marshal feature for '%s', %w", path, err)) {
						return
					}

					continue
				}

				br := bytes.NewReader(feature)
				rsc, err := ioutil.NewReadSeekCloser(br)

				if err != nil {
					if !yield(nil, fmt.Errorf("Failed to create new ReadSeekCloser for '%s', %w", path, err)) {
						return
					}

					continue
				}

				if it.filters != nil {

					ok, err := ApplyFilters(ctx, rsc, it.filters)

					if err != nil {
						rsc.Close()
						if !yield(nil, fmt.Errorf("Failed to apply filters for '%s', %w", path, err)) {
							return
						}

						continue
					}

					if !ok {
						rsc.Close()
						continue
					}
				}

				rec := NewRecord(path, rsc)

				if !yield(rec, nil) {
					return
				}
			}

		}
	}
}

// Seen() returns the total number of records processed so far.
func (it *FeatureCollectionIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *FeatureCollectionIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *FeatureCollectionIterator) Close() error {
	return nil
}
