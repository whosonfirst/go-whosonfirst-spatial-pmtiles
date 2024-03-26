package http

import (
	"errors"
	"html/template"
	_ "log"
	gohttp "net/http"

	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-spatial/app"
)

type PointInPolygonHandlerOptions struct {
	Templates        *template.Template
	InitialLatitude  float64
	InitialLongitude float64
	InitialZoom      int
	MapProvider      string
	MaxBounds        string
	LeafletTileURL   string
}

type PointInPolygonHandlerTemplateVars struct {
	InitialLatitude  float64
	InitialLongitude float64
	InitialZoom      int
	MaxBounds        string
	MapProvider      string
	LeafletTileURL   string
	Placetypes       []*placetypes.WOFPlacetype
}

func PointInPolygonHandler(spatial_app *app.SpatialApplication, opts *PointInPolygonHandlerOptions) (gohttp.Handler, error) {

	t := opts.Templates.Lookup("pointinpolygon")

	if t == nil {
		return nil, errors.New("Missing pointinpolygon template")
	}

	iterator := spatial_app.Iterator

	pt_list, err := placetypes.Placetypes()

	if err != nil {
		return nil, err
	}

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		if iterator.IsIndexing() {
			gohttp.Error(rsp, "indexing records", gohttp.StatusServiceUnavailable)
			return
		}

		// important if we're trying to use this in a Lambda/API Gateway context

		rsp.Header().Set("Content-Type", "text/html; charset=utf-8")

		vars := PointInPolygonHandlerTemplateVars{
			InitialLatitude:  opts.InitialLatitude,
			InitialLongitude: opts.InitialLongitude,
			InitialZoom:      opts.InitialZoom,
			LeafletTileURL:   opts.LeafletTileURL,
			MaxBounds:        opts.MaxBounds,
			MapProvider:      opts.MapProvider,
			Placetypes:       pt_list,
		}

		err := t.Execute(rsp, vars)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	h := gohttp.HandlerFunc(fn)
	return h, nil
}
