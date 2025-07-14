package properties

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

type Hierarchy map[string]int64
type Hierarchies []Hierarchy

func EnsureHierarchy(ctx context.Context, feature []byte) ([]byte, error) {

	pt_rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_PLACETYPE)

	if !pt_rsp.Exists() {
		return feature, wof_properties.MissingProperty(wof_properties.PATH_WOF_PLACETYPE)
	}

	pt := pt_rsp.String()

	id_rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_ID)

	if !id_rsp.Exists() {
		return feature, wof_properties.MissingProperty(wof_properties.PATH_WOF_ID)
	}

	id := id_rsp.Int()

	pt_keys := []string{
		fmt.Sprintf("%s_id", pt),
	}

	if pt == "custom" {

		alt_rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_PLACETYPE_ALT)

		for _, r := range alt_rsp.Array() {
			pt_keys = append(pt_keys, fmt.Sprintf("%s_id", r.String()))
		}
	}

	hierarchies := make([]Hierarchy, 0)

	hier_rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_HIERARCHY)

	if hier_rsp.Exists() {

		for _, possible := range hier_rsp.Array() {

			h := make(map[string]int64)

			for k, r := range possible.Map() {

				v, exists := h[k]

				if exists && v != r.Int() {
					return nil, fmt.Errorf("Hierarchy key '%s' already set with value '%d' (trying to set '%d')", k, v, r.Int())
				}

				h[k] = r.Int()
			}

			hierarchies = append(hierarchies, h)
		}
	}

	if len(hierarchies) == 0 {
		h := make(map[string]int64)
		hierarchies = append(hierarchies, h)
	}

	for _, h := range hierarchies {

		for _, k := range pt_keys {
			h[k] = id
		}
	}

	feature, err := sjson.SetBytes(feature, wof_properties.PATH_WOF_HIERARCHY, hierarchies)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_WOF_HIERARCHY, err)
	}

	return feature, nil
}
