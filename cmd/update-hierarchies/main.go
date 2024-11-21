package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"github.com/whosonfirst/go-whosonfirst-spatial/app/hierarchy/update"
)

func main() {

	ctx := context.Background()
	err := update.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
