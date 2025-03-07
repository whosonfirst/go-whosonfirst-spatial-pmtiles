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

Here's an example of the creating a compatible SQLite database for all the [administative data in Canada](https://github.com/whosonfirst-data/whosonfirst-data-admin-ca) using the `wof-sqlite-index` tool which is part of the [go-whosonfirst-database-sqlite](https://github.com/whosonfirst/go-whosonfirst-database-sqlite) package:

```
$> ./bin/wof-sqlite-index \
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
	-database-uri 'sqlite://sqlite3?dsn=/usr/local/data/ca-alt.db' \
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

## Database URIs and "drivers"

Database URIs for the `go-whosonfirst-spatial-sqlite` package take the form of:

```
"sqlite://" + {DATABASE_SQL_ENGINE} + "?dsn=" + {DATABASE_SQL_DSN}
```

Where `DATABASE_SQL` refers to the build-in [database/sql](https://pkg.go.dev/database/sql) package.

For example:

```
sqlite://sqlite3?dsn=test.db
```

By default this package bundles support for the [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) driver but does NOT enable it by default. You will need to pass in the `-tag mattn` argument when building tools to enable it. This is the default behaviour in the `cli` Makefile target for building binary tools.

If you want or need to use the [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) driver take a look at the [database_mattn.go](database_mattn.go) file for an example of how you might go about enabling it. As of this writing the `modernc.org/sqlite` package is not bundled with this package because it adds ~200MB of code to the `vendor` directory.

## Example

```
package main

import (
	"context"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
	"github.com/whosonfirst/go-whosonfirst-spatial/properties"
	"github.com/whosonfirst/go-whosonfirst-spr"
)

func main() {

	database_uri := "sqlite://sqlite3?dsn=whosonfirst.db"
	properties_uri := "sqlite://sqlite3?dsn=whosonfirst.db"
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

$> make cli
go build -tags mattn -ldflags="-s -w" -mod vendor -o bin/http-server cmd/http-server/main.go
go build -tags mattn -ldflags="-s -w" -mod vendor -o bin/grpc-server cmd/grpc-server/main.go
go build -tags mattn -ldflags="-s -w" -mod vendor -o bin/grpc-client cmd/grpc-client/main.go
go build -tags mattn -ldflags="-s -w" -mod vendor -o bin/update-hierarchies cmd/update-hierarchies/main.go
go build -tags mattn -ldflags="-s -w" -mod vendor -o bin/pip cmd/pip/main.go
go build -tags mattn -ldflags="-s -w" -mod vendor -o bin/intersects cmd/intersects/main.go

### pip

Documentation for the `pip` tool has been moved in to [cmd/pip/README.md](cmd/pip/README.md)

### intersects

Documentation for the `pip` tool can be found in [cmd/intersects/README.md](cmd/intersects/README.md)

### http-server

Documentation for the `pip` tool has been moved in to [cmd/http-server/README.md](cmd/http-server/README.md)

### grpc-server

Documentation for the `pip` tool has been moved in to [cmd/grpc-server/README.md](cmd/grpc-server/README.md)

### grpc-client

Documentation for the `pip` tool has been moved in to [cmd/grpc-client/README.md](cmd/grpc-client/README.md)

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-spatial-www
* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc
* https://github.com/whosonfirst/go-whosonfirst-database
* https://github.com/whosonfirst/go-whosonfirst-database-sqlite
* https://github.com/whosonfirst/go-reader
