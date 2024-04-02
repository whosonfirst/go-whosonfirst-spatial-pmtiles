package www

import (
	"errors"
	"html/template"
	"net/http"
)

type IndexHandlerOptions struct {
	Templates *template.Template
}

func IndexHandler(opts *IndexHandlerOptions) (http.Handler, error) {

	t := opts.Templates.Lookup("index")

	if t == nil {
		return nil, errors.New("Missing 'index' template")
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		rsp.Header().Set("Content-Type", "text/html; charset=utf-8")

		err := t.Execute(rsp, nil)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	h := http.HandlerFunc(fn)
	return h, nil
}
