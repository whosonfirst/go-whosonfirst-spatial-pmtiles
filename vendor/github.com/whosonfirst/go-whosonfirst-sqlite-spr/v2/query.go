package spr

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aaronland/go-pagination"
	sql_pagination "github.com/aaronland/go-pagination-sql"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

// QueryPaginated will iterate over all the rows for 'q' in batches determined by 'pg_opts' and return a `spr.StandardPlacesResults` and `pagination.Results`
// instance for the results.
func QueryPaginated(ctx context.Context, conn *sql.DB, pg_opts pagination.Options, q string, args ...interface{}) (spr.StandardPlacesResults, pagination.Results, error) {

	rsp, err := sql_pagination.QueryPaginated(conn, pg_opts, q, args...)

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to query database, %w", err)
	}

	rows := rsp.Rows()
	pg := rsp.Results()

	spr_results := make([]spr.StandardPlacesResult, 0)

	for rows.Next() {

		result_spr, err := RetrieveSPRWithRows(ctx, rows)

		if err != nil {
			return nil, nil, fmt.Errorf("Failed to retrieve SPR from row, %w", err)
		}

		spr_results = append(spr_results, result_spr)
	}

	err = rows.Err()

	if err != nil {
		return nil, nil, fmt.Errorf("There was a problem retrieving rows from the database, %w", err)
	}

	results := &SQLiteResults{
		Places: spr_results,
	}

	return results, pg, nil
}
