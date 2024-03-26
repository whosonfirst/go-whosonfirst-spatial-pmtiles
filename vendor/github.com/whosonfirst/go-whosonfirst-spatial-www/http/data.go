package http

// TBD: make this part of whosonfirst/go-reader package...

import (
	"fmt"
	"io"
	gohttp "net/http"

	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func NewDataHandler(r reader.Reader) (gohttp.Handler, error) {

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		path := req.URL.Path

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			e := fmt.Errorf("Failed to parse %s, %w", path, err)
			gohttp.Error(rsp, e.Error(), gohttp.StatusBadRequest)
			return
		}

		rel_path, err := uri.Id2RelPath(id, uri_args)

		if err != nil {
			e := fmt.Errorf("Failed to derive path for %d, %w", id, err)
			gohttp.Error(rsp, e.Error(), gohttp.StatusBadRequest)
			return
		}

		ctx := req.Context()
		fh, err := r.Read(ctx, rel_path)

		if err != nil {
			e := fmt.Errorf("Failed to load %s, %w", rel_path, err)
			gohttp.Error(rsp, e.Error(), gohttp.StatusBadRequest)
			return
		}

		rsp.Header().Set("Content-Type", "application/json")

		_, err = io.Copy(rsp, fh)

		if err != nil {
			e := fmt.Errorf("Failed to copy %s, %w", rel_path, err)
			gohttp.Error(rsp, e.Error(), gohttp.StatusBadRequest)
			return
		}

		return
	}

	h := gohttp.HandlerFunc(fn)
	return h, nil
}
