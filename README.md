# go-whosonfirst-spatial-pmtiles

Go package to implement the `whosonfirst/go-whosonfirst-spatial` interfaces using a Protomaps `.pmtiles` database.

## Documentation

Documentation is incomplete at this time.

## This is work in progress

There may be bugs. Notably, as written:
 
* The underlying spatial queries performed on features derived from PMTiles data are done using in-memory [whosonfirst/go-whosonfirst-spatial-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite) instances.
* The `go-whosonfirst-spatial-pmtiles` package implements both the [whosonfirst/go-whosonfirst-spatial](https://github.com/whosonfirst/go-whosonfirst-spatial) and [whosonfirst/go-reader](https://github.com/whosonfirst/go-reader) interfaces however in order to support the latter caching must be enabled in the spatial database URI constructor (see below). Caching is necessary to maintain a local cache of features mapped to any given Who's On First ID. This is really only important if you need to to return GeoJSON responses (rather than the default Standard Place Response) or you are using an application derived from `go-whosonfirst-spatial-www` which tries to load GeoJSON features from itself.
* GeoJSON features for large, administrative areas (states, countries, etc.) are likely to be clipped to the tile boundary that contains them. Likewise because features are cached so a read request for a place with a large surface area (say the United States) will return the geometry for the first tile that contains it. This can lead to bizarre results or potential information leakage or both.
* As is often the case with any kind of caching there are probably still "edge cases" to account for and improvements to implement.
* The following WOF properties are decoded from their MVT encoding: `wof:belongsto`, `wof:supersedes`, `wof:superseded_by` and `wof:hierarchy`. The first three are core properties in the [standard place response](https://github.com/whosonfirst/go-whosonfirst-spr) definition; the third is an optional property that can be included in a PMTiles database using the `-append-spr-property` flag in the `features` tool discussed below. Support for custom decoders are not available yet.
* Alternate geometry files are not supported yet.

## Producing a Who's On First -enabled Protomaps tile database

As of this writing producing a Who's On First -enabled Protomaps tile database is a manual two-step process.

### Tippecanoe

The first step is to produce a MBTiles database derived from Who's On First data. There are a variety of ways you might accomplish this but as a convenience you can use the `features` tool which is part of the [whosonfirst/go-whosonfirst-tippecanoe](https://github.com/whosonfirst/go-whosonfirst-tippecanoe) package. For example:

```
$> bin/features \
	-writer-uri 'constant://?val=jsonl://?writer=stdout://' \
	/usr/local/data/sfomuseum-data-whosonfirst/ \
	
	| tippecanoe -P -z 12 -pf -pk -o /usr/local/data/wof.mbtiles
```

Things to note: The `-pf` and `-pk` flags to ensure that no features are dropped and the `-z 12` flag to store everything at zoom 12 which is a good trade-off to minimize the number of places to query (ray cast) in any given tile and the total number of tiles to produce and store.

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
| database | The name of the Protomaps tiles database | yes | Ensure that this value does _not_ include a `.pmtiles` extension. |
| layer | The name of the MVT layer containing your tile data | no | Default is to assume the same name as the value of `database`. |
| pmtiles-cache-size | The size, in megabytes, of the pmtiles cache | no | Default is 64. |
| zoom | The zoom level to perform point-in-polygon queries at | no | Default is 12. |
| enable-cache | Enable caching of WOF features. | no | Default is false however you will need to enable it if you want to use the `-property` flag to append additional properties to results emitted by the `query` tool (discussed below). |
| cache-ttl | The number of seconds that items in the cache should persist | no | Default is 300. |
| feature-cache-uri | A valid URI template containing a `gocloud.dev/docstore` collection URI where GeoJSON features should be cached | no | Support for `mem://` URIs is enabled by default. The template MUST contain a `{key}` element. Default is `mem://pmtiles_features/{key}`. |

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

     db, _ := database.NewSpatialDatabase(ctx, "pmtiles://?tiles=file:///usr/local/data&database=wof&enable_cache=true")

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
