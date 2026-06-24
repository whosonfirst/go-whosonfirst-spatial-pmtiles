package auth

import (
	"context"
	"net/http"

	"github.com/aaronland/go-http/v4/response"
)

func init() {
	ctx := context.Background()
	RegisterAuthenticator(ctx, "none", NewNoneAuthenticator)
}

// type NoneAuthenticator implements the Authenticator interface that always returns a "not authorized" error.
type NoneAuthenticator struct {
	Authenticator
}

// NewNoneAuthenticator implements the Authenticator interface that always returns a "not authorized" error.
// configured by 'uri' which is expected to take the form of:
//
//	none://
func NewNoneAuthenticator(ctx context.Context, uri string) (Authenticator, error) {
	a := &NoneAuthenticator{}
	return a, nil
}

// WrapHandler returns 'h' unchanged.
func (a *NoneAuthenticator) WrapHandler(h http.Handler) http.Handler {
	return h
}

// GetAccountForRequest returns an stub `Account` instance.
func (a *NoneAuthenticator) GetAccountForRequest(req *http.Request) (Account, error) {
	return nil, NotAuthorized{}
}

// SigninHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *NoneAuthenticator) SigninHandler() http.Handler {
	return response.NotImplementedHandler()
}

// SignoutHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *NoneAuthenticator) SignoutHandler() http.Handler {
	return response.NotImplementedHandler()
}

// SignoutHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *NoneAuthenticator) SignupHandler() http.Handler {
	return response.NotImplementedHandler()
}
