package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"github.com/whosonfirst/go-whosonfirst-spatial/app/intersects"
)

func main() {

	ctx := context.Background()
	err := intersects.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
