# grpc-client

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

## Example

```
$> ./bin/grpc-client -latitude 37.621131 -longitude -122.384292 | jq '.places[]["name"]'
"San Francisco International Airport"
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc
