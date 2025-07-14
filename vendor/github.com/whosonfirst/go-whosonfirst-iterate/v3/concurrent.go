// Package iterator provides methods and utilities for iterating over a collection of records
// (presumed but not required to be Who's On First records) from a variety of sources and dispatching
// processing to user-defined callback functions.
package iterate

import (
	"context"
	"fmt"
	_ "io"
	"iter"
	"log/slog"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/whosonfirst/go-whosonfirst-uri"
)

// concurrentIterator implements the `Iterator` interface wrapping an exiting
// `Iterator` instance and performing additional file-matching checks and retry-on-error
// handling.
type concurrentIterator struct {
	Iterator
	iterator Iterator
	// The count of documents that have been processed.
	seen int64
	// Boolean value indicating whether records are still being iterated.
	iterating *atomic.Bool
	// The number maximum (CPU) processes to used to process documents simultaneously.
	max_procs int
	// A `regexp.Regexp` instance used to test and exclude (if matching) the paths of documents as they are iterated through.
	exclude_paths *regexp.Regexp
	// Exclude Who's On First style "alternate geometry" file paths.
	exclude_alt_files bool
	// A `regexp.Regexp` instance used to test and include (if matching) the paths of documents as they are iterated through.
	include_paths *regexp.Regexp
	// The maximum numbers of attempts to iterate a source. Default is 1.
	max_attempts int
	// The number of seconds to wait between retry attempts. Default is 10.
	retry_after int
	// Skip records (specifically their relative URI) that have already been processed
	dedupe bool
	// Lookup table to track records (specifically their relative URI) that have been processed
	dedupe_map *sync.Map
}

