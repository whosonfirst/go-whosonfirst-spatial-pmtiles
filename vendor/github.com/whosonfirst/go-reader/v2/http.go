package reader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/whosonfirst/go-ioutil"
)

// HTTPReader is a struct that implements the `Reader` interface for reading documents from an HTTP(S) resource.
type HTTPReader struct {
	Reader
	url        *url.URL
	throttle   <-chan time.Time
	user_agent string
}

func init() {

	ctx := context.Background()

	schemes := []string{
		"http",
		"https",
	}

	for _, s := range schemes {

		err := RegisterReader(ctx, s, NewHTTPReader)

		if err != nil {
			panic(err)
		}
	}
}

// NewStdinReader returns a new `Reader` instance for reading documents from an HTTP(s) resource,
// configured by 'uri' in the form of:
//
//	http(s)://{HOST}?{PARAMS}
//
// Where {PARAMS} can be:
// * user-agent - An optional user agent string to include with requests.
func NewHTTPReader(ctx context.Context, uri string) (Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	rate := time.Second / 3
	throttle := time.Tick(rate)

	r := HTTPReader{
		throttle: throttle,
		url:      u,
	}

	q := u.Query()
	ua := q.Get("user-agent")

	if ua != "" {
		r.user_agent = ua
	}

	return &r, nil
}

// Exists returns a boolean value indicating whether 'path' already exists.
func (r *HTTPReader) Exists(ctx context.Context, uri string) (bool, error) {

	<-r.throttle

	u, _ := url.Parse(r.url.String())
	u.Path = filepath.Join(u.Path, uri)

	url := u.String()

	req, err := http.NewRequest(http.MethodHead, url, nil)

	if err != nil {
		return false, fmt.Errorf("Failed to create new request, %w", err)
	}

	if r.user_agent != "" {
		req.Header.Set("User-Agent", r.user_agent)
	}

	cl := &http.Client{}

	rsp, err := cl.Do(req)

	if err != nil {
		return false, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

// Read will open a `io.ReadSeekCloser` for the resource located at 'uri'.
func (r *HTTPReader) Read(ctx context.Context, uri string) (io.ReadSeekCloser, error) {

	<-r.throttle

	u, _ := url.Parse(r.url.String())
	u.Path = filepath.Join(u.Path, uri)

	url := u.String()

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new request, %w", err)
	}

	if r.user_agent != "" {
		req.Header.Set("User-Agent", r.user_agent)
	}

	cl := &http.Client{}

	rsp, err := cl.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute request, %w", err)
	}

	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %s", rsp.Status)
	}

	fh, err := ioutil.NewReadSeekCloser(rsp.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new ReadSeekCloser, %w", err)
	}

	return fh, nil
}

// ReaderURI returns 'uri'.
func (r *HTTPReader) ReaderURI(ctx context.Context, uri string) string {
	return uri
}
