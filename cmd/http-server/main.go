package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
	"github.com/whosonfirst/go-whosonfirst-spatial-www/app/server"
)

func main() {

	ctx := context.Background()
	err := server.Run(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
