// package spr provides an interface which defines the minimum set of methods that a system working with a collection of Who's On First (WOF) must implement for any given record. Not all records are the same so the SPR interface is meant to serve as a baseline for common data that describes every record.
//
// The `StandardPlacesResult` (SPR) interface defines the _minimum_ set of methods that a system working with a collection of Who's On First (WOF) must implement for any given record. Not all records are the same so the SPR interface is meant to serve as a baseline for common data that describes every record.
//
// The `StandardPlacesResults` takes the Flickr [standard photo response](https://code.flickr.net/2008/08/19/standard-photos-response-apis-for-civilized-age) as its inspiration which was designed to be the minimum amount of information about a Flickr photo necessary to display that photo with proper attribution and a link back to the photo page itself. The `StandardPlacesResults` aims to achieve the same thing for WOF records.
//
// Being a [Go language interface type](https://www.alexedwards.net/blog/interfaces-explained) the SPR is _not_ designed as a data exchange method. Any given implementation of the SPR _may_ allow its internal data to be exported or serialized (for example, as JSON) but this is not a requirement.
//
// For a concrete example of a package that implements the `SPR` have a look at the [go-whosonfirst-sqlite-spr](https://github.com/whosonfirst/go-whosonfirst-sqlite-spr) package.
package spr
