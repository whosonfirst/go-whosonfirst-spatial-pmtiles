# go-whosonfirst-spatial-pip

Opionated point-in-polygon operations for `go-whosonfirst-spatial` packages.

## IMPORTANT

This is work in progress. Documentation to follow.

_If you're reading this it means the documentation below is out of date._

## Background

This package exports point-in-polygon (PIP) applications using the `whosonfirst/go-whosonfirst-spatial` interfaces.

The code in this package does not contain any specific implementation of those interfaces so when invoked on its own it won't work as expected.

The code in this package is designed to be imported by _other_ code that also loads the relevant packages that implement the `whosonfirst/go-whosonfirst-spatial` interfaces. For example, here is the `query` application using the `whosonfirst/go-whosonfirst-spatial-sqlite` package. This application is part of the [go-whosonfirst-spatial-pip-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-pip-sqlite) package:

```
package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-spatial-pip/query"
)

func main() {

	ctx := context.Background()

	fs, _ := query.NewQueryApplicationFlagSet(ctx)
	app, _ := query.NewQueryApplication(ctx)

	app.RunWithFlagSet(ctx, fs)
}
```

The idea is that this package defines code to implement opinionated applications without specifying an underlying database implementation (interface).

As of this writing this package exports two "applications":

* The "Query" application performs a basic point-in-polygon (PIP) query, with optional "standard places response" (SPR) filters.

* The "Update" application accepts a series of Who's On First (WOF) records and attempts to assign a "parent" ID and hierarchy by performing one or more PIP operations for that record's centroid and potential ancestors (derived from its placetype). If successful the application also tries to "write" the updated feature to a target that implements the `whosonfirst/go-writer` interface.

Although there is a substantial amount of overlap, conceptually, between the two applications not all those similarities have been reconciled. These include:

* The "Update" application will, optionally, attempt to populate (or index) a spatial database when it starts. The "Query" application does not yet.

* The "Query" application is designed to run in a number of different "modes". These are: As a command line application; As a standalone HTTP server; As an AWS Lambda function. The "Update" application currently only runs as a command line application.

Both applications also use the `whosonfirst/go-whosonfirst-spatial/flags` package for common "spatial" application flags. In practice this tends to be more confusing than not so that may change too.

## Applications

_The examples shown here assume applications that have been built with the [whosonfirst/go-whosonfirst-spatial-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite). Although there are sample applications bundled in this package's `examples` folder because they don't load anything that implements the `go-whosonfirst-spatial` interfaces they won't work. They are included as reference implementations._

