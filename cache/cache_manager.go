package cache

import (
	"context"
)

type CacheManager interface {
	CacheFeature(context.Context, []byte) (*FeatureCache, error)
	GetFeatureCache(context.Context, string) (*FeatureCache, error)
	Close() error
}
