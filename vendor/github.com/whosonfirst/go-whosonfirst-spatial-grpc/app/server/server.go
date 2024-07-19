package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	grpc_server "github.com/whosonfirst/go-whosonfirst-spatial-grpc/server"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
	app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, logger *log.Logger) error {

	fs, err := DefaultFlagSet()

	if err != nil {
		return fmt.Errorf("Failed to derive default flag set, %w", err)
	}

	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	opts, err := RunOptionsFromFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to derive options from flagset, %w", err)
	}

	return RunWithOptions(ctx, opts, logger)
}

func RunWithOptions(ctx context.Context, opts *RunOptions, logger *log.Logger) error {

	spatial_opts := &app.SpatialApplicationOptions{
		SpatialDatabaseURI:     opts.SpatialDatabaseURI,
		PropertiesReaderURI:    opts.PropertiesReaderURI,
		IteratorURI:            opts.IteratorURI,
		EnableCustomPlacetypes: opts.EnableCustomPlacetypes,
		CustomPlacetypes:       opts.CustomPlacetypes,
		IsWhosOnFirst:          opts.IsWhosOnFirst,
	}

	spatial_app, err := app.NewSpatialApplication(ctx, spatial_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new spatial application, %w", err)
	}

	if len(opts.IteratorSources) > 0 {

		err = spatial_app.IndexPaths(ctx, opts.IteratorSources...)

		if err != nil {
			return fmt.Errorf("Failed to index paths, %v", err)
		}
	}

	spatial_server, err := grpc_server.NewSpatialServer(spatial_app)

	if err != nil {
		return fmt.Errorf("Failed to create spatial server, %v", err)
	}

	grpc_server := grpc.NewServer()

	spatial.RegisterSpatialServer(grpc_server, spatial_server)

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	log.Printf("Listening on %s\n", addr)

	lis, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpc_server.Serve(lis)
	return nil
}
