package pip

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	"github.com/whosonfirst/go-whosonfirst-spatial/pip"
)

func Run(ctx context.Context, logger *log.Logger) error {

	fs, err := DefaultFlagSet(ctx)

	if err != nil {
		return fmt.Errorf("Failed to create application flag set, %v", err)
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	done_ch := make(chan bool)

	go func() {

		err := spatial_app.Iterator.IterateURIs(ctx, opts.Sources...)

		if err != nil {
			logger.Printf("Failed to iterate URIs, %v", err)
		}

		done_ch <- true
	}()

	switch mode {

	case "cli":

		props := opts.Properties

		<-done_ch

		req := &pip.PointInPolygonRequest{
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

		var rsp interface{}

		pip_rsp, err := pip.QueryPointInPolygon(ctx, spatial_app, req)

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

		handler := func(ctx context.Context, req *pip.PointInPolygonRequest) (interface{}, error) {
			return pip.QueryPointInPolygon(ctx, spatial_app, req)
		}

		lambda.Start(handler)

	default:
		return fmt.Errorf("Invalid or unsupported mode '%s'", mode)
	}

	return nil
}
