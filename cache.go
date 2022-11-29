package pmtiles

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"gocloud.dev/docstore"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

// JSON-encoding features is not ideal given the known performance issues around marshaling
// and unmarshaling JSON but everything else fails at some stage with type issues so this will
// do for now.

type FeatureCache struct {
	Created int64  `json:"created"`
	Id      string `json:"id"` // this is a string rather than int64 because it might include an alt label
	Body    string `json:"body"`
}

type TileCache struct {
	Created  int64    `json:"created"`
	Path     string   `json:"path"`
	Features []string `json:"features"`
}

type CacheManager struct {
	keys_map           *sync.Map
	feature_collection *docstore.Collection
	tile_collection    *docstore.Collection
	logger             *log.Logger
	ticker             *time.Ticker
	feature_counts     map[string]int64
	counts_mu          *sync.RWMutex
}

type CacheManagerOptions struct {
	FeatureCollection *docstore.Collection
	TileCollection    *docstore.Collection
	Logger            *log.Logger
	CacheTTL          int
}

func NewCacheManager(ctx context.Context, opts *CacheManagerOptions) *CacheManager {

	keys_map := new(sync.Map)

	feature_counts := make(map[string]int64)
	counts_mu := new(sync.RWMutex)

	m := &CacheManager{
		feature_collection: opts.FeatureCollection,
		tile_collection:    opts.TileCollection,
		logger:             opts.Logger,
		keys_map:           keys_map,
		feature_counts:     feature_counts,
		counts_mu:          counts_mu,
	}

	cache_ttl := opts.CacheTTL

	now := time.Now()
	then := now.Add(time.Duration(-cache_ttl) * time.Second)

	m.pruneCaches(ctx, then)

	ticker := time.NewTicker(time.Duration(cache_ttl) * time.Second)
	m.ticker = ticker

	go func() {

		for {
			select {
			case t := <-ticker.C:
				m.pruneCaches(ctx, t)
			}
		}
	}()

	return m
}

