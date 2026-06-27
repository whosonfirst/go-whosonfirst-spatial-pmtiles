package database

// Implement the whosonfirst/go-writer/v3.Writer interface.

import (
	"context"
	"io"
	"log"
)

func (db *NullSpatialDatabase) Write(ctx context.Context, key string, r io.ReadSeeker) (int64, error) {
	return 0, nil
}

func (db *NullSpatialDatabase) Close(ctx context.Context) error {
	return nil
}

func (db *NullSpatialDatabase) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

func (db *NullSpatialDatabase) Flush(ctx context.Context) error {
	return nil
}

func (db *NullSpatialDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
