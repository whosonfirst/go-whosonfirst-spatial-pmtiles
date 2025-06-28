package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sfomuseum/runtimevar"
)

const AUTHORIZATION_HEADER string = "Authentication"

var re_auth = regexp.MustCompile(`Bearer\s+((?:[_\-A-Za-z0-9+/]*={0,3})\.(?:[_\-A-Za-z0-9+/]*={0,3})\.([_\-A-Za-z0-9+/]*={0,3}))$`)

func init() {
	ctx := context.Background()
	RegisterAuthenticator(ctx, "jwt", NewJWTAuthenticator)
}

// type JWTAuthenticatorClaims are the custom claims for Authorization requests.
type JWTAuthenticatorClaims struct {
	// The unique ID associated with this account.
	AccountId int64 `json:"account_id"`
	// The name associated with this account.
	AccountName string `json:"account_name"`
	jwt.RegisteredClaims
}

// type JWTAuthenticator implements the Authenticator interface to require a valid JSON Web Token (JWT) be passed
// with all requests.
type JWTAuthenticator struct {
	Authenticator
	secret               string
	authorization_header string
}

// NewJWTAuthenticator implements the Authenticator interface to ensure that requests contain a `Authorization: Bearer {JWT_TOKEN}` HTTP
// header configured by 'uri' which is expected to take the form of:
//
//	jwt://{SECRET}
//
// Where {SECRET} is expected to be the shared JWT signing secret passed by HTTP requests. Or:
//
//	jwt://runtimevar?runtimevar-uri={GOCLOUD_DEV_RUNTIMEVAR_URI}
//
// Where {GOCLOUD_DEV_RUNTIMEVAR_URI} is a valid `gocloud.dev/runtimevar` URI used to dereference the JWT signing secret.
// Under the hood this method using the `github.com/sfomuseum/runtimevar.StringVar` method to dereference runtimevar URIs.
//
// By default a `JWTAuthenticator` instance looks for JWT Bearer tokens in the HTTP "Authorization" header. This behaviour
// can be customized by passing an "authorization-header" query parameter in 'uri'. For example:
//
//	jwt://?authorization-header=X-Custom-AuthHeader
func NewJWTAuthenticator(ctx context.Context, uri string) (Authenticator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	secret := u.Host

	if secret == "runtimevar" {

		runtimevar_uri := q.Get("runtimevar-uri")

		s, err := runtimevar.StringVar(ctx, runtimevar_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive secret from runtimevar, %w", err)
		}

		secret = strings.TrimSpace(s)
	}

	if secret == "" {
		return nil, fmt.Errorf("Missing or invalid secret")
	}

	authorization_header := AUTHORIZATION_HEADER

	if q.Has("authorization-header") {
		authorization_header = q.Get("authorization-header")
	}

	a := &JWTAuthenticator{
		secret:               secret,
		authorization_header: authorization_header,
	}

	return a, nil
}

// WrapHandler returns
func (a *JWTAuthenticator) WrapHandler(next http.Handler) http.Handler {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		_, err := a.GetAccountForRequest(req)

		if err != nil {
			slog.Error("Failed to derive account", "error", err)
			http.Error(rsp, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(rsp, req)
		return
	}

	return http.HandlerFunc(fn)
}

// GetAccountForRequest returns an stub `Account` instance for requests that contain a valid `Authorization: Bearer {JWT_TOKEN}` HTTP header (or a custom header if defined in the `JWTAuthenticator` constuctor URI).
func (a *JWTAuthenticator) GetAccountForRequest(req *http.Request) (Account, error) {

	var acct Account

	auth_header := req.Header.Get(a.authorization_header)

	if !re_auth.MatchString(auth_header) {
		slog.Error("Authorization header mismatch", "header", a.authorization_header, "value", auth_header)
		return nil, fmt.Errorf("Invalid auth header")
	}

	m := re_auth.FindStringSubmatch(auth_header)
	str_token := m[1]

	parse_func := func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	}

	parse_opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithExpirationRequired(),
		jwt.WithPaddingAllowed(),
	}

	token, err := jwt.ParseWithClaims(str_token, &JWTAuthenticatorClaims{}, parse_func, parse_opts...)

	if err != nil {
		return nil, err
	} else if claims, ok := token.Claims.(*JWTAuthenticatorClaims); ok {
		acct = NewAccount(claims.AccountId, claims.AccountName)
	} else {
		return nil, fmt.Errorf("Unknown claims type, cannot proceed")
	}

	if acct.Id() == 0 {
		return nil, fmt.Errorf("Missing account ID")
	}

	if acct.Name() == "" {
		return nil, fmt.Errorf("Missing account name")
	}

	return acct, nil
}

// SigninHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *JWTAuthenticator) SigninHandler() http.Handler {
	return notImplementedHandler()
}

// SignoutHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *JWTAuthenticator) SignoutHandler() http.Handler {
	return notImplementedHandler()
}

// SignoutHandler returns an `http.Handler` instance that returns an HTTP "501 Not implemented" error.
func (a *JWTAuthenticator) SignupHandler() http.Handler {
	return notImplementedHandler()
}
