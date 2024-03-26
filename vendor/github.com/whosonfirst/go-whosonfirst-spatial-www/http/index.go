package http

import (
	"errors"
	"html/template"
	gohttp "net/http"
)

type IndexHandlerOptions struct {
	Templates *template.Template
}

func IndexHandler(opts *IndexHandlerOptions) (gohttp.Handler, error) {

	t := opts.Templates.Lookup("index")

	if t == nil {
		return nil, errors.New("Missing 'index' template")
	}

	fn := func(rsp gohttp.ResponseWriter, req *gohttp.Request) {

		rsp.Header().Set("Content-Type", "text/html; charset=utf-8")

		err := t.Execute(rsp, nil)

		if err != nil {
			gohttp.Error(rsp, err.Error(), gohttp.StatusInternalServerError)
			return
		}

		return
	}

	h := gohttp.HandlerFunc(fn)
	return h, nil
}
