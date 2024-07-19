package client

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/request"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/pip"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
)

func Run(ctx context.Context, logger *log.Logger) error {

	fs, err := DefaultFlagSet()

	if err != nil {
		return fmt.Errorf("Failed to create default flagset, %w", err)
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

	pip_req := &pip.PointInPolygonRequest{
		Latitude:            opts.Latitude,
		Longitude:           opts.Longitude,
		Placetypes:          opts.Placetypes,
		Geometries:          opts.Geometries,
		AlternateGeometries: opts.AlternateGeometries,
		IsCurrent:           opts.IsCurrent,
		IsCeased:            opts.IsCeased,
		IsDeprecated:        opts.IsDeprecated,
		IsSuperseded:        opts.IsSuperseded,
		IsSuperseding:       opts.IsSuperseding,
		InceptionDate:       opts.InceptionDate,
		CessationDate:       opts.CessationDate,
		Properties:          opts.Properties,
		Sort:                opts.Sort,
	}

	spatial_req, err := request.NewPointInPolygonRequest(pip_req)

	if err != nil {
		return fmt.Errorf("Failed to create spatial PIP request, %w", err)
	}

	var grpc_opts []grpc.DialOption
	grpc_opts = append(grpc_opts, grpc.WithInsecure())

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)

	conn, err := grpc.Dial(addr, grpc_opts...)

	if err != nil {
		return fmt.Errorf("Failed to dial '%s', %v", addr, err)
	}

	defer conn.Close()

	client := spatial.NewSpatialClient(conn)

	stream, err := client.PointInPolygon(ctx, spatial_req)

	if err != nil {
		return fmt.Errorf("Failed to perform point in polygon operation, %w", err)
	}

	writers := make([]io.Writer, 0)

	if opts.Stdout {
		writers = append(writers, os.Stdout)
	}

	if opts.Null {
		writers = append(writers, io.Discard)
	}

	wr := io.MultiWriter(writers...)

	enc := json.NewEncoder(wr)
	err = enc.Encode(stream)

	if err != nil {
		return fmt.Errorf("Failed to encode stream, %v", err)
	}

	return nil
}
