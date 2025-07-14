package iterate

import (
	"io"
)

// Record is a struct wrapping the details of records processed by a `whosonfirst/go-whosonfirst-iterate/v3.Iterator` instance.
type Record struct {
	// Path is the URI of the record. This will vary from one `whosonfirst/go-whosonfirst-iterate/v3.Iterator`
	// implementation to the next.
	Path string
	// Body is an `io.ReadSeekCloser` containing the body of the record.
	Body io.ReadSeekCloser
}

// NewRecord returns a new `Record` instance wrapping 'path' and 'r'.
func NewRecord(path string, r io.ReadSeekCloser) *Record {

	rec := &Record{
		Path: path,
		Body: r,
	}

	return rec
}
