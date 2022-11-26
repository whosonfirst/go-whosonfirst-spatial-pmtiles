# go-whosonfirst-spatial

## Documentation

Documentation, particularly proper Go documentation, is incomplete at this time.

## Motivation

`go-whosonfirst-spatial` is an attempt to de-couple the various components that make up the [go-whosonfirst-pip-v2](https://github.com/whosonfirst/go-whosonfirst-pip-v2) package – indexing, storage, querying and serving – in to separate packages in order to allow for more flexibility.

It is the "base" package that defines provider-agnostic, but WOF-specific, interfaces for a limited set of spatial queries and reading properties.

These interfaces are then implemented in full or in part by provider-specific classes. For example, an in-memory RTree index or SQLite:

* https://github.com/whosonfirst/go-whosonfirst-spatial-rtree
* https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite

Building on that there are equivalent base packages for "server" implementations, like:

* https://github.com/whosonfirst/go-whosonfirst-spatial-www
* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc

The idea is that all of these pieces can be _easily_ combined in to purpose-fit applications.  As a practical matter it's mostly about trying to identify and package the common pieces in to as few lines of code as possible so that they might be combined with an application-specific `import` statement. For example:

```
import (
         _ "github.com/whosonfirst/go-whosonfirst-spatial-MY-SPECIFIC-REQUIREMENTS"
)
```

Here is a concrete example, implementing a point-in-polygon service over HTTP using a SQLite backend:

* https://github.com/whosonfirst/go-whosonfirst-spatial-www-sqlite/blob/main/cmd/server/main.go

It is part of the overall goal of:

* Staying out people's database or delivery choices (or needs)
* Supporting as many databases (and delivery (and indexing) choices) as possible
* Not making database B a dependency (in the Go code) in order to use database A, as in not bundling everything in a single mono-repo that becomes bigger and has more requirements over time.

That's the goal, anyway. I am still working through the implementation details.

Functionally the `go-whosonfirst-spatial-` packages should be equivalent to `go-whosonfirst-pip-v2` as in there won't be any functionality _removed_.

## Example

```
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-spatial/app"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"		
	_ "github.com/whosonfirst/go-whosonfirst-spatial-rtree"	
)

func main() {

	fl, _ := flags.CommonFlags()
	flags.Parse(fl)

	flags.ValidateCommonFlags(fl)

	paths := fl.Args()
	
	ctx := context.Background()

	spatial_app, _ := app.NewSpatialApplicationWithFlagSet(ctx, fl)
	spatial_app.IndexPaths(ctx, paths...)

	c, _ := geo.NewCoordinate(-122.395229, 37.794906)
	f, _ := filter.NewSPRFilter()

	spatial_db := spatial_app.SpatialDatabase
	spatial_results, _ := spatial_db.PointInPolygon(ctx, c, f)

	body, _ := json.Marshal(spatial_results)
	fmt.Println(string(body))
}
```

_Error handling omitted for brevity._

## Concepts

### Applications

_Please write me_

### Database

_Please write me_

### Filters

_Please write me_

### Indices

_Please write me_

### Standard Places Response (SPR)

_Please write me_

## Implementations

* https://github.com/whosonfirst/go-whosonfirst-spatial-rtree
* https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite

## Servers and clients

* https://github.com/whosonfirst/go-whosonfirst-spatial-www
* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc

## Services

* https://github.com/whosonfirst/go-whosonfirst-spatial-pip
* https://github.com/whosonfirst/go-whosonfirst-spatial-hierarchy