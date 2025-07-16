package database

import (
	"context"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/dhconnelly/rtreego"
	gocache "github.com/patrickmn/go-cache"
	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func init() {
	ctx := context.Background()
	RegisterSpatialDatabase(ctx, "rtree", NewRTreeSpatialDatabase)
}

type RTreeCache struct {
	Geometry *geojson.Geometry        `json:"geometry"`
	SPR      spr.StandardPlacesResult `json:"properties"`
}

// PLEASE DISCUSS WHY patrickm/go-cache AND NOT whosonfirst/go-cache HERE

type RTreeSpatialDatabase struct {
	SpatialDatabase
	index_alt_files bool
	rtree           *rtreego.Rtree
	gocache         *gocache.Cache
	mu              *sync.RWMutex
	strict          bool
}

type RTreeSpatialIndex struct {
	Rect      *rtreego.Rect
	Id        string
	FeatureId string
	IsAlt     bool
	AltLabel  string
}

func (i *RTreeSpatialIndex) Bounds() rtreego.Rect {
	return *i.Rect
}

type RTreeResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
	Places                    []spr.StandardPlacesResult `json:"places"`
}

func (r *RTreeResults) Results() []spr.StandardPlacesResult {
	return r.Places
}

func NewRTreeSpatialDatabase(ctx context.Context, uri string) (SpatialDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	strict := true

	if q.Get("strict") == "false" {
		strict = false
	}

	expires := 0 * time.Second
	cleanup := 0 * time.Second

	str_exp := q.Get("default_expiration")
	str_cleanup := q.Get("cleanup_interval")

	if str_exp != "" {

		int_expires, err := strconv.Atoi(str_exp)

		if err != nil {
			return nil, err
		}

		expires = time.Duration(int_expires) * time.Second
	}

	if str_cleanup != "" {

		int_cleanup, err := strconv.Atoi(str_cleanup)

		if err != nil {
			return nil, err
		}

		cleanup = time.Duration(int_cleanup) * time.Second
	}

	index_alt_files := false

	str_index_alt := q.Get("index_alt_files")

	if str_index_alt != "" {

		index_alt, err := strconv.ParseBool(str_index_alt)

		if err != nil {
			return nil, err
		}

		index_alt_files = index_alt
	}

	gc := gocache.New(expires, cleanup)

	rtree := rtreego.NewTree(2, 25, 50)

	mu := new(sync.RWMutex)

	db := &RTreeSpatialDatabase{
		rtree:           rtree,
		index_alt_files: index_alt_files,
		gocache:         gc,
		strict:          strict,
		mu:              mu,
	}

	return db, nil
}
