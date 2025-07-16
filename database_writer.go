package pmtiles

// Implement the whosonfirst/go-writer/v3.Writer interface.

import (
	"context"
	"io"
	"log"

	"github.com/whosonfirst/go-whosonfirst-spatial"	
)

// Write implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance (by invoking the `IndexFeature` method).
func (r *PMTilesSpatialDatabase) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	return 0, spatial.ErrNotImplemented
}

// WriterURI implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance
func (r *PMTilesSpatialDatabase) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

// Flush implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *PMTilesSpatialDatabase) Flush(ctx context.Context) error {
	return nil
}

// Close implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *PMTilesSpatialDatabase) Close(ctx context.Context) error {
	return nil
}

// SetLogger implements the whosonfirst/go-writer interface so that the database itself can be used as a
// writer.Writer instance. This method is a no-op and simply returns `nil`.
func (r *PMTilesSpatialDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
