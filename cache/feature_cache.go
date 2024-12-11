package cache

import (
	"fmt"
	"time"
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
