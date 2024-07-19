# Point-in-polygon "hierarchy resolvers"

At a high-level a point-in-polygon "hierarchy resolver" consists of (4) parts:

* Given a GeoJSON Feature use its geometry to derive the most appropriate centroid for performing a point-in-polygon query
* Perform a point-in-polygon query for a centroid, excluding results using criteria defined by zero or more filters.
* Convert the list of candidate results (derived from the point-in-polygon query) in to a single result using a callback function.
* Apply updates derived from the final result to the original GeoJSON Feature using a callback function.

These functionalities are implemented by the `hierarchy.PointInPolygonHierarchyResolver` package. In addition to wrapping all those moving pieces the `hierachy` package also exports a handful of predefined callback functions to use for filtering results and applying updates.

Importantly, hierarchy resolvers are not responsible for reading Who's On First documents, writing updates to those documents or populating the spatial databases used to perform point-in-polygon operations. These tasks are left to other bits of code. The principal goal of a hierarchy resolver is to perform a point-in-polygon operation, resolve multiple overlapping candidates down to a single result and then generate/apply updates (to a source document) derive from that result.

## Example

The following examples describe how to use the `hierarchy.PointInPolygonHierarchyResolver` package in abbreviated (incomplete) and annotated code. These example do not reflect all the functionality of the `hierarchy.PointInPolygonHierarchyResolver` package. For details consult the [Go reference documentation](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-spatial/hierarchy).

_Note: For the sake of brevity all error-handling has been removed from these examples._

### Basic

This example demonstrates how to use the `hierarchy.PointInPolygonHierarchyResolver` package with a set of "core" Who's On First documents.

```
import (
       "context"

       "github.com/sfomuseum/go-sfomuseum-mapshaper"
       "github.com/whosonfirst/go-whosonfirst-spatial/database"
       _ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
       "github.com/whosonfirst/go-whosonfirst-spatial/filter"
       "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy"
       hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy/filter"       
)

func main() {

	ctx := context.Background()
	
	// The Mapshaper "client" (and its associated "server") is not required by a point-in-polygon
	// hierarchy resolver but is included in this example for the sake of thoroughness. If present
	// it will be used to derive the centroid for a GeoJSON Feature using the Mapshape "inner point"
	// command. Both the "client" and "server" components are part of the [sfomuseum/go-sfomuseum-mapshaper](#)
	// package but setting up and running the "server" component is out of scope for this document.
	// Basically Mapshaper's "inner point" functonality can't be ported to Go fast enough.
	//
	// If the mapshaper client is `nil` then there are a variety of other heuristics that will be
	// used, based on the content of the input GeoJSON Feature, to derive a candidate centroid to
	// be used for point-in-polygon operations.
        mapshaper_cl, _ := mapshaper.NewClient(ctx, "http://localhost:8080")

	// Create a new spatial database instance. For the sake of this example it
	// is assumed that the database has already been populated with records.
        spatial_db, _ := database.NewSpatialDatabase(ctx, "sqlite://?dsn=modernc://cwd/example.db")

	// Create configuration options for hierarchy resolver
	resolver_opts := &hierarchy.PointInPolygonHierarchyResolverOptions{
		Database:             spatial_db,
	        Mapshaper:            mapshaper_cl,
        }

	// Create the hierarchy resolver itself
        resolver, _ := hierarchy.NewPointInPolygonHierarchyResolver(ctx, resolver_opts)

	// Create zero or more filters to prune point-in-polygon results with, in this case
	// only return records whose `mz:is_current` property is "1".
        pip_inputs := &filter.SPRInputs{
                IsCurrent: []int64{1},
        }

	// Instantiate a predefined results callback function that returns the first result in a list
	// of candidates but does not trigger an error if that list is empty.
        results_cb := hierarchy_filter.FirstButForgivingSPRResultsFunc

	// Instantiate a predefined update callback that will return a dictionary populated with the 
	// following properties from the final point-in-polygon result (derived from `results_cb`):
	// wof:parent_id, wof:hierarchy, wof:country
	update_cb := hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()

	// Where body is assumed to be a valid Who's On First style GeoJSON Feature
	var body []byte

	// Invoke the hierarchy resolver's `PointInPolygonAndUpdate` method using `body` as the input
	// parameter.
	updates, _ := resolver.PointInPolygonAndUpdate(ctx, pip_inputs, results_cb, update_cb, body)

	// Apply updates to body here
}	
```

