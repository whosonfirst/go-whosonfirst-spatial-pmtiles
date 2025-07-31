#!/bin/sh

SOURCES=""		# for example: -s 'sfomuseum-data://?prefix=sfomuseum-data-whosonfirst'
TARGET=""		# for example: -t s3blob://bucket?region=us-east-1&credentials=iam:

WRITE_FEATURES=""
HELP=""

NAME="whosonfirst"	# for example -n whosonfirst"
ITERATOR="githuborg:///tmp"
ZOOM="12"

LAYER_NAME=""	# tippecanoe layer name
PROPERTIES=""	# for example: -p 'wof:hierarchy wof:concordances'

FORGIVING=""

while getopts "i:l:n:p:s:t:z:fFh" opt; do
    case "$opt" in
	f)
	    WRITE_FEATURES=1
	    ;;
	F)
	    FORGIVING=1
	    ;;
	h)
	    HELP=1
	    ;;
	i)
	    ITERATOR=$OPTARG
	    ;;
	l)
	    LAYER_NAME=$OPTARG
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

if [ "${FORGIVING}" != "" ]
then
    FEATURES_ARGS="${FEATURES_ARGS} -forgiving"
fi
			 
for PROP in ${PROPERTIES}
do
    FEATURES_ARGS="${FEATURES_ARGS} -spr-append-property ${PROP}"
done

for SRC in ${SOURCES}
do
    FEATURES_ARGS="${FEATURES_ARGS} ${SRC}"
done

TIPPECANOE_ARGS="-Q -P -z ${ZOOM} -pf -pk"

if [ "${LAYER_NAME}" != "" ]
then
	TIPPECANOE_ARGS="${TIPPECANOE_ARGS} -n ${LAYER_NAME}"
fi

TIPPECANOE_ARGS="${TIPPECANOE_ARGS} -t /usr/local/data -o /usr/local/data/${NAME}.pmtiles"

echo "wof-tippecanoe-features ${FEATURES_ARGS} | tippecanoe ${TIPPECANOE_ARGS}"

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
    
    tippecanoe ${TIPPECANOE_ARGS} /usr/local/data/features.jsonl

else 
    wof-tippecanoe-features ${FEATURES_ARGS} | tippecanoe ${TIPPECANOE_ARGS}
fi

if [ $? -ne 0 ]
then

    echo "Failed to create PMTiles database"
    exit 1
fi

if [ "${TARGET}" != "" ]
then
    
    copy-uri -source-uri file:///usr/local/data/${NAME}.pmtiles -target-uri ${TARGET}

    if [ -f /usr/local/data/features.jsonl ]
    then
	copy-uri -source-uri file:///usr/local/data/features.jsonl -target-uri ${TARGET}
    fi
fi

exit 0
