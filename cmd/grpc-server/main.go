package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/app/server"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
)

func main() {

	ctx := context.Background()
	err := server.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run client, %v", err)
	}
}
