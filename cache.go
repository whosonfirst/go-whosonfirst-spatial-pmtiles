package pmtiles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"gocloud.dev/docstore"
)

const FEATURES_CACHE_TABLE string = "pmtiles_features"

// JSON-encoding features is not ideal given the known performance issues around marshaling
// and unmarshaling JSON but everything else fails at some stage with type issues so this will
// do for now. Note that we are using json-iterator/go as a orb/geojson custom marshaler/unmarsheler
// whic his defined in pmtiles.go

type FeatureCache struct {
	Created int64  `json:"created"`
	Id      string `json:"id"` // this is a string rather than int64 because it might include an alt label
	Body    string `json:"body"`
}

type CacheManager struct {
	feature_collection *docstore.Collection
	tile_collection    *docstore.Collection
	logger             *log.Logger
	ticker             *time.Ticker
}

type CacheManagerOptions struct {
	FeatureCollection *docstore.Collection
	Logger            *log.Logger
	CacheTTL          int
}

func NewCacheManager(ctx context.Context, opts *CacheManagerOptions) *CacheManager {

	m := &CacheManager{
		feature_collection: opts.FeatureCollection,
		logger:             opts.Logger,
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

func (m *CacheManager) CacheFeatureCollection(ctx context.Context, fc *geojson.FeatureCollection) ([]string, error) {

	features, err := UniqueFeatures(ctx, fc.Features)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive unique features for FeatureCollection, %w", err)
	}

	return m.CacheFeatures(ctx, features)
}

func (m *CacheManager) CacheFeatures(ctx context.Context, features [][]byte) ([]string, error) {

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

		go func(idx int, f []byte) {

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

func (m *CacheManager) CacheFeature(ctx context.Context, body []byte) (*FeatureCache, error) {

	if m.feature_collection == nil {
		return nil, fmt.Errorf("No feature collection defined")
	}

	fc, err := NewFeatureCache(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to create feature cache, %w", err)
	}

	// m.logger.Printf("cache feature %s\n", fc.Id)

	err = m.feature_collection.Put(ctx, fc)

	if err != nil {
		return nil, fmt.Errorf("Failed to store feature cache for %s, %w", fc.Id, err)
	}

	return fc, nil
}

func (m *CacheManager) GetFeatureCache(ctx context.Context, id string) (*FeatureCache, error) {

	if m.feature_collection == nil {
		return nil, fmt.Errorf("No feature collection defined")
	}

	fc := FeatureCache{
		Id: id,
	}

	// m.logger.Printf("get feature %s\n", fc.Id)

	err := m.feature_collection.Get(ctx, &fc)

	if err != nil {
		return nil, fmt.Errorf("Failed to get feature cache for %s, %w", id, err)
	}

	return &fc, nil
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

func (m *CacheManager) pruneCaches(ctx context.Context, t time.Time) {
	go m.pruneFeatureCache(ctx, t)
}

func (m *CacheManager) pruneFeatureCache(ctx context.Context, t time.Time) error {

	if m.feature_collection == nil {
		return nil
	}

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

			err := m.feature_collection.Delete(ctx, &fc)

			if err != nil {
				m.logger.Printf("Failed to delete feature cache %s, %v", fc.Id, err)
			}
		}
	}

	return nil
}

func (m *CacheManager) Close(ctx context.Context) error {
	m.ticker.Stop()

	if m.feature_collection != nil {
		m.feature_collection.Close()
	}

	if m.tile_collection != nil {
		m.tile_collection.Close()
	}

	return nil
}

func UniqueFeatures(ctx context.Context, features []*geojson.Feature) ([][]byte, error) {

	seen := make(map[string]bool)
	unique_features := make([][]byte, 0)

	mu := new(sync.RWMutex)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done_ch := make(chan bool)
	err_ch := make(chan error)
	f_ch := make(chan []byte)

	for idx, f := range features {

		go func(idx int, f *geojson.Feature) {

			defer func() {
				done_ch <- true
			}()

			body, err := f.MarshalJSON()

			if err != nil {
				err_ch <- fmt.Errorf("Failed to marshal feature at offset %d, %w", idx, err)
				return
			}

			f_id, err := FeatureIdFromBytes(body)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to derive feature at offset %d, %w", idx, err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			_, exists := seen[f_id]

			if exists {
				return
			}

			seen[f_id] = true

			f_ch <- body
			return

		}(idx, f)
	}

	remaining := len(features)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return nil, err
		case f := <-f_ch:
			unique_features = append(unique_features, f)
		}
	}

	return unique_features, nil
}

func FeatureId(feature *geojson.Feature) (string, error) {

	body, err := feature.MarshalJSON()

	if err != nil {
		return "", fmt.Errorf("Failed to marshal JSON for feature, %w", err)
	}

	return FeatureIdFromBytes(body)
}

func FeatureIdFromBytes(body []byte) (string, error) {

	id, err := properties.Id(body)

	if err != nil {
		return "", fmt.Errorf("Failed to derive ID from feature, %w", err)
	}

	alt_label, err := properties.AltLabel(body)

	if err != nil {
		return "", fmt.Errorf("Failed to derive alt label from feature, %w", err)
	}

	str_id := fmt.Sprintf("%d", id)

	if alt_label != "" {
		str_id = fmt.Sprintf("%s-alt-%s", str_id, alt_label)
	}

	return str_id, nil
}

func NewFeatureCache(body []byte) (*FeatureCache, error) {

	f_id, err := FeatureIdFromBytes(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive feature ID, %w", err)
	}

	now := time.Now()
	ts := now.Unix()

	fc := &FeatureCache{
		Created: ts,
		Id:      f_id,
		Body:    string(body),
	}

	return fc, nil
}
