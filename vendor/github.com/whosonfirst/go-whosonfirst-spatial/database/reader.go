package database

import (
	"context"
	"fmt"
	"io"

	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
)

func IndexDatabaseWithReader(ctx context.Context, db SpatialDatabase, r io.Reader) error {

	index_func := func(ctx context.Context, body []byte, geom_type string) error {

		switch geom_type {
		case "Polygon", "MultiPolygon":
			return db.IndexFeature(ctx, body)
		default:
			return nil
		}
	}

	body, err := io.ReadAll(r)

	if err != nil {
		return fmt.Errorf("Failed to read document, %w", err)
	}

	geom_type, err := geometry.Type(body)

	if err == nil {
		return index_func(ctx, body, geom_type)
	}

	// Check to see if this is a FeatureCollection

	t_rsp := gjson.GetBytes(body, "type")

	if !t_rsp.Exists() || t_rsp.String() != "FeatureCollection" {
		return fmt.Errorf("Failed to derive geometry type for record, %w", err)
	}

	// If it is process each file

	fc, err := geojson.UnmarshalFeatureCollection(body)

	if err != nil {
		return fmt.Errorf("Failed to unmarshal record in to a FeatureCollection, %w", err)
	}

	for i, f := range fc.Features {

		f_body, err := f.MarshalJSON()

		if err != nil {
			return fmt.Errorf("Failed to marshal Feature at offset %d in record, %w", i, err)
		}

		geom_type := f.Geometry.GeoJSONType()

		err = index_func(ctx, f_body, geom_type)

		if err != nil {
			return fmt.Errorf("Failed to index Feature at offset %d in record, %w", i, err)
		}
	}

	return nil
}
