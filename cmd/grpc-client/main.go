package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/app/client"	
)

func main() {

	ctx := context.Background()
	logger := log.Default()

	err := client.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run client, %v", err)
	}
}
