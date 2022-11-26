// package geometry provides methods for deriving and interpreting Who's On First geometries.
package geometry

import (
	"fmt"
	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/gjson"
)

// Geometry() will return a `paulmach/orb/geojson.Geometry` instance derived from 'body'.
func Geometry(body []byte) (*geojson.Geometry, error) {

	rsp := gjson.GetBytes(body, "geometry")

	if !rsp.Exists() {
		return nil, fmt.Errorf("Failed to derive geometry for feature")
	}

	str_geom := rsp.String()

	geom, err := geojson.UnmarshalGeometry([]byte(str_geom))

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal geometry for feature, %w", err)
	}

	return geom, nil
}
