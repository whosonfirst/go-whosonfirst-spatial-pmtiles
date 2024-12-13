package document

import (
	"context"
	_ "log"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ...
func Flatten(ctx context.Context, body []byte) ([]byte, error) {

	var flattened []byte
	var err error

	rsp := gjson.ParseBytes(body)

	for _, details := range rsp.Map() {

		for k, v := range details.Map() {

			flattened, err = sjson.SetBytes(flattened, k, v.Value())

			if err != nil {
				return nil, err
			}
		}

	}

	return flattened, nil
}
