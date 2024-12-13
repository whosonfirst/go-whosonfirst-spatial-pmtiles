package document

// Now that the "v1" spelunker has been retired, this code should be considered deprecated.

import (
	"context"
)

// PrepareSpelunkerV1Document prepares a Who's On First document for indexing with the
// "v1" Elasticsearch (v2.x) schema. For details please consult:
// https://github.com/whosonfirst/es-whosonfirst-schema/tree/master/schema/2.4
func PrepareSpelunkerV1Document(ctx context.Context, body []byte) ([]byte, error) {

	prepped, err := ExtractProperties(ctx, body)

	if err != nil {
		return nil, err
	}

	return AppendSpelunkerV1Properties(ctx, prepped)
}

// AppendSpelunkerV1Properties appends properties specific to the v1" Elasticsearch (v2.x) schema
// to a Who's On First document for. For details please consult:
// https://github.com/whosonfirst/es-whosonfirst-schema/tree/master/schema/2.4
func AppendSpelunkerV1Properties(ctx context.Context, body []byte) ([]byte, error) {

	var err error

	body, err = AppendNameStats(ctx, body)

	if err != nil {
		return nil, err
	}

	body, err = AppendConcordancesStats(ctx, body)

	if err != nil {
		return nil, err
	}

	body, err = AppendPlacetypeDetails(ctx, body)

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
