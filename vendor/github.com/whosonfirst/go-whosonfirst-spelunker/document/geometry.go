package document

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// ExtractGeometry returns the "geometrry" element of a Who's On First document as a JSON-encoded byte array.
func ExtractGeometry(ctx context.Context, body []byte) ([]byte, error) {

	geom_rsp := gjson.GetBytes(body, "geometry")

	if !geom_rsp.Exists() {
		return nil, fmt.Errorf("Missing geometry element.")
	}

	geom_body, err := json.Marshal(geom_rsp.Value())

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal geometry element, %w", err)
	}

	return geom_body, nil

}
