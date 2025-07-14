package export

import (
	"context"
	"fmt"
	"net/url"
)

type WhosOnFirstExporter struct {
	Exporter
}

func init() {

	ctx := context.Background()

	err := RegisterExporter(ctx, "whosonfirst", NewWhosOnFirstExporter)

	if err != nil {
		panic(err)
	}
}

func NewWhosOnFirstExporter(ctx context.Context, uri string) (Exporter, error) {

	_, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	ex := WhosOnFirstExporter{}
	return &ex, nil
}

func (ex *WhosOnFirstExporter) Export(ctx context.Context, feature []byte) (bool, []byte, error) {
	return Export(ctx, feature)
}
