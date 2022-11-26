package query

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-server"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/http/api"
	"github.com/whosonfirst/go-whosonfirst-spatial/app"
	spatial_flags "github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"log"
	gohttp "net/http"
)

type RunOptions struct {
	Logger  *log.Logger
	FlagSet *flag.FlagSet
}

func Run(ctx context.Context, logger *log.Logger) error {

	fs, err := DefaultFlagSet(ctx)

	if err != nil {
		return fmt.Errorf("Failed to create application flag set, %v", err)
	}

	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	opts := &RunOptions{
		Logger:  logger,
		FlagSet: fs,
	}

	return RunWithOptions(ctx, opts)
}

func RunWithOptions(ctx context.Context, opts *RunOptions) error {

	fs := opts.FlagSet
	logger := opts.Logger

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "WHOSONFIRST")

	if err != nil {
		return err
	}

	err = spatial_flags.ValidateCommonFlags(fs)

	if err != nil {
		return err
	}

	err = spatial_flags.ValidateQueryFlags(fs)

	if err != nil {
		return err
	}

	err = spatial_flags.ValidateIndexingFlags(fs)

	if err != nil {
		return err
	}

	spatial_app, err := app.NewSpatialApplicationWithFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to create new spatial application, %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	uris := fs.Args()

	done_ch := make(chan bool)

	go func() {

		err := spatial_app.Iterator.IterateURIs(ctx, uris...)

		if err != nil {
			logger.Printf("Failed to iterate URIs, %v", err)
		}

		done_ch <- true
	}()

	switch mode {

	case "cli":

		props, err := lookup.MultiStringVar(fs, spatial_flags.PROPERTIES)

		if err != nil {
			return err
		}

		<-done_ch

		req, err := pip.NewPointInPolygonRequestFromFlagSet(fs)

		if err != nil {
			return fmt.Errorf("Failed to create SPR filter, %v", err)
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

	case "server":

		pip_opts := &api.PointInPolygonHandlerOptions{
			EnableGeoJSON: enable_geojson,
		}

		pip_handler, err := api.PointInPolygonHandler(spatial_app, pip_opts)

		if err != nil {
			return fmt.Errorf("Failed to create PIP handler, %v", err)
		}

		mux := gohttp.NewServeMux()
		mux.Handle("/", pip_handler)

		s, err := server.NewServer(ctx, server_uri)

		if err != nil {
			return fmt.Errorf("Failed to create server for '%s', %v", server_uri, err)
		}

		log.Printf("Listening for requests at %s\n", s.Address())

		err = s.ListenAndServe(ctx, mux)

		if err != nil {
			return fmt.Errorf("Failed to serve requests for '%s', %v", server_uri, err)
		}

	default:
		return fmt.Errorf("Invalid or unsupported mode '%s'", mode)
	}

	return nil
}