func (m *CacheManager) CacheFeatures(ctx context.Context, features []*geojson.Feature) ([]string, error) {

	feature_ids := make([]string, len(features))

	type cache_rsp struct {
		offset int
		id     string
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

	_, exists := m.keys_map.Load(tc.Path)

	if exists {

		go func() {

			mod := docstore.Mods{
				"Created": tc.Created,
			}

			err := m.tile_collection.Update(ctx, tc, mod)

			if err != nil {
				m.logger.Printf("Failed to update tile cache for %s, %v", tc.Path, err)
			}
		}()

	} else {

		m.logger.Printf("cache tile %s\n", tc.Path)

		m.counts_mu.Lock()
		defer m.counts_mu.Unlock()

		err = m.tile_collection.Put(ctx, tc)

		if err != nil {
			return nil, fmt.Errorf("Failed to store tile cache for %s, %w", path, err)
		}

		m.keys_map.Store(tc.Path, tc.Created)

		for _, id := range tc.Features {

			count, exists := m.feature_counts[id]

			if exists {
				count = count + 1
			} else {
				count = int64(1)
			}

			m.feature_counts[id] = count
		}

	}

	return tc, err
}

func (m *CacheManager) CacheFeature(ctx context.Context, feature *geojson.Feature) (*FeatureCache, error) {

	fc, err := NewFeatureCache(feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to create feature cache, %w", err)
	}

	_, exists := m.keys_map.Load(fc.Id)

	if exists {

		mod := docstore.Mods{
			"Created": fc.Created,
		}

		err := m.feature_collection.Update(ctx, fc, mod)

		if err != nil {
			m.logger.Printf("Failed to update feature cache for %s, %v", fc.Id, err)
		}

	} else {

		m.logger.Printf("cache feature %s\n", fc.Id)

		err = m.feature_collection.Put(ctx, fc)

		if err != nil {
			return nil, fmt.Errorf("Failed to store feature cache, %w", err)
		}

		m.keys_map.Store(fc.Id, fc.Created)
	}

	return fc, nil
}

func (m *CacheManager) GetFeatureCache(ctx context.Context, id string) (*FeatureCache, error) {

	fc := FeatureCache{
		Id: id,
	}

	// m.logger.Printf("get feature %s\n", fc.Id)

	err := m.feature_collection.Get(ctx, &fc)

	if err != nil {
		return nil, fmt.Errorf("Failed to get feature cache for %d, %w", id, err)
	}

	return &fc, nil
}

func (m *CacheManager) GetTileCache(ctx context.Context, path string) (*TileCache, error) {

	tc := TileCache{
		Path: path,
	}

	// m.logger.Printf("get tile %s\n", tc.Path)

	err := m.tile_collection.Get(ctx, &tc)

	if err != nil {
		return nil, fmt.Errorf("Failed to get tile cache for %s, %w", path, err)
	}

	return &tc, nil
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

	// cache this too?

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

		go func(idx int, id string) {

			defer func() {
				done_ch <- true
			}()

			fc, err := m.GetFeatureCache(ctx, id)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to retrieve feature cache for %d, %w", id, err)
				return
			}

			f, err := m.UnmarshalFeatureCache(ctx, fc)

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

func (m *CacheManager) pruneCaches(ctx context.Context, t time.Time) {
	go m.pruneTileCache(ctx, t)
	go m.pruneFeatureCache(ctx, t)
}

func (m *CacheManager) pruneTileCache(ctx context.Context, t time.Time) error {

	m.logger.Printf("Prune feature cache older that %v\n", t)

	ts := t.Unix()

	q := m.tile_collection.Query()
	q = q.Where("Created", "<=", ts)

	iter := q.Get(ctx)

	defer iter.Stop()

	for {

		var tc TileCache

		err := iter.Next(ctx, &tc)

		if err == io.EOF {
			break
		} else if err != nil {
			m.logger.Printf("Failed to get next iterator, %v", err)
		} else {

			m.logger.Printf("Prune %s\n", tc.Path)
			err := m.tile_collection.Delete(ctx, &tc)

			if err != nil {
				m.logger.Printf("Failed to delete feature cache %s, %v", tc.Path, err)
			}

			m.keys_map.Delete(tc.Path)

			m.counts_mu.Lock()

			for _, id := range tc.Features {

				count, exists := m.feature_counts[id]

				if exists {
					count = count - 1
				}

				if count <= 0 {
					delete(m.feature_counts, id)
				} else {
					m.feature_counts[id] = count
				}
			}

			m.counts_mu.Unlock()
		}
	}

	return nil
}

func (m *CacheManager) pruneFeatureCache(ctx context.Context, t time.Time) error {

	m.logger.Printf("Prune tile cache older that %v\n", t)

	ts := t.Unix()

	q := m.feature_collection.Query()
	q = q.Where("Created", "<=", ts)

	iter := q.Get(ctx)

	defer iter.Stop()

	for {

		var fc FeatureCache

		err := iter.Next(ctx, &fc)

		if err == io.EOF {
			break
		} else if err != nil {
			m.logger.Printf("Failed to get next iterator, %v", err)
		} else {

			m.counts_mu.RLock()
			count, exists := m.feature_counts[fc.Id]
			m.counts_mu.RUnlock()

			if exists && count > 0 {
				m.logger.Printf("Feature %s still has %d existing pointers, do not prune yet\n", fc.Id, count)
				continue
			}

			m.logger.Printf("Prune feature %s\n", fc.Id)
			err := m.feature_collection.Delete(ctx, &fc)

			if err != nil {
				m.logger.Printf("Failed to delete feature cache %s, %v", fc.Id, err)
			}

			m.keys_map.Delete(fc.Id)
		}
	}

	return nil
}

func (m *CacheManager) Close(ctx context.Context) error {
	m.ticker.Stop()
	m.feature_collection.Close()
	m.tile_collection.Close()
	return nil
}

func NewFeatureCache(feature *geojson.Feature) (*FeatureCache, error) {

	body, err := feature.MarshalJSON()

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal JSON for feature, %w", err)
	}

	id, err := properties.Id(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive ID from feature, %w", err)
	}

	alt_label, err := properties.AltLabel(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive alt label from feature, %w", err)
	}

	str_id := fmt.Sprintf("%d", id)

	if alt_label != "" {
		str_id = fmt.Sprintf("%s-alt-%s", str_id, alt_label)
	}

	now := time.Now()
	ts := now.Unix()

	fc := &FeatureCache{
		Created: ts,
		Id:      str_id,
		Body:    string(body),
	}

	return fc, nil
}

func NewTileCache(path string, feature_ids []string) (*TileCache, error) {

	now := time.Now()
	ts := now.Unix()

	c := &TileCache{
		Created:  ts,
		Path:     path,
		Features: feature_ids,
	}

	return c, nil
}
