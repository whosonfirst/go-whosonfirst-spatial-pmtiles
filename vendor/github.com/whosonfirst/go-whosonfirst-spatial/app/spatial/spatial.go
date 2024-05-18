package spatial

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/sfomuseum/go-flags/lookup"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

type SpatialApplication struct {
	mode             string
	SpatialDatabase  database.SpatialDatabase
	PropertiesReader reader.Reader
	Iterator         *iterator.Iterator
	Timings          []*timings.SinceResponse
	Monitor          timings.Monitor
	mu               *sync.RWMutex
}

func NewSpatialApplicationWithFlagSet(ctx context.Context, fl *flag.FlagSet) (*SpatialApplication, error) {

	spatial_db_uri, err := lookup.StringVar(fl, flags.SPATIAL_DATABASE_URI)

	if err != nil {
		return nil, fmt.Errorf("Failed look up '%s' flag, %w", flags.SPATIAL_DATABASE_URI, err)
	}

	spatial_db, err := database.NewSpatialDatabase(ctx, spatial_db_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed instantiate spatial database, %v", err)
	}

	// Set up properties reader

	var properties_reader reader.Reader

	properties_reader_uri, _ := lookup.StringVar(fl, flags.PROPERTIES_READER_URI)

	if properties_reader_uri != "" {

		use_spatial_uri := fmt.Sprintf("{%s}", flags.SPATIAL_DATABASE_URI)

		if properties_reader_uri == use_spatial_uri {

			spatial_database_uri, err := lookup.StringVar(fl, flags.SPATIAL_DATABASE_URI)

			if err != nil {
				return nil, fmt.Errorf("Failed to retrieve %s flag", flags.SPATIAL_DATABASE_URI)
			}

			properties_reader_uri = spatial_database_uri
		}

		r, err := reader.NewReader(ctx, properties_reader_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create properties reader, %v", err)
		}

		properties_reader = r
	}

	if properties_reader == nil {
		properties_reader = spatial_db
	}

	// Set up iterator (to index records at start up if necessary)

	iter, err := NewIteratorWithFlagSet(ctx, fl, spatial_db)

	if err != nil {
		return nil, fmt.Errorf("Failed to instantiate iterator, %v", err)
	}

	// Enable custom placetypes

	err = AppendCustomPlacetypesWithFlagSet(ctx, fl)

	if err != nil {
		return nil, fmt.Errorf("Failed to append custom placetypes, %v", err)
	}

	// Set up monitor (for tracking indexing)

	m, err := timings.NewMonitor(ctx, "since://")

	if err != nil {
		return nil, fmt.Errorf("Failed to create timings monitor, %w", err)
	}

	app_timings := make([]*timings.SinceResponse, 0)

	r, wr := io.Pipe()

	scanner := bufio.NewScanner(r)

	err = m.Start(ctx, wr)

	if err != nil {
		return nil, fmt.Errorf("Failed to start timings monitor, %w", err)
	}

	mu := new(sync.RWMutex)

	sp := SpatialApplication{
		SpatialDatabase:  spatial_db,
		PropertiesReader: properties_reader,
		Iterator:         iter,
		Timings:          app_timings,
		Monitor:          m,
		mu:               mu,
	}

	go func() {

		for scanner.Scan() {

			go func(body []byte) {

				var tr *timings.SinceResponse
				err := json.Unmarshal(body, &tr)

				if err != nil {
					slog.Warn("Failed to decoder since response", "error", err)
					return
				}

				sp.mu.Lock()
				sp.Timings = append(sp.Timings, tr)
				sp.mu.Unlock()

			}(scanner.Bytes())
		}
	}()

	return &sp, nil
}

func (p *SpatialApplication) Close(ctx context.Context) error {
	p.SpatialDatabase.Disconnect(ctx)
	p.Monitor.Stop(ctx)
	return nil
}

func (p *SpatialApplication) IndexPaths(ctx context.Context, paths ...string) error {

	go func() {

		// TO DO: put this somewhere so that it can be triggered by signal(s)
		// to reindex everything in bulk or incrementally

		t1 := time.Now()

		err := p.Iterator.IterateURIs(ctx, paths...)

		if err != nil {
			slog.Error("failed to index paths", "error", err)
			os.Exit(1)
		}

		slog.Info("finished indexing", "time", time.Since(t1))
		debug.FreeOSMemory()
	}()

	// set up some basic monitoring and feedback stuff

	go func() {

		c := time.Tick(1 * time.Second)

		for _ = range c {

			if !p.Iterator.IsIndexing() {
				continue
			}

			slog.Info("indexing records", "indexed", p.Iterator.Seen)
		}
	}()

	return nil
}
