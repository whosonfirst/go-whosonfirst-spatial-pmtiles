# go-whosonfirst-iterate

Go package for iterating through collections  of Who's On First documents.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-iterate.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-iterate/v3)

## Example

Version 3.x of this package introduce major, backward-incompatible changes from earlier releases. That said, migragting from version 2.x to 3.x should be relatively straightforward as a the _basic_ concepts are still the same but (hopefully) simplified. Where version 2.x relied on defining a custom callback for looping over records version 3.x use Go's [iter.Seq2](https://pkg.go.dev/iter) iterator construct to yield records as they are encountered.

For example:

```
import (
	"context"
	"flag"
	"log"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
)

func main() {

     	var iterator_uri string

	flag.StringVar(&iterator_uri, "iterator-uri", "repo://". "A registered whosonfirst/go-whosonfirst-iterate/v3.Iterator URI.")
	ctx := context.Background()
	
	iter, _:= iterate.NewIterator(ctx, iterator_uri)
	defer iter.Close()
	
	paths := flag.Args()
	
	for rec, _ := range iter.Iterate(ctx, paths...) {
	    	defer rec.Body.Close()
		log.Printf("Indexing %s\n", rec.Path)
	}
}
```

_Error handling removed for the sake of brevity._

### Version 2.x (the old way)

This is how you would do the same thing using the older version 2.x code:

```
package main

import (
       "context"
       "flag"
       "io"
       "log"

       "github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"       
       "github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func main() {

	emitter_uri := flag.String("emitter-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI")
	
     	flag.Parse()

	ctx := context.Background()

	emitter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {
		log.Printf("Indexing %s\n", path)
		return nil
	}

	iter, _ := iterator.NewIterator(ctx, *emitter_uri, cb)

	uris := flag.Args()
	iter.IterateURIs(ctx, uris...)
}
```

## Iterators

Iterators are defined as a standalone packages implementing the `Iterator` interface:

```
// Iterator defines an interface for iterating through collections  of Who's On First documents.
type Iterator interface {
	// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in one or more URIs.
	Iterate(context.Context, ...string) iter.Seq2[*Record, error]
	// Seen() returns the total number of records processed so far.
	Seen() int64
	// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
	IsIterating() bool
	// Close performs any implementation specific tasks before terminating the iterator.
	Close() error	
}
```

Then, at the package level, they are "registered" with the `iterate` package so that they can be invoked using a simple declarative URI syntax. For example:

```
func init() {
	ctx := context.Background()
	err := RegisterIterator(ctx, "cwd", NewCwdIterator)

	if err != nil {
		panic(err)
	}
}
```

And then:

```
it, err := iterate.NewIterator(ctx, "cwd://")
```

Importantly, `Iterator` implementations that are "registered" are wrapped in a second (internal) `Iterator` implementation that provides for concurrent processing, retries and regular-expression based file inclusion and exclusion rules. These criteria are defined using query parameters appended to the initial iterator URI that are prefixed with an "_" character. For example:

```
it, err := iterate.NewIterator(ctx, "cwd://?_exclude=.*\.txt$")
```

The following iterators schemes are supported by default:

### cwd://

`CwdIterator` implements the `Iterator` interface for crawling records in the current working directory.

### directory://

`DirectoryIterator` implements the `Iterator` interface for crawling records in a directory.

### featurecollection://

`FeatureCollectionIterator` implements the `Iterator` interface for crawling features in a GeoJSON FeatureCollection record.

### file://

`FileIterator` implements the `Iterator` interface for crawling individual file records.

### filelist://

`FileListIterator` implements the `Iterator` interface for crawling records listed in a "file list" (a plain text newline-delimted list of files).

### fs://

`FSIterator` implements the `Iterator` interface for crawling records listed in a `fs.FS` instance. For example:

```
import (
	"context"
	"flag"
	"io/fs"	
	"log"
	
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
)

func main() {

     	var iterator_uri string

	flag.StringVar(&iterator_uri, "iterator-uri", "fs://". "A registered whosonfirst/go-whosonfirst-iterate/v3.Iterator URI.")
	ctx := context.Background()

	// Your fs.FS goes here
	var your_fs fs.FS
	
	iter, _:= iterate.NewFSIterator(ctx, iterator_uri, fs)

	for rec, _ := range iter.Iterate(ctx, ".") {
	    	defer rec.Body.Close()
		log.Printf("Indexing %s\n", rec.Path)
	}
}
```

Notes:

* The `go-whosonfirst-iterate-fs/v3` implementation does NOT register itself with the `whosonfirst/go-whosonfirst-iterate.RegisterIterator` method and is NOT instantiated using the `whosonfirst/go-whosonfirst-iterate.NewIterator` method since `fs.FS` instances can not be defined as URI constructs.
* Under the hood the `NewFSIterator` is wrapping a `FSIterator` instance in a `whosonfirst/go-whosonfirst-iterate.concrurrentIterator` instance to provide for throttling, filtering and other common (configurable) operations.

### geojsonl://

`GeojsonLIterator` implements the `Iterator` interface for crawling features in a line-separated GeoJSON record.

### null://

`NullIterator` implements the `Iterator` interface for appearing to crawl records but not doing anything.

### repo://

`RepoIterator` implements the `Iterator` interface for crawling records in a Who's On First style data directory.


## Query parameters

The following query parameters are honoured by all `iterate.Iterator` instances:

| Name | Value | Required | Notes
| --- | --- | --- | --- |
| include | String | No | One or more query filters (described below) to limit documents that will be processed. |
| exclude | String | No | One or more query filters (described below) for excluding documents from being processed. |

