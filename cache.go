package pmtiles

import (
	"github.com/paulmach/orb/geojson"
	"time"
)

type TileFeaturesCache struct {
	Created  int64              `json:"created"`
	Path     string             `json:"path"`
	Features []*geojson.Feature `json:"features"`
}

func NewTileFeatureCache(path string, features []*geojson.Feature) *TileFeaturesCache {
	now := time.Now()

	c := &TileFeaturesCache{
		Created:  now.Unix(),
		Path:     path,
		Features: features,
	}

	return c
}
