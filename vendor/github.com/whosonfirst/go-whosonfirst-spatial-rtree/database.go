package rtree

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dhconnelly/rtreego"
	gocache "github.com/patrickmn/go-cache"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/timer"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	ctx := context.Background()
	database.RegisterSpatialDatabase(ctx, "rtree", NewRTreeSpatialDatabase)
}

type RTreeCache struct {
	Geometry *geojson.Geometry        `json:"geometry"`
	SPR      spr.StandardPlacesResult `json:"properties"`
}

// PLEASE DISCUSS WHY patrickm/go-cache AND NOT whosonfirst/go-cache HERE

type RTreeSpatialDatabase struct {
	database.SpatialDatabase
	Logger          *log.Logger
	Timer           *timer.Timer
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

func (i *RTreeSpatialIndex) Bounds() *rtreego.Rect {
	return i.Rect
}

type RTreeResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
	Places                    []spr.StandardPlacesResult `json:"places"`
}

func (r *RTreeResults) Results() []spr.StandardPlacesResult {
	return r.Places
}

func NewRTreeSpatialDatabase(ctx context.Context, uri string) (database.SpatialDatabase, error) {

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

	logger := log.Default()

	rtree := rtreego.NewTree(2, 25, 50)

	mu := new(sync.RWMutex)

	t := timer.NewTimer()

	db := &RTreeSpatialDatabase{
		Logger:          logger,
		Timer:           t,
		rtree:           rtree,
		index_alt_files: index_alt_files,
		gocache:         gc,
		strict:          strict,
		mu:              mu,
	}

	return db, nil
}

func (r *RTreeSpatialDatabase) Close(ctx context.Context) error {
	return nil
}

func (r *RTreeSpatialDatabase) IndexFeature(ctx context.Context, body []byte) error {

	is_alt := alt.IsAlt(body)
	alt_label, _ := properties.AltLabel(body)

	if is_alt && !r.index_alt_files {
		return nil
	}

	if is_alt && alt_label == "" {
		return fmt.Errorf("Invalid alt label")
	}

	err := r.setCache(ctx, body)

	if err != nil {
		return fmt.Errorf("Failed to cache feature, %w", err)
	}
	
	feature_id, err := properties.Id(body)

	if err != nil {
		return fmt.Errorf("Failed to derive ID, %w", err)
	}

	str_id := strconv.FormatInt(feature_id, 10)

	// START OF put me in go-whosonfirst-feature/geometry

	geojson_geom, err := geometry.Geometry(body)

	if err != nil {
		return fmt.Errorf("Failed to derive geometry, %w", err)
	}

	orb_geom := geojson_geom.Geometry()

	bounds := make([]orb.Bound, 0)

	switch orb_geom.GeoJSONType() {

	case "MultiPolygon":

		for _, poly := range orb_geom.(orb.MultiPolygon) {

			for _, ring := range poly {
				bounds = append(bounds, ring.Bound())
			}
		}

	case "Polygon":

		for _, ring := range orb_geom.(orb.Polygon) {
			bounds = append(bounds, ring.Bound())
		}
	default:
		bounds = append(bounds, orb_geom.Bound())
	}

	// END OF put me in go-whosonfirst-feature/geometry

	for i, bbox := range bounds {

		sp_id, err := spatial.SpatialIdWithFeature(body, i)

		if err != nil {
			return fmt.Errorf("Failed to derive spatial ID, %v", err)
		}

		min := bbox.Min
		max := bbox.Max

		min_x := min[0]
		min_y := min[1]

		max_x := max[0]
		max_y := max[1]

		llat := max_y - min_y
		llon := max_x - min_x

		pt := rtreego.Point{min_x, min_y}
		rect, err := rtreego.NewRect(pt, []float64{llon, llat})

		if err != nil {

			if r.strict {
				return fmt.Errorf("Failed to derive rtree bounds, %w", err)
			}

			r.Logger.Printf("%s failed indexing, (%v). Strict mode is disabled, so skipping.", sp_id, err)
			return nil
		}

		r.Logger.Printf("index %s %v", sp_id, rect)

		sp := &RTreeSpatialIndex{
			Rect:      rect,
			Id:        sp_id,
			FeatureId: str_id,
			IsAlt:     is_alt,
			AltLabel:  alt_label,
		}

		r.mu.Lock()
		r.rtree.Insert(sp)

		r.mu.Unlock()
	}

	return nil
}

/*

TO DO: figure out suitable comparitor

/ DeleteWithComparator removes an object from the tree using a custom
// comparator for evaluating equalness. This is useful when you want to remove
// an object from a tree but don't have a pointer to the original object
// anymore.
func (tree *Rtree) DeleteWithComparator(obj Spatial, cmp Comparator) bool {
	n := tree.findLeaf(tree.root, obj, cmp)

// Comparator compares two spatials and returns whether they are equal.
type Comparator func(obj1, obj2 Spatial) (equal bool)

func defaultComparator(obj1, obj2 Spatial) bool {
	return obj1 == obj2
}

*/

