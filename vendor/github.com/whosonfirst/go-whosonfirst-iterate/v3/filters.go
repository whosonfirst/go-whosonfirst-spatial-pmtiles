package iterate

import (
	"context"
	"io"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3/filters"
)

// ApplyFilters is a convenience methods to test whether 'r' matches all the filters defined
// by 'f' and also "rewinds" 'r' before returning.
func ApplyFilters(ctx context.Context, r io.ReadSeeker, f filters.Filters) (bool, error) {

	ok, err := f.Apply(ctx, r)

	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	_, err = r.Seek(0, 0)

	if err != nil {
		return true, err
	}

	return true, nil
}
