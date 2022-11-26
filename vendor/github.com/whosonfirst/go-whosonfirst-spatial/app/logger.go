package app

import (
	"context"
	"flag"
	_ "github.com/sfomuseum/go-flags/lookup"
	_ "github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"io"
	"log"
	"os"
)

func NewApplicationLoggerWithFlagSet(ctx context.Context, fl *flag.FlagSet) (*log.Logger, error) {

	writers := []io.Writer{
		os.Stdout,
	}

	mw := io.MultiWriter(writers...)

	logger := log.New(mw, "[spatial] ", log.Lshortfile)
	return logger, nil
}
