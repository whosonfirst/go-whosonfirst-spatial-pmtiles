package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	_ "gocloud.dev/blob/s3blob"
	_ "github.com/aaronland/gocloud-blob-s3"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/app/query"
	"log"
)

func main() {

	ctx := context.Background()

	logger := log.Default()

	err := query.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run PIP application, %v", err)
	}

}
