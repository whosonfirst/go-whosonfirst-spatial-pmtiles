package api

import (
	"encoding/json"
	"github.com/aaronland/go-http-sanitize"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip"
	spatial_app "github.com/whosonfirst/go-whosonfirst-spatial/app"
	"github.com/whosonfirst/go-whosonfirst-spr-geojson"
	"github.com/sfomuseum/go-timings"
	"log"
	"net/http"
)

const timingsPIPHandler string = "PIP handler"

const timingsPIPQuery string = "PIP handler query"

const timingsPIPFeatureCollection string = "PIP handler feature collection"

const timingsPIPProperties string = "PIP handler properties"

const GEOJSON string = "application/geo+json"

type PointInPolygonHandlerOptions struct {
	EnableGeoJSON bool
	Logger *log.Logger
	LogTimings bool
}

func PointInPolygonHandler(app *spatial_app.SpatialApplication, opts *PointInPolygonHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		if req.Method != "POST" {
			http.Error(rsp, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}

		if app.Iterator.IsIndexing() {
			http.Error(rsp, "Indexing records", http.StatusServiceUnavailable)
			return
		}

		app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPHandler)
		
		defer func(){

			app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPHandler)

			if opts.LogTimings {
				
				for _, t := range app.Timings {
					opts.Logger.Println(t)
				}
			}
		}()
		
		var pip_req *pip.PointInPolygonRequest

		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&pip_req)

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
		
		pip_rsp, err := pip.QueryPointInPolygon(ctx, app, pip_req)

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

		if len(pip_req.Properties) > 0 {
			
			props_opts := &spatial.PropertiesResponseOptions{
				Reader:       app.PropertiesReader,
				Keys:         pip_req.Properties,
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
