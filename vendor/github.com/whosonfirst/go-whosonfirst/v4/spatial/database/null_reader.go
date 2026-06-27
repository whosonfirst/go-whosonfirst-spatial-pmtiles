package database

// Implement the whosonfirst/go-reader/v2.Reader interface.

import (
	"context"
	"fmt"
	"io"
)

func (db *NullSpatialDatabase) Read(ctx context.Context, str_uri string) (io.ReadSeekCloser, error) {
	return nil, fmt.Errorf("Not found")
}

func (db *NullSpatialDatabase) Exists(ctx context.Context, str_uri string) (bool, error) {
	return false, nil
}

func (db *NullSpatialDatabase) ReaderURI(ctx context.Context, str_uri string) string {
	return str_uri
}
