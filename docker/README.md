# docker

Dockerfile and build tool to generate Protomaps PMTiles databases from one or more Who's On First repositories.

## Build

```
$> docker build -t whosonfirst-spatial-pmtiles .
```

## Example

### Writing data to a local file using Docker volumes

```
$> docker run whosonfirst-spatial-pmtiles -v ${LOCAL_DIRECTORY}:/usr/local/data \   
	/usr/local/bin/build.sh \
	-n sfomuseum \
	-s 'sfomuseum-data://?prefix=sfomuseum-data-maps'
```

### Writing data to the container and then copying to a target

```
$> docker run whosonfirst-spatial-pmtiles \
	/usr/local/bin/build.sh \
	-n sfomuseum \
	-s 'sfomuseum-data://?prefix=sfomuseum-data-maps' \
	-t s3blob://${BUCKET_NAME}?region={REGION}&credentials=iam:
```

### Writing data from two different organizations

```
$> docker run whosonfirst-spatial-pmtiles \
	/usr/local/bin/build.sh \
	-n whosonfirst_sfo \
	-s 'sfomuseum-data://?prefix=sfomuseum-data-architecture whosonfirst-data://?prefix=whosonfirst-data-admin-' \
	-t s3blob://{BUCKET_NAME}?&region={REGION}&credentials=iam:
```

## build.sh

### Flags

| Flag | Value | Required | Notes |
| --- | --- | --- | --- |
| -n | The name of the final PMTiles database | yes | This should not contain a file extension |
| -s | One or more strings containing source URIs that are compatible with the `-i`(terator) flag, described below | yes | |
| -i | A valid `whosonfirst/go-whosonfirst-iterate/v2` URI | no | Default is `org:///tmp` which will attempt to iterate through records in one or more repositories that are part of a GitHub organization |
| -t | A valid `gocloud.dev/blob.Bucket` URI where the final PMTiles database will be copied | no | If not specified the final PMTiles database will be written to `/usr/local/data` and it is assumed that directory will be mounted on a local volume. The following `gocloud.dev/blob` schemes are supported by default: `file://`, `s3://`, `s3blob://` |
| -z | The zoom level to create tiles at | no | Default is 12 |

The `s3blob://` scheme is distinguished from the `s3://` scheme in its ability to specify [AWS credentials using query parameters](https://github.com/aaronland/gocloud-blob-s3#credentials).

## See also

* https://github.com/felt/tippecanoe
* https://github.com/whosonfirst/go-whosonfirst-tippecanoe
* https://github.com/protomaps/go-pmtiles
* https://github.com/aaronland/gocloud-blob
* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-whosonfirst-iterate-organization