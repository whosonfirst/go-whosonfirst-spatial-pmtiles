package database

// Implement the whosonfirst/go-reader/v2.Reader interface.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func (r *RTreeSpatialDatabase) Read(ctx context.Context, str_uri string) (io.ReadSeekCloser, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI %s, %w", str_uri, err)
	}

	// TO DO : ALT STUFF HERE

	str_id := strconv.FormatInt(id, 10)

	sp := &RTreeSpatialIndex{
		FeatureId: str_id,
		AltLabel:  "",
	}

	cache_item, err := r.retrieveCache(ctx, sp)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve cache, %w", err)
	}

	// START OF this is dumb

	enc_spr, err := json.Marshal(cache_item.SPR)

	if err != nil {
		return nil, fmt.Errorf("Failed to marchal cache record, %w", err)
	}

	var props map[string]interface{}

	err = json.Unmarshal(enc_spr, &props)

	if err != nil {
		return nil, err
	}

	// END OF this is dumb

	orb_geom := cache_item.Geometry.Geometry()
	f := geojson.NewFeature(orb_geom)

	if err != nil {
		return nil, err
	}

	f.Properties = props

	enc_f, err := f.MarshalJSON()

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(enc_f)
	return ioutil.NewReadSeekCloser(br)
}

func (r *RTreeSpatialDatabase) Exists(ctx context.Context, str_uri string) (bool, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return false, fmt.Errorf("Failed to parse URI %s, %w", str_uri, err)
	}

	// TO DO : ALT STUFF HERE

	str_id := strconv.FormatInt(id, 10)

	sp := &RTreeSpatialIndex{
		FeatureId: str_id,
		AltLabel:  "",
	}

	_, err = r.retrieveCache(ctx, sp)

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (r *RTreeSpatialDatabase) ReaderURI(ctx context.Context, str_uri string) string {
	return str_uri
}