func (r *RTreeSpatialDatabase) RemoveFeature(ctx context.Context, id string) error {

	obj := &RTreeSpatialIndex{
		Rect: nil,
		Id:   id,
	}

	comparator := func(obj1, obj2 rtreego.Spatial) bool {

		// 2021/10/12 11:17:11 COMPARE 1: '101737491#:0' 2: '101737491'
		// log.Printf("COMPARE 1: '%v' 2: '%v'\n", obj1.(*RTreeSpatialIndex).Id, obj2.(*RTreeSpatialIndex).Id)

		obj1_id := obj1.(*RTreeSpatialIndex).Id
		obj2_id := obj2.(*RTreeSpatialIndex).Id

		return strings.HasPrefix(obj1_id, obj2_id)
	}

	ok := r.rtree.DeleteWithComparator(obj, comparator)

	if !ok {
		return fmt.Errorf("Failed to remove %s from rtree", id)
	}

	return nil
}

func (r *RTreeSpatialDatabase) PointInPolygon(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) (spr.StandardPlacesResults, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rsp_ch := make(chan spr.StandardPlacesResult)
	err_ch := make(chan error)
	done_ch := make(chan bool)

	results := make([]spr.StandardPlacesResult, 0)
	working := true

	go r.PointInPolygonWithChannels(ctx, rsp_ch, err_ch, done_ch, coord, filters...)

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case <-done_ch:
			working = false
		case rsp := <-rsp_ch:
			results = append(results, rsp)
		case err := <-err_ch:
			return nil, err
		default:
			// pass
		}

		if !working {
			break
		}
	}

	spr_results := &RTreeResults{
		Places: results,
	}

	return spr_results, nil
}

func (r *RTreeSpatialDatabase) PointInPolygonWithChannels(ctx context.Context, rsp_ch chan spr.StandardPlacesResult, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	defer func() {
		done_ch <- true
	}()

	rows, err := r.getIntersectsByCoord(coord)

	if err != nil {
		err_ch <- err
		return
	}

	r.inflateResultsWithChannels(ctx, rsp_ch, err_ch, rows, coord, filters...)
	return
}

func (r *RTreeSpatialDatabase) PointInPolygonCandidates(ctx context.Context, coord *orb.Point, filters ...spatial.Filter) ([]*spatial.PointInPolygonCandidate, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rsp_ch := make(chan *spatial.PointInPolygonCandidate)
	err_ch := make(chan error)
	done_ch := make(chan bool)

	candidates := make([]*spatial.PointInPolygonCandidate, 0)
	working := true

	go r.PointInPolygonCandidatesWithChannels(ctx, rsp_ch, err_ch, done_ch, coord, filters...)

	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case <-done_ch:
			working = false
		case rsp := <-rsp_ch:
			candidates = append(candidates, rsp)
		case err := <-err_ch:
			return nil, err
		default:
			// pass
		}

		if !working {
			break
		}
	}

	return candidates, nil
}

func (r *RTreeSpatialDatabase) PointInPolygonCandidatesWithChannels(ctx context.Context, rsp_ch chan *spatial.PointInPolygonCandidate, err_ch chan error, done_ch chan bool, coord *orb.Point, filters ...spatial.Filter) {

	defer func() {
		done_ch <- true
	}()

	intersects, err := r.getIntersectsByCoord(coord)

	if err != nil {
		err_ch <- err
		return
	}

	for _, raw := range intersects {

		sp := raw.(*RTreeSpatialIndex)

		// bounds := sp.Bounds()

		c := &spatial.PointInPolygonCandidate{
			Id:        sp.Id,
			FeatureId: sp.FeatureId,
			AltLabel:  sp.AltLabel,
			// FIX ME
			// Bounds:   bounds,
		}

		rsp_ch <- c
	}

	return
}

func (r *RTreeSpatialDatabase) getIntersectsByCoord(coord *orb.Point) ([]rtreego.Spatial, error) {

	lat := coord.Y()
	lon := coord.X()

	pt := rtreego.Point{lon, lat}
	rect, err := rtreego.NewRect(pt, []float64{0.0001, 0.0001}) // how small can I make this?

	if err != nil {
		return nil, fmt.Errorf("Failed to derive rtree bounds, %w", err)
	}

	return r.getIntersectsByRect(rect)
}

func (r *RTreeSpatialDatabase) getIntersectsByRect(rect *rtreego.Rect) ([]rtreego.Spatial, error) {

	// to do: timings that don't slow everything down the way
	// go-whosonfirst-timer does now (20170915/thisisaaronland)

	results := r.rtree.SearchIntersect(rect)
	return results, nil
}

