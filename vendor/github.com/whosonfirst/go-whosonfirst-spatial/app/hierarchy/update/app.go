package update

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/hierarchy"
	hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy/filter"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer/v3"
	"github.com/whosonfirst/go-writer/v3"
)

type UpdateApplicationOptions struct {
	Writer             writer.Writer
	WriterURI          string
	Exporter           export.Exporter
	ExporterURI        string
	MapshaperServerURI string
	SpatialDatabase    database.SpatialDatabase
	SpatialDatabaseURI string
	ToIterator         string
	FromIterator       string
	SPRFilterInputs    *filter.SPRInputs
	SPRResultsFunc     hierarchy_filter.FilterSPRResultsFunc                   // This one chooses one result among many (or nil)
	PIPUpdateFunc      hierarchy.PointInPolygonHierarchyResolverUpdateCallback // This one constructs a map[string]interface{} to update the target record (or not)
}

type UpdateApplicationPaths struct {
	To   []string
	From []string
}

type UpdateApplication struct {
	to                  string
	from                string
	tool                *hierarchy.PointInPolygonHierarchyResolver
	writer              writer.Writer
	exporter            export.Exporter
	spatial_db          database.SpatialDatabase
	sprResultsFunc      hierarchy_filter.FilterSPRResultsFunc
	sprFilterInputs     *filter.SPRInputs
	hierarchyUpdateFunc hierarchy.PointInPolygonHierarchyResolverUpdateCallback
	logger              *log.Logger
}

func (app *UpdateApplication) Run(ctx context.Context, paths *UpdateApplicationPaths) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// These are the data we are indexing to HIERARCHY from

	err := app.IndexSpatialDatabase(ctx, paths.From...)

	if err != nil {
		return err
	}

	// These are the data we are HIERARCHY-ing

	to_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read '%s', %v", path, err)
		}

		_, err = app.UpdateAndPublishFeature(ctx, body)

		if err != nil {
			return fmt.Errorf("Failed to update feature for '%s', %v", path, err)
		}

		return nil
	}

	to_iter, err := iterator.NewIterator(ctx, app.to, to_cb)

	if err != nil {
		return fmt.Errorf("Failed to create new HIERARCHY (to) iterator for input, %v", err)
	}

	err = to_iter.IterateURIs(ctx, paths.To...)

	if err != nil {
		return err
	}

	// This is important for something things like
	// whosonfirst/go-writer-featurecollection
	// (20210219/thisisaaronland)

	return app.writer.Close(ctx)
}

func (app *UpdateApplication) IndexSpatialDatabase(ctx context.Context, uris ...string) error {

	from_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		geom_type, err := geometry.Type(body)

		if err != nil {
			return fmt.Errorf("Failed to derive geometry type for %s, %w", path, err)
		}

		switch geom_type {
		case "Polygon", "MultiPolygon":
			return app.spatial_db.IndexFeature(ctx, body)
		default:
			return nil
		}
	}

	from_iter, err := iterator.NewIterator(ctx, app.from, from_cb)

	if err != nil {
		return fmt.Errorf("Failed to create spatial (from) iterator, %v", err)
	}

	err = from_iter.IterateURIs(ctx, uris...)

	if err != nil {
		return fmt.Errorf("Failed to iteratre URIs, %w", err)
	}

	return nil
}

// UpdateAndPublishFeature will invoke the `PointInPolygonAndUpdate` method using the `hierarchy.PointInPolygonHierarchyResolver` instance
// associated with 'app' using 'body' as its input. If successful and there are changes the result will be published using the `PublishFeature`
// method.
func (app *UpdateApplication) UpdateAndPublishFeature(ctx context.Context, body []byte) ([]byte, error) {

	has_changed, new_body, err := app.UpdateFeature(ctx, body)

	if err != nil {
		return nil, fmt.Errorf("Failed to update feature, %w", err)
	}

	if has_changed {

		new_body, err = app.PublishFeature(ctx, new_body)

		if err != nil {
			return nil, fmt.Errorf("Failed to publish feature, %w", err)
		}
	}

	return new_body, nil
}

// UpdateFeature will invoke the `PointInPolygonAndUpdate` method using the `hierarchy.PointInPolygonHierarchyResolver` instance
// associated with 'app' using 'body' as its input.
func (app *UpdateApplication) UpdateFeature(ctx context.Context, body []byte) (bool, []byte, error) {

	return app.tool.PointInPolygonAndUpdate(ctx, app.sprFilterInputs, app.sprResultsFunc, app.hierarchyUpdateFunc, body)
}

// PublishFeature exports 'body' using the `whosonfirst/go-writer/v3` instance associated with 'app'.
func (app *UpdateApplication) PublishFeature(ctx context.Context, body []byte) ([]byte, error) {

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
