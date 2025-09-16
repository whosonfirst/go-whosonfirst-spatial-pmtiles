package auth

import (
	go_http "net/http"
)

// EnsureAccountHandler is a middleware `net/http` handler that wraps 'next' and ensures that the
// authenticator.GetAccountForRequest method does not return an error.
func EnsureAccountHandler(authenticator Authenticator, next go_http.Handler) go_http.Handler {

	fn := func(rsp go_http.ResponseWriter, req *go_http.Request) {

		_, err := authenticator.GetAccountForRequest(req)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(rsp, req)
	}

	return go_http.HandlerFunc(fn)
}
