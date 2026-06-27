package iterate

import (
	"context"
	"fmt"
	"io"
	"os"
)

// STDIN is a constant value signaling that a record was read from `STDIN` and has no URI (path).
const STDIN string = "STDIN"

// ReaderWithPath returns a new `io.ReadSeekCloser` instance derived from 'abs_path'.
func ReaderWithPath(ctx context.Context, abs_path string) (io.ReadSeekCloser, error) {

	if abs_path == STDIN {
		return os.Stdin, nil
	}

	r, err := os.Open(abs_path)

	if err != nil {
		return nil, fmt.Errorf("Failed to open %s, %w", abs_path, err)
	}

	return r, nil
}
