package app

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"io"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

type SpatialApplication struct {
	mode             string
	SpatialDatabase  database.SpatialDatabase
	PropertiesReader reader.Reader
	Iterator         *iterator.Iterator
	Logger           *log.Logger
	Timings          []*timings.SinceResponse
	Monitor          timings.Monitor
	mu               *sync.RWMutex
}

func NewSpatialApplicationWithFlagSet(ctx context.Context, fl *flag.FlagSet) (*SpatialApplication, error) {

	logger, err := NewApplicationLoggerWithFlagSet(ctx, fl)

	if err != nil {
		return nil, err
	}

	spatial_db, err := NewSpatialDatabaseWithFlagSet(ctx, fl)

	if err != nil {
		return nil, fmt.Errorf("Failed instantiate spatial database, %v", err)
	}

	properties_r, err := NewPropertiesReaderWithFlagsSet(ctx, fl)

	if err != nil {
		return nil, fmt.Errorf("Failed to create properties reader, %v", err)
	}

	if properties_r == nil {
		properties_r = spatial_db
	}

	iter, err := NewIteratorWithFlagSet(ctx, fl, spatial_db)

	if err != nil {
		return nil, fmt.Errorf("Failed to instantiate iterator, %v", err)
	}

	err = AppendCustomPlacetypesWithFlagSet(ctx, fl)

	if err != nil {
		return nil, fmt.Errorf("Failed to append custom placetypes, %v", err)
	}

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
		PropertiesReader: properties_r,
		Iterator:         iter,
		Logger:           logger,
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
					logger.Printf("Failed to decoder since response, %w", err)
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
			p.Logger.Fatalf("failed to index paths because %s", err)
		}

		t2 := time.Since(t1)

		p.Logger.Printf("finished indexing in %v", t2)
		debug.FreeOSMemory()
	}()

	// set up some basic monitoring and feedback stuff

	go func() {

		c := time.Tick(1 * time.Second)

		for _ = range c {

			if !p.Iterator.IsIndexing() {
				continue
			}

			p.Logger.Printf("indexing %d records indexed", p.Iterator.Seen)
		}
	}()

	return nil
}
