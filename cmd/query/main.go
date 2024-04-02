package main

import (
	_ "github.com/aaronland/gocloud-blob-s3"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	_ "gocloud.dev/blob/s3blob"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial/app/pip"
	"log"
)

func main() {

	ctx := context.Background()

	logger := log.Default()

	err := pip.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run PIP application, %v", err)
	}

}
