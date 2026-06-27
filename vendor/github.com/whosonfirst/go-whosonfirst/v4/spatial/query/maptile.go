package query

import (
	"fmt"

	"github.com/paulmach/orb/maptile"
)

// https://github.com/paulmach/orb/blob/v0.11.1/maptile/tile.go#L30

type Tile struct {
	Zoom uint32 `json:"zoom"`
	X    uint32 `json:"x"`
	Y    uint32 `json:"y"`
}

type MapTileSpatialQuery struct {
	Tile                *Tile    `json:"tile,omitempty"`
	Placetypes          []string `json:"placetypes,omitempty"`
	Geometries          string   `json:"geometries,omitempty"`
	AlternateGeometries []string `json:"alternate_geometries,omitempty"`
	IsCurrent           []int64  `json:"is_current,omitempty"`
	IsCeased            []int64  `json:"is_ceased,omitempty"`
	IsDeprecated        []int64  `json:"is_deprecated,omitempty"`
	IsSuperseded        []int64  `json:"is_superseded,omitempty"`
	IsSuperseding       []int64  `json:"is_superseding,omitempty"`
	InceptionDate       string   `json:"inception_date,omitempty"`
	CessationDate       string   `json:"cessation_date,omitempty"`
	Properties          []string `json:"properties,omitempty"`
	Sort                []string `json:"sort,omitempty"`
}

func (q *MapTileSpatialQuery) MapTile() (*maptile.Tile, error) {

	if q.Tile == nil {
		return nil, fmt.Errorf("Missing tile")
	}

	if q.Tile.Zoom == 0 {
		return nil, fmt.Errorf("Invalid tile zoom value")
	}

	if q.Tile.X == 0 {
		return nil, fmt.Errorf("Invalid tile x value")
	}

	if q.Tile.Y == 0 {
		return nil, fmt.Errorf("Invalid tile y value")
	}

	z := maptile.Zoom(q.Tile.Zoom)
	t := maptile.New(q.Tile.X, q.Tile.Y, z)

	return &t, nil
}

func (q *MapTileSpatialQuery) SpatialQuery() *SpatialQuery {

	sp_q := &SpatialQuery{
		Placetypes:          q.Placetypes,
		Geometries:          q.Geometries,
		AlternateGeometries: q.AlternateGeometries,
		IsCurrent:           q.IsCurrent,
		IsCeased:            q.IsCeased,
		IsDeprecated:        q.IsDeprecated,
		IsSuperseded:        q.IsSuperseded,
		InceptionDate:       q.InceptionDate,
		CessationDate:       q.CessationDate,
		Properties:          q.Properties,
		Sort:                q.Sort,
	}

	return sp_q
}
