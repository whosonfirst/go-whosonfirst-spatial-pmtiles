package main

import (
	_ "gocloud.dev/blob/fileblob"
)

import (
	"context"
	"flag"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"log"
)

func main() {

	db_uri := flag.String("spatial-database-uri", "", "")

	flag.Parse()

	ctx := context.Background()

	db, err := pmtiles.NewPMTilesDatabase(ctx, *db_uri)

	if err != nil {
		log.Fatalf("Failed to create database, %v", err)
	}

	lat := 37.621131
	lon := -122.384292

	coord := &orb.Point{lon, lat}

	rsp, err := db.PointInPolygon(ctx, coord)

	if err != nil {
		log.Fatalf("Failed to perform point in polygon, %v", err)
	}

	for _, r := range rsp.Results() {
		fmt.Println(r.Id(), r.Name())
	}
}
