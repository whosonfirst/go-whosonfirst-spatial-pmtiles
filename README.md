# go-whosonfirst-spatial-pmtiles

Go package to implement the `whosonfirst/go-whosonfirst-spatial` interfaces using a Protomaps `.pmtiles` database.

## Documentation

Documentation is incomplete at this time. There may be bugs. Notably, as written:

## This is work in progress

* The code and examples do not do anything to clip the geometries or minimize the size or volume of data in the `tippecanoe` MBTiles or `protomaps` PMTiles databases.
* It is possible to define a caching layer for GeoJSON features associated with a given PMTiles tile path but there is no cache invalidation or expiry.
* All tile lookups are performed at zoom 9.

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
| feature-cache-uri | A valid `gocloud.dev/docstore` collection URI | no | Support for `memdocstore://` URIs is enabled by default. |

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
       "github.com/whosonfirst/go-whosonfirst-spatial"
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

## Tools

### query

```
> ./bin/query \
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

* https://github.com/whosonfirst/go-whosonfirst-spatial-www-pmtiles
* https://github.com/whosonfirst/go-whosonfirst-tippecanoe
* https://github.com/protomaps/go-pmtiles
* https://github.com/felt/tippecanoe