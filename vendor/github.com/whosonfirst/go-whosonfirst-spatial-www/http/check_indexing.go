package http

import (
	gohttp "net/http"

	app "github.com/whosonfirst/go-whosonfirst-spatial/application"
)

// CheckIndexingHandler() returns a `http.Handler` that will check with 'app.Iterator' is currently
// indexing records. If it is the handler will return an HTTP 503 error. If it is not it will pass
// the request on to 'next'.
func CheckIndexingHandler(app *app.SpatialApplication, next gohttp.Handler) gohttp.Handler {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		if app.IsIndexing() {
			gohttp.Error(rsp, "Service unavailable: indexing", gohttp.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(rsp, req)
	}

	return gohttp.HandlerFunc(fn)
}
