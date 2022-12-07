package main

// go run cmd/pmtile/main.go -tiles file:///tmp -database test -x 655 -y 1585 | jq

import (
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/docstore/memdocstore"
)

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/maptile"
	"github.com/protomaps/go-pmtiles/pmtiles"
	"log"
	"os"
)

func main() {

	var tile_path string
	var database string

	var z int
	var x int
	var y int

	flag.StringVar(&tile_path, "tiles", "", "")
	flag.StringVar(&database, "database", "", "")

	flag.IntVar(&z, "z", 12, "")
	flag.IntVar(&x, "x", 0, "")
	flag.IntVar(&y, "y", 0, "")

	flag.Parse()

	ctx := context.Background()

	logger := log.Default()
	cache_size := 64

	loop, err := pmtiles.NewLoop(tile_path, logger, cache_size, "")

	if err != nil {
		log.Fatalf("Failed to create pmtiles.Loop, %v", err)
	}

	loop.Start()

	zm := maptile.Zoom(uint32(z))
	t := maptile.New(uint32(x), uint32(y), zm)

	path := fmt.Sprintf("/%s/%d/%d/%d.mvt", database, z, x, y)

	status_code, _, body := loop.Get(ctx, path)

	if status_code != 200 {
		log.Fatalf("%s returns status code %d\n", path, status_code)
	}

	layers, err := mvt.UnmarshalGzipped(body)

	if err != nil {
		log.Fatalf("Failed to unmarshal gzipped body for %s, %v", path, err)
	}

	layers.ProjectToWGS84(t)

	fc := layers.ToFeatureCollections()

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(fc)

	if err != nil {
		log.Fatalf("Failed to encode feature collection for %s, %v", path, err)
	}
}
