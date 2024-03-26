# go-http-tangramjs

`go-http-tangramjs` is an HTTP middleware package for including Tangram.js (v0.21.1) assets in web applications.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-http-tangramjs.svg)](https://pkg.go.dev/github.com/sfomuseum/go-http-tangramjs)

`go-http-tangramjs` is an HTTP middleware package for including Tangram.js assets in web applications. It exports two principal methods: 

* `tangramjs.AppendAssetHandlers(*http.ServeMux)` which is used to append HTTP handlers to a `http.ServeMux` instance for serving Tangramjs JavaScript files, and related assets.
* `tangramjs.AppendResourcesHandler(http.Handler, *TangramJSOptions)` which is used to rewrite any HTML produced by previous handler to include the necessary markup to load Tangramjs JavaScript files and related assets.

## Examples

### Basic

```
package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/aaronland/go-http-tangramjs"
)

//go:embed *.html
var FS embed.FS

func ExampleHandler(templates *template.Template) (http.Handler, error) {

	t := templates.Lookup("example")

	fn := func(rsp http.ResponseWriter, req *http.Request) {
		err := t.Execute(rsp, nil)
		return
	}

	return http.HandlerFunc(fn), nil
}

func main() {

	api_key := "****"
	style_url := "/tangram/refill-style.zip"

	t, _ := template.ParseFS(FS, "*.html")

	mux := http.NewServeMux()
	
	map_handler, _:= ExampleHandler(t)

	tangramjs_opts := tangramjs.DefaultTangramJSOptions()
	tangramjs_opts.NextzenOptions.APIKey = api_key
	tangramjs_opts.NextzenOptions.StyleURL = style_url

	tangramjs.AppendAssetHandlers(mux, tangramjs_opts)
	
	map_handler = tangramjs.AppendResourcesHandler(map_handler, tangramjs_opts)

	mux.Handle("/", map_handler)

	endpoint := "localhost:8080"
	log.Printf("Listening for requests on %s\n", endpoint)

	http.ListenAndServe(endpoint, mux)
}
```

_Error handling omitted for the sake of brevity._

You can see an example of this application by running the [cmd/example](cmd/example/main.go) application. You can do so by invoking the `example` Makefile target. For example:

```
$> make example APIKEY=XXXXXX
go run -mod vendor cmd/example/main.go -api-key XXXXXX
2021/05/05 15:31:06 Listening for requests on localhost:8080
```

The when you open the URL `http://localhost:8080` in a web browser you should see the following:

![](docs/images/go-http-tangramjs-example.png)

### With local "tilepacks"

It is also possible to use the `go-http-tangramjs` with local Nextzen "tilepack" MBTiles (or SQLite) databases using tools and methods provided by the [tilezen/go-tilepacks](https://github.com/tilezen/go-tilepacks) package.

Unfortunately it is not possible to enable this functionality with a single "wrapper" method and requires some additional code that you'll need to add manually to your application.

Here is an abbreviated and annotated example of how to use a local "tilepack" in conjunction with the `go-tilepacks`' [MbtilesHandler](https://github.com/tilezen/go-tilepacks/blob/master/http/mbtiles.go) HTTP handler to serve tiles directly from your application. Error handling has been removed for the sake of brevity.

```
import (
       "flag"
	"github.com/aaronland/go-http-tangramjs"
	tiles_http "github.com/tilezen/go-tilepacks/http"
	"github.com/tilezen/go-tilepacks/tilepack"
       "net/http"
       "strings"
)

func main() {

	// These are the normal flags you might set to configure Tangram.js
	
	nextzen_api_key := flag.String("nextzen-api-key", "xxxxxx", "A valid Nextzen API key")
	nextzen_style_url := flag.String("nextzen-style-url", "/tangram/refill-style.zip", "A valid Nextzen style URL")

	// These next two flags are necessary if you want to enable local tile-serving.
	// The first flag is the path to a Nextzen MVT "tilepack" as created by the `go-tilepacks/cmd/build` tool.
	// The second flag is the relative URL that Tangram.js will use to load MVT data.
	
	tilepack_db := flag.String("nextzen-tilepack-database", "", "The path to a valid MBTiles database (tilepack) containing Nextzen MVT tiles.")
	tilepack_uri := flag.String("nextzen-tilepack-uri", "/tilezen/vector/v1/512/all/{z}/{x}/{y}.mvt", "The relative URI to serve Nextzen MVT tiles from a MBTiles database (tilepack).")

	flag.Parse()

	// Create a new ServeMux for handlings requests and 
	// append Tangram.js assets handlers to the mux
	
	mux := http.NewServeMux()	

	// Create a default TangramJSOptions instance
	// Assign the API key and style URL as usual
	
	tangramjs_opts := tangramjs.DefaultTangramJSOptions()
	tangramjs_opts.NextzenOptions.APIKey = *nextzen_api_key
	tangramjs_opts.NextzenOptions.StyleURL = *nextzen_style_url

	tangramjs.AppendAssetHandlers(mux, tangramjs_opts)
	
	// If tilepack_db is not empty the first thing we want to is
	// update the tile URL that Tangram.js will use to find vector
	// data to point to the value of the `tilepack_uri` flag
	
	if *tilepack_db != "" {
		tangramjs_opts.NextzenOptions.TileURL = *tilepack_uri
	}

	// Create the `net/http` handler that will serve your appliction
	// In this example "YourWWWHandler" is just a placeholder
	
	www_handler, _ := YourWWWHandler()

	// Append the Tangram.js resources to your application handler
	// and configure the mux to serve it
	www_handler = tangramjs.AppendResourcesHandler(www_handler, tangramjs_opts)
	mux.Handle("/", www_handler)

	// If tilepack_db is not empty then use it to create a tilepack.MbtilesReader
	// instance. Then use that reader to create a `net/http` handler for serving
	// tile requests. If you are configuring a mux handler to listen for requests
	// on "/" it is important that you configure your local tile handler afterward.
	
	if *tilepack_db != "" {

		tiles_reader, _ := tilepack.NewMbtilesReader(*tilepack_db)
		tiles_handler := tiles_http.MbtilesHandler(tiles_reader)

		// In this example we take the first leaf from the value of the `tilepack_uri`
		// flag ("/tilezen/") and use it to configure the mux instance for serving
		// tile requests.
		
		u := strings.TrimLeft(*tilepack_uri, "/")
		p := strings.Split(u, "/")
		path_tiles := fmt.Sprintf("/%s/", p[0])

		mux.Handle(path_tiles, tiles_handler)
	}

	// Serve mux here...
}
```

You can see a working example of this in the [aaronland/go-marc/cmd/marc-034d](https://github.com/aaronland/go-marc/blob/main/cmd/marc-034d/main.go) application.

The creation of Nextzen vector tile "tilepack" databases is out of scope for this document. Please consult the documentation for the [go-tilepacks build tool](https://github.com/tilezen/go-tilepacks#build) for details on creating custom databases.

It is currently only possible to serve tiles from a single "tilepack" database.

It is not possible to dynamically limit the map to the zoom range and tile extent of a given "tilepack" database. Yet. I'm working on it.

There are precompiled databases with global tile coverage for zoom levels 1-10, 11 and 12 available on the Internet Archive:

* [Global tiles, zoom levels 1 through 10](https://archive.org/details/nextzen-world-2019-1-10) (1.8GB)
* [Global tiles, zoom level 11](https://archive.org/details/nextzen-world-2019-1-10) (3.5GB)
* [Global tiles, zoom level 12](https://archive.org/details/nextzen-world-2019-1-10) (7.9GB)

## See also

* https://github.com/tangrams/tangram
* https://github.com/aaronland/go-http-leaflet
* https://github.com/tilezen/go-tilepacks