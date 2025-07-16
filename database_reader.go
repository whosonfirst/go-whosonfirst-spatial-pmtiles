package pmtiles

// Implement the whosonfirst/go-reader/v2.Reader interface.

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func (db *PMTilesSpatialDatabase) Read(ctx context.Context, path string) (io.ReadSeekCloser, error) {

	if !db.enable_feature_cache {
		return nil, spatial.ErrNotFound
	}

	id, uri_args, err := uri.ParseURI(path)

	if err != nil {
		return nil, fmt.Errorf("Failed to path %s, %w", path, err)
	}

	fname, err := uri.Id2Fname(id, uri_args)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive filename from %s, %w", path, err)
	}

	fname = strings.Replace(fname, ".geojson", "", 1)

	fc, err := db.cache_manager.GetFeatureCache(ctx, fname)

	if err != nil {
		return nil, fmt.Errorf("Failed to read feature from cache for %s, %w", path, err)
	}

	r := strings.NewReader(fc.Body)

	rsc, err := ioutil.NewReadSeekCloser(r)

	if err != nil {
		return nil, fmt.Errorf("Failed to create ReadSeekCloser for %s, %w", path, err)
	}

	return rsc, nil
}

func (db *PMTilesSpatialDatabase) Exists(ctx context.Context, path string) (bool, error) {

	if !db.enable_feature_cache {
		return false, spatial.ErrNotFound
	}

	id, uri_args, err := uri.ParseURI(path)

	if err != nil {
		return false, fmt.Errorf("Failed to path %s, %w", path, err)
	}

	fname, err := uri.Id2Fname(id, uri_args)

	if err != nil {
		return false, fmt.Errorf("Failed to derive filename from %s, %w", path, err)
	}

	fname = strings.Replace(fname, ".geojson", "", 1)

	_, err = db.cache_manager.GetFeatureCache(ctx, fname)

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (db *PMTilesSpatialDatabase) ReaderURI(ctx context.Context, path string) string {
	return path
}
