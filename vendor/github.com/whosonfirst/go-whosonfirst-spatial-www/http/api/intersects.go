package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/aaronland/go-http-sanitize"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	spatial_app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	"github.com/whosonfirst/go-whosonfirst-spatial/query"
	"github.com/whosonfirst/go-whosonfirst-spr-geojson"
)

const timingsIntersectsHandler string = "PIP handler"

const timingsIntersectsQuery string = "PIP handler query"

const timingsIntersectsFeatureCollection string = "PIP handler feature collection"

const timingsIntersectsProperties string = "PIP handler properties"

type IntersectsHandlerOptions struct {
	EnableGeoJSON bool
	LogTimings    bool
}

func IntersectsHandler(app *spatial_app.SpatialApplication, opts *IntersectsHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		logger := slog.Default()

		ctx := req.Context()

		if req.Method != "POST" {
			http.Error(rsp, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}

		if app.IsIndexing() {
			http.Error(rsp, "Indexing records", http.StatusServiceUnavailable)
			return
		}

		app.Monitor.Signal(ctx, timings.SinceStart, timingsIntersectsHandler)

		defer func() {

			app.Monitor.Signal(ctx, timings.SinceStop, timingsIntersectsHandler)

			if opts.LogTimings {

				for _, t := range app.Timings {
					logger.Debug("Timings", "timing", t)
				}
			}
		}()

		intersects_fn, err := query.NewSpatialFunction(ctx, "intersects://")

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		var intersects_query *query.SpatialQuery

		dec := json.NewDecoder(req.Body)
		err = dec.Decode(&intersects_query)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		accept, err := sanitize.HeaderString(req, "Accept")

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		if accept == GEOJSON && !opts.EnableGeoJSON {
			http.Error(rsp, "GeoJSON output is not supported", http.StatusBadRequest)
			return
		}

		app.Monitor.Signal(ctx, timings.SinceStart, timingsIntersectsQuery)

		intersects_rsp, err := query.ExecuteQuery(ctx, app.SpatialDatabase, intersects_fn, intersects_query)

		app.Monitor.Signal(ctx, timings.SinceStop, timingsIntersectsQuery)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		if opts.EnableGeoJSON && accept == GEOJSON {

			app.Monitor.Signal(ctx, "Start PIP handler feature collection")

			opts := &geojson.AsFeatureCollectionOptions{
				Reader: app.SpatialDatabase,
				Writer: rsp,
			}

			app.Monitor.Signal(ctx, timings.SinceStart, timingsIntersectsFeatureCollection)

			err := geojson.AsFeatureCollection(ctx, intersects_rsp, opts)

			app.Monitor.Signal(ctx, timings.SinceStop, timingsIntersectsFeatureCollection)

			if err != nil {
				http.Error(rsp, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if len(intersects_query.Properties) > 0 {

			props_opts := &spatial.PropertiesResponseOptions{
				Reader:       app.PropertiesReader,
				Keys:         intersects_query.Properties,
				SourcePrefix: "properties",
			}

			app.Monitor.Signal(ctx, timings.SinceStart, timingsIntersectsProperties)

			props_rsp, err := spatial.PropertiesResponseResultsWithStandardPlacesResults(ctx, props_opts, intersects_rsp)

			app.Monitor.Signal(ctx, timings.SinceStop, timingsIntersectsProperties)

			if err != nil {
				http.Error(rsp, err.Error(), http.StatusInternalServerError)
				return
			}

			enc := json.NewEncoder(rsp)
			err = enc.Encode(props_rsp)

			if err != nil {
				http.Error(rsp, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		enc := json.NewEncoder(rsp)
		err = enc.Encode(intersects_rsp)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	intersects_handler := http.HandlerFunc(fn)
	return intersects_handler, nil
}
