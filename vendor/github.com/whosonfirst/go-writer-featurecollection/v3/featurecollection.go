package featurecollection

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-writer/v3"	
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
	closed bool
}

func NewFeatureCollectionWriter(ctx context.Context, uri string) (writer.Writer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	wr_uri := q.Get("writer")

	if wr_uri == "" {
		return nil, fmt.Errorf("Missing ?writer= parameter")
	}

	wr, err := writer.NewWriter(ctx, wr_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create writer for '%s', %w", wr_uri, err)
	}

	mu := new(sync.RWMutex)

	fc := &FeatureCollectionWriter{
		writer: wr,
		mu:     mu,
		count:  int64(0),
	}

	return fc, nil
}

func NewFeatureCollectionWriterWithWriter(ctx context.Context, wr io.Writer) (writer.Writer, error) {

	io_wr, err := writer.NewIOWriterWithWriter(ctx, wr)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new IOWriter, %w", err)
	}

	mu := new(sync.RWMutex)

	fc := &FeatureCollectionWriter{
		writer: io_wr,
		mu:     mu,
		count:  int64(0),
	}

	return fc, nil
}

func (fc *FeatureCollectionWriter) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	body, err := io.ReadAll(fh)

	if err != nil {
		return 0, fmt.Errorf("Failed to read  filehandle, %w", err)
	}

	_, err = geojson.UnmarshalFeature(body)

	if err != nil {
		return 0, fmt.Errorf("Failed to unmarshal GeoJSON feature, %w", err)
	}

	fc.mu.Lock()

	defer func() {
		atomic.AddInt64(&fc.count, 1)		
		fc.mu.Unlock()
	}()

	var preamble string

	if atomic.LoadInt64(&fc.count) == 0 {
		preamble = `{"type":"FeatureCollection", "features":[`
	} else {
		preamble = `,`
	}

	sr := strings.NewReader(preamble + string(body))

	i, err := fc.writer.Write(ctx, key, sr)

	if err != nil {
		return 0, fmt.Errorf("Failed write body, %w", err)
	}

	return i, nil
}

func (fc *FeatureCollectionWriter) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

func (fc *FeatureCollectionWriter) Flush(ctx context.Context) error {
	return nil
}

func (fc *FeatureCollectionWriter) Close(ctx context.Context) error {

	if fc.closed {
		return fmt.Errorf("Feature collection writer has already been closed")
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	var body string

	if atomic.LoadInt64(&fc.count) == 0 {
		body = `{"type":"FeatureCollection", "features":[]}`
	} else {
		body = `]}`
	}

	sr := strings.NewReader(body)
	_, err := fc.writer.Write(ctx, "", sr)

	if err != nil {
		return fmt.Errorf("Failed to write closure, %w", err)
	}

	fc.closed = true	
	return nil
}

func (fc *FeatureCollectionWriter) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