// NewConcurrentIterator() returns a new `Iterator` instance derived from 'iterator_uri' and 'it'. The former is expected
// to be a valid `whosonfirst/go-whosonfirst-iterate/v3.Iterator` URI defined by the following parameters:
// * `?_max_procs=` Explicitly set the number maximum processes to use for iterating documents simultaneously. (Default is the value of `runtime.NumCPU()`.)
// * `?_exclude=` A valid regular expresion used to test and exclude (if matching) the paths of documents as they are iterated through.
// * `?_exclude_alt_files= A boolean value indicating whether Who's On First style "alternate geometry" file paths should be excluded. (Default is false.)
// * `?_include=` A valid regular expresion used to test and include (if matching) the paths of documents as they are iterated through.
// * `?_dedupe=` A boolean value to track and skip records (specifically their relative URI) that have already been processed.
// * `?_retry=` A boolean value indicating whether failed iterators should be retried. (Default is false.)
// * `?_max_attempts=` The number of times to retry a failed iterator. (Default is 1.)
// * `?_retry_after=` The number of seconds to wait before retrying a failed iterator. (Default is 10.)
// These parameters will be used to wrap and perform additional checks when iterating through documents using 'it'.
func NewConcurrentIterator(ctx context.Context, iterator_uri string, it Iterator) (Iterator, error) {

	u, err := url.Parse(iterator_uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	max_procs := runtime.NumCPU()

	retry := false
	max_attempts := 1
	retry_after := 10 // seconds

	if q.Has("_max_procs") {

		max, err := strconv.ParseInt(q.Get("_max_procs"), 10, 64)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_max_procs' parameter, %w", err)
		}

		max_procs = int(max)
	}

	if q.Has("_retry") {

		v, err := strconv.ParseBool(q.Get("_retry"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_retry' parameter, %w", err)
		}

		retry = v
	}

	if retry {

		if q.Has("_max_retries") {

			v, err := strconv.Atoi(q.Get("_max_retries"))

			if err != nil {
				return nil, fmt.Errorf("Failed to parse '_max_retries' parameter, %w", err)
			}

			max_attempts = v
		}

		if q.Has("_retry_after") {

			v, err := strconv.Atoi(q.Get("_retry_after"))

			if err != nil {
				return nil, fmt.Errorf("Failed to parse '_retry_after' parameter, %w", err)
			}

			retry_after = v
		}
	}

	i := &concurrentIterator{
		iterator:     it,
		seen:         int64(0),
		iterating:    new(atomic.Bool),
		max_procs:    max_procs,
		max_attempts: max_attempts,
		retry_after:  retry_after,
	}

	if q.Has("_include") {

		re_include, err := regexp.Compile(q.Get("_include"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_include' parameter, %w", err)
		}

		i.include_paths = re_include
	}

	if q.Has("_exclude") {

		re_exclude, err := regexp.Compile(q.Get("_exclude"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_exclude' parameter, %w", err)
		}

		i.exclude_paths = re_exclude
	}

	if q.Has("_exclude_alt") {

		v, err := strconv.ParseBool(q.Get("_exclude_alt"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_exclude_alt' parameter, %w", err)
		}

		i.exclude_alt_files = v
	}

	if q.Has("_dedupe") {

		v, err := strconv.ParseBool(q.Get("_dedupe"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_dedupe' parameter, %w", err)
		}

		if v {
			i.dedupe = true
			i.dedupe_map = new(sync.Map)
		}

	}

	return i, nil
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *concurrentIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	return func(yield func(rec *Record, err error) bool) {

		t1 := time.Now()

		defer func() {
			slog.Debug("Time to process paths", "count", len(uris), "time", time.Since(t1))
		}()

		it.iterating.Swap(true)
		defer it.iterating.Swap(false)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		procs := it.max_procs
		throttle := make(chan bool, procs)

		for i := 0; i < procs; i++ {
			throttle <- true
		}

		done_ch := make(chan bool)
		err_ch := make(chan error)
		rec_ch := make(chan *Record)

		remaining := len(uris)

		for _, uri := range uris {

			go func(uri string) {

				logger := slog.Default()
				logger = logger.With("uri", uri)

				t2 := time.Now()

				defer func() {
					logger.Debug("Time to iterate uri", "time", time.Since(t2))
				}()

				<-throttle

				defer func() {
					throttle <- true
					done_ch <- true
				}()

				select {
				case <-ctx.Done():
					return
				default:
					// pass
				}

				// The number of records processed across all attempts (it.max_attempts)
				it_counter := int64(0)
				attempts := 0

				do_iter := func() error {

					// The number of records processed in this attempt
					local_counter := int64(1)

					for rec, err := range it.iterator.Iterate(ctx, uri) {

						if err != nil {
							logger.Error("Iterator failed", "counter", it_counter, "local counter", local_counter, "error", err)
							return err
						}

						if it_counter > local_counter {
							logger.Debug("Local counter < counter, skipping", "counter", it_counter, "local counter", local_counter)
							local_counter += 1
							continue
						}

						local_counter += 1
						it_counter += 1

						atomic.AddInt64(&it.seen, 1)

						ok, err := it.shouldYieldRecord(ctx, rec)

						if err != nil {
							continue
						}

						if !ok {
							rec.Body.Close()
							continue
						}

						logger.Debug("Yield record", "counter", it_counter, "local counter", local_counter, "path", rec.Path)
						rec_ch <- rec
					}

					return nil
				}

				for attempts < it.max_attempts {

					attempts += 1
					err := do_iter()

					if err != nil {

						logger.Error("Iterator failed", "attempts", attempts, "max attempts", it.max_attempts, "error", err)

						if it.retry_after == 0 || attempts >= it.max_attempts {
							err_ch <- err
							break
						}

						tts := it.retry_after * attempts
						time.Sleep(time.Duration(tts) * time.Second)
					}
				}

			}(uri)
		}

		for remaining > 0 {
			select {
			case <-done_ch:
				remaining -= 1
			case err := <-err_ch:
				if !yield(nil, err) {
					return
				}
			case rec := <-rec_ch:
				if !yield(rec, nil) {
					return
				}
			default:
				// pass
			}
		}

	}
}

// Seen() returns the total number of records processed so far.
func (it concurrentIterator) Seen() int64 {
	return atomic.LoadInt64(&it.seen)
}

// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
func (it concurrentIterator) IsIterating() bool {
	return it.iterating.Load()
}

// Close performs any implementation specific tasks before terminating the iterator.
func (it concurrentIterator) Close() error {
	return it.iterator.Close()
}

func (it concurrentIterator) shouldYieldRecord(ctx context.Context, rec *Record) (bool, error) {

	if it.include_paths != nil {

		if !it.include_paths.MatchString(rec.Path) {
			return false, nil
		}
	}

	if it.exclude_paths != nil {

		if it.exclude_paths.MatchString(rec.Path) {
			return false, nil
		}
	}

	if it.exclude_alt_files {

		is_alt, err := uri.IsAltFile(rec.Path)

		if err != nil {
			return false, err
		}

		if is_alt {
			return false, nil
		}
	}

	if it.dedupe {

		id, uri_args, err := uri.ParseURI(rec.Path)

		if err != nil {
			return false, fmt.Errorf("Failed to parse %s, %w", rec.Path, err)
		}

		rel_path, err := uri.Id2RelPath(id, uri_args)

		if err != nil {
			return false, fmt.Errorf("Failed to derive relative path for %s, %w", rec.Path, err)
		}

		_, seen := it.dedupe_map.LoadOrStore(rel_path, true)

		if seen {
			slog.Debug("Skip record", "path", rel_path)
			return false, nil
		}
	}

	return true, nil
}
