package reader

import (
	"context"
	"fmt"
	"net/url"
	"os"
)

func init() {

	ctx := context.Background()

	err := RegisterReader(ctx, "cwd", NewCwdReader)

	if err != nil {
		panic(err)
	}

}

// NewFileReader returns a new `FileReader` instance for reading documents from the current
// working directory, configured by 'uri' in the form of:
//
//	cwd://
func NewCwdReader(ctx context.Context, uri string) (Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	cwd, err := os.Getwd()

	if err != nil {
		return nil, fmt.Errorf("Failed to determine current working directory, %v", err)
	}

	u.Scheme = "fs"
	u.Path = cwd

	return NewFileReader(ctx, u.String())
}
