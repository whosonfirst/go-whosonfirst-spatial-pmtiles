package update

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/hierarchy"
	hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy/filter"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer/v3"
	"github.com/whosonfirst/go-writer/v3"
)

// updateApplication is a struct to wrap the details of (optionally) populating a spatial
// database and updating the hierarchies of (n) files derived from an iterator including
// writing (publishing) the updated records.
type updateApplication struct {
	resolver            *hierarchy.PointInPolygonHierarchyResolver
	writer              writer.Writer
	exporter            export.Exporter
	export_opts         *export.Options
	spatial_db          database.SpatialDatabase
	sprResultsFunc      hierarchy_filter.FilterSPRResultsFunc
	sprFilterInputs     *filter.SPRInputs
	hierarchyUpdateFunc hierarchy.PointInPolygonHierarchyResolverUpdateCallback
}

func (app *updateApplication) Run(ctx context.Context, sources map[string][]string, targets map[string][]string) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// These are the data we are indexing hierarchies FROM

	sources_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		slog.Debug("Process source", "path", path)
		return database.IndexDatabaseWithReader(ctx, app.spatial_db, r)
	}

	// These are the data we are hierarchy-ing TO

	targets_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		slog.Debug("Process target", "path", path)

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read '%s', %v", path, err)
		}

		_, err = app.updateAndPublishFeature(ctx, body)

		if err != nil {
			return fmt.Errorf("Failed to update feature for '%s', %v", path, err)
		}

		return nil
	}

	iterate := func(ctx context.Context, iter_map map[string][]string, iter_cb emitter.EmitterCallbackFunc) error {

		for iter_uri, iter_sources := range iter_map {

			iter, err := iterator.NewIterator(ctx, iter_uri, iter_cb)

			if err != nil {
				return fmt.Errorf("Failed to create iterator for %s, %w", iter_uri, err)
			}

			err = iter.IterateURIs(ctx, iter_sources...)

			if err != nil {
				return fmt.Errorf("Failed to iterate sources for %s, %w", iter_uri, err)
			}
		}

		return nil
	}

	err := iterate(ctx, sources, sources_cb)

	if err != nil {
		return err
	}

	err = iterate(ctx, targets, targets_cb)

	if err != nil {
		return err
	}

	// This is important for something things like
	// whosonfirst/go-writer-featurecollection
	// (20210219/thisisaaronland)

	return app.writer.Close(ctx)
}

// UpdateAndPublishFeature will invoke the `PointInPolygonAndUpdate` method using the `hierarchy.PointInPolygonHierarchyResolver` instance
// associated with 'app' using 'body' as its input. If successful and there are changes the result will be published using the `PublishFeature`
// method.
func (app *updateApplication) updateAndPublishFeature(ctx context.Context, body []byte) ([]byte, error) {

	has_changed, new_body, err := app.updateFeature(ctx, body)

	if err != nil {
		return nil, fmt.Errorf("Failed to update feature, %w", err)
	}

	// But really, has the record _actually_ changed?

	if has_changed {

		has_changed, err = export.ExportChanged(new_body, body, app.export_opts, io.Discard)

		if err != nil {
			return nil, fmt.Errorf("Failed to determine if export has changed post update, %w", err)
		}
	}

	if has_changed {

		new_body, err = app.publishFeature(ctx, new_body)

		if err != nil {
			return nil, fmt.Errorf("Failed to publish feature, %w", err)
		}
	}

	return new_body, nil
}

// UpdateFeature will invoke the `PointInPolygonAndUpdate` method using the `hierarchy.PointInPolygonHierarchyResolver` instance
// associated with 'app' using 'body' as its input.
func (app *updateApplication) updateFeature(ctx context.Context, body []byte) (bool, []byte, error) {

	return app.resolver.PointInPolygonAndUpdate(ctx, app.sprFilterInputs, app.sprResultsFunc, app.hierarchyUpdateFunc, body)
}

// PublishFeature exports 'body' using the `whosonfirst/go-writer/v3` instance associated with 'app'.
func (app *updateApplication) publishFeature(ctx context.Context, body []byte) ([]byte, error) {

	new_body, err := app.exporter.Export(ctx, body)

	if err != nil {
		return nil, err
	}

	_, err = wof_writer.WriteBytes(ctx, app.writer, new_body)

	if err != nil {
		return nil, err
	}

	return new_body, nil
}