func (r *RTreeSpatialDatabase) inflateResultsWithChannels(ctx context.Context, rsp_ch chan spr.StandardPlacesResult, err_ch chan error, possible []rtreego.Spatial, c *orb.Point, filters ...spatial.Filter) {

	seen := make(map[string]bool)

	mu := new(sync.RWMutex)
	wg := new(sync.WaitGroup)

	for _, row := range possible {

		sp := row.(*RTreeSpatialIndex)
		wg.Add(1)

		go func(sp *RTreeSpatialIndex) {

			sp_id := sp.Id
			feature_id := sp.FeatureId

			t1 := time.Now()

			defer func() {
				r.Timer.Add(ctx, sp_id, "time to inflate", time.Since(t1))
			}()

			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			mu.RLock()
			_, ok := seen[feature_id]
			mu.RUnlock()

			if ok {
				return
			}

			mu.Lock()
			seen[feature_id] = true
			mu.Unlock()

			t2 := time.Now()

			cache_item, err := r.retrieveCache(ctx, sp)

			r.Timer.Add(ctx, sp_id, "time to retrieve cache", time.Since(t2))

			if err != nil {
				r.Logger.Printf("Failed to retrieve cache for %s, %v", sp_id, err)
				return
			}

			s := cache_item.SPR

			t3 := time.Now()

			for _, f := range filters {

				err = filter.FilterSPR(f, s)

				if err != nil {
					// r.Logger.Debug("SKIP %s because filter error %s", sp_id, err)
					return
				}
			}

			r.Timer.Add(ctx, sp_id, "time to filter", time.Since(t3))

			t4 := time.Now()

			geom := cache_item.Geometry

			orb_geom := geom.Geometry()
			geom_type := orb_geom.GeoJSONType()

			contains := false

			switch geom_type {
			case "Polygon":
				contains = planar.PolygonContains(orb_geom.(orb.Polygon), *c)
			case "MultiPolygon":
				contains = planar.MultiPolygonContains(orb_geom.(orb.MultiPolygon), *c)
			default:
				r.Logger.Printf("Geometry has unsupported geometry type '%s'", geom.Type)
			}

			r.Timer.Add(ctx, sp_id, "time to test geometry", time.Since(t4))

			if !contains {
				// r.Logger.Debug("SKIP %s because does not contain coord (%v)", sp_id, c)
				return
			}

			rsp_ch <- s
		}(sp)
	}

	wg.Wait()
}

func (r *RTreeSpatialDatabase) setCache(ctx context.Context, body []byte) error {

	s, err := spr.WhosOnFirstSPR(body)

	if err != nil {
		return err
	}

	geom, err := geometry.Geometry(body)

	if err != nil {
		return fmt.Errorf("Failed to derive geometry for feature, %w", err)
	}

	alt_label, err := properties.AltLabel(body)

	if err != nil {
		return fmt.Errorf("Failed to derive alt label, %w", err)
	}

	feature_id, err := properties.Id(body)

	if err != nil {
		return fmt.Errorf("Failed to derive feature ID, %w", err)
	}

	cache_key := fmt.Sprintf("%d:%s", feature_id, alt_label)

	cache_item := &RTreeCache{
		Geometry: geom,
		SPR:      s,
	}

	r.gocache.Set(cache_key, cache_item, -1)
	return nil
}

func (r *RTreeSpatialDatabase) retrieveCache(ctx context.Context, sp *RTreeSpatialIndex) (*RTreeCache, error) {

	cache_key := fmt.Sprintf("%s:%s", sp.FeatureId, sp.AltLabel)

	cache_item, ok := r.gocache.Get(cache_key)

	if !ok {
		return nil, fmt.Errorf("Invalid cache ID '%s'", cache_key)
	}

	return cache_item.(*RTreeCache), nil
}

// whosonfirst/go-reader interface

func (r *RTreeSpatialDatabase) Read(ctx context.Context, str_uri string) (io.ReadSeekCloser, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI %s, %w", str_uri, err)
	}

	// TO DO : ALT STUFF HERE

	str_id := strconv.FormatInt(id, 10)

	sp := &RTreeSpatialIndex{
		FeatureId: str_id,
		AltLabel:  "",
	}

	cache_item, err := r.retrieveCache(ctx, sp)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve cache, %w", err)
	}

	// START OF this is dumb

	enc_spr, err := json.Marshal(cache_item.SPR)

	if err != nil {
		return nil, fmt.Errorf("Failed to marchal cache record, %w", err)
	}

	var props map[string]interface{}

	err = json.Unmarshal(enc_spr, &props)

	if err != nil {
		return nil, err
	}

	// END OF this is dumb

	orb_geom := cache_item.Geometry.Geometry()
	f := geojson.NewFeature(orb_geom)

	if err != nil {
		return nil, err
	}

	f.Properties = props

	enc_f, err := f.MarshalJSON()

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(enc_f)
	return ioutil.NewReadSeekCloser(br)
}

func (r *RTreeSpatialDatabase) ReaderURI(ctx context.Context, str_uri string) string {
	return str_uri
}

// whosonfirst/go-writer interface

func (r *RTreeSpatialDatabase) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {
	return 0, fmt.Errorf("Not implemented")
}

func (r *RTreeSpatialDatabase) WriterURI(ctx context.Context, str_uri string) string {
	return str_uri
}

func (r *RTreeSpatialDatabase) Flush(ctx context.Context) error {
	return nil
}

// Close defined above

func (r *RTreeSpatialDatabase) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
