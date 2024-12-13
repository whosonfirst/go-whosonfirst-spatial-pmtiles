package document

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// ExtractProperties returns the "properties" element of a Who's On First document as a JSON-encoded byte array.
func ExtractProperties(ctx context.Context, body []byte) ([]byte, error) {

	props_rsp := gjson.GetBytes(body, "properties")

	if !props_rsp.Exists() {
		return nil, fmt.Errorf("Missing properties element.")
	}

	props_body, err := json.Marshal(props_rsp.Value())

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal properties element, %w", err)
	}

	return props_body, nil

}
