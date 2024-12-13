package document

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func AppendConcordancesMachineTags(ctx context.Context, body []byte) ([]byte, error) {

	root := gjson.ParseBytes(body)

	props_rsp := gjson.GetBytes(body, "properties")

	if props_rsp.Exists() {
		root = props_rsp
	}

	concordances_rsp := root.Get("wof:concordances")

	if !concordances_rsp.Exists() {
		return body, nil
	}

	tags := make([]string, 0)

	for k, v := range concordances_rsp.Map() {
		mt := fmt.Sprintf("%s=%s", k, v.String())
		tags = append(tags, mt)
	}

	updates := map[string]interface{}{
		"wof:concordances_machinetags": tags,
	}

	var err error

	for k, v := range updates {

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

// AppendConcordancesStats appends statistics about the `wof:concordances` properties in a Who's On First document.
// Specifically:
// * An array containing the set of source prefixes for concordances
// * The total number of concordances in a record.
func AppendConcordancesStats(ctx context.Context, body []byte) ([]byte, error) {

	root := gjson.ParseBytes(body)

	props_rsp := gjson.GetBytes(body, "properties")

	if props_rsp.Exists() {
		root = props_rsp
	}

	concordances_rsp := root.Get("wof:concordances")

	if !concordances_rsp.Exists() {
		return body, nil
	}

	sources := make([]string, 0)

	for k, _ := range concordances_rsp.Map() {
		sources = append(sources, k)
	}

	stats := map[string]interface{}{
		"wof:concordances_sources":  sources,
		"counts:concordances_total": len(sources),
	}

	var err error

	for k, v := range stats {

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
