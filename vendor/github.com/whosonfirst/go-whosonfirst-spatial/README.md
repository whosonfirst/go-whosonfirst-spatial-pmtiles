# go-whosonfirst-spatial

Go package defining interfaces for Who's On First specific spatial operations.

## Documentation

Documentation, particularly proper Go documentation, is incomplete at this time.

## Motivation

The goal of the `go-whosonfirst-spatial` package is to de-couple the various components that made up the now-deprecated [go-whosonfirst-pip-v2](https://github.com/whosonfirst/go-whosonfirst-pip-v2) package – indexing, storage, querying and serving – in to separate packages in order to allow for more flexibility.

It is the "base" package that defines provider-agnostic, but WOF-specific, interfaces for a limited set of spatial queries and reading properties.

These interfaces are then implemented in full or in part by provider-specific classes. For example, an in-memory RTree index or a SQLite database or even a Protomaps database:

* https://github.com/whosonfirst/go-whosonfirst-spatial-rtree
* https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite
* https://github.com/whosonfirst/go-whosonfirst-spatial-pmtiles

_You may have noticed the absence of an equivalent `go-whosonfirst-spatial-postgis` or even `go-whosonfirst-spatial-mysql` implementation. That's only because I've been focusing on implementations with fewer requirements, dependencies and less overhead to set up and maintain. There is no reason there couldn't be implementations for either database and some day I hope there will be._

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

* https://github.com/whosonfirst/go-whosonfirst-spatial-www/blob/main/application/server
* https://github.com/whosonfirst/go-whosonfirst-spatial-www-sqlite/blob/main/cmd/server/main.go

It is part of the overall goal of:

* Staying out people's database or delivery choices (or needs)
* Supporting as many databases (and delivery (and indexing) choices) as possible
* Not making database `B` a dependency (in the Go code) in order to use database `A`, as in not bundling everything in a single mono-repo that becomes bigger and has more requirements over time.

Importantly this package does not implement any actual spatial functionality. It defines the interfaces that are implemented by other packages which allows code to function without the need to consider the underlying mechanics of how spatial operations are being performed.

The layout of this package remains in flux and is likely to change. Things have almost settled but not quite yet.

## Interfaces

### SpatialIndex

```
type SpatialIndex interface {
	IndexFeature(context.Context, []byte) error
	RemoveFeature(context.Context, string) error
	PointInPolygon(context.Context, *orb.Point, ...Filter) (spr.StandardPlacesResults, error)
	PointInPolygonCandidates(context.Context, *orb.Point, ...Filter) ([]*PointInPolygonCandidate, error)
	PointInPolygonWithChannels(context.Context, chan spr.StandardPlacesResult, chan error, chan bool, *orb.Point, ...Filter)
	PointInPolygonCandidatesWithChannels(context.Context, chan *PointInPolygonCandidate, chan error, chan bool, *orb.Point, ...Filter)
	Disconnect(context.Context) error
}
```

_Where `orb.*` and `spr.*` refer to the [paulmach/orb](https://github.com/paulmach/orb) and [whosonfirst/go-whosonfirst-flags](https://github.com/whosonfirst/go-whosonfirst-flags) packages respectively._

### SpatialDatabase

```
type SpatialDatabase interface {
	reader.Reader
	writer.Writer
	spatial.SpatialIndex
}
```

_Where `reader.Reader` and `writer.Writer` are the [whosonfirst/go-reader](https://pkg.go.dev/github.com/whosonfirst/go-reader#Reader) and [whosonfirst/go-writer](https://pkg.go.dev/github.com/whosonfirst/go-writer#Writer) interfaces, respectively._

### Filter

```
type Filter interface {
	HasPlacetypes(flags.PlacetypeFlag) bool
	MatchesInception(flags.DateFlag) bool
	MatchesCessation(flags.DateFlag) bool
	IsCurrent(flags.ExistentialFlag) bool
	IsDeprecated(flags.ExistentialFlag) bool
	IsCeased(flags.ExistentialFlag) bool
	IsSuperseded(flags.ExistentialFlag) bool
	IsSuperseding(flags.ExistentialFlag) bool
	IsAlternateGeometry(flags.AlternateGeometryFlag) bool
	HasAlternateGeometry(flags.AlternateGeometryFlag) bool
}
```

_Where `flags.*` refers to the [whosonfirst/go-whosonfirst-flags](https://github.com/whosonfirst/go-whosonfirst-flags) package._

## Implementations

* https://github.com/whosonfirst/go-whosonfirst-spatial-rtree
* https://github.com/whosonfirst/go-whosonfirst-spatial-sqlite
* https://github.com/whosonfirst/go-whosonfirst-spatial-pmtiles

## Servers and clients

### WWW

* https://github.com/whosonfirst/go-whosonfirst-spatial-www
* https://github.com/whosonfirst/go-whosonfirst-spatial-www-sqlite
* https://github.com/whosonfirst/go-whosonfirst-spatial-www-pmtiles

### gRPC

* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc
* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc-sqlite
* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc-pmtiles

## Services and Operations

* https://github.com/whosonfirst/go-whosonfirst-spatial-pip
* https://github.com/whosonfirst/go-whosonfirst-spatial-hierarchy

## See also

* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-whosonfirst-flags
* https://github.com/whosonfirst/go-reader
* https://github.com/whosonfirst/go-writer
* https://github.com/paulmach/orb
