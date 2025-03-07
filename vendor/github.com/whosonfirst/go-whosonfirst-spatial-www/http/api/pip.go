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

const timingsPIPHandler string = "PIP handler"

const timingsPIPQuery string = "PIP handler query"

const timingsPIPFeatureCollection string = "PIP handler feature collection"

const timingsPIPProperties string = "PIP handler properties"

type PointInPolygonHandlerOptions struct {
	EnableGeoJSON bool
	LogTimings    bool
}

func PointInPolygonHandler(app *spatial_app.SpatialApplication, opts *PointInPolygonHandlerOptions) (http.Handler, error) {

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

		app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPHandler)

		defer func() {

			app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPHandler)

			if opts.LogTimings {

				for _, t := range app.Timings {
					logger.Debug("Timings", "timing", t)
				}
			}
		}()

		pip_fn, err := query.NewSpatialFunction(ctx, "pip://")

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		var pip_query *query.SpatialQuery

		dec := json.NewDecoder(req.Body)
		err = dec.Decode(&pip_query)

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

		app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPQuery)

		pip_rsp, err := query.ExecuteQuery(ctx, app.SpatialDatabase, pip_fn, pip_query)

		app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPQuery)

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

			app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPFeatureCollection)

			err := geojson.AsFeatureCollection(ctx, pip_rsp, opts)

			app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPFeatureCollection)

			if err != nil {
				http.Error(rsp, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if len(pip_query.Properties) > 0 {

			props_opts := &spatial.PropertiesResponseOptions{
				Reader:       app.PropertiesReader,
				Keys:         pip_query.Properties,
				SourcePrefix: "properties",
			}

			app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPProperties)

			props_rsp, err := spatial.PropertiesResponseResultsWithStandardPlacesResults(ctx, props_opts, pip_rsp)

			app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPProperties)

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
		err = enc.Encode(pip_rsp)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	pip_handler := http.HandlerFunc(fn)
	return pip_handler, nil
}
