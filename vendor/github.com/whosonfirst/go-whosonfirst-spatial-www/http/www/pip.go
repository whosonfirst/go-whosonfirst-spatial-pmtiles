package www

import (
	"errors"
	"html/template"
	_ "log"
	"net/http"

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

func PointInPolygonHandler(spatial_app *app.SpatialApplication, opts *PointInPolygonHandlerOptions) (http.Handler, error) {

	t := opts.Templates.Lookup("pointinpolygon")

	if t == nil {
		return nil, errors.New("Missing pointinpolygon template")
	}

	iterator := spatial_app.Iterator

	pt_list, err := placetypes.Placetypes()

	if err != nil {
		return nil, err
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		if iterator.IsIndexing() {
			http.Error(rsp, "indexing records", http.StatusServiceUnavailable)
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
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	h := http.HandlerFunc(fn)
	return h, nil
}
