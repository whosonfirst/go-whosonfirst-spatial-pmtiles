package cache

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"gocloud.dev/docstore"
)

type DocstoreCacheManager struct {
	feature_collection *docstore.Collection
	tile_collection    *docstore.Collection
	ticker             *time.Ticker
}

type DocstoreCacheManagerOptions struct {
	FeatureCollection *docstore.Collection
	CacheTTL          int
}

func NewDocstoreCacheManager(ctx context.Context, opts *DocstoreCacheManagerOptions) CacheManager {

	m := &DocstoreCacheManager{
		feature_collection: opts.FeatureCollection,
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

func (m *DocstoreCacheManager) CacheFeature(ctx context.Context, body []byte) (*FeatureCache, error) {

	if m.feature_collection == nil {
		return nil, fmt.Errorf("No feature collection defined")
	}

	fc, err := NewFeatureCache(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to create feature cache, %w", err)
	}

	slog.Debug("Store in feature cache", "id", fc.Id)

	err = m.feature_collection.Put(ctx, fc)

	if err != nil {
		return nil, fmt.Errorf("Failed to store feature cache for %s, %w", fc.Id, err)
	}

	return fc, nil
}

func (m *DocstoreCacheManager) GetFeatureCache(ctx context.Context, id string) (*FeatureCache, error) {

	if m.feature_collection == nil {
		return nil, fmt.Errorf("No feature collection defined")
	}

	fc := FeatureCache{
		Id: id,
	}

	err := m.feature_collection.Get(ctx, &fc)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve feature from cache for %s, %w", id, err)
	}

	return &fc, nil
}

func (m *DocstoreCacheManager) pruneCaches(ctx context.Context, t time.Time) {
	go m.pruneFeatureCache(ctx, t)
}

func (m *DocstoreCacheManager) pruneFeatureCache(ctx context.Context, t time.Time) error {

	if m.feature_collection == nil {
		return nil
	}

	slog.Debug("Prune tile cache", "older than", t)

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
			slog.Error("Failed to get next iterator", "error", err)
		} else {

			slog.Debug("Remove from feature cache", "id", fc.Id, "created", fc.Created)

			err := m.feature_collection.Delete(ctx, &fc)

			if err != nil {
				slog.Error("Failed to delete from feature cache", "id", fc.Id, "error", err)
			}
		}
	}

	return nil
}

func (m *DocstoreCacheManager) Close() error {

	m.ticker.Stop()

	if m.feature_collection != nil {
		m.feature_collection.Close()
	}

	if m.tile_collection != nil {
		m.tile_collection.Close()
	}

	return nil
}
