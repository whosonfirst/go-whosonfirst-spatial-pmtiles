package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// type Authenticator is a simple interface for	enforcing authentication in HTTP handlers.
type Authenticator interface {
	// WrapHandler wraps a `http.Handler` with any implementation-specific middleware.
	WrapHandler(http.Handler) http.Handler
	// GetAccountForRequest returns an `Account` instance  for an HTTP request.
	GetAccountForRequest(*http.Request) (Account, error)
	// SigninHandler returns a `http.Handler` for implementing account signin.
	SigninHandler() http.Handler
	// SignoutHandler returns a `http.Handler` for implementing account signout.
	SignoutHandler() http.Handler
	// SignupHandler returns a `http.Handler` for implementing account signups.
	SignupHandler() http.Handler
	// SetLogger assigns a `log.Logger` instance.
	SetLogger(*log.Logger)
}

var authenticator_roster roster.Roster

// AuthenticatorInitializationFunc is a function defined by individual authenticator package and used to create
// an instance of that authenticator
type AuthenticatorInitializationFunc func(ctx context.Context, uri string) (Authenticator, error)

// RegisterAuthenticator registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Authenticator` instances by the `NewAuthenticator` method.
func RegisterAuthenticator(ctx context.Context, scheme string, init_func AuthenticatorInitializationFunc) error {

	err := ensureAuthenticatorRoster()

	if err != nil {
		return err
	}

	return authenticator_roster.Register(ctx, scheme, init_func)
}

func ensureAuthenticatorRoster() error {

	if authenticator_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		authenticator_roster = r
	}

	return nil
}

// NewAuthenticator returns a new `Authenticator` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `AuthenticatorInitializationFunc`
// function used to instantiate the new `Authenticator`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterAuthenticator` method.
func NewAuthenticator(ctx context.Context, uri string) (Authenticator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := authenticator_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(AuthenticatorInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureAuthenticatorRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range authenticator_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
