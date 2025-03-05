package pmtiles

import (
	"testing"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"log/slog"
	"io"
	"os"
	
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"	
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
)

func TestDatabase(t *testing.T) {

	rel_path := "fixtures/sf.pmtiles"
	abs_path, err := filepath.Abs(rel_path)

	if err != nil {
		t.Fatalf("Failed to derive absolute path for %s, %v", rel_path, err)
	}

	root := filepath.Dir(abs_path)
	fname := filepath.Base(abs_path)

	fname = strings.Replace(fname, ".pmtiles", "", 1)
	
	db_uri := fmt.Sprintf("pmtiles://?tiles=file://%s&database=%s&zoom=13&enable_cache=true&layer=whosonfirst", root, fname)
	
	ctx := context.Background()
	
	db, err := database.NewSpatialDatabase(ctx, db_uri)
	
	if err != nil {
		t.Fatalf("Failed to create new spatial database for %s, %v", db_uri, err)
	}

	err = db.Disconnect(ctx)

	if err != nil {
		t.Fatalf("Failed to disconnect database, %v", err)
	}
}

func TestPointInPolygon(t *testing.T) {

	rel_path := "fixtures/sf.pmtiles"
	abs_path, err := filepath.Abs(rel_path)

	if err != nil {
		t.Fatalf("Failed to derive absolute path for %s, %v", rel_path, err)
	}

	root := filepath.Dir(abs_path)
	fname := filepath.Base(abs_path)

	fname = strings.Replace(fname, ".pmtiles", "", 1)
	
	db_uri := fmt.Sprintf("pmtiles://?tiles=file://%s&database=%s&zoom=13&enable_cache=true&layer=whosonfirst", root, fname)
	
	ctx := context.Background()
	
	db, err := database.NewSpatialDatabase(ctx, db_uri)
	
	if err != nil {
		t.Fatalf("Failed to create new spatial database for %s, %v", db_uri, err)
	}

	defer db.Disconnect(ctx)

	lat := 37.759415
	lon := -122.414647

	pt := orb.Point([2]float64{ lon, lat })

	i, err := filter.NewSPRInputs()
	
	if err != nil {
		t.Fatalf("Failed to create SPR inputs, %v", err)
	}
	
	i.IsCurrent = []int64{ 1 }
	
	f, err := filter.NewSPRFilterFromInputs(i)
	
	if err != nil {
		t.Fatalf("Failed to create SPR filter from inputs, %v", err)
	}
	
	rsp , err := db.PointInPolygon(ctx, &pt, f)

	if err != nil {
		t.Fatalf("Failed to perform point in polygon query, %v", err)
	}

	results := rsp.Results()
	count := len(results)

	expected := 9

	if count != expected {
		t.Fatalf("Unexpected count (%d), expected %d", count, expected)
	}

	/*
	slog.Info("count", "c", count)
	
	for _, r := range results {
		slog.Info("r", "id", r.Id(), "name", r.Name())
	}
	*/
}

func TestIntersects(t *testing.T) {
	
	rel_path := "fixtures/sf.pmtiles"
	abs_path, err := filepath.Abs(rel_path)

	if err != nil {
		t.Fatalf("Failed to derive absolute path for %s, %v", rel_path, err)
	}

	root := filepath.Dir(abs_path)
	fname := filepath.Base(abs_path)

	fname = strings.Replace(fname, ".pmtiles", "", 1)
	
	db_uri := fmt.Sprintf("pmtiles://?tiles=file://%s&database=%s&zoom=13&enable_cache=true&layer=whosonfirst", root, fname)
	
	ctx := context.Background()
	
	db, err := database.NewSpatialDatabase(ctx, db_uri)
	
	if err != nil {
		t.Fatalf("Failed to create new spatial database for %s, %v", db_uri, err)
	}

	defer db.Disconnect(ctx)

	r, err := os.Open("fixtures/1108830809.geojson")

	if err != nil {
		t.Fatalf("Failed to open 1108830809 for reading, %v", err)
	}

	defer r.Close()

	body, err := io.ReadAll(r)

	if err != nil {
		t.Fatalf("Failed to read 1108830809, %v", err)
	}

	f, err := geojson.UnmarshalFeature(body)

	if err != nil {
		t.Fatalf("Failed to unmarshal 1108830809, %v", err)
	}

	geom := f.Geometry
	
	rsp, err := db.Intersects(ctx, geom)

	if err != nil {
		t.Fatalf("Failed to perform intersects query, %v", err)
	}

	results := rsp.Results()
	count := len(results)

	/*
	expected := 9

	if count != expected {
		t.Fatalf("Unexpected count (%d), expected %d", count, expected)
	}
	*/

	slog.Info("count", "c", count)
	
	for _, r := range results {
		slog.Info("r", "id", r.Id(), "name", r.Name())
	}

	
}
