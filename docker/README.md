# docker

## Build

```
$> docker build -t whosonfirst-spatial-pmtiles .
```

## Example

```
$> docker run whosonfirst-spatial-pmtiles /usr/local/bin/build.sh -n sfomuseum -s 'sfomuseum-data://?prefix=sfomuseum-data-maps' -t file:///tmp
```
