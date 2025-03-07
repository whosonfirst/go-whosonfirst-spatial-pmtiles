# pip

Perform an point-in-polygon operation for an input latitude, longitude coordinate and on a set of Who's on First records stored in a spatial database.

```
$> ./bin/pip -h
Perform an point-in-polygon operation for an input latitude, longitude coordinate and on a set of Who's on First records stored in a spatial database.
Usage:
	 ./bin/pip [options]
Valid options are:

  -alternate-geometry value
    	One or more alternate geometry labels (wof:alt_label) values to filter results by.
  -cessation string
    	A valid EDTF date string.
  -custom-placetypes string
    	A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.
  -enable-custom-placetypes
    	Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.
  -geometries string
    	Valid options are: all, alt, default. (default "all")
  -inception string
    	A valid EDTF date string.
  -is-ceased value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-current value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-deprecated value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-superseded value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -is-superseding value
    	One or more existential flags (-1, 0, 1) to filter results by.
  -iterator-uri value
    	Zero or more URIs denoting data sources to use for indexing the spatial database at startup. URIs take the form of {ITERATOR_URI} + "#" + {PIPE-SEPARATED LIST OF ITERATOR SOURCES}. Where {ITERATOR_URI} is expected to be a registered whosonfirst/go-whosonfirst-iterate/v2 iterator (emitter) URI and {ITERATOR SOURCES} are valid input paths for that iterator. Supported whosonfirst/go-whosonfirst-iterate/v2 iterator schemes are: cwd://, directory://, featurecollection://, file://, filelist://, geojsonl://, null://, repo://.
  -latitude float
    	A valid latitude.
  -longitude float
    	A valid longitude.
  -mode string
    	Valid options are: cli, lambda. (default "cli")
  -placetype value
    	One or more place types to filter results by.
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. Available options are: [fs:// null:// pmtiles:// repo:// sqlite:// stdin://]. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.
  -property value
    	One or more Who's On First properties to append to each result.
  -sort-uri value
    	Zero or more whosonfirst/go-whosonfirst-spr/sort URIs.
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: [pmtiles:// rtree:// sqlite://] (default "rtree://")
  -verbose
    	Enable verbose (debug) logging.
```

## Example

```
$> ./bin/pip \
	-spatial-database-uri 'pmtiles://?tiles=file:///usr/local/data&database=wof' \
	-latitude 37.621131 \
	-longitude -122.384292 \
	-sort-uri name:// \

| jq '.places[]["wof:name"]'

"94128"
"California"
"Earth"
"North America"
"San Francisco International Airport"
"San Mateo"
"United States"
```

## Support for gocloud.dev `blob` and `docstore` interfaces

By default this package enables support the [gocloud.dev/blob/fileblob](https://gocloud.dev/howto/blob/#local) and [gocloud.dev/docstore/memdocstore](https://gocloud.dev/howto/docstore/#mem) packages for reading files from the local filesystem and caching data in memory respectively.

In order to add support for additional implementations you will need to clone the relevant code and add any required `import` declarations. For example this is how you would support for reading data from an S3 bucket to the `query` tool using the [s3blob](https://gocloud.dev/howto/blob/#s3) package:

```
package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-whosonfirst-spatial/app/pip"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"	
        _ "gocloud.dev/blob/s3blob"	
)

func main() {

	ctx := context.Background()

	logger := log.Default()

	err := pip.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run PIP application, %v", err)
	}

}
```
