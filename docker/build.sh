#!/bin/sh

SOURCE=""	# for example: 'sfomuseum-data://?prefix=sfomuseum-data-whosonfirst'
NAME=""		# for example whosonfirst"
TARGET=""	# for example: s3blob://bucket?region=us-east-1&credentials=iam:

ITERATOR="org:///tmp"
ZOOM="12"

while getopts "i:n:s:t:z:" opt; do
    case "$opt" in
	i)
	    ITERATOR=$OPTARG
	    ;;
	n)
	    NAME=$OPTARG
	    ;;
	s )
	    SOURCE=$OPTARG
	    ;;
	t )
	    TARGET=$OPTARG
	    ;;
	z )
	    ZOOM=$OPTARG
	    ;;
	: )
	    echo "WHAT"
	    ;;
    esac
done

echo "Import ${SOURCE} FROM ${ITERATOR} as ${NAME} and copy to ${TARGET}"

wof-tippecanoe-features \
    -as-spr \
    -require-polygons \
    -spr-append-property wof:hierarchy \
    -writer-uri 'constant://?val=jsonl://?writer=stdout://' \
    -iterator-uri ${ITERATOR} ${SOURCE} \
    | tippecanoe -P -z ${ZOOM} -pf -pk -o /usr/local/data/${NAME}.mbtiles

pmtiles convert /usr/local/data/${NAME}.mbtiles /usr/local/data/${NAME}.pmtiles

rm -f /usr/local/data/${NAME}.mbtiles

if [ "${TARGET}" != "" ]
then
	copy-uri -source-uri file:///usr/local/data/${NAME}.pmtiles -target-uri ${TARGET}
fi

