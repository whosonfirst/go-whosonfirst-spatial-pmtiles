package http

import (
	"log"
	gohttp "net/http"
	"time"

	"github.com/protomaps/go-pmtiles/pmtiles"
)

// TileHandler provides a `http.Handler` for serving Protomaps tile requests using 'server'.
func TileHandler(server *pmtiles.Server, logger *log.Logger) gohttp.Handler {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		start := time.Now()

		status_code, headers, body := server.Get(req.Context(), req.URL.Path)

		for k, v := range headers {
			rsp.Header().Set(k, v)
		}

		rsp.WriteHeader(status_code)
		rsp.Write(body)

		go logger.Printf("[%d] served %s in %s", status_code, req.URL.Path, time.Since(start))
	}

	return gohttp.HandlerFunc(fn)
}
