package api

// TBD: make this part of whosonfirst/go-reader package...

import (
	"fmt"
	"io"
	"net/http"

	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func NewDataHandler(r reader.Reader) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		path := req.URL.Path

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			e := fmt.Errorf("Failed to parse %s, %w", path, err)
			http.Error(rsp, e.Error(), http.StatusBadRequest)
			return
		}

		rel_path, err := uri.Id2RelPath(id, uri_args)

		if err != nil {
			e := fmt.Errorf("Failed to derive path for %d, %w", id, err)
			http.Error(rsp, e.Error(), http.StatusBadRequest)
			return
		}

		ctx := req.Context()
		fh, err := r.Read(ctx, rel_path)

		if err != nil {
			e := fmt.Errorf("Failed to load %s, %w", rel_path, err)
			http.Error(rsp, e.Error(), http.StatusBadRequest)
			return
		}

		rsp.Header().Set("Content-Type", "application/json")

		_, err = io.Copy(rsp, fh)

		if err != nil {
			e := fmt.Errorf("Failed to copy %s, %w", rel_path, err)
			http.Error(rsp, e.Error(), http.StatusBadRequest)
			return
		}

		return
	}

	h := http.HandlerFunc(fn)
	return h, nil
}
