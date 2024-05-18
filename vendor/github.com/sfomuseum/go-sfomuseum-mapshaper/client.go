package mapshaper

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"	
)

// As in:
// https://github.com/sfomuseum/docker-sfomuseum-mapshaper or
// https://github.com/sfomuseum/go-sfomuseum-mapshaper
//
// For example:
// docker run -it -p 8080:8080 mapshaper mapshaper-server -server-uri http://0.0.0.0:8080

type Client struct {
	address string
	client  *http.Client
}

func NewLocalClient(ctx context.Context) (*Client, error) {
	return NewClient(ctx, "http://localhost:8080")
}

func NewClient(ctx context.Context, address string) (*Client, error) {

	http_client := &http.Client{}

	s := &Client{
		address: address,
		client:  http_client,
	}

	return s, nil
}

func (s *Client) Ping() (bool, error) {

	u, err := url.Parse(s.address)

	if err != nil {
		return false, err
	}

	u.Path = "/api/ping"

	req, err := http.NewRequest("GET", u.String(), nil)

	if err != nil {
		return false, err
	}

	rsp, err := s.client.Do(req)

	if err != nil {
		return false, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return false, errors.New(rsp.Status)
	}

	return true, nil
}

func (s *Client) AppendCentroids(ctx context.Context, fc *geojson.FeatureCollection) (*geojson.FeatureCollection, error) {

	centroid_fc, err := s.ExecuteMethod(ctx, "/api/innerpoint", fc)

	if err != nil {
		return nil, err
	}

	for idx, centroid_f := range centroid_fc.Features {

		pt := centroid_f.Geometry.(orb.Point)

		lon := pt[0]
		lat := pt[1]

		fc.Features[idx].Properties["mps:latitude"] = lon
		fc.Features[idx].Properties["mps:longitude"] = lat
	}

	return fc, nil
}

func (s *Client) ExecuteMethod(ctx context.Context, method string, fc *geojson.FeatureCollection) (*geojson.FeatureCollection, error) {

	u, err := url.Parse(s.address)

	if err != nil {
		return nil, err
	}

	u.Path = method

	body, err := fc.MarshalJSON()

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	rsp, err := s.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	rsp_body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	return geojson.UnmarshalFeatureCollection(rsp_body)
}
