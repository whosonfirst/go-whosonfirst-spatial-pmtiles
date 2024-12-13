package document

import (
	"context"
	"errors"
	"fmt"
	_ "log"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
)

// AppendPlacetypeDetails appends addition properties related to the `wof:placetype` and `wof:placetype_alt` properties in a Who's On First record.
// Specifically:
// * The unique placetype ID for a placetype
// * The set of string names (including "alternate" placetypes) associated with a placetype
func AppendPlacetypeDetails(ctx context.Context, body []byte) ([]byte, error) {

	root := gjson.ParseBytes(body)

	props_rsp := gjson.GetBytes(body, "properties")

	if props_rsp.Exists() {
		root = props_rsp
	}

	pt_rsp := root.Get("wof:placetype")

	if !pt_rsp.Exists() {
		return nil, errors.New("Missing wof:placetype property")
	}

	str_pt := pt_rsp.String()

	if !placetypes.IsValidPlacetype(str_pt) {
		return body, nil
	}

	pt, err := placetypes.GetPlacetypeByName(str_pt)

	if err != nil {
		return nil, err
	}

	placetype_names := []string{
		pt.Name,
	}

	alt_rsp := root.Get("wof:placetype_alt")

	for _, a := range alt_rsp.Array(){
		placetype_names = append(placetype_names, a.String())
	}

	details := map[string]interface{}{
		"wof:placetype_id":    pt.Id,
		"wof:placetype_names": placetype_names,
	}

	for k, v := range details {

		path := k

		if props_rsp.Exists() {
			path = fmt.Sprintf("properties.%s", k)
		}

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return nil, err
		}
	}

	return body, nil
}
