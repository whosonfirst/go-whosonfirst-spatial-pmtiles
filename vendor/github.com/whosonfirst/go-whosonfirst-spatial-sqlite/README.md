# go-whosonfirst-spatial-sqlite

SQLite-backed implementation of the go-whosonfirst-spatial interfaces.

## Important

This is work in progress. It may change still. The goal is to have a package that conforms to the [database.SpatialDatabase](https://github.com/whosonfirst/go-whosonfirst-spatial#spatialdatabase) interface using [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) and SQLite's [RTree](https://www.sqlite.org/rtree.html) extension.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-spatial-sqlite.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-spatial-sqlite)

Documentation is incomplete.

## Databases

This code depends on (4) tables as indexed by the `go-whosonfirst-sqlite-features` package:

* [rtree](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#rtree) - this table is used to perform point-in-polygon spatial queries.
* [spr](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#spr) - this table is used to generate [standard place response](#) (SPR) results.
* [properties](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#properties) - this table is used to append extra properties (to the SPR response) for `spatial.PropertiesResponseResults` responses.
* [geojson](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#geojson) - this table is used to satisfy the `whosonfirst/go-reader.Reader` requirements in the `spatial.SpatialDatabase` interface. It is meant to be a simple ID to bytes (or filehandle) lookup rather than a data structure that is parsed or queried.

Here's an example of the creating a compatible SQLite database for all the [administative data in Canada](https://github.com/whosonfirst-data/whosonfirst-data-admin-ca) using the `wof-sqlite-index-features` tool which is part of the [go-whosonfirst-sqlite-features-index](https://github.com/whosonfirst/go-whosonfirst-sqlite-features-index) package:

```
$> ./bin/wof-sqlite-index-features \
	-index-alt-files \
	-spatial-tables \
	-timings \
	-dsn /usr/local/ca-alt.db \
	-mode repo:// \
	/usr/local/data/whosonfirst-data-admin-ca/

13:09:44.642004 [wof-sqlite-index-features] STATUS time to index rtree (11860) : 30.469010289s
13:09:44.642136 [wof-sqlite-index-features] STATUS time to index geometry (11860) : 5.155172377s
13:09:44.642141 [wof-sqlite-index-features] STATUS time to index properties (11860) : 4.631908497s
13:09:44.642143 [wof-sqlite-index-features] STATUS time to index spr (11860) : 19.160260741s
13:09:44.642146 [wof-sqlite-index-features] STATUS time to index all (11860) : 1m0.000182571s
13:10:44.642848 [wof-sqlite-index-features] STATUS time to index spr (32724) : 39.852608874s
13:10:44.642861 [wof-sqlite-index-features] STATUS time to index rtree (32724) : 57.361318918s
13:10:44.642864 [wof-sqlite-index-features] STATUS time to index geometry (32724) : 10.242155898s
13:10:44.642868 [wof-sqlite-index-features] STATUS time to index properties (32724) : 10.815961878s
13:10:44.642871 [wof-sqlite-index-features] STATUS time to index all (32724) : 2m0.000429956s
```

And then...

```
$> ./bin/pip \
	-database-uri 'sqlite://?dsn=/usr/local/data/ca-alt.db' \
	-latitude 45.572744 \
	-longitude -73.586295
| jq \
| grep wof:id

2020/12/16 13:25:32 Time to point in polygon, 395.201983ms
      "wof:id": "85633041",
      "wof:id": "85874359",
      "wof:id": "1108955735",
      "wof:id": "85874359",
      "wof:id": "85633041",
      "wof:id": "890458661",
      "wof:id": "136251273",
      "wof:id": "136251273",
      "wof:id": "85633041",
      "wof:id": "136251273",
      "wof:id": "85633041",
```

_TBW: Indexing tables on start-up._

## Example

```
package main

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spatial/properties"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

func main() {

	database_uri := "sqlite://?dsn=whosonfirst.db"
	properties_uri := "sqlite://?dsn=whosonfirst.db"
	latitude := 37.616951
	longitude := -122.383747

	props := []string{
		"wof:concordances",
		"wof:hierarchy",
		"sfomuseum:*",
	}

	ctx := context.Background()
	
	db, _ := database.NewSpatialDatabase(ctx, *database_uri)
	pr, _ := properties.NewPropertiesReader(ctx, *properties_uri)
	
	c, _ := geo.NewCoordinate(*longitude, *latitude)
	f, _ := filter.NewSPRFilter()
	r, _ := db.PointInPolygon(ctx, c, f)

	r, _ = pr.PropertiesResponseResultsWithStandardPlacesResults(ctx, r, props)

	enc, _ := json.Marshal(r)
	fmt.Println(string(enc))
}
```

_Error handling removed for the sake of brevity._

## Filters

_To be written_

## Tools

### pip

```
$> ./bin/pip -h
  -alternate-geometry value
    	One or more alternate geometry labels (wof:alt_label) values to filter results by.
  -cessation-date string
    	A valid EDTF date string.
  -custom-placetypes string
    	A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.
  -enable-custom-placetypes
    	Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.
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
  -latitude float
    	A valid latitude.
  -longitude float
    	A valid longitude.
  -placetype value
    	One or more place types to filter results by.
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. Available options are: [file:// fs:// null://]
  -property value
    	One or more Who's On First properties to append to each result.
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: [sqlite://]
  -verbose
    	Be chatty.
```

For example:

```
$> ./bin/pip \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/sfomuseum-data-architecture.db' \
	-latitude 37.616951 \
	-longitude -122.383747 \
	-properties 'wof:hierarchy' \
	-properties 'sfomuseum:*' \
| jq

{
  "properties": [
    {
      "mz:is_ceased": 1,
      "mz:is_current": 0,
      "mz:is_deprecated": 0,
      "mz:is_superseded": 1,
      "mz:is_superseding": 1,
      "mz:latitude": 37.617475,
      "mz:longitude": -122.383371,
      "mz:max_latitude": 37.61950174060331,
      "mz:max_longitude": -122.38139655218178,
      "mz:min_latitude": 37.61615511156664,
      "mz:min_longitude": -122.3853565208227,
      "mz:uri": "https://data.whosonfirst.org/115/939/616/5/1159396165.geojson",
      "sfomuseum:is_sfo": 1,
      "sfomuseum:placetype": "terminal",
      "sfomuseum:terminal_id": "CENTRAL",
      "wof:country": "US",
      "wof:hierarchy": [
        {
          "building_id": 1159396339,
          "campus_id": 102527513,
          "continent_id": 102191575,
          "country_id": 85633793,
          "county_id": 102087579,
          "locality_id": 85922583,
          "neighbourhood_id": -1,
          "region_id": 85688637,
          "wing_id": 1159396165
        }
      ],
      "wof:id": 1159396165,
      "wof:lastmodified": 1547232162,
      "wof:name": "Central Terminal",
      "wof:parent_id": 1159396339,
      "wof:path": "115/939/616/5/1159396165.geojson",
      "wof:placetype": "wing",
      "wof:repo": "sfomuseum-data-architecture",
      "wof:superseded_by": [
        1159396149
      ],
      "wof:supersedes": [
        1159396171
      ]
    },

    ... and so on
   }
]   
```

#### Filters

##### Existential flags

It is possible to filter results by one or more existential flags (`-is-current`, `-is-ceased`, `-is-deprecated`, `-is-superseded`, `-is-superseding`). For example, this query for a point at SFO airport returns 24 possible candidates:

```
$> ./bin/pip \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/sfom-arch.db' \
	-latitude 37.616951 \
	-longitude -122.383747

| jq | grep wof:id | wc -l

2020/12/17 17:01:16 Time to point in polygon, 38.131108ms
      24
```

But when filtered using the `-is-current 1` flag there is only a single result:

```
> ./bin/pip \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/sfom-arch.db' \
	-latitude 37.616951 \
	-longitude -122.383747 \
	-is-current 1

| jq

2020/12/17 17:00:11 Time to point in polygon, 46.401411ms
{
  "places": [
    {
      "wof:id": "1477855655",
      "wof:parent_id": "1477855607",
      "wof:name": "Terminal 2 Main Hall",
      "wof:country": "US",
      "wof:placetype": "concourse",
      "mz:latitude": 37.617044,
      "mz:longitude": -122.383533,
      "mz:min_latitude": 37.61569458544746,
      "mz:min_longitude": 37.617044,
      "mz:max_latitude": -122.3849257355292,
      "mz:max_longitude": -122.38294919235318,
      "mz:is_current": 1,
      "mz:is_deprecated": 0,
      "mz:is_ceased": 1,
      "mz:is_superseded": 0,
      "mz:is_superseding": 1,
      "wof:path": "147/785/565/5/1477855655.geojson",
      "wof:repo": "sfomuseum-data-architecture",
      "wof:lastmodified": 1569430965
    }
  ]
}
```

##### Alternate geometries

You can also filter results to one or more specific alternate geometry labels. For example here are the `quattroshapes` and `whosonfirst-reversegeo` geometries for a point in the city of Montreal, using a SQLite database created from the `whosonfirst-data-admin-ca` database:

```
$> ./bin/pip \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/ca-alt.db' \
	-latitude 45.572744 \
	-longitude -73.586295 \
	-alternate-geometry quattroshapes \
	-alternate-geometry whosonfirst-reversegeo

| jq | grep wof:name

2020/12/17 16:52:08 Time to point in polygon, 419.727612ms
      "wof:name": "136251273 alt geometry (quattroshapes)",
      "wof:name": "85633041 alt geometry (whosonfirst-reversegeo)",
      "wof:name": "85874359 alt geometry (quattroshapes)",
```

Note: These examples assumes a database that was previously indexed using the [whosonfirst/go-whosonfirst-sqlite-features](https://github.com/whosonfirst/go-whosonfirst-sqlite-features) `wof-sqlite-index-features` tool. For example:

```
$> ./bin/wof-sqlite-index-features \
	-rtree \
	-spr \
	-properties \
	-dsn /tmp/test.db
	-mode repo:// \
	/usr/local/data/sfomuseum-data-architecture/
```

The exclude alternate geometries from query results pass the `-geometries default` flag:

```
$> ./bin/pip \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/ca-alt.db' \
	-latitude 45.572744 \
	-longitude -73.586295 \
	-geometries default

| jq | grep wof:name

2020/12/17 17:07:31 Time to point in polygon, 405.430776ms
      "wof:name": "Canada",
      "wof:name": "Saint-Leonard",
      "wof:name": "Quartier Port-Maurice",
      "wof:name": "Montreal",
      "wof:name": "Quebec",
```

To limit query results to _only_ alternate geometries pass the `-geometries alternate` flag:

```
$> ./bin/pip \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/ca-alt.db' \
	-latitude 45.572744 \
	-longitude -73.586295 \
	-geometries alternate

2020/12/17 17:07:39 Time to point in polygon, 366.347365ms
      "wof:name": "85874359 alt geometry (quattroshapes)",
      "wof:name": "85633041 alt geometry (naturalearth)",
      "wof:name": "85633041 alt geometry (naturalearth-display-terrestrial-zoom6)",
      "wof:name": "136251273 alt geometry (whosonfirst)",
      "wof:name": "136251273 alt geometry (quattroshapes)",
      "wof:name": "85633041 alt geometry (whosonfirst-reversegeo)",
```

#### Remote databases

Support for remotely-hosted SQLite databases is available. For example:

```
$> go run -mod vendor cmd/pip/main.go \
	-spatial-database-uri 'sqlite://?dsn=http://localhost:8080/sfomuseum-architecture.db' \
	-latitude 37.616951 \
	-longitude -122.383747 \
	-is-current 1 \

| json_pp | grep "wof:name"

         "wof:name" : "Terminal Two Arrivals",
         "wof:name" : "Terminal 2",
         "wof:name" : "SFO Terminal Complex",
         "wof:name" : "Terminal 2 Main Hall",
         "wof:name" : "SFO Terminal Complex",
```

_Big thanks to @psanford 's [sqlitevfshttp](https://github.com/psanford/sqlite3vfshttp) package for making this possible._

### http-server

```
> ./bin/http-server -h
  -authenticator-uri string
    	A valid sfomuseum/go-http-auth URI. (default "null://")
  -cors-allow-credentials
    	Allow HTTP credentials to be included in CORS requests.
  -cors-origin value
    	One or more hosts to allow CORS requests from; may be a comma-separated list.
  -custom-placetypes string
    	A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.
  -enable-cors
    	Enable CORS headers for data-related and API handlers.
  -enable-custom-placetypes
    	Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.
  -enable-geojson
    	Enable GeoJSON output for point-in-polygon API calls.
  -enable-gzip
    	Enable gzip-encoding for data-related and API handlers.
  -enable-www
    	Enable the interactive /debug endpoint to query points and display results.
  -is-wof
    	Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents. (default true)
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported schemes are: directory://, featurecollection://, file://, filelist://, geojsonl://, null://, repo://. (default "repo://")
  -leaflet-initial-latitude float
    	The initial latitude for map views to use. (default 37.616906)
  -leaflet-initial-longitude float
    	The initial longitude for map views to use. (default -122.386665)
  -leaflet-initial-zoom int
    	The initial zoom level for map views to use. (default 14)
  -leaflet-max-bounds string
    	An optional comma-separated bounding box ({MINX},{MINY},{MAXX},{MAXY}) to set the boundary for map views.
  -log-timings
    	Emit timing metrics to the application's logger
  -map-provider-uri string
    	A valid aaronland/go-http-maps/provider URI. (default "leaflet://?leaflet-tile-url=https://tile.openstreetmap.org/{z}/{x}/{y}.png")
  -path-api string
    	The root URL for all API handlers (default "/api")
  -path-data string
    	The URL for data (GeoJSON) handler (default "/data")
  -path-ping string
    	The URL for the ping (health check) handler (default "/health/ping")
  -path-pip string
    	The URL for the point in polygon web handler (default "/point-in-polygon")
  -path-prefix string
    	Prepend this prefix to all assets (but not HTTP handlers). This is mostly for API Gateway integrations.
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. Available options are: [fs:// null:// repo:// sqlite:// stdin://]. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.
  -server-uri string
    	A valid aaronland/go-http-server URI. (default "http://localhost:8080")
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: [rtree:// sqlite://] (default "rtree://")
```

For example:

```
$> bin/http-server \
	-enable-www \
	-spatial-database-uri 'sqlite:///?dsn=modernc:///usr/local/data/sfomuseum-data-architecture.db'
```

A couple things to note:

* The SQLite databases specified in the `sqlite:///?dsn` string are expected to minimally contain the `rtree` and `spr` and `properties` tables confirming to the schemas defined in the [go-whosonfirst-sqlite-features](https://github.com/whosonfirst/go-whosonfirst-sqlite-features). They are typically produced by the [go-whosonfirst-sqlite-features-index](https://github.com/whosonfirst/go-whosonfirst-sqlite-features-index) package. See the documentation in the [go-whosonfirst-spatial-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite) package for details.

When you visit `http://localhost:8080` in your web browser you should see something like this:

![](docs/images/server.png)

If you don't need, or want, to expose a user-facing interface simply remove the `-enable-www` and `-nextzen-apikey` flags. For example:

```
$> bin/http-server \
	-spatial-database-uri 'sqlite:///?dsn=modernc:///usr/local/data/sfomuseum-data-architecture.db' 
```

And then to query the point-in-polygon API you would do something like this:

```
$> curl -X POST -s 'http://localhost:8080/api/point-in-polygon' -d '{"latitude":37.61701894316063, "longitude":-122.3866653442383}'

{
  "places": [
    {
      "wof:id": 1360665043,
      "wof:parent_id": -1,
      "wof:name": "Central Parking Garage",
      "wof:placetype": "wing",
      "wof:country": "US",
      "wof:repo": "sfomuseum-data-architecture",
      "wof:path": "136/066/504/3/1360665043.geojson",
      "wof:superseded_by": [],
      "wof:supersedes": [
        1360665035
      ],
      "mz:uri": "https://data.whosonfirst.org/136/066/504/3/1360665043.geojson",
      "mz:latitude": 37.616332,
      "mz:longitude": -122.386047,
      "mz:min_latitude": 37.61498599208708,
      "mz:min_longitude": -122.38779093748578,
      "mz:max_latitude": 37.61767331604971,
      "mz:max_longitude": -122.38429192207244,
      "mz:is_current": 0,
      "mz:is_ceased": 1,
      "mz:is_deprecated": 0,
      "mz:is_superseded": 0,
      "mz:is_superseding": 1,
      "wof:lastmodified": 1547232156
    }
    ... and so on
}    
```

By default, results are returned as a list of ["standard places response"](https://github.com/whosonfirst/go-whosonfirst-spr/) (SPR) elements. You can also return results as a GeoJSON `FeatureCollection` by passing the `-enable-geojson` flag to the server and including a `format=geojson` query parameter with requests. For example:


```
$> bin/http-server \
	-enable-geojson \
	-spatial-database-uri 'sqlite:///?dsn=modernc:///usr/local/data/sfomuseum-data-architecture.db'
```

And then:

```
$> curl -s -XPOST -H 'Accept: application/geo+json' 'http://localhost:8080/api/point-in-polygon' -d '{"latitude":37.61701894316063,"longitude":-122.3866653442383 }'

{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {
        "type": "MultiPolygon",
        "coordinates": [ ...omitted for the sake of brevity ]
      },
      "properties": {
        "mz:is_ceased": 1,
        "mz:is_current": 0,
        "mz:is_deprecated": 0,
        "mz:is_superseded": 0,
        "mz:is_superseding": 1,
        "mz:latitude": 37.616332,
        "mz:longitude": -122.386047,
        "mz:max_latitude": 37.61767331604971,
        "mz:max_longitude": -122.38429192207244,
        "mz:min_latitude": 37.61498599208708,
        "mz:min_longitude": -122.38779093748578,
        "mz:uri": "https://data.whosonfirst.org/136/066/504/3/1360665043.geojson",
        "wof:country": "US",
        "wof:id": 1360665043,
        "wof:lastmodified": 1547232156,
        "wof:name": "Central Parking Garage",
        "wof:parent_id": -1,
        "wof:path": "136/066/504/3/1360665043.geojson",
        "wof:placetype": "wing",
        "wof:repo": "sfomuseum-data-architecture",
        "wof:superseded_by": [],
        "wof:supersedes": [
          1360665035
        ]
      }
    }
    ... and so on
  ]
}  
```

### grpc-server

```
$> ./bin/grpc-server -h
  -custom-placetypes string
    	A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.
  -enable-custom-placetypes
    	Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.
  -host string
    	The host to listen for requests on (default "localhost")
  -is-wof
    	Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents. (default true)
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported schemes are: directory://, featurecollection://, file://, filelist://, geojsonl://, null://, repo://. (default "repo://")
  -port int
    	The port to listen for requests on (default 8082)
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. Available options are: [fs:// null:// repo:// sqlite:// stdin://]. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: [rtree:// sqlite://] (default "rtree://")
```	

For example:

```
$> ./bin/grpc-server -spatial-database-uri 'sqlite://?dsn=modernc:///usr/local/data/arch.db' 
2024/07/19 10:52:47 Listening on localhost:8082
```

And then in another terminal:

```
$> ./bin/grpc-client -latitude 37.621131 -longitude -122.384292 | jq '.places[]["name"]'
"San Francisco International Airport"
```

### grpc-client

```
$> ./bin/grpc-client -h
  -alternate-geometry value
    	One or more alternate geometry labels (wof:alt_label) values to filter results by.
  -cessation string
    	A valid EDTF date string.
  -geometries string
    	Valid options are: all, alt, default. (default "all")
  -host string
    	The host of the gRPC server to connect to. (default "localhost")
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
  -latitude float
    	A valid latitude.
  -longitude float
    	A valid longitude.
  -null
    	Emit results to /dev/null
  -placetype value
    	One or more place types to filter results by.
  -port int
    	The port of the gRPC server to connect to. (default 8082)
  -property value
    	One or more Who's On First properties to append to each result.
  -sort-uri value
    	Zero or more whosonfirst/go-whosonfirst-spr/sort URIs.
  -stdout
    	Emit results to STDOUT (default true)
```

For example:

```
$> ./bin/grpc-client -latitude 37.621131 -longitude -122.384292 | jq '.places[]["name"]'
"San Francisco International Airport"
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-spatial-www
* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc
* https://github.com/whosonfirst/go-whosonfirst-sqlite
* https://github.com/whosonfirst/go-whosonfirst-sqlite-features
* https://github.com/whosonfirst/go-whosonfirst-sqlite-features-index
* https://github.com/whosonfirst/go-reader
