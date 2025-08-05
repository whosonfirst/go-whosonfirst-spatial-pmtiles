# go-whosonfirst-format

Standardised GeoJSON formatting for Whos On First files.

Usable as both a library and a binary.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-format.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-format)

## Library usage

```golang
import (
       "fmt"
       "io/ioutil"
       
       "github.com/whosonfirst/go-whosonfirst-format"
)

func main() {

	inputBytes, _ := ioutil.ReadFile(inputPath)
	outputBytes, _ := format.FormatBytes(feature)
	
	fmt.Printf("%s", outputBytes)
}
```

_Error handling removed for the sake of brevity._

There is also a `FormatFeature` method which takes as its input	a `paulmach/orb/geojson.Feature` object (this is what the `FormatBytes` method is calling under the hood in order to ensure valid GeoJSON input).

## Tools

```shell
$> make cli
go build -mod vendor -ldflags="-s -w" -o bin/wof-format ./cmd/wof-format/main.go
```

### wof-format

Format Who's On First records in files or on STDIN.

```shell
$> ./bin/wof-format -h
Usage: ./bin/wof-format [-check] [input]

Either provide a WOF feature on stdin:
cat input.geojson | ./bin/wof-format > output.geojson

Or provide a path to a WOF feature as the first argument:
./bin/wof-format input.geojson > output.geojson

Optional arguments:

  -check
    	exits silently with a non-zero status code if the file is not correctly formatted
  -output string
    	write the output to a file instead of stdout
  -overwrite
    	overwrite the input file with the formatted output
```

For example:

```shell
$> cat input.geojson | ./bin/wof-format > output.geojson
```

## WASM (Javascript)

The `wof-format` functionality is also available as a JavaScript compatible WebAssembly (WASM) binary. The binary comes precompiled with this package but if you need or want to rebuild it the simplest way to do that is to use the handy `wasmjs` Makefile target.

```shell
$> make wasmjs
GOOS=js GOARCH=wasm \
		go build -mod vendor -ldflags="-s -w" \
		-o www/wasm/wof_format.wasm \
		cmd/wof-format-wasm/main.go
```

To test the binary open the web application in `www/index.html` using the HTTP server of your choice. I like to use [aaronland/go-http-fileserver](https://github.com/aaronland/go-http-fileserver) but that's mostly because I wrote it. Any old web server will do. For example:

```shell
$> fileserver -root www
2025/06/28 06:50:46 Serving www and listening for requests on http://localhost:8080
```

And then when you open your browser to `http://localhost:8080` you'll see a simple two-pane web application where you can enter a document to format on the left and see the result on the right, like this:

![](docs/images/whosonfirst-format-wasm.png)