The following query paramters are honoured for `iterate.Iterator` URIs passed to the `iterator.NewIterator` method:

| Name | Value | Required | Notes
| --- | --- | --- | --- |
| _max_procs | Int | No | _To be written_ |
| _include | String (a valid regular expression) for paths (uris) to include for processing. | No | _To be written_ |
| _exclude | String (a valid regular expression) for paths (uris) to exclude from processing. | No | _To be written_ |
| _exclude_alt | Bool | No | If true do not process "alternate geometry" files. |
| _retry | Bool | No | A boolean flag signaling that if a URI being walked fails it should be retried. Used in conjunction with the `_max_retries` and `_retry_after` parameters. |
| _max_retries | Int | No | The maximum number of attempts to walk any given URI. Defaults to "1" and the `_retry` parameter _must_ evaluate to a true value in order to change the default. |
| _retry_after | Int | The number of seconds to wait between attempts to walk any given URI. Defaults to "10" (seconds) and the `_retry` parameter _must_ evaluate to a true value in order to change the default. |
| _dedupe | Bool | No | A boolean value to track and skip records (specifically their relative URI) that have already been processed. |
| _with_stats | Bool | No | Boolean flag indicating whether stats should be logged. Default is true. |
| _stats_interval | Int | No | The number of seconds between stats logging events. Default is 60. |
| _stats_level | String | No | The (slog/log) level at which stats are logged. Default is INFO. |

## Filters

### QueryFilters

You can also specify inline queries by appending one or more `include` or `exclude` parameters to a `iterate.Iterator` URI, where the value is a string in the format of:

```
{PATH}={REGULAR EXPRESSION}
```

Paths follow the dot notation syntax used by the [tidwall/gjson](https://github.com/tidwall/gjson) package and regular expressions are any valid [Go language regular expression](https://golang.org/pkg/regexp/). Successful path lookups will be treated as a list of candidates and each candidate's string value will be tested against the regular expression's [MatchString](https://golang.org/pkg/regexp/#Regexp.MatchString) method.

For example:

```
repo://?include=properties.wof:placetype=region
```

You can pass multiple query parameters. For example:

```
repo://?include=properties.wof:placetype=region&include=properties.wof:name=(?i)new.*
```

The default query mode is to ensure that all queries match but you can also specify that only one or more queries need to match by appending a `include_mode` or `exclude_mode` parameter where the value is either "ANY" or "ALL".

## Tools

```
$> make cli
go build -mod vendor -o bin/count cmd/count/main.go
go build -mod vendor -o bin/emit cmd/emit/main.go
```

### count

Count files in one or more whosonfirst/go-whosonfirst-iterate/v3 iterator sources.

```
$> ./bin/count -h
Count files in one or more whosonfirst/go-whosonfirst-iterate/v3.Iterator sources.
Usage:
	 ./bin/count [options] uri(N) uri(N)
Valid options are:

  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v3.Iterator URI. Supported iterator URI schemes are: cwd://,directory://,featurecollection://,file://,filelist://,geojsonl://,null://,repo:// (default "repo://")
```

For example:

```
$> ./bin/count fixtures
2025/06/23 08:26:59 INFO Counted records count=37 time=9.216979ms
```

### emit

Emit records in one or more whosonfirst/go-whosonfirst-iterate/v3.Iterator sources as structured data.

```
$> ./bin/emit -h
Emit records in one or more whosonfirst/go-whosonfirst-iterate/v3.Iterator sources as structured data.
Usage:
	 ./bin/emit [options] uri(N) uri(N)
Valid options are:

  -geojson
    	Emit features as a well-formed GeoJSON FeatureCollection record.
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v3.Iterator URI. Supported iterator URI schemes are: cwd://,directory://,featurecollection://,file://,filelist://,geojsonl://,null://,repo:// (default "repo://")
  -json
    	Emit features as a well-formed JSON array.
  -null
    	Publish features to /dev/null
  -stdout
    	Publish features to STDOUT. (default true)
```

For example:

```
$> ./bin/emit \
	-iterator-uri 'repo://?include=properties.sfomuseum:placetype=museum' \
	-geojson \	
	fixtures \

| jq '.features[]["properties"]["wof:id"]'

1360391311
1360391313
1360391315
1360391317
1360391321
1360391323
1360391325
1360391327
1360391329
...and so on
```

## Notes about writing your own `iterate.Iterator` implementation.

Under the hood all `iterate.Iterate` instances are wrapped using the (private) `concurrentIterator` implementation. This is the code that implements throttling, file matching and other common tasks. That happens automatically when code calls `iterate.NewIterator` but you do need to make sure that you "register" your custom implementation, for example:

```
package custom

import (
	"context"

	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
)

func init() {

	ctx := context.Background()
	err := iterate.RegisterIterator(ctx, "custom", YourCustomIterator)

	if err != nil {
		panic(err)
	}
}

type CustomIterator struct {
	iterate.Iterator
}

func NewCustomIterator(ctx context.Context, uri string) (iterate.Iterator, error) {
	it := &CustomIterator{}
	return it, nil
}

// The rest of the iterate.Iterator interfece goes here...
```

## Other implementations

* https://github.com/whosonfirst/go-whosonfirst-iterate-bucket
* https://github.com/whosonfirst/go-whosonfirst-iterate-git
* https://github.com/whosonfirst/go-whosonfirst-iterate-reader
* https://github.com/whosonfirst/go-whosonfirst-iterate-sql

## See also

* https://github.com/aaronland/go-json-query
* https://github.com/aaronland/go-roster
* https://pkg.go.dev/iter