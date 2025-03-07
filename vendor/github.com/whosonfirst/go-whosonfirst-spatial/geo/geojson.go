package geo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

func BoundingBoxToFeature(str_bbox string, is_latlon bool) (*geojson.Feature, error) {

	str_bbox = strings.TrimSpace(str_bbox)
	str_parts := strings.Split(str_bbox, ",")

	if len(str_parts) != 4 {
		return nil, fmt.Errorf("Invalid number of parts, %d", len(str_parts))
	}

	parts := make([]float64, 4)

	for idx, str_i := range str_parts {

		str_i = strings.TrimSpace(str_i)

		i, err := strconv.ParseFloat(str_i, 64)

		if err != nil {
			return nil, fmt.Errorf("Invalid float '%s', %w", str_i, err)
		}

		parts[idx] = i
	}

	var min_x float64
	var min_y float64
	var max_x float64
	var max_y float64

	if is_latlon {
		min_x = parts[1]
		min_y = parts[0]
		max_x = parts[3]
		max_y = parts[2]
	} else {
		min_x = parts[0]
		min_y = parts[1]
		max_x = parts[2]
		max_y = parts[3]
	}

	min := orb.Point{min_x, min_y}
	max := orb.Point{max_x, max_y}

	b := orb.Bound{min, max}
	poly := b.ToPolygon()

	bbox := [4]float64{
		min.X(),
		min.Y(),
		max.X(),
		max.Y(),
	}

	f := geojson.NewFeature(poly)
	// Without this orb/geojson returns a null element for 'properties'
	// which makes tools like geojson.io sad...
	f.Properties = map[string]interface{}{"bbox": bbox}

	return f, nil
}