### Query

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
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Supported schemes are: directory://, featurecollection://, file://, filelist://, geojsonl://, repo://. (default "repo://")
  -latitude float
    	A valid latitude.
  -longitude float
    	A valid longitude.
  -mode string
    	... (default "cli")
  -placetype value
    	One or more place types to filter results by.
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. Available options are: [file:// fs:// null://]
  -property value
    	One or more Who's On First properties to append to each result.
  -server-uri string
    	... (default "http://localhost:8080")
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: [rtree://]
  -verbose
    	Be chatty.
```

#### Command line

```
$> ./bin/query \
	-spatial-database-uri 'sqlite://?dsn=/usr/local/data/arch.db' \
	-latitude 37.616951 \
	-longitude -122.383747 \
	-is-current 1

| jq '.["places"][]["wof:id"]'

"1729792685"
"1729792433"
```

#### Server

```
$> ./bin/query -mode server -spatial-database-uri 'sqlite://?dsn=/usr/local/data/arch.db'
```

And in another terminal:

```
$> curl -s -XPOST \
	http://localhost:8080/ \
	-d '{"latitude":37.616951,"longitude":-122.383747,"is_current":[1]}' \

| jq '.["places"][]["wof:id"]'

"1729792685"
"1729792433"
```

#### Lambda (using container images)

##### Running locally

_This assumes that you have packaged the `query` tool as a container image. For an example of this take a look at the [whosonfirst/go-whosonfirst-pip-sqlite](https://github.com/whosonfirst/go-whosonfirst-spatial-pip-sqlite) package. Note that the `go-whosonfirst-spatial-pip-sqlite` package bundles a SQLite database inside the container image itself._

```
$> docker run -e PIP_MODE=lambda -e PIP_SPATIAL_DATABASE_URI=sqlite://?dsn=/usr/local/data/query.db -p 9000:8080 point-in-polygon:latest /main
time="2021-03-11T01:19:37.994" level=info msg="exec '/main' (cwd=/go, handler=)"
```

And then in another terminal:

```
$> curl -s -XPOST \
	"http://localhost:9000/2015-03-31/functions/function/invocations" \
	-d '{"latitude":37.616951,"longitude":-122.383747,"is_current":[1]}' \

| jq '.["places"][]["wof:id"]'

"1729792685"
"1729792433"
```

##### Running in AWS

Update your container (see above) to a your AWS ECS repository. Create a new AWS Lambda function and configure it to use your container.

Ensure the following image configuration variables are assigned:

| Name | Value |
| --- | --- |
| CMD override | /main |

Ensure the following environment variables are assigned:

| Name | Value |
| --- | --- |
| PIP_MODE | lambda |
| PIP_SPATIAL_URI | sqlite://?dsn=/usr/local/data/query.db |

Create a test like this and invoke it:

```
{
  "latitude": 37.616951,
  "longitude": -122.383747,
  "is_current": [
    1
  ]
}
```

##### Running in AWS with API Gateway

Ensure the following environment variables are assigned:

| Name | Value |
| --- | --- |
| PIP_MODE | server |
| PIP_SERVER_URI | lambda:// |
| PIP_SPATIAL_URI | sqlite://?dsn=/usr/local/data/query.db |

_To be written_

### Update

Perform point-in-polygon (PIP), and related update, operations on a set of Who's on First records.

```
$> ./bin/point-in-polygon -h
Perform point-in-polygon (PIP), and related update, operations on a set of Who's on First records.
Usage:
	 ./bin/point-in-polygon [options] uri(N) uri(N)
Valid options are:

  -exporter-uri string
    	A valid whosonfirst/go-whosonfirst-export URI. (default "whosonfirst://")
  -is-ceased value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-current value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-deprecated value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-superseded value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -is-superseding value
    	One or more existential flags (-1, 0, 1) to filter PIP results.
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI scheme. This is used to identify WOF records to be PIP-ed. (default "repo://")
  -mapshaper-server string
    	A valid HTTP URI pointing to a sfomuseum/go-sfomuseum-mapshaper server endpoint. (default "http://localhost:8080")
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial URI. This is the database of spatial records that will for PIP-ing.
  -spatial-iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI scheme. This is used to identify WOF records to be indexed in the spatial database. (default "repo://")
  -spatial-source value
    	One or more URIs to be indexed in the spatial database (used for PIP-ing).
  -writer-uri string
    	A valid whosonfirst/go-writer URI. This is where updated records will be written to. (default "null://")
```

#### Command line

For example:

```
> ./bin/point-in-polygon \
	-writer-uri 'featurecollection://?writer=stdout://' \
	-spatial-database-uri 'sqlite://?dsn=:memory:' \
	-spatial-iterator-uri 'repo://?include=properties.mz:is_current=1' \
	-spatial-source /usr/local/data/sfomuseum-data-architecture \
	-iterator-uri 'repo://?include=properties.mz:is_current=1' \
	/usr/local/data/sfomuseum-data-publicart \
| jq '.features[]["properties"]["wof:parent_id"]' \
| sort \
| uniq \

-1
1159162825
1159162827
1477855979
1477855987
1477856005
1729791967
1729792389
1729792391
1729792433
1729792437
1729792459
1729792483
1729792489
1729792551
1729792577
1729792581
1729792643
1729792645
1729792679
1729792685
1729792689
1729792693
1729792699
```

So what's going on here?

The first thing we're saying is: Write features that have been PIP-ed to the [featurecollection](https://github.com/whosonfirst/go-writer-featurecollection) writer (which in turn in writing it's output to [STDOUT](https://github.com/whosonfirst/go-writer).

```
	-writer-uri 'featurecollection://?writer=stdout://'
```

Then we're saying: Create a new in-memory [SQLite spatial database](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite) to use for performing PIP operations.

```
	-spatial-database-uri 'sqlite://?dsn=:memory:' 
```

We're also going to create this spatial database on-the-fly by reading records in the `sfomuseum-data-architecture` respository selecting only records with a `mz:is_current=1` property.

```
	-spatial-iterator-uri 'repo://?include=properties.mz:is_current=1'
	-spatial-source /usr/local/data/sfomuseum-data-architecture 
```

If we already had a pre-built [SQLite database](https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite#databases) we could specify it like this:

```
	-spatial-database-uri 'sqlite://?dsn=/path/to/sqlite.db' 
```

Next we define our _input_ data. This is the data is going to be PIP-ed. We are going to read records from the `sfomuseum-data-publicart` repository selecting only records with a `mz:is_current=1` property.

```
	-iterator-uri 'repo://?include=properties.mz:is_current=1' 
	/usr/local/data/sfomuseum-data-publicart 
```

Finally we pipe the results (a GeoJSON `FeatureCollection` string output to STDOUT) to the `jq` tool for filtering out `wof:parent_id` properties and then to the `sort` and `uniq` utlities to format the results.

```
| jq '.features[]["properties"]["wof:parent_id"]' 
| sort 
| uniq 
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-spatial-rtree
* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-whosonfirst-exporter
* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-writer
* https://github.com/sfomuseum/go-sfomuseum-mapshaper
