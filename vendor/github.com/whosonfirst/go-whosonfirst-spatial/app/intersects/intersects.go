package intersects

/*

$> ./bin/intersects -geometry-source flag -geometry-type bbox -geometry-value '-122.408061,37.601617,-122.354907,37.640167' -iterator-uri repo://#/usr/local/data/sfomuseum-data-whosonfirst | jq -r '.places[]["wof:name"]'
2025/03/07 08:55:43 INFO time to index paths (1) 26.923220625s

Earth
94128
Burlingame
San Mateo
San Francisco International Airport
California
United States
North America

*/

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	app "github.com/whosonfirst/go-whosonfirst-spatial/application"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
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

	intersects_fn, err := query.NewSpatialFunction(ctx, "intersects://")

	if err != nil {
		return fmt.Errorf("Failed to create intersects function, %w", err)
	}

	switch opts.Mode {

	case "cli":

		props := opts.Properties

		<-done_ch

		// bounding box
		// geojson
		// wkt
		//
		// flag
		// stdin
		// file

		var geom_raw []byte
		var geom *geojson.Geometry

		switch opts.GeometrySource {
		case "file":

			r, err := os.Open(opts.GeometryValue)

			if err != nil {
				return fmt.Errorf("Failed to open %s for reading, %w", opts.GeometryValue, err)
			}

			defer r.Close()

			body, err := io.ReadAll(r)

			if err != nil {
				return fmt.Errorf("Failed to read data from %s, %w", opts.GeometryValue, err)
			}

			geom_raw = body

		case "stdin":

			body, err := io.ReadAll(os.Stdin)

			if err != nil {
				return fmt.Errorf("Failed to read from STDIN, %w", err)
			}

			geom_raw = body
		default:
			geom_raw = []byte(opts.GeometryValue)
		}

		switch opts.GeometryType {
		case "geojson":

			f, err := geojson.UnmarshalFeature(geom_raw)

			if err != nil {
				return fmt.Errorf("Failed to unmarshal GeoJSON, %w", err)
			}

			orb_geom := f.Geometry
			geom = geojson.NewGeometry(orb_geom)

		case "wkt":

			orb_geom, err := wkt.Unmarshal(string(geom_raw))

			if err != nil {
				return fmt.Errorf("Failed to unmarshal WKT, %w", err)
			}

			geom = geojson.NewGeometry(orb_geom)

		default:

			is_latlon := false

			f, err := geo.BoundingBoxToFeature(string(geom_raw), is_latlon)

			if err != nil {
				return fmt.Errorf("Failed to parse bounding box, %w", err)
			}

			orb_geom := f.Geometry
			geom = geojson.NewGeometry(orb_geom)
		}

		intersects_q := &query.SpatialQuery{
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

		intersects_rsp, err := query.ExecuteQuery(ctx, spatial_app.SpatialDatabase, intersects_fn, intersects_q)

		if err != nil {
			return fmt.Errorf("Failed to perform intersects query, %v", err)
		}

		rsp = intersects_rsp

		if len(props) > 0 {

			props_opts := &spatial.PropertiesResponseOptions{
				Reader:       spatial_app.PropertiesReader,
				Keys:         props,
				SourcePrefix: "properties",
			}

			props_rsp, err := spatial.PropertiesResponseResultsWithStandardPlacesResults(ctx, props_opts, intersects_rsp)

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

		handler := func(ctx context.Context, intersects_q *query.SpatialQuery) (interface{}, error) {
			return query.ExecuteQuery(ctx, spatial_app.SpatialDatabase, intersects_fn, intersects_q)
		}

		lambda.Start(handler)

	default:
		return fmt.Errorf("Invalid or unsupported mode '%s'", mode)
	}

	return nil
}
