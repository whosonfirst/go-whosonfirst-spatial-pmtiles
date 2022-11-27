package api

import (
	"context"
	"fmt"
	"github.com/aaronland/go-artisanal-integers/client"
	"github.com/cenkalti/backoff/v4"
	"github.com/tidwall/gjson"
	"go.uber.org/ratelimit"
	"io"
	"log"
	"net/http"
	"net/url"
)

func init() {
	ctx := context.Background()
	client.RegisterClient(ctx, "brooklynintegers", NewAPIClient)
}

type APIClient struct {
	client.Client
	isa          string
	http_client  *http.Client
	Scheme       string
	Host         string
	Endpoint     string
	rate_limiter ratelimit.Limiter
}

type APIError struct {
	Code    int64
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

type APIResponse struct {
	raw []byte
}

func (rsp *APIResponse) Int() (int64, error) {

	ints := gjson.GetBytes(rsp.raw, "integers.0.integer")

	if !ints.Exists() {
		return -1, fmt.Errorf("Failed to generate any integers")
	}

	i := ints.Int()
	return i, nil
}

func (rsp *APIResponse) Stat() string {

	r := gjson.GetBytes(rsp.raw, "stat")

	if !r.Exists() {
		return ""
	}

	return r.String()
}

func (rsp *APIResponse) Ok() (bool, error) {

	stat := rsp.Stat()

	if stat == "ok" {
		return true, nil
	}

	return false, rsp.Error()
}

func (rsp *APIResponse) Error() error {

	c := gjson.GetBytes(rsp.raw, "error.code")
	m := gjson.GetBytes(rsp.raw, "error.message")

	if !c.Exists() {
		return fmt.Errorf("Failed to parse error code")
	}

	if !m.Exists() {
		return fmt.Errorf("Failed to parse error message")
	}

	err := APIError{
		Code:    c.Int(),
		Message: m.String(),
	}

	return &err
}

func NewAPIClient(ctx context.Context, uri string) (client.Client, error) {

	http_client := &http.Client{}
	rl := ratelimit.New(10) // please make this configurable

	cl := &APIClient{
		Scheme:       "https",
		Host:         "api.brooklynintegers.com",
		Endpoint:     "rest/",
		http_client:  http_client,
		rate_limiter: rl,
	}

	return cl, nil
}

func (client *APIClient) NextInt(ctx context.Context) (int64, error) {

	params := url.Values{}
	method := "brooklyn.integers.create"

	var next_id int64

	cb := func() error {

		rsp, err := client.executeMethod(ctx, method, &params)

		if err != nil {
			return err
		}

		i, err := rsp.Int()

		if err != nil {
			log.Println(err)
			return err
		}

		next_id = i
		return nil
	}

	bo := backoff.NewExponentialBackOff()

	err := backoff.Retry(cb, bo)

	if err != nil {
		return -1, fmt.Errorf("Failed to execute method (%s), %w", method, err)
	}

	return next_id, nil
}

func (client *APIClient) executeMethod(ctx context.Context, method string, params *url.Values) (*APIResponse, error) {

	client.rate_limiter.Take()

	url := client.Scheme + "://" + client.Host + "/" + client.Endpoint

	params.Set("method", method)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to create request (%s), %w", url, err)
	}

	req.URL.RawQuery = (*params).Encode()

	req.Header.Add("Accept-Encoding", "gzip")

	rsp, err := client.http_client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to create request (%s), %w", url, err)
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to read response, %w", err)
	}

	r := APIResponse{
		raw: body,
	}

	return &r, nil
}
