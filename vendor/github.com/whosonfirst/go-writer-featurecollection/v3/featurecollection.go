package featurecollection

import (
	"context"
	"fmt"
	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-writer/v3"
	"io"
	"log"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

func init() {

	ctx := context.Background()

	err := writer.RegisterWriter(ctx, "featurecollection", NewFeatureCollectionWriter)

	if err != nil {
		panic(err)
	}
}

type FeatureCollectionWriter struct {
	writer.Writer
	writer writer.Writer
	mu     *sync.RWMutex
	count  int64
}

func NewFeatureCollectionWriter(ctx context.Context, uri string) (writer.Writer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	wr_uri := q.Get("writer")

	if wr_uri == "" {
		return nil, fmt.Errorf("Missing ?writer= parameter")
	}

	wr, err := writer.NewWriter(ctx, wr_uri)

	if err != nil {
		return nil, err
	}

	mu := new(sync.RWMutex)

	fc := &FeatureCollectionWriter{
		writer: wr,
		mu:     mu,
		count:  int64(0),
	}

	return fc, nil
}

func (fc *FeatureCollectionWriter) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	body, err := io.ReadAll(fh)

	if err != nil {
		return 0, err
	}

	_, err = geojson.UnmarshalFeature(body)

	if err != nil {
		return 0, err
	}

	fc.mu.Lock()

	defer func() {
		fc.mu.Unlock()
		atomic.AddInt64(&fc.count, 1)
	}()

	var preamble string

	if atomic.LoadInt64(&fc.count) == 0 {
		preamble = `{"type":"FeatureCollection", "features":[`
	} else {
		preamble = `,`
	}

	sr := strings.NewReader(preamble + string(body))

	return fc.writer.Write(ctx, key, sr)
}

func (fc *FeatureCollectionWriter) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

func (fc *FeatureCollectionWriter) Flush(ctx context.Context) error {
	return nil
}

func (fc *FeatureCollectionWriter) Close(ctx context.Context) error {

	body := `]}`

	if atomic.LoadInt64(&fc.count) == 0 {
		body = `{"type":"FeatureCollection", "features":[]}`
	}

	sr := strings.NewReader(body)
	_, err := fc.writer.Write(ctx, "", sr)

	if err != nil {
		return err
	}

	return nil
}

func (fc *FeatureCollectionWriter) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
