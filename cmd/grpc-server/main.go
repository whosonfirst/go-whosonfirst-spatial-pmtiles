package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"

	"github.com/whosonfirst/go-whosonfirst/v4/app/spatial/grpc/server"
)

func main() {

	ctx := context.Background()
	err := server.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run client, %v", err)
	}
}
