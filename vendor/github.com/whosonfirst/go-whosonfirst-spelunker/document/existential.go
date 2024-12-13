package document

import (
	"context"
	"fmt"
	_ "log/slog"
	
	"github.com/sfomuseum/go-edtf"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func AppendExistentialDetails(ctx context.Context, body []byte) ([]byte, error) {

	root := gjson.ParseBytes(body)
	props_rsp := gjson.GetBytes(body, "properties")

	if props_rsp.Exists() {
		root = props_rsp
	}

	is_deprecated := 0
	is_ceased := 0

	deprecated_rsp := root.Get("edtf:deprecated")

	if deprecated_rsp.Exists() {

		if deprecated_rsp.String() != edtf.UNKNOWN && deprecated_rsp.String() != edtf.UNKNOWN_2012 {
			is_deprecated = 1
		}
	}

	ceased_rsp := root.Get("edtf:cessation")

	if ceased_rsp.Exists() {
		
		if ceased_rsp.String() != edtf.UNKNOWN && ceased_rsp.String() != edtf.UNKNOWN_2012 {
			is_ceased = 1
		}
	}

	to_assign := map[string]interface{}{
		"mz:is_deprecated": is_deprecated,
		"mz:is_ceased":     is_ceased,
	}

	var err error

	for k, v := range to_assign {

		path := k

		if props_rsp.Exists() {
			path = fmt.Sprintf("properties.%s", k)
		}

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return nil, fmt.Errorf("Failed to assign %s (%d), %w", path, v, err)
		}
	}

	return body, nil
}
