#!/bin/sh

SOURCES=""		# for example: -s 'sfomuseum-data://?prefix=sfomuseum-data-whosonfirst'
TARGET=""		# for example: -t s3blob://bucket?region=us-east-1&credentials=iam:

NAME="whosonfirst"	# for example -n whosonfirst"
ITERATOR="org:///tmp"
ZOOM="12"

PROPERTIES=""	# for example: -p 'wof:hierarchy wof:concordances'

while getopts "i:n:p:s:t:z:" opt; do
    case "$opt" in
	i)
	    ITERATOR=$OPTARG
	    ;;
	n)
	    NAME=$OPTARG
	    ;;
	p)
	    PROPERTIES=$OPTARG
	    ;;
	s )
	    SOURCES=$OPTARG
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

FEATURES_ARGS="-as-spr -require-polygons -writer-uri constant://?val=jsonl://?writer=stdout:// -iterator-uri ${ITERATOR}"

for PROP in ${PROPERTIES}
do
    FEATURES_ARGS="${FEATURES_ARGS} -spr-append-property ${PROP}"
done

for SRC in ${SOURCES}
do
    FEATURES_ARGS="${FEATURES_ARGS} ${SRC}"
done

echo "wof-tippecanoe-features ${FEATURES_ARGS} | tippecanoe -P -z ${ZOOM} -pf -pk -o /usr/local/data/${NAME}.mbtiles"

wof-tippecanoe-features ${FEATURES_ARGS} | tippecanoe -P -z ${ZOOM} -pf -pk -o /usr/local/data/${NAME}.mbtiles

pmtiles convert /usr/local/data/${NAME}.mbtiles /usr/local/data/${NAME}.pmtiles

rm -f /usr/local/data/${NAME}.mbtiles

if [ "${TARGET}" != "" ]
then
	copy-uri -source-uri file:///usr/local/data/${NAME}.pmtiles -target-uri ${TARGET}
fi

