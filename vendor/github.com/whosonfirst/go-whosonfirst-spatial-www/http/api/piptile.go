package api

/*

$> curl -s -X POST http://localhost:8080/api/point-in-polygon-with-tile \
	-d '{ "is_current": [ 1 ], "tile": { "zoom": 14, "x": 2622, "y": 6341 } }' \
	| jq -r '.features[]["properties"]["mz:is_current"]' \
	| wc -l

180

*/

import (
	"encoding/json"
	"log/slog"
	"net/http"

	spatial_app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	"github.com/whosonfirst/go-whosonfirst-spatial/maptile"
	"github.com/whosonfirst/go-whosonfirst-spatial/query"
)

type PointInPolygonTileHandlerOptions struct{}

func PointInPolygonTileHandler(app *spatial_app.SpatialApplication, opts *PointInPolygonTileHandlerOptions) (http.Handler, error) {

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

		var tile_query *query.MapTileSpatialQuery

		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&tile_query)

		if err != nil {
			logger.Error("Failed to decode map tile spatial query", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if tile_query.Tile == nil {
			logger.Debug("Tile query is missing tile.")
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}
		
		map_t, err := tile_query.MapTile()

		if err != nil {
			logger.Error("Failed to derive map tile from query", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		spatial_q := tile_query.SpatialQuery()

		fc, err := maptile.PointInPolygonCandidateFeaturesFromTile(ctx, app.SpatialDatabase, spatial_q, map_t)

		if err != nil {
			logger.Error("Failed to derive candidate features from map tile", "error", err)
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-type", GEOJSON)

		enc := json.NewEncoder(rsp)
		err = enc.Encode(fc)

		if err != nil {
			logger.Error("Failed to marshal FeatureCollection results", "error", err)
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	piptile_handler := http.HandlerFunc(fn)
	return piptile_handler, nil
}
