package update

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/sfomuseum/go-sfomuseum-mapshaper"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/hierarchy"
	"github.com/whosonfirst/go-writer/v3"
)

func Run(ctx context.Context, logger *log.Logger) error {

	fs, err := DefaultFlagSet(ctx)

	if err != nil {
		fmt.Errorf("Failed to create application flag set, %w", err)
	}

	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	opts, err := RunOptionsFromFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to derive run options, %w", err)
	}

	return RunWithOptions(ctx, opts, logger)
}

func RunWithOptions(ctx context.Context, opts *RunOptions, logger *log.Logger) error {

	// Note that the bulk of this method is simply taking opts and using it to
	// instantiate all the different pieces necessary for the updateApplication
	// type to actually do the work of updating hierarchies.

	var ex export.Exporter
	var wr writer.Writer
	var spatial_db database.SpatialDatabase

	if opts.Exporter != nil {
		ex = opts.Exporter
	} else {

		_ex, err := export.NewExporter(ctx, opts.ExporterURI)

		if err != nil {
			return fmt.Errorf("Failed to create exporter for '%s', %v", opts.ExporterURI, err)
		}

		ex = _ex
	}

	// In addition to the "exporter" we also create a default options instance
	// that is used by the updateApplication instance to do a final check whether
	// or not records have actually been updated (beyond just incrementing the
	// lastmodified date).

	export_opts, err := export.NewDefaultOptions(ctx)

	if err != nil {
		return fmt.Errorf("Failed to create export options, %w", err)
	}

	if opts.Writer != nil {
		wr = opts.Writer
	} else {
		_wr, err := writer.NewWriter(ctx, opts.WriterURI)

		if err != nil {
			return fmt.Errorf("Failed to create writer for '%s', %v", opts.WriterURI, err)
		}

		wr = _wr
	}

	if opts.SpatialDatabase != nil {
		spatial_db = opts.SpatialDatabase
	} else {

		_db, err := database.NewSpatialDatabase(ctx, opts.SpatialDatabaseURI)

		if err != nil {
			return fmt.Errorf("Failed to create spatial database for '%s', %v", opts.SpatialDatabaseURI, err)
		}

		spatial_db = _db
	}

	// All of this mapshaper stuff can't be retired/replaced fast enough...
	// (20210222/thisisaaronland)

	var ms_client *mapshaper.Client

	if opts.MapshaperServerURI != "" {

		// Set up mapshaper endpoint (for deriving centroids during PIP operations)
		// Make sure it's working

		client, err := mapshaper.NewClient(ctx, opts.MapshaperServerURI)

		if err != nil {
			return fmt.Errorf("Failed to create mapshaper client for '%s', %v", opts.MapshaperServerURI, err)
		}

		ok, err := client.Ping()

		if err != nil {
			return fmt.Errorf("Failed to ping '%s', %v", opts.MapshaperServerURI, err)
		}

		if !ok {
			return fmt.Errorf("'%s' returned false", opts.MapshaperServerURI)
		}

		ms_client = client
	}

	update_cb := opts.PIPUpdateFunc

	if update_cb == nil {
		update_cb = hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()
	}

	resolver_opts := &hierarchy.PointInPolygonHierarchyResolverOptions{
		Database:  spatial_db,
		Mapshaper: ms_client,
	}

	resolver, err := hierarchy.NewPointInPolygonHierarchyResolver(ctx, resolver_opts)

	if err != nil {
		return fmt.Errorf("Failed to create PIP tool, %v", err)
	}

	// This is where the actual work happens

	app := &updateApplication{
		to:                  opts.ToIterator,
		from:                opts.FromIterator,
		spatial_db:          spatial_db,
		resolver:            resolver,
		exporter:            ex,
		export_opts:         export_opts,
		writer:              wr,
		sprFilterInputs:     opts.SPRFilterInputs,
		sprResultsFunc:      opts.SPRResultsFunc,
		hierarchyUpdateFunc: update_cb,
	}

	paths := &updateApplicationPaths{
		To:   opts.To,
		From: opts.From,
	}

	return app.Run(ctx, paths)
}
