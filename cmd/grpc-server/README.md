# grpc-server

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
    	A valid whosonfirst/go-reader.Reader URI. Available options are: [fs:// null:// pmtiles:// repo:// sqlite:// stdin://]. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.
  -spatial-database-uri string
    	A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: [pmtiles:// rtree:// sqlite://] (default "rtree://")
```

## Example

```
$> bin/grpc-server \
	-spatial-database-uri 'pmtiles://?tiles=file:///usr/local/data&database=wof'
2024/07/19 11:13:44 Listening on localhost:8082	
```

And then in another terminal:

```
$> ./bin/grpc-client -latitude 37.621131 -longitude -122.384292 | jq '.places[]["name"]'
"San Francisco International Airport"
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial-grpc