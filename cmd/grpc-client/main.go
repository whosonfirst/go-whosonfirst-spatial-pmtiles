package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/app/client"
)

func main() {

	ctx := context.Background()
	err := client.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run client, %v", err)
	}
}
