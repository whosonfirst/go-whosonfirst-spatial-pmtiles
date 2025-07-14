# go-whosonfirst-validate

Go package for validating Who's On First documents

## Documentation

Documentation is incomplete.

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-validate.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-validate)

## Background

So far this is a tool built and developed out of expediency (translation: why is `x` broken?). So far it validates the things that have needed to be validated.

## Example

```
import (
       "github.com/whosonfirst/go-whosonfirst-validate"
)

func main() {

     r, _ := os.Open("/path/to/feature.geojson")

     body, _ := validate.EnsureValidGeoJSON(r)
     
     validate_opts := default.ValidateOptions()
     validate.ValidateWithOptions(body, opts)
```

_Error handling omitted for the sake of brevity._

## Tools

### wof-validate

This tool will attempt to load all the (principal) WOF documents (using `go-whosonfirst-geojson-v2`) passed to it using a `go-whosonfirst-iterate` iterator.

```
$> ./bin/wof-validate -h
Validate the contents of one or more whosonfirst/go-whosonfirst-iterate/v2 data sources.
Usage:
	 ./bin/wof-validate path(N) path(N)
Valid arguments are:
  -all
    	Enable all validation checks.
  -edtf
    	Validate edtf: properties. (default true)
  -is-current
    	Validate mz:is_current property.
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v2 URI (default "repo://")
  -name
    	Validate wof:name property. (default true)
  -names
    	Validate WOF/RFC 5646 names.
  -placetype
    	Validate wof:placetype property. (default true)
  -repo
    	Validate wof:repo property. (default true)
  -verbose
    	Be chatty about what's happening.
```

For example:

```
$> ./bin/wof-validate /usr/local/data/whosonfirst-data
...time passes
```

Assuming everything loads successfully you won't see any output (unless you've passed the `-verbose` flag (in which case you'll see _a lot_ of output)).

Or this:

```
$> ./bin/wof-validate -names /usr/local/data/whosonfirst-data
error: Failed to parse name tag for /usr/local/data/whosonfirst-data/data/112/585/728/5/1125857285.geojson, because Failed to parse language tag 'eng_v_variant'
```

## WASM

The `wof-validate` functionality is also available as a JavaScript compatible WebAssembly (WASM) binary. The binary comes precompiled with this package but if you need or want to rebuild it the simplest way to do that is to use the handy `wasmjs` Makefile target.

```shell
$> make wasmjs
GOOS=js GOARCH=wasm \
		go build -mod vendor -ldflags="-s -w" -tags wasmjs \
		-o www/wasm/wof_validate.wasm \
		cmd/wof-validate-wasm/main.go
```

To test the binary open the web application in `www/index.html` using the HTTP server of your choice. I like to use [aaronland/go-http-fileserver](https://github.com/aaronland/go-http-fileserver) but that's mostly because I wrote it. Any old web server will do. For example:

```shell
$> fileserver -root www
2025/06/28 06:50:46 Serving www and listening for requests on http://localhost:8080
```

And then when you open your web browser to `http://localhost:8080` you'll see something like this:

![](docs/images/whosonfirst-validate.png)

The default document included in the example is _invalid_ so when you click the "Validate" button you should see an error like this:

![](docs/images/whosonfirst-validate-error.png)

If you add a `"wof:repo":"whosonfirst-data-example"` property and resubmit the document the validation will succeed.

![](docs/images/whosonfirst-validate-valid.png)

## See also

* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-whosonfirst-feature
* https://github.com/whosonfirst/go-whosonfirst-names