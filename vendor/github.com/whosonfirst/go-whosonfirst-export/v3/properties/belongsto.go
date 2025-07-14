package properties

import (
	"context"
	"slices"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func EnsureBelongsTo(ctx context.Context, feature []byte) ([]byte, error) {

	belongsto := make([]int64, 0)

	wofid_rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_ID)

	if !wofid_rsp.Exists() {
		return nil, wof_properties.MissingProperty(wof_properties.PATH_WOF_ID)
	}

	wofid := wofid_rsp.Int()

	// Load the existing belongsto array, if it exists

	belongsToRsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_BELONGSTO)

	if belongsToRsp.Exists() {

		belongsToRsp.ForEach(func(key gjson.Result, value gjson.Result) bool {
			if value.Type == gjson.Number {
				id := value.Int()
				belongsto = append(belongsto, id)
			}

			return true
		})
	}

	rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_HIERARCHY)

	if rsp.Exists() {

		ids := make([]int64, 0)

		for _, h := range rsp.Array() {
			h.ForEach(func(key gjson.Result, value gjson.Result) bool {

				if value.Type == gjson.Number {

					id := value.Int()

					if id > 0 && id != wofid {
						ids = append(ids, id)
					}
				}

				return true
			})
		}

		// Add all the IDs we've not seen before
		for _, id := range ids {
			if !slices.Contains(belongsto, id) {
				belongsto = append(belongsto, id)
			}
		}

		// Remove all the IDs we no longer want in the list - in reverse,
		// because Golang.
		for i := len(belongsto) - 1; i >= 0; i-- {
			id := belongsto[i]

			if !slices.Contains(ids, id) {
				belongsto = append(belongsto[:i], belongsto[i+1:]...)
			}
		}
	}

	tz_rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_TIMEZONES)

	if tz_rsp.Exists() {

		for _, i := range tz_rsp.Array() {

			id := i.Int()
			if !slices.Contains(belongsto, id) {
				belongsto = append(belongsto, id)
			}
		}
	}

	feature, err := sjson.SetBytes(feature, wof_properties.PATH_WOF_BELONGSTO, belongsto)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_WOF_BELONGSTO, err)
	}

	return feature, nil
}
