package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func CacheFeatureCollection(ctx context.Context, m CacheManager, fc *geojson.FeatureCollection) ([]string, error) {

	features, err := UniqueFeatures(ctx, fc.Features)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive unique features for FeatureCollection, %w", err)
	}

	return CacheFeatures(ctx, m, features)
}

func CacheFeatures(ctx context.Context, m CacheManager, features [][]byte) ([]string, error) {

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

func UnmarshalFeatureCache(ctx context.Context, fc *FeatureCache) (*geojson.Feature, error) {

	var feature *geojson.Feature

	r := strings.NewReader(fc.Body)

	dec := json.NewDecoder(r)
	err := dec.Decode(&feature)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode feature, %w", err)
	}

	return feature, nil
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
