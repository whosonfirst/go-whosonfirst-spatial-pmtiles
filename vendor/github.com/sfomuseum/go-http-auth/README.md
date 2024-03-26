# go-http-auth

Go package to provide a simple interface for enforcing authentication in HTTP handlers.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-http-auth.svg)](https://pkg.go.dev/github.com/sfomuseum/go-http-auth)

## Important

This work is nearly settled (at least for a "v1" release) but may still change.

## Motivation

The `go-http-auth` package aims to provide a simple interface for enforcing authentication in Go-based web applications in a way that those applications don't need to know anyt
hing about how authentication happens.

```
// type Authenticator is a simple interface for	enforcing authentication in HTTP handlers.
type Authenticator interface {
	// WrapHandler wraps a `http.Handler` with any implementation-specific middleware.
	WrapHandler(http.Handler) http.Handler
	// GetAccountForRequest returns an `Account` instance  for an HTTP request.
	GetAccountForRequest(*http.Request) (*Account, error)
	// SigninHandler returns a `http.Handler` for implementing account signin.
	SigninHandler() http.Handler
	// SignoutHandler returns a `http.Handler` for implementing account signout.
	SignoutHandler() http.Handler
	// SignupHandler returns a `http.Handler` for implementing account signups.
	SignupHandler() http.Handler
	// SetLogger assigns a `log.Logger` instance.
	SetLogger(*log.Logger)
}
```

All the web application (specifically its `net/http.Handler` instances) know is that it has an implementation of the `auth.Authenticator` interface which exposes a `GetAccountForRequest` method which returns an `auth.Account` struct or an error.

```
// type Account is a struct that defines minimal information for an account.
type Account struct {
	// The unique ID associated with this account.
	Id int64 `json:"id"`
	// The name associated with this account.
	Name string `json:"name"`
}
```

Account structs are _not_ meant to be accurate reflections of the underlying implementation's account structure but rather a minimalist struct containing on an account name and unique identifier.

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

		log.Printf("Authenticated as %s (%d)\n", acct.Name, acct.Id)
		
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

For concrete examples, have a look at the code for the [NullHandler](null.go) and [NoneHandler](none.go) which always return a "null" user and a "not authorized" error, respectively.
