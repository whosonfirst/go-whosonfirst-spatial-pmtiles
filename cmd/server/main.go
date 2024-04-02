package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/app/server"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := server.Run(ctx, logger)

	if err != nil {
		logger.Fatal(err)
	}
}
