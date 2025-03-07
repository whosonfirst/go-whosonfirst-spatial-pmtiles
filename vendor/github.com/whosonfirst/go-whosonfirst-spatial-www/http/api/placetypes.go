package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/whosonfirst/go-whosonfirst-placetypes"
)

func NewPlacetypesHandler() (http.Handler, error) {

	pt_list, err := placetypes.Placetypes()

	if err != nil {
		return nil, err
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		logger := slog.Default()

		rsp.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(pt_list)

		if err != nil {
			logger.Error("Failed to marshal placetypes", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	h := http.HandlerFunc(fn)
	return h, nil
}
