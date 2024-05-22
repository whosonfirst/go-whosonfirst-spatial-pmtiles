package pip

import (
	"context"
	"fmt"

	"github.com/sfomuseum/go-timings"
	spatial_app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-spr/v2/sort"
)

const timingsPIPQuery string = "PIP query"

const timingsPIPQueryPointInPolygon string = "PIP query point in polygon"

const timingsPIPQuerySort string = "PIP query sort"

func QueryPointInPolygon(ctx context.Context, app *spatial_app.SpatialApplication, req *PointInPolygonRequest) (spr.StandardPlacesResults, error) {

	app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPQuery)
	defer app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPQuery)

	c, err := geo.NewCoordinate(req.Longitude, req.Latitude)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new coordinate, %v", err)
	}

	f, err := NewSPRFilterFromPointInPolygonRequest(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to create point in polygon filter from request, %w", err)
	}

	var principal_sorter sort.Sorter
	var follow_on_sorters []sort.Sorter

	for idx, uri := range req.Sort {

		s, err := sort.NewSorter(ctx, uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create sorter for '%s', %w", uri, err)
		}

		if idx == 0 {
			principal_sorter = s
		} else {
			follow_on_sorters = append(follow_on_sorters, s)
		}
	}

	app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPQueryPointInPolygon)

	db := app.SpatialDatabase
	rsp, err := db.PointInPolygon(ctx, c, f)

	app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPQueryPointInPolygon)

	if err != nil {
		return nil, fmt.Errorf("Failed to perform point in polygon query, %w", err)
	}

	if principal_sorter != nil {

		app.Monitor.Signal(ctx, timings.SinceStart, timingsPIPQuerySort)

		sorted, err := principal_sorter.Sort(ctx, rsp, follow_on_sorters...)

		app.Monitor.Signal(ctx, timings.SinceStop, timingsPIPQuerySort)

		if err != nil {
			return nil, fmt.Errorf("Failed to sort results, %w", err)
		}

		rsp = sorted
	}

	app.Monitor.Signal(ctx, "complete point in polygon")
	return rsp, nil
}
