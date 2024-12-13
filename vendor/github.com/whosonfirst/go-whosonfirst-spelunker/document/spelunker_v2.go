package document

import (
	"context"
	"fmt"
	_ "log/slog"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// PrepareSpelunkerV2Document prepares a Who's On First document for indexing with the
// "v2" OpenSearch (v2.x) schema. For details please consult:
func PrepareSpelunkerV2Document(ctx context.Context, body []byte) ([]byte, error) {

	props, err := ExtractProperties(ctx, body)

	if err != nil {
		return nil, err
	}

	geom, err := ExtractGeometry(ctx, body)

	if err != nil {
		return nil, err
	}

	type_rsp := gjson.GetBytes(geom, "type")

	if type_rsp.Exists() {

		props, err = sjson.SetBytes(props, "geom:type", type_rsp.String())

		if err != nil {
			return nil, fmt.Errorf("Failed to set geom:type, %w", err)
		}
	}

	return AppendSpelunkerV2Properties(ctx, props)
}

// AppendSpelunkerV2Properties appends properties specific to the v2" OpenSearch (v2.x) schema
// to a Who's On First document for. For details please consult:
func AppendSpelunkerV2Properties(ctx context.Context, body []byte) ([]byte, error) {

	var err error

	body, err = AppendNameStats(ctx, body)

	if err != nil {
		return nil, err
	}

	body, err = AppendConcordancesStats(ctx, body)

	if err != nil {
		return nil, err
	}

	body, err = AppendConcordancesMachineTags(ctx, body)

	if err != nil {
		return nil, err
	}

	body, err = AppendPlacetypeDetails(ctx, body)

	if err != nil {
		return nil, err
	}

	body, err = AppendExistentialDetails(ctx, body)

	if err != nil {
		return nil, err
	}

	body, err = AppendEDTFRanges(ctx, body)

	if err != nil {
		return nil, err
	}

	// to do: categories and machine tags...

	return body, nil
}
