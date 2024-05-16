package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"github.com/whosonfirst/go-whosonfirst-spatial/app/hierarchy/update"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := update.Run(ctx, logger)

	if err != nil {
		logger.Fatal(err)
	}
}
