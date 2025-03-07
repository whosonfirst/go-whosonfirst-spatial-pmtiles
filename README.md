# go-whosonfirst-spatial-pmtiles

Go package to implement the `whosonfirst/go-whosonfirst-spatial` interfaces using a Protomaps `.pmtiles` database.

## Background

[A global point-in-polygon service using a static 8GB data file](https://millsfield.sfomuseum.org/blog/2022/12/19/pmtiles-pip/)

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

As of this writing producing a Who's On First -enabled Protomaps tile database is a manual two-step process. The first step is to produce a line-separated feed of GeoJSON files. The second step is to stream that feed in to the `tippecanoe` application to generate a PMTiles database.

There are a variety of ways you might accomplish the first step but as a convenience you can use the `features` tool which is part of the [whosonfirst/go-whosonfirst-tippecanoe](https://github.com/whosonfirst/go-whosonfirst-tippecanoe) package. For example:

```
$> bin/features \
	-writer-uri 'constant://?val=jsonl://?writer=stdout://' \
	/usr/local/data/sfomuseum-data-whosonfirst/ \
	
	| tippecanoe -P -z 12 -pf -pk -o /usr/local/data/wof.pmtiles
```

Things to note: The `-pf` and `-pk` flags to ensure that no features are dropped and the `-z 12` flag to store everything at zoom 12 which is a good trade-off to minimize the number of places to query (ray cast) in any given tile and the total number of tiles to produce and store.

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
       "context"
       "fmt"
       
       "github.com/paulmach/orb"       
       "github.com/whosonfirst/go-whosonfirst-spatial/database"
       _ "github.com/whosonfirst/go-whosonfirst-spatial-pmtiles"       
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

```
$> make cli
go build -mod readonly -ldflags="-s -w" -o bin/pmtile cmd/pmtile/main.go
go build -mod readonly -ldflags="-s -w" -o bin/http-server cmd/http-server/main.go
go build -mod readonly -ldflags="-s -w" -o bin/grpc-server cmd/grpc-server/main.go
go build -mod readonly -ldflags="-s -w" -o bin/grpc-client cmd/grpc-client/main.go
go build -mod readonly -ldflags="-s -w" -o bin/update-hierarchies cmd/update-hierarchies/main.go
go build -mod readonly -ldflags="-s -w" -o bin/pip cmd/pip/main.go
go build -mod readonly -ldflags="-s -w" -o bin/intersects cmd/intersects/main.go
```

### pip

Documentation for the `pip` tool has been moved in to [cmd/pip/README.md](cmd/pip/README.md).

### intersects

Documentation for the `intersects` tool can be found [cmd/intersects/README.md](cmd/intersects/README.md).

## Web interface(s)

### http-server

Documentation for the `http-server` tool has been moved in to [cmd/http-server/README.md](cmd/http-server/README.md).

## gRPC interface(s)

### grpc-server

Documentation for the `http-server` tool has been moved in to [cmd/grpc-server/README.md](cmd/grpc-server/README.md).

### grpc-client

Documentation for the `http-server` tool has been moved in to [cmd/grpc-client/README.md](cmd/grpc-client/README.md).

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite
* https://github.com/whosonfirst/go-whosonfirst-spatial-www
* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc
* https://github.com/whosonfirst/go-whosonfirst-tippecanoe
* https://github.com/protomaps/go-pmtiles
* https://github.com/felt/tippecanoe
