#!/bin/sh

SOURCES=""		# for example: -s 'sfomuseum-data://?prefix=sfomuseum-data-whosonfirst'
TARGET=""		# for example: -t s3blob://bucket?region=us-east-1&credentials=iam:

WRITE_FEATURES=""
HELP=""

NAME="whosonfirst"	# for example -n whosonfirst"
ITERATOR="org:///tmp"
ZOOM="12"

PROPERTIES=""	# for example: -p 'wof:hierarchy wof:concordances'

while getopts "i:n:p:s:t:z:fh" opt; do
    case "$opt" in
	f)
	    WRITE_FEATURES=1
	    ;;
	h)
	    HELP=1
	    ;;
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

if [ "${HELP}" = "1" ]
then
    echo "Print this message"
    exit 0
fi

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

echo "wof-tippecanoe-features ${FEATURES_ARGS} | tippecanoe -P -z ${ZOOM} -pf -pk -o /usr/local/data/${NAME}.pmtiles"

if [ "${WRITE_FEATURES}" = "1" ]
then
    
    wof-tippecanoe-features ${FEATURES_ARGS} > /usr/local/data/features.jsonl
    
    COUNT=`cat /usr/local/data/features.jsonl | wc -l`
    echo "Wrote ${COUNT} features"
    
    if [ $? -ne 0 ]
    then
	echo "Failed to write all features"
	exit 1;
    fi
    
    tippecanoe -P -z ${ZOOM} -pf -pk -o /usr/local/data/${NAME}.pmtiles /usr/local/data/features.jsonl

else 
    wof-tippecanoe-features ${FEATURES_ARGS} | tippecanoe -P -z ${ZOOM} -pf -pk -o /usr/local/data/${NAME}.pmtiles
fi

if [ $? -ne 0 ]
then

    echo "Failed to create PMTiles database"
    exit 1
fi

if [ "${TARGET}" != "" ]
then
    
    copy-uri -source-uri file:///usr/local/data/${NAME}.pmtiles -target-uri ${TARGET}
    # copy-uri -source-uri file:///usr/local/data/${NAME}.mbtiles -target-uri ${TARGET}

    if [ -f /usr/local/data/features.jsonl ]
    then
	copy-uri -source-uri file:///usr/local/data/features.jsonl -target-uri ${TARGET}
    fi
fi

exit 0
