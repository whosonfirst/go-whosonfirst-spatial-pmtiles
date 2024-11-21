package server

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"

	grpc_server "github.com/whosonfirst/go-whosonfirst-spatial-grpc/server"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
	app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	"google.golang.org/grpc"
)

func Run(ctx context.Context) error {

	fs, err := DefaultFlagSet()

	if err != nil {
		return fmt.Errorf("Failed to derive default flag set, %w", err)
	}

	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	opts, err := RunOptionsFromFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to derive options from flagset, %w", err)
	}

	return RunWithOptions(ctx, opts)
}

func RunWithOptions(ctx context.Context, opts *RunOptions) error {

	spatial_opts := &app.SpatialApplicationOptions{
		SpatialDatabaseURI:     opts.SpatialDatabaseURI,
		PropertiesReaderURI:    opts.PropertiesReaderURI,
		EnableCustomPlacetypes: opts.EnableCustomPlacetypes,
		CustomPlacetypes:       opts.CustomPlacetypes,
	}

	spatial_app, err := app.NewSpatialApplication(ctx, spatial_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new spatial application, %w", err)
	}

	go func() {

		err := spatial_app.IndexDatabaseWithIterators(ctx, opts.IteratorSources)

		if err != nil {
			slog.Error("Failed to index database", "error", err)
		}
	}()

	spatial_server, err := grpc_server.NewSpatialServer(spatial_app)

	if err != nil {
		return fmt.Errorf("Failed to create spatial server, %v", err)
	}

	grpc_server := grpc.NewServer()

	spatial.RegisterSpatialServer(grpc_server, spatial_server)

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	slog.Info("Listening for requests", "address", addr)

	lis, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpc_server.Serve(lis)
	return nil
}
