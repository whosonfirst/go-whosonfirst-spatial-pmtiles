package auth

import (
	"context"
	"log"
	"net/http"
)

func init() {
	ctx := context.Background()
	RegisterAuthenticator(ctx, "null", NewNullAuthenticator)
}

// type NullAuthenticator implements the Authenticator interface such that no authentication is performed.
type NullAuthenticator struct {
	Authenticator
}

// NewNullAuthenticator implements the Authenticator interface such that no authentication is performed
// configured by 'uri' which is expected to take the form of:
//
//	null://
func NewNullAuthenticator(ctx context.Context, uri string) (Authenticator, error) {
	a := &NullAuthenticator{}
	return a, nil
}

// WrapHandler returns 'h' unchanged.
func (a *NullAuthenticator) WrapHandler(h http.Handler) http.Handler {
	return h
}

// GetAccountForRequest returns an stub `Account` instance.
func (a *NullAuthenticator) GetAccountForRequest(req *http.Request) (*Account, error) {

	acct := &Account{
		Id:   0,
		Name: "Null",
	}

	return acct, nil
}

// SigninHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *NullAuthenticator) SigninHandler() http.Handler {
	return notImplementedHandler()
}

// SignoutHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *NullAuthenticator) SignoutHandler() http.Handler {
	return notImplementedHandler()
}

// SignoutHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *NullAuthenticator) SignupHandler() http.Handler {
	return notImplementedHandler()
}

// SetLogger is a no-op and does nothing.
func (a *NullAuthenticator) SetLogger(logger *log.Logger) {
	// no-op
}
