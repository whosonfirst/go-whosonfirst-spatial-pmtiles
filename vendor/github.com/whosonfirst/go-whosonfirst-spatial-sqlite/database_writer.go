package sqlite

// Implement the whosonfirst/go-writer/v3.Writer interface.

import (
	"context"
	"io"
	"log"
)

// Write implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance (by invoking the `IndexFeature` method).
func (r *SQLiteSpatialDatabase) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	body, err := io.ReadAll(fh)

	if err != nil {
		return 0, err
	}

	err = r.IndexFeature(ctx, body)

	if err != nil {
		return 0, err
	}

	return int64(len(body)), nil
}

// WriterURI implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance
func (r *SQLiteSpatialDatabase) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

// Flush implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *SQLiteSpatialDatabase) Flush(ctx context.Context) error {
	return nil
}

// Close implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *SQLiteSpatialDatabase) Close(ctx context.Context) error {
	return r.db.Close()
}

// SetLogger implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *SQLiteSpatialDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
