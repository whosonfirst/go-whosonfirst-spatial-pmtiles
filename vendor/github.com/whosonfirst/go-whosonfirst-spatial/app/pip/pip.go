package pip

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	// "github.com/whosonfirst/go-whosonfirst-spatial/pip"
	"github.com/whosonfirst/go-whosonfirst-spatial/query"
)

func Run(ctx context.Context) error {

	fs, err := DefaultFlagSet(ctx)

	if err != nil {
		return fmt.Errorf("Failed to create application flag set, %v", err)
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

	if opts.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	done_ch := make(chan bool)

	go func() {

		err := spatial_app.IndexDatabaseWithIterators(ctx, opts.IteratorSources)

		if err != nil {
			slog.Error("Failed to index database", "error", err)
		}

		done_ch <- true
	}()

	fn, err := query.NewSpatialFunction(ctx, "pip://")

	if err != nil {
		return fmt.Errorf("Failed to create point in polygon function, %w", err)
	}

	switch mode {

	case "cli":

		props := opts.Properties

		<-done_ch

		pt := orb.Point([2]float64{opts.Longitude, opts.Latitude})
		geom := geojson.NewGeometry(pt)

		req := &query.SpatialQuery{
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

		var rsp interface{}

		pip_rsp, err := query.ExecuteQuery(ctx, spatial_app.SpatialDatabase, fn, req)

		if err != nil {
			return fmt.Errorf("Failed to query, %v", err)
		}

		rsp = pip_rsp

		if len(props) > 0 {

			props_opts := &spatial.PropertiesResponseOptions{
				Reader:       spatial_app.PropertiesReader,
				Keys:         props,
				SourcePrefix: "properties",
			}

			props_rsp, err := spatial.PropertiesResponseResultsWithStandardPlacesResults(ctx, props_opts, pip_rsp)

			if err != nil {
				return fmt.Errorf("Failed to generate properties response, %v", err)
			}

			rsp = props_rsp
		}

		enc, err := json.Marshal(rsp)

		if err != nil {
			return fmt.Errorf("Failed to marshal results, %v", err)
		}

		fmt.Println(string(enc))

	case "lambda":

		<-done_ch

		handler := func(ctx context.Context, req *query.SpatialQuery) (interface{}, error) {
			return query.ExecuteQuery(ctx, spatial_app.SpatialDatabase, fn, req)
		}

		lambda.Start(handler)

	default:
		return fmt.Errorf("Invalid or unsupported mode '%s'", mode)
	}

	return nil
}
