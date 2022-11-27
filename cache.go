package pmtiles

import (
	"encoding/json"
	"fmt"
	"github.com/paulmach/orb/geojson"
	"strings"
	"time"
)

// JSON-encoding features is not ideal given the known performance issues around marshaling
// and unmarshaling JSON but everything else fails at some stage with type issues so this will
// do for now.

type TileFeaturesCache struct {
	Created      int64  `json:"created"`
	LastAccessed int64  `json:"last_accessed"`
	Path         string `json:"path"`
	Features     string `json:"features"`
}

func NewTileFeatureCache(path string, features []*geojson.Feature) (*TileFeaturesCache, error) {

	enc, err := json.Marshal(features)

	if err != nil {
		return nil, fmt.Errorf("Failed to encode features, %w", err)
	}

	now := time.Now()

	c := &TileFeaturesCache{
		Created:  now.Unix(),
		Path:     path,
		Features: string(enc),
	}

	return c, nil
}

func (tc *TileFeaturesCache) UnmarshalFeatures() ([]*geojson.Feature, error) {

	var features []*geojson.Feature

	r := strings.NewReader(tc.Features)

	dec := json.NewDecoder(r)
	err := dec.Decode(&features)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode features, %w", err)
	}

	return features, nil
}
