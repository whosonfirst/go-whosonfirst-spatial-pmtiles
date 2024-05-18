# go-sfomuseum-mapshaper

Go package for interacting with the mapserver-cli tool.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-sfomuseum-mapshaper.svg)](https://pkg.go.dev/github.com/sfomuseum/go-sfomuseum-mapshaper)

Documentation is incomplete.

## Tools

### server

A simple HTTP server to expose the mapserver-cli tool. Currently, only the '-points inner' functionality is exposed.

```
$> ./bin/server -h
A simple HTTP server to expose the mapserver-cli tool. Currently, only the '-points inner' functionality is exposed.
Usage:
	 ./bin/server [options]

Valid options are:
  -allowed-origins string
    	A comma-separated list of hosts to allow CORS requests from.
  -enable-cors
    	Enable support for CORS headers
  -mapshaper-path string
    	The path to your mapshaper binary. (default "/usr/local/bin/mapshaper")
  -server-uri string
    	A valid aaronland/go-http-server URI. (default "http://localhost:8080")
  -uploads-max-bytes int
    	The maximum allowed size (in bytes) for uploads. (default 1048576)
```

## Docker

```
$> docker build -t mapshaper-server .

$> docker run -it -p 8080:8080 -e MAPSHAPER_SERVER_URI=http://0.0.0.0:8080 mapshaper-server /usr/local/bin/mapshaper-server

$> curl -s http://localhost:8080/api/innerpoint \
	-d @fixtures/1745882083.geojson \

| jq '.features[].geometry'

{
  "type": "Point",
  "coordinates": [
    -122.38875600604932,
    37.61459515528007
  ]
}
```

## AWS

### Lambda

It is possible to run the `mapshaper-server` tool as an AWS Lambda Function URL.

1. Start by create a container image by running the `docker build -t mapshaper-server .` command.
2. Upload the container to an AWS ECS repository.
3. Create a new Lambda function and configure to use the container image you've just uploaded to ECS.
4. Update the "Image Configuration" of your Lambda function and assign the following container override: `/usr/local/bin/mapshaper-server, -server-uri, functionurl://`
5. Configure your function to a "Function URL". The details of whether your function URL requires authentication or not are left to you to decide.

That's it. You can test your function URL like this (where `{FUNCTION_URL_ID}` and `{AWS_REGION}` should be replace with the relevant values in your specific function URL):

```
$> curl -s https://{FUNCTION_URL_ID}.lambda-url.{AWS_REGION}.on.aws/api/innerpoint \
	-d @fixtures/1745882083.geojson \
	
| jq '.features[].geometry'
	
{
  "type": "Point",
  "coordinates": [
    -122.38875600604932,
    37.61459515528007
  ]
}
```

## See also

* https://github.com/mbloch/mapshaper