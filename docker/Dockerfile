FROM golang:1.21-alpine as builder

RUN apk update && apk upgrade \
    && apk add --no-cache git gcc libc-dev \
    #
    && git clone https://github.com/whosonfirst/go-whosonfirst-tippecanoe.git go-whosonfirst-tippecanoe \
    && cd go-whosonfirst-tippecanoe \
    && go build -mod vendor -o /usr/local/bin/wof-tippecanoe-features cmd/features/main.go \
    && cd - \
    && rm -rf /go-whosonfirst-tippecanoe \
    #
    && git clone https://github.com/aaronland/gocloud-blob.git gocloud-blob \
    && cd gocloud-blob \
    && go build -mod vendor -o /usr/local/bin/copy-uri cmd/copy-uri/main.go \
    && cd - \
    && rm -rf /gocloud-blob
    
FROM alpine

RUN mkdir /usr/local/data

VOLUME /usr/local/data

RUN apk update && apk upgrade \
    && apk add --no-cache git g++ make libgcc libstdc++ sqlite-libs sqlite-dev zlib-dev bash \
    #
    && git clone https://github.com/felt/tippecanoe.git tippecanoe \
    && cd tippecanoe \
    && make -j \
    && make install \
    && cd - \
    && rm -rf /tippecanoe \
    && apk del git g++ make sqlite-dev

COPY --from=builder /usr/local/bin/wof-tippecanoe-features /usr/local/bin/wof-tippecanoe-features
COPY --from=builder /usr/local/bin/copy-uri /usr/local/bin/copy-uri

COPY build.sh /usr/local/bin/build.sh