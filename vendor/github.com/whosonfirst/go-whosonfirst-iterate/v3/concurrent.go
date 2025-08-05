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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
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
	// Boolean flag indicating whether stats should be logged. Default is true.
	with_stats bool
	// The inteval at which stats are logged. Default is 60 seconds.
	stats_interval time.Duration
	// The level at which stats are logged. Default is INFO.
	stats_level slog.Level
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
// * `?_with_stats=` Boolean flag indicating whether stats should be logged. Default is true.
// * `?_stats_interval=` The number of seconds between stats logging events. Default is 60.
// * `?_stas_level=` The (slog/log) level at which stats are logged. Default is INFO.
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

	with_stats := true
	stats_interval := 1 * time.Minute
	stats_level := slog.LevelInfo

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
		iterator:       it,
		seen:           int64(0),
		iterating:      new(atomic.Bool),
		max_procs:      max_procs,
		max_attempts:   max_attempts,
		retry_after:    retry_after,
		with_stats:     with_stats,
		stats_interval: stats_interval,
		stats_level:    stats_level,
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

	if q.Has("_with_stats") {

		v, err := strconv.ParseBool(q.Get("_with_stats"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_with_stats' parameter, %w", err)
		}

		i.with_stats = v
	}

	if q.Has("_stats_interval") {

		v, err := strconv.Atoi(q.Get("_stats_interval"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '_stats_interval' parameter, %w", err)
		}

		i.stats_interval = time.Duration(v) * time.Second
	}

	if q.Has("_stats_level") {

		switch strings.ToUpper(q.Get("_stats_level")) {
		case "DEBUG":
			i.stats_level = slog.LevelDebug
		case "INFO":
			i.stats_level = slog.LevelInfo
		case "WARN":
			i.stats_level = slog.LevelWarn
		case "ERROR":
			i.stats_level = slog.LevelError
		default:
			return nil, fmt.Errorf("Invalid or unsupport log level for stats")
		}

		slog.Info("BUELLER", "level", i.stats_level)
	}

	return i, nil
}

func (it *concurrentIterator) showStats(ctx context.Context, t1 time.Time) {

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	slog.Log(ctx, it.stats_level,
		"Iterator stats",
		"elapsed", time.Since(t1),
		"seen", it.Seen(),
		"allocated", humanize.Bytes(m.Alloc),
		"total allocated", humanize.Bytes(m.TotalAlloc),
		"sys", humanize.Bytes(m.Sys),
		"numgc", m.NumGC,
	)
}

// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in 'uris'.
func (it *concurrentIterator) Iterate(ctx context.Context, uris ...string) iter.Seq2[*Record, error] {

	return func(yield func(rec *Record, err error) bool) {

		t1 := time.Now()

		if it.with_stats {

			ticker := time.NewTicker(it.stats_interval)
			ticker_done := make(chan bool)

			defer func() {
				ticker.Stop()
				ticker_done <- true
			}()

			go func() {

				for {
					select {
					case <-ticker_done:
						it.showStats(ctx, t1)
						return
					case <-ticker.C:
						it.showStats(ctx, t1)
					}
				}
			}()
		}

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
				t2 := time.Now()

				var it_counter int64
				attempts := 0

				defer func() {
					logger.Debug("Run garbage collector")
					runtime.GC()

					logger.Debug("Time to iterate uri", "time", time.Since(t2))
				}()

				<-throttle

				defer func() {
					throttle <- true
					done_ch <- true
				}()

				logger_uri, err := ScrubURI(uri)

				if err != nil {
					slog.Error("Failed to scrub URI", "error", err)
					return
				}

				logger = logger.With("uri", logger_uri)

				// END OF put me in a funciton somewhere...

				select {
				case <-ctx.Done():
					return
				default:
					// pass
				}

				atomic.StoreInt64(&it_counter, 0)

				do_iter := func(target_uri string) error {

					logger_uri, err := ScrubURI(target_uri)

					if err != nil {
						return err
					}

					// The number of records processed in this attempt
					var local_counter int64
					atomic.StoreInt64(&local_counter, 0)

					logger.Debug("Iterate target", "target uri", logger_uri, "counter", atomic.LoadInt64(&it_counter), "local counter", atomic.LoadInt64(&local_counter))

					for rec, err := range it.iterator.Iterate(ctx, target_uri) {

						if err != nil {
							logger.Error("Iterator failed", "counter", atomic.LoadInt64(&it_counter), "local counter", atomic.LoadInt64(&local_counter), "error", err)
							return err
						}

						if atomic.LoadInt64(&it_counter) > atomic.LoadInt64(&local_counter) {
							logger.Debug("Iterator counter > local counter, skipping", "path", rec.Path, "counter", atomic.LoadInt64(&it_counter), "local counter", atomic.LoadInt64(&local_counter))
							atomic.AddInt64(&local_counter, 1)
							rec.Body.Close()
							continue
						}

						atomic.AddInt64(&local_counter, 1)
						atomic.AddInt64(&it_counter, 1)
						atomic.AddInt64(&it.seen, 1)

						ok, err := it.shouldYieldRecord(ctx, rec)

						if err != nil {
							logger.Warn("Failed to determine if record should yield", "path", rec.Path, "error", err)
							rec.Body.Close()
							continue
						}

						if !ok {
							rec.Body.Close()
							continue
						}

						rec_ch <- rec
					}

					return nil
				}

				for attempts < it.max_attempts {

					logger.Debug("Do iter", "attempt", attempts, "max attempts", it.max_attempts, "counter", atomic.LoadInt64(&it_counter))

					attempts += 1
					err := do_iter(uri)

					if err == nil {
						logger.Debug("Iteration successful", "attempt", attempts, "max attempts", it.max_attempts, "counter", atomic.LoadInt64(&it_counter))
						break
					} else {

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
