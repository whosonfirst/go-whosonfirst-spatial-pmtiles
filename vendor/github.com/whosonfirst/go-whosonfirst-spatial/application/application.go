package application

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
)

// SpatialApplication a bunch of different operations related to indexing and querying spatial
// data in to a single "container" struct with pointers to underlying instances (like a SpatialDatabase)
// as well as a handful of methods for automating common operations (like indexing records). To be
// honest, I am kind of ambivalent about this but it is handy for creating spatial "applications"
// (like point-in-polygon operations) when the underlying spatial database is an in-memory RTree and
// everything needs to be indexed before the operation in question can occur. When you create a new
// SpatialApplication instance (by calling `NewSpatialApplication`) here's what happens:
//   - A new `database.SpatialDatabase` instance is created and made available as a public 'SpatialDatabase' property .
//   - A new `reader.Reader` instance is created and made available a public 'PropertiesReader' property. This reader
//     is intended for use by methods like `PropertiesResponseResultsWithStandardPlacesResults` which is what appends
//     custom  properties to SPR responses (for example, in a point-in-polygon result set).
//   - If custom placetypes are defined these will be loaded an appended to the default `whosonfirst/go-whosonfirst-placetype`
//     specification. This can be useful if you are working with non-standard Who's On First style records that define
//     their own placetypes.
type SpatialApplication struct {
	SpatialDatabase  database.SpatialDatabase
	PropertiesReader reader.Reader
	Timings          []*timings.SinceResponse
	Monitor          timings.Monitor
	mu               *sync.RWMutex
	indexing         int64
	indexed          int64
}

// SpatialApplicationOptions defines properties used to instantiate a new `SpatialApplication` instance.
type SpatialApplicationOptions struct {
	// A valid `whosonfirst/go-whosonfirst-spatial/database` URI.
	SpatialDatabaseURI string
	// A valid `whosonfirst/go-reader` URI.
	PropertiesReaderURI string
	// EnableCustomPlacetypes signals that custom placetypes should be appended to the default placetype specification.
	EnableCustomPlacetypes bool
	// A JSON-encoded `whosonfirst/go-whosonfirst-placetypes.WOFPlacetypeSpecification` definition with custom placetypes.
	CustomPlacetypes string
}

// NewSpatialApplication returns a new `SpatialApplication` instance.
func NewSpatialApplication(ctx context.Context, opts *SpatialApplicationOptions) (*SpatialApplication, error) {

	spatial_db, err := database.NewSpatialDatabase(ctx, opts.SpatialDatabaseURI)

	if err != nil {
		return nil, fmt.Errorf("Failed instantiate spatial database, %v", err)
	}

	// Set up properties reader

	var properties_reader reader.Reader

	if opts.PropertiesReaderURI != "" {

		use_spatial_uri := "{spatial-database-uri}"

		if opts.PropertiesReaderURI == use_spatial_uri {
			opts.PropertiesReaderURI = opts.SpatialDatabaseURI
		}

		r, err := reader.NewReader(ctx, opts.PropertiesReaderURI)

		if err != nil {
			return nil, fmt.Errorf("Failed to create properties reader, %v", err)
		}

		properties_reader = r
	}

	if properties_reader == nil {
		properties_reader = spatial_db
	}

	// Enable custom placetypes

	if opts.EnableCustomPlacetypes {

		custom_placetypes := opts.CustomPlacetypes

		// Alternate sources for custom placetypes are not supported yet - once they
		// are the corresponding flag in the flags/common.go package should be reenabled
		// (20210324/thisisaaronland)

		custom_placetypes_source := ""

		var custom_reader io.Reader

		if custom_placetypes_source == "" {
			custom_reader = strings.NewReader(custom_placetypes)
		} else {
			// whosonfirst/go-reader or ... ?
		}

		spec, err := placetypes.NewWOFPlacetypeSpecificationWithReader(custom_reader)

		if err != nil {
			return nil, fmt.Errorf("Failed to create place specification with reader, %w", err)
		}

		err = placetypes.AppendPlacetypeSpecification(spec)

		if err != nil {
			return nil, fmt.Errorf("Failed to append placetypes specification, %w", err)
		}
	}

	// START OF set up monitor (for tracking indexing)

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

	// END OF set up monitor (for tracking indexing)

	mu := new(sync.RWMutex)

	sp := SpatialApplication{
		SpatialDatabase:  spatial_db,
		PropertiesReader: properties_reader,
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

// Close() will terminate an spatial database connections and stop any internal timing monitors.
func (p *SpatialApplication) Close(ctx context.Context) error {
	p.SpatialDatabase.Disconnect(ctx)
	p.Monitor.Stop(ctx)
	return nil
}

func (p *SpatialApplication) IndexDatabaseWithIterators(ctx context.Context, sources map[string][]string) error {
	return database.IndexDatabaseWithIterators(ctx, p.SpatialDatabase, sources)
}

func (p *SpatialApplication) IsIndexing() bool {

	if atomic.LoadInt64(&p.indexing) > 0 {
		return true
	}

	return false
}

func (p *SpatialApplication) Indexed() int64 {
	return atomic.LoadInt64(&p.indexed)
}
