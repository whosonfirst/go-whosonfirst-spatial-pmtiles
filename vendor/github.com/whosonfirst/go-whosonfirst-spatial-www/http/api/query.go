package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/whosonfirst/go-whosonfirst-spatial/query"
)

func SpatialQueryFromRequest(req *http.Request) (*query.SpatialQuery, error) {

	var q *query.SpatialQuery

	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&q)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode query, %w", err)
	}

	if q.Geometry == nil {
		return nil, fmt.Errorf("Query is missing geometry")
	}

	return q, nil
}
