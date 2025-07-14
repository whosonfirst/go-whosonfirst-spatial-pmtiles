package properties

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsureSrcGeom(ctx context.Context, feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, wof_properties.PATH_SRC_GEOM)

	if rsp.Exists() {
		return feature, nil
	}

	feature, err := sjson.SetBytes(feature, wof_properties.PATH_SRC_GEOM, "unknown")

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_SRC_GEOM, err)
	}

	return feature, nil
}

func EnsureGeomHash(ctx context.Context, feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, wof_properties.PATH_GEOMETRY)

	if !rsp.Exists() {
		return nil, wof_properties.MissingProperty(wof_properties.PATH_GEOMETRY)
	}

	enc, err := json.Marshal(rsp.Value())

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal %s property, %w", wof_properties.PATH_GEOMETRY, err)
	}

	hash := md5.Sum(enc)
	geom_hash := hex.EncodeToString(hash[:])

	feature, err = sjson.SetBytes(feature, wof_properties.PATH_WOF_GEOMHASH, geom_hash)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_WOF_GEOMHASH, err)
	}

	return feature, nil
}

func EnsureGeomCoords(ctx context.Context, feature []byte) ([]byte, error) {

	// https://github.com/paulmach/orb/blob/master/geojson/feature.go
	// https://github.com/paulmach/orb/blob/master/planar/area.go

	f, err := geojson.UnmarshalFeature(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal feature, %w", err)
	}

	centroid, area := planar.CentroidArea(f.Geometry)

	feature, err = sjson.SetBytes(feature, wof_properties.PATH_GEOM_LATITUDE, centroid.Y())

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_GEOM_LATITUDE, err)
	}

	feature, err = sjson.SetBytes(feature, wof_properties.PATH_GEOM_LONGITUDE, centroid.X())

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_GEOM_LONGITUDE, err)
	}

	feature, err = sjson.SetBytes(feature, wof_properties.PATH_GEOM_AREA, area)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_GEOM_AREA, err)
	}

	bounds := f.Geometry.Bound()

	min := bounds.Min
	max := bounds.Max

	minx := min.X()
	miny := min.Y()
	maxx := max.X()
	maxy := max.Y()

	bbox := []float64{
		minx,
		miny,
		maxx,
		maxy,
	}

	str_bbox := fmt.Sprintf("%.06f,%.06f,%.06f,%.06f", minx, miny, maxx, maxy)

	feature, err = sjson.SetBytes(feature, wof_properties.PATH_GEOM_BBOX, str_bbox)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_GEOM_BBOX, err)
	}

	feature, err = sjson.SetBytes(feature, wof_properties.PATH_BBOX, bbox)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_BBOX, err)
	}

	return feature, nil
}
