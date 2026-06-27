package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"github.com/whosonfirst/go-whosonfirst/v4/app/spatial/intersects"
)

func main() {

	ctx := context.Background()
	err := intersects.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
