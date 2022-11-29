#!/bin/sh

SOURCE=""	# for example: 'sfomuseum-data://?prefix=sfomuseum-data-whosonfirst'
NAME=""		# for example whosonfirst"
TARGET=""	# for example: s3blob://bucket?region=us-east-1&credentials=iam:

while getopts "n:s:t:" opt; do
    case "$opt" in
	n)
	    NAME=$OPTARG
	    ;;
	s )
	    SOURCE=$OPTARG
	    ;;
	t )
	    TARGET=$OPTARG
	    ;;
	: )
	    echo "WHAT"
	    ;;
    esac
done

echo "Import ${SOURCE} as ${NAME} and copy to ${TARGET}"

wof-tippecanoe-features -as-spr -require-polygons \
    -writer-uri 'constant://?val=jsonl://?writer=stdout://' \
    -iterator-uri 'org:///tmp' \
    ${SOURCE} \
    | tippecanoe -P -z 12 -pf -pk -o /usr/local/data/${NAME}.mbtiles

pmtiles convert /usr/local/data/${NAME}.mbtiles /usr/local/data/${NAME}.pmtiles

copy-uri -source-uri file:///usr/local/data/${NAME}.pmtiles -target-uri ${TARGET}
