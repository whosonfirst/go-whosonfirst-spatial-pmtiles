package geo

import (
	"fmt"
	"math"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// UpdateLineStringIfCrossingAntimeridian will rewrite GeoJSON geometry LineString coordinates as MultiLineString
// coordinates with (2) or more LineString elements if the origin path crosses the antimeridian. For example: flights
// From Hong Kong to San Francisco.
func UpdateLineStringIfCrossingAntimeridian(body []byte) (bool, []byte, error) {

	type_rsp := gjson.GetBytes(body, "geometry.type")

	if type_rsp.String() != "LineString" {
		return false, nil, fmt.Errorf("Geometry is not LineString")
	}

	linestrings := make([][][2]float64, 0)
	current := make([][2]float64, 0)

	coords_rsp := gjson.GetBytes(body, "geometry.coordinates")

	var last_lon float64

	for idx, pt_rsp := range coords_rsp.Array() {

		lon := pt_rsp.Get("0").Float()
		lat := pt_rsp.Get("1").Float()

		if idx > 0 {

			crossing := IsCrossingAntimeridian(last_lon, lon)

			if crossing {
				linestrings = append(linestrings, current)
				current = make([][2]float64, 0)
			}

		}

		current = append(current, [2]float64{lon, lat})
		last_lon = lon
	}

	linestrings = append(linestrings, current)

	if len(linestrings) == 1 {
		return false, body, nil
	}

	var err error

	body, err = sjson.SetBytes(body, "geometry.type", "MultiLineString")

	if err != nil {
		return true, nil, err
	}

	body, err = sjson.SetBytes(body, "geometry.coordinates", linestrings)

	if err != nil {
		return true, nil, err
	}

	return true, body, nil
}

// IsCrossingAntimeridian determines whether two longitudes cross the antimeridian.
func IsCrossingAntimeridian(lon1, lon2 float64) bool {
	return math.Abs(lon1-lon2) > 180 || math.Abs(lon2-lon1) > 180
}
