package sqlite

// Implement the whosonfirst/go-reader/v2.Reader interface.

import (
	"context"
	"fmt"
	"io"
	"strings"
	"database/sql"
	
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

// Read implements the whosonfirst/go-reader interface so that the database itself can be used as a
// reader.Reader instance (reading features from the `geojson` table.
func (r *SQLiteSpatialDatabase) Read(ctx context.Context, str_uri string) (io.ReadSeekCloser, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return nil, err
	}

	// TO DO : ALT STUFF HERE

	q := fmt.Sprintf("SELECT body FROM %s WHERE id = ?", r.geojson_table.Name())

	row := r.db.QueryRowContext(ctx, q, id)

	var body string

	err = row.Scan(&body)

	if err != nil {
		return nil, err
	}

	sr := strings.NewReader(body)
	fh, err := ioutil.NewReadSeekCloser(sr)

	if err != nil {
		return nil, err
	}

	return fh, nil
}

// Exists returns a boolean value indicating whether 'str_uri` exists.
func (r *SQLiteSpatialDatabase) Exists(ctx context.Context, str_uri string) (bool, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return false, err
	}

	// TO DO : ALT STUFF HERE

	q := fmt.Sprintf("SELECT 1 FROM %s WHERE id = ?", r.geojson_table.Name())

	row := r.db.QueryRowContext(ctx, q, id)

	var one int

	err = row.Scan(&one)

	if err != nil {

		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}
	
// ReadURI implements the whosonfirst/go-reader interface so that the database itself can be used as a
// reader.Reader instance
func (r *SQLiteSpatialDatabase) ReaderURI(ctx context.Context, str_uri string) string {
	return str_uri
}
