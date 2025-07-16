package database

// Implement the whosonfirst/go-writer/v3.Writer interface.

import (
	"context"
	"fmt"
	"io"
	"log"
)

func (r *RTreeSpatialDatabase) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {
	return 0, fmt.Errorf("Not implemented")
}

func (r *RTreeSpatialDatabase) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

func (r *RTreeSpatialDatabase) Flush(ctx context.Context) error {
	return nil
}

func (r *RTreeSpatialDatabase) Close(ctx context.Context) error {
	return nil
}

func (r *RTreeSpatialDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
