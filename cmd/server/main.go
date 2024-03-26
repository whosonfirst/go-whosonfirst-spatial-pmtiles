package main

import (
	"context"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/application/server"
	"log"
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := server.Run(ctx, logger)

	if err != nil {
		logger.Fatal(err)
	}
}
