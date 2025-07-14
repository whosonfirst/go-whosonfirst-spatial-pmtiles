package iterate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"sync/atomic"

	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

func init() {
	ctx := context.Background()
	err := RegisterIterator(ctx, "geojsonl", NewGeoJSONLIterator)

	if err != nil {
		panic(err)
	}
}

// GeoJSONLIterator implements the `Iterator` interface for crawling features in a line-separated GeoJSON record.
type GeoJSONLIterator struct {
	Iterator
	// filters is a `filters.Filters` instance used to include or exclude specific records from being crawled.
	filters filters.Filters
	// seen is the count of documents that have been processed.
	seen int64
	// iterating is a boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
}

// NewGeojsonLIterator() returns a new `GeojsonLIterator` instance configured by 'uri' in the form of:
//
//	geojsonl://?{PARAMETERS}
//
// Where {PARAMETERS} may be:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewGeoJSONLIterator(ctx context.Context, uri string) (Iterator, error) {

	f, err := filters.NewQueryFiltersFromURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create filters from query, %w", err)
	}

	it := &GeoJSONLIterator{
		filters:   f,
		seen:      int64(0),
		iterating: new(atomic.Bool),
	}

	return it, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *GeoJSONLIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	return func(yield func(rec *Record, err error) bool) {

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		for _, uri := range uris {

			r, err := ReaderWithPath(ctx, uri)

			if err != nil {
				if !yield(nil, fmt.Errorf("Failed to create reader for '%s', %w", uri, err)) {
					return
				}
			}

			defer r.Close()

			// see this - we're using ReadLine because it's entirely possible
			// that the raw GeoJSON (LS) will be too long for bufio.Scanner
			// see also - https://golang.org/pkg/bufio/#Reader.ReadLine
			// (20170822/thisisaaronland)

			reader := bufio.NewReader(r)
			raw := bytes.NewBuffer([]byte(""))

			i := 0

			for {

				select {
				case <-ctx.Done():
					break
				default:
					// pass
				}

				path := fmt.Sprintf("%s#%d", uri, i)
				i += 1

				fragment, is_prefix, err := reader.ReadLine()

				if err == io.EOF {
					break
				}

				if err != nil {
					if !yield(nil, fmt.Errorf("Failed to read line at '%s', %w", path, err)) {
						return
					}
				}

				raw.Write(fragment)

				if is_prefix {
					continue
				}

				br := bytes.NewReader(raw.Bytes())
				rsc, err := ioutil.NewReadSeekCloser(br)

				if err != nil {
					if !yield(nil, fmt.Errorf("Failed to create new ReadSeekCloser for '%s', %w", path, err)) {
						return
					}
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

				raw.Reset()
			}
		}
	}

}

// Seen() returns the total number of records processed so far.
func (it *GeoJSONLIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it *GeoJSONLIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it *GeoJSONLIterator) Close() error {
	return nil
}
