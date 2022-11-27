package pmtiles

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/paulmach/orb/geojson"
	wof_id "github.com/whosonfirst/go-whosonfirst-id"
	"gocloud.dev/docstore"
	"strconv"
	"strings"
	"time"
	"log"
)

// JSON-encoding features is not ideal given the known performance issues around marshaling
// and unmarshaling JSON but everything else fails at some stage with type issues so this will
// do for now.

type FeatureCache struct {
	Created int64  `json:"created"`
	Id      int64  `json:"id"`
	Body    string `json:"body"`
}

type TileCache struct {
	Created      int64   `json:"created"`
	LastAccessed int64   `json:"last_accessed"`
	Path         string  `json:"path"`
	Features     []int64 `json:"features"`
}

type CacheManager struct {
	feature_collection *docstore.Collection
	tile_collection    *docstore.Collection
}

func NewCacheManager(feature_collection *docstore.Collection, tile_collection *docstore.Collection) *CacheManager {

	m := &CacheManager{
		feature_collection: feature_collection,
		tile_collection:    tile_collection,
	}

	return m
}

func (m *CacheManager) CacheFeatures(ctx context.Context, features []*geojson.Feature) ([]int64, error) {

	feature_ids := make([]int64, len(features))

	type cache_rsp struct {
		offset int
		id     int64
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done_ch := make(chan bool)
	err_ch := make(chan error)
	rsp_ch := make(chan *cache_rsp)

	for idx, f := range features {

		go func(idx int, f *geojson.Feature) {

			defer func() {
				done_ch <- true
			}()

			c, err := m.CacheFeature(ctx, f)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to cache feature at offset %d, %w", idx, err)
				return
			}

			rsp_ch <- &cache_rsp{
				offset: idx,
				id:     c.Id,
			}

		}(idx, f)
	}

	remaining := len(feature_ids)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return nil, fmt.Errorf("Failed to cache features, %w", err)
		case rsp := <-rsp_ch:
			feature_ids[rsp.offset] = rsp.id
		}
	}

	return feature_ids, nil
}

func (m *CacheManager) CacheTile(ctx context.Context, path string, features []*geojson.Feature) (*TileCache, error) {

	feature_ids, err := m.CacheFeatures(ctx, features)

	if err != nil {
		return nil, fmt.Errorf("Failed to cache features for %s, %w", path, err)
	}

	tc, err := NewTileCache(path, feature_ids)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new tile cache for %s, %w", path, err)
	}

	err = m.tile_collection.Put(ctx, tc)

	if err != nil {
		return nil, fmt.Errorf("Failed to store tile cache for %s, %w", path, err)
	}

	return tc, err
}

func (m *CacheManager) CacheFeature(ctx context.Context, feature *geojson.Feature) (*FeatureCache, error) {

	c, err := NewFeatureCache(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to create feature cache, %w", err)
	}

	err = m.feature_collection.Put(ctx, c)

	if err != nil {
		return nil, fmt.Errorf("Failed to store feature cache, %w", err)
	}

	return c, nil
}

func (m *CacheManager) GetFeatureCache(ctx context.Context, id int64) (*FeatureCache, error) {

	fc := &FeatureCache{
		Id: id,
	}

	err := m.feature_collection.Get(ctx, &fc)

	if err != nil {
		return nil, fmt.Errorf("Failed to get feature cache for %d, %w", id, err)
	}

	return fc, nil
}

func (m *CacheManager) GetTileCache(ctx context.Context, path string) (*TileCache, error) {

	tc := &TileCache{
		Path: path,
	}

	err := m.tile_collection.Get(ctx, &tc)

	if err != nil {
		return nil, fmt.Errorf("Failed to get tile cache for %s, %w", path, err)
	}

	return tc, nil
}

func (m *CacheManager) UnmarshalFeatureCache(ctx context.Context, fc *FeatureCache) (*geojson.Feature, error) {

	var feature *geojson.Feature

	r := strings.NewReader(fc.Body)

	dec := json.NewDecoder(r)
	err := dec.Decode(&feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode feature, %w", err)
	}

	return feature, nil
}

func (m *CacheManager) UnmarshalTileCache(ctx context.Context, tc *TileCache) ([]*geojson.Feature, error) {

	features := make([]*geojson.Feature, len(tc.Features))

	type cache_rsp struct {
		offset  int
		feature *geojson.Feature
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)
	rsp_ch := make(chan *cache_rsp)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for idx, id := range tc.Features {

		go func(idx int, id int64) {

			defer func() {
				done_ch <- true
			}()

			fc := FeatureCache{
				Id: id,
			}

			err := m.feature_collection.Get(ctx, &fc)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to retrieve feature cache for %d, %w", id, err)
				return
			}

			f, err := m.UnmarshalFeatureCache(ctx, &fc)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to unmarshal feature for %d, %w", id, err)
				return
			}

			rsp_ch <- &cache_rsp{
				offset:  idx,
				feature: f,
			}

		}(idx, id)
	}

	remaining := len(features)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return nil, fmt.Errorf("Failed to retrieve features for tile, %w", err)
		case rsp := <-rsp_ch:
			features[rsp.offset] = rsp.feature
		}
	}

	return features, nil
}

func NewFeatureCache(feature *geojson.Feature) (*FeatureCache, error) {

	var id int64

	props := feature.Properties

	props_id := fmt.Sprintf("%s", props["wof:id"])

	id, err := strconv.ParseInt(props_id, 10, 64)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse %s, %w", props_id, err)
	}

	if id < wof_id.EARTH {
		return nil, fmt.Errorf("Invalid ID %d", id)
	}

	body, err := feature.MarshalJSON()

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal JSON for ID %d, %w", id, err)
	}

	now := time.Now()
	ts := now.Unix()

	fc := &FeatureCache{
		Created: ts,
		Id:      id,
		Body:    string(body),
	}

	return fc, nil
}

func NewTileCache(path string, feature_ids []int64) (*TileCache, error) {

	now := time.Now()
	ts := now.Unix()

	c := &TileCache{
		Created:  ts,
		Path:     path,
		Features: feature_ids,
	}

	return c, nil
}
