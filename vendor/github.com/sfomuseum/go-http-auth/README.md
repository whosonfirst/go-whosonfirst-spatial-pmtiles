# go-http-auth

Go package to provide a simple interface for enforcing authentication in HTTP handlers.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-http-auth.svg)](https://pkg.go.dev/github.com/sfomuseum/go-http-auth)

## Motivation

The `go-http-auth` package aims to provide a simple interface for enforcing authentication in Go-based web applications in a way that those applications don't need to know anyt
hing about how authentication happens.

```
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
}
```

All the web application (specifically its `net/http.Handler` instances) know is that it has an implementation of the `auth.Authenticator` interface which exposes a `GetAccountForRequest` method which returns an implementation of the `auth.Account` interface or an error.

```
// type Account is an interface that defines minimal information for an account.
type Account interface {
	// The unique ID associated with this account.	
	Id() int64
	// The name associated with this account.
	Name() string
}
```

Account interfaces/implementations meant to be accurate reflections of the underlying implementation's account structure but rather a minimalist struct containing on an account name and unique identifier.

Here is a simplified example, with error handling omitted for brevity, that uses the built-in `null://` Authenticator implementation (that will always return a `Account` instance with ID "0"):

```
package main

import (
     "context"	
     "log"
     "net/http"

     "github.com/sfomuseum/go-http-auth"
)

type HandlerOptions struct {
     Authenticator auth.Authenticator
}

func Handler(opts *HandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		acct, err := opts.Authenticator.GetAccountForRequest(req)

		if err != nil {
			switch err.(type) {
			case auth.NotLoggedIn:
				signin_handler := opts.Authenticator.SigninHandler()
				signin_handler.ServeHTTP(rsp, req)
				return

			default:
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		log.Printf("Authenticated as %s (%d)\n", acct.Name(), acct.Id())
		
		... carry on with handler
	}

	return http.HandlerFunc(fn), nil
}

func main (){

     ctx := context.Background()
     
     authenticator, _ := auth.NewAuthenticator(ctx, "null://")

     handler_opts := &HandlerOptions{
     	Authenticator: authenticator,
     }
     
     handler, _ := Handler(handler_opts)
     handler = authenticator.WrapHandler(handler)
     
     mux := http.NewServeMux()

     mux.Handle("/", handler)

     http.ListenAndServe(":8080", mux)
}
```

Note the use the of the `authenticator.WrapHandler()` method. This is a `net/http` "middleware" handler where implementation-specific logic for checking whether a user is authenticated is expected to happen. That information is expected to be retrievable using the `GetAccountForRequest` method in the subsequent handlers that `WrapHandler` serves. The details of how authentication information is stored and retrieved between handlers is left to individual implmentations.

For concrete examples, have a look at the code for [cmd/example/main.go](cmd/example/main.go).

## Authenticators

The following `Authenticator` implementations are included by default:

### jwt://

`JWTAuthenticator` implements the `Authenticator` interface to ensure that requests contain a "Authorization: Bearer {JWT_TOKEN}"` HTTP header configured by 'uri' which is expected to take the form of:

```
jwt://{SECRET}
```

Where `{SECRET}` is expected to be the shared JWT signing secret passed by HTTP requests. Or:

```
jwt://runtimevar?runtimevar-uri={GOCLOUD_DEV_RUNTIMEVAR_URI}
```

Where `{GOCLOUD_DEV_RUNTIMEVAR_URI}` is a valid [gocloud.dev/runtimevar](https://godoc.org/gocloud.dev/runtimevar/) URI used to dereference the JWT signing secret. Under the hood this method using the [sfomuseum/runtimevar.StringVar](https://github.com/sfomuseum/runtimevar) method to dereference runtimevar URIs.

By default a `JWTAuthenticator` instance looks for JWT Bearer tokens in the HTTP "Authorization" header. This behaviour can be customized by passing an "authorization-header" query parameter in 'uri'. For example:

```
jwt://?authorization-header=X-Custom-AuthHeader
```

JWT payloads are expected to conform to the `JWTAuthenticatorClaims` struct:

```
type JWTAuthenticatorClaims struct {
	// The unique ID associated with this account.
	AccountId   int64  `json:"account_id"`
	// The name associated with this account.	
	AccountName string `json:"account_name"`
	jwt.RegisteredClaims
}
```

### none://

`NoneAuthenticator` implements the `Authenticator` interface and always returns a "Not authorized" error. It is instantiated with the following URI construct:

```
none://
```

### null://

`NullAuthenticator` implements the `Authenticator` interface such that no authentication is performed. It is instantiated with the following URI construct:

```
null://
```

### sharedsecret://

`SharedSecretAuthenticator` implements the `Authenticator` interface to require a simple shared secret be passed with all requests. This is not a sophisticated handler. There are no nonces or hashing of requests or anything like that. It is a bare-bones supplementary authentication handler for environments that already implement their own measures of access control. It is instantiated with the following URI construct:

```
sharedsecret://{SECRET}
```

Where `{SECRET}` is expected to be the shared secret passed by HTTP requests in the `X-Shared-Secret` header.

## See also

* https://datatracker.ietf.org/doc/html/rfc7519
* https://jwt.io/introduction


