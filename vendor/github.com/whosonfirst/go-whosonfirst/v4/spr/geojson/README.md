# go-whosonfirst-spr-geojson

Go package for translating Who's On First Standard Places Results (SPR) to GeoJSON.

## Important

Work in progress. Documentation to follow.

## Example

### Using with spr.StandardPlacesResults

```
import (
	"context"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr-geojson"
	"os"
)

func main() {

	ctx := context.Background()

	results, _ := SomethingThatReturnsAStandardPlacesResults()

	reader_uri := "fs://usr/local/data/whosonfirst-data-admin-us"
	r, _ := reader.NewReader(ctx, reader_uri)
	
	opts := &geojson.AsFeatureCollectionOptions{
		Reader: r:
		Writer: os.Stdout,
	}

	geojson.AsFeatureCollection(ctx, results, opts)
}
```

### Using with JSON-encoded spr.StandardPlacesResults

```
import (
	"bufio"
	"context"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr-geojson"
	"io/ioutil"
	"os"
)

func main() {

	reader_uri := "fs://usr/local/data/whosonfirst-data-admin-us"
	path := "places.#.wof:path"

	ctx := context.Background()

	r, _ := reader.NewReader(ctx, reader_uri)

	resolver_func := geojson.JSONPathResolverFunc(path)

	as_opts := &geojson.AsFeatureCollectionOptions{
		Reader: r:
		Writer: os.Stdout,
		JSONPathResolver: resolver_func,		
	}

	reader := bufio.NewReader(os.Stdin)
	body, _ := ioutil.ReadAll(reader)

	geojson.AsFeatureCollectionWithJSON(ctx, body, path, as_opts)
}
```

Here's how you might use this with "plain old" GeoJSON records that use a WOF-style ID/URI structure:

```
import (
	"bufio"
	"context"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr-geojson"
	"io/ioutil"
	"os"
)

func main() {

	reader_uri := "fs://usr/local/data/woeplanet-state-us/data"
	path := "places.#.spr:id"

	ctx := context.Background()

	r, _ := reader.NewReader(ctx, reader_uri)

	cb = func(ctx context.Context, path string) (string, error) {
		return geojson.WhosOnFirstPathWithString(path)
	}

	resolver_func := geojson.JSONPathResolverFuncWithCallback(path, cb)
	
	as_opts := &geojson.AsFeatureCollectionOptions{
		Reader: r:
		Writer: os.Stdout,
		JSONPathResolver: resolver_func,		
	}

	reader := bufio.NewReader(os.Stdin)
	body, _ := ioutil.ReadAll(reader)

	geojson.AsFeatureCollectionWithJSON(ctx, body, as_opts)
}
```

Here's what's going on:

* We're using data from the [woeplanet-state-us](https://github.com/woeplanet-data/woeplanet-state-us) repository. In this case we have a local checkout (as defined by the `reader_uri` variable).
* We're assuming that `woeplanet-data` SPR records are being passed in with a JSON encoding mapping the `spr.Id()` method to a `spr:id` property.
* W're also taking it for granted that `woeplanet-data` records have numeric IDs and are stored using the same nested directory structure as Who's On First records.
* The `geojson.JSONPathResolverFunc` returns a list of WOEPlanet IDs (using the `places.#.spr:id` syntax defined in the `path` variable).
* Each WOEPlanet Id is transformed in to a Who's On First style path using the callback defined by the `cb` variable.
* Theses are the final paths that are passed to the `reader.Reader` instance for loading the actual GeoJSON record.

## Tools

To build binary versions of these tools run the `cli` Makefile target. For example:

```
$> make cli
go build -mod vendor -o bin/as-geojson cmd/as-geojson/main.go
```

## as-geojson

```
$> ./bin/as-geojson -h
Usage of ./bin/as-geojson:
  -path string
    	A valid tidwall/gjson query path for finding the path for each element in your SPR response. (default "places.#.wof:path")
  -reader-uri string
    	A valid whosonfirst/go-reader URI.
```

The `as-geojson` tool reads a JSON-encoded Standard Places Results (SPR) document on `STDIN` and outputs a GeoJSON `FeatureCollection` document on `STDOUT`.

Internally it uses a [whosonfirst/go-reader](#) `Reader` instance to resolve SPR paths to Who's On First records.

For example, here's how you might use the `as-geojson` tool in concert with a tool like [go-whosonfirst-spatial-sqlite](#)'s `query` which outputs SPR results as JSON and the `jq` tool for querying the final GeoJSON `FeatureCollection`.

In this example the `query` tool is using a SQLite database to generate SPR results. The `as-geojson` tool is using a local filesystem reader to resolve SPR paths in to Who's On First records.

```
$> /usr/local/go-whosonfirst-spatial-sqlite/bin/query \
	-database-uri 'sqlite://?dsn=/usr/local/data/ca-alt.db' \
	-latitude 45.572744 \
	-longitude -73.586295 \

| ./bin/as-geojson \
	-reader-uri fs:///usr/local/data/whosonfirst-data-admin-ca/data \

| jq '.features[]["properties"]["wof:id"]'

85874359
1108955735
85874359
136251273
85633041
136251273
85633041
136251273
85633041
890458661
85633041
```

## Readers

This package uses the [go-reader.Reader](https://github.com/whosonfirst/go-reader) interface for retrieving data associated with a Who's On First record. The only reader that is available by default, with this package, is the [local filesystem reader](https://github.com/whosonfirst/go-reader#fs) (`fs://`). If you need to use other `go-reader.Reader` implementations in your code you will need to import them explicitly.

## See also

* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-reader