### Custom placetypes

This example demonstrates how to the `hierarchy.PointInPolygonHierarchyResolver` package with a set of Who's On First style documents that contain custom placetypes (defined in a separate property from the default `wof:placetype` property).

```
import (
       "context"

       "github.com/sfomuseum/go-sfomuseum-mapshaper"
       _ "github.com/sfomuseum/go-sfomuseum-placetypes"
       "github.com/whosonfirst/go-whosonfirst-placetypes"
       "github.com/whosonfirst/go-whosonfirst-spatial/database"
       _ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
       "github.com/whosonfirst/go-whosonfirst-spatial/filter"
       "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy"
       hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy/filter"       
)

func main() {

        mapshaper_cl, _ := mapshaper.NewClient(ctx, "http://localhost:8080")
	
        spatial_db, _ := database.NewSpatialDatabase(ctx, "sqlite://?dsn=modernc://cwd/example.db")

	// Create a new custom placetypes definition. In this case the standard Who's On First places
	// definition supplemented with custom placetypes used by SFO Museum. This is used to derive
	// the list of (custom) ancestors associated with any given (custom) placetype.
	pt_def, _ := placetypes.NewDefinition(ctx, "sfomuseum://")

	// Append the custom placetypes definition to the hierarchy resolver options AND explicitly
	// disable placetype filtering (removing candidates that are not ancestors of the placetype
	// of the Who's On First GeoJSON Feature being PIP-ed.
	//
	// If you don't disable default placetype filtering you will need to ensure that the `wof:placetype`
	// property of the features in the spatial database are manually reassigned to take the form
	// of "PLACETYPE" + "#" + "PLACETYPE_DEFINITION_URI", for example: "airport#sfomuseum://"
	//
	// More accurately though the requirement is less that you need to alter the values in the underlying
	// database so much as ensure that the value returned by the `Placetype` method of each `StandardPlacesResult`
	// (SPR) candidate result produced during a point-in-polygon operation is formatted that way. There is
	// more than one way to do this but as a practical matter it's probably easiest to store that (formatted)
	// value in the database IF the database itself is transient. There can be no guarantees thought that
	// a change like this won't have downstream effects on the rest of your code.
	//
	// If you're curious the "PLACETYPE" + "#" + "PLACETYPE_DEFINITION_URI" syntax is parsed by the
	// code used to create placetype filter flags in the `whosonfirst/go-whosonfirst-flags` package and
	// used to load custom definitions on the fly to satisfy tests.
	//
	// Basically, custom placetypes make things more complicated because they are... well, custom. At a
	// certain point it may simply be easier to disable default placetype checks in your own custom results
	// filtering callback function.
	resolver_opts := &hierarchy.PointInPolygonHierarchyResolverOptions{
		Database:             spatial_db,
	        Mapshaper:            mapshaper_cl,
		PlacetypesDefinition: pt_def,
                SkipPlacetypeFilter:  true,
        }

        resolver, _ := hierarchy.NewPointInPolygonHierarchyResolver(ctx, resolver_opts)

        pip_inputs := &filter.SPRInputs{
                IsCurrent: []int64{1},
        }

	// In the case of SFO Museum related records here is a custom results callback that implements its
	// own placetype and floor level checking.
        results_cb := sfom_hierarchy.ChoosePointInPolygonCandidateStrict

	update_cb := hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()
	
	var body []byte
	
	updates, _ := resolver.PointInPolygonAndUpdate(ctx, pip_inputs, results_cb, update_cb, body)

	// Apply updates to body here

}
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-placetypes
* https://github.com/whosonfirst/go-whosonfirst-flags