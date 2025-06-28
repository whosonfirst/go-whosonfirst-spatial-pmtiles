package client

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/request"
	"github.com/whosonfirst/go-whosonfirst-spatial-grpc/spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spatial/query"
	"google.golang.org/grpc"
)

func Run(ctx context.Context) error {

	fs, err := DefaultFlagSet()

	if err != nil {
		return fmt.Errorf("Failed to create default flagset, %w", err)
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

	pt, err := geo.NewCoordinate(opts.Longitude, opts.Latitude)

	if err != nil {
		return err
	}

	geom := geojson.NewGeometry(pt)

	pip_q := &query.SpatialQuery{
		Geometry:            geom,
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

	spatial_req, err := request.NewPointInPolygonRequest(pip_q)

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
