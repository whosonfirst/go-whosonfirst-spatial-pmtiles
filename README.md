# go-whosonfirst-spatial-pmtiles

Go package to implement the `whosonfirst/go-whosonfirst-spatial` interfaces using a Protomaps `.pmtiles` database.

## Documentation

Documentation is incomplete at this time.

## This is work in progress

There may be bugs. Notably, as written:
 
* All tile lookups are performed at zoom 9.
* The underlying spatial queries performed on features derived from PMTiles data are done using in-memory [whosonfirst/go-whosonfirst-spatial-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite) instances. These instances are not cached at this time.

## Producing a Who's On First -enabled Protomaps tile database

As of this writing producing a Who's On First -enabled Protomaps tile database is a manual two-step process.

### Tippecanoe

The first step is to produce a MBTiles database derived from Who's On First data. There are a variety of ways you might accomplish this but as a convenience you can use the `features` tool which is part of the [whosonfirst/go-whosonfirst-tippecanoe](https://github.com/whosonfirst/go-whosonfirst-tippecanoe) package. For example:

```
$> bin/features \
	-writer-uri 'constant://?val=featurecollection://?writer=stdout://' \
	/usr/local/data/sfomuseum-data-whosonfirst/ \
	
	| tippecanoe -zg -o /usr/local/data/wof.mbtiles
```

### PMTiles

Next, use the `pmtiles` tool which is part of the [protomaps/go-pmtiles](https://github.com/protomaps/go-pmtiles#creating-pmtiles-archives) package to convert the MBTiles database to a Protomaps PMTiles database:

```
$> pmtiles /usr/local/data/wof.mbtiles /usr/local/data/wof.pmtiles
```

## Spatial Database URIs

Spatial database URIs for Protomaps PMTiles databases take the form of:

```
pmtiles://?{QUERY_PARAMETERS}
```

### Query parameters

| Name | Value | Required | Notes |
| --- | --- | --- | --- |
| tiles | A valid `gocloud.dev/blob` bucket URI | yes | Support for `file://` URIs is enabled by default. |
| database | The name of the Protomaps tiles database | yes | Ensure that this value does _not_ include a `.pmtiles` extension |
| pmtiles-cache-size | The size, in megabytes, of the pmtiles cache | no | Default is 64 |
| enable-feature-cache | Enable caching of WOF features associated with a tile path | no | Default is false |
| feature-cache-uri | A valid `gocloud.dev/docstore` collection URI | no | Support for `mem://` URIs is enabled by default |
| feature-cache-ttl | The number of seconds that items in the feature cache should persist | no | Default is 300 |

For example:

```
pmtiles://?tiles=file:///usr/local/data&database=wof
```

## Example

```
import (
       _ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
)

import (
       "context"
       "fmt"
       "github.com/paulmach/orb"       
       "github.com/whosonfirst/go-whosonfirst-spatial/database"
)

func main(){

     ctx := context.Background()

     db, _ := database.NewSpatialDatabase(ctx, "pmtiles://?tiles=file:///usr/local/data&database=wof")

     lat := 37.621131
     lon := -122.384292
     
     pt := orb.Point{ lon, lat }
     
     spr, _ := db.PointInPolygon(ctx, pt)

     for _, r := range spr.Results(){
     	fmt.Printf("%s %s\n", r.Id(), r.Name())
     }
}
```

_Error handling omitted for the sake of brevity._

## Tools

### query

```
$> ./bin/query -h
  -alternate-geometry value
    	One or more alternate geometry labels (wof:alt_label) values to filter results by.
  -cessation-date string
    	A valid EDTF date string.
  -custom-placetypes string
    	A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.
  -enable-custom-placetypes
    	Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.
  -enable-geojson
    	...
  -geometries string
    	Valid options are: all, alt, default. (default "all")
  -inception-date string
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
  -is-wof
    	Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents. (default true)
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported schemes are: directory://, featurecollection://, file://, filelist://, geojsonl://, null://, repo://. (default "repo://")
  -latitude float
    	A valid latitude.
  -longitude float
    	A valid longitude.
  -mode string
    	... (default "cli")
  -placetype value
    	One or more place types to filter results by.
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. Available options are: [fs:// null:// repo:// sqlite:// stdin://]
  -property value
    	One or more Who's On First properties to append to each result.
  -server-uri string
    	... (default "http://localhost:8080")
  -sort-uri value
    	Zero or more whosonfirst/go-whosonfirst-spr/v2/sort URIs.
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: [pmtiles:// sqlite://]
  -verbose
    	Be chatty.
```

#### Example

```
$> ./bin/query \
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
       _ "gocloud.dev/blob/s3blob"
)

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/app/query"
	"log"
)

func main() {

	ctx := context.Background()

	logger := log.Default()

	err := query.Run(ctx, logger)

	if err != nil {
		logger.Fatalf("Failed to run PIP application, %v", err)
	}

}

```

## Web interface(s)

The [whosonfirst/go-whosonfirst-spatial-www-pmtiles](https://github.com/whosonfirst/go-whosonfirst-spatial-www-pmtiles) package provides supports for exposing access to a PMTiles-enabled spatial database over HTTP.

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-spatial-pip
* https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite
* https://github.com/whosonfirst/go-whosonfirst-spatial-www
* https://github.com/whosonfirst/go-whosonfirst-spatial-www-pmtiles
* https://github.com/whosonfirst/go-whosonfirst-tippecanoe
* https://github.com/protomaps/go-pmtiles
* https://github.com/felt/tippecanoe
