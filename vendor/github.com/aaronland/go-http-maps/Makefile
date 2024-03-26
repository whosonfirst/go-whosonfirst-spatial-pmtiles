GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
CWD=$(shell pwd)

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/server cmd/server/main.go

debug-tangram:
	go run -mod $(GOMOD) cmd/server/main.go \
	-map-provider tangram \
	-nextzen-apikey $(APIKEY) \
	-leaflet-enable-draw \
	-javascript-at-eof \
	-rollup-assets

debug-tilepack:
	go run -mod $(GOMOD) cmd/server/main.go \
	-map-provider tangram \
	-tilezen-enable-tilepack \
	-tilezen-tilepack-path /usr/local/data/sf.db \
	-leaflet-enable-draw

debug-protomaps:
	go run -mod $(GOMOD) cmd/server/main.go \
	-map-provider protomaps \
	-protomaps-serve-tiles \
	-protomaps-bucket-uri file://$(CWD)/fixtures \
	-protomaps-database sfo \
	-protomaps-paint-rules-uri file://$(CWD)/fixtures/protomaps.rules.paint.js \
	-protomaps-label-rules-uri file://$(CWD)/fixtures/protomaps.rules.label.js \
	-leaflet-enable-draw \
	-javascript-at-eof \
	-rollup-assets

debug-leaflet:
	go run -mod $(GOMOD) cmd/server/main.go \
	-map-provider leaflet \
	-leaflet-tile-url https://tile.openstreetmap.org/{z}/{x}/{y}.png \
	-leaflet-enable-draw \
	-javascript-at-eof \
	-rollup-assets

debug-null:
	go run -mod $(GOMOD) cmd/server/main.go \
	-map-provider null \
	-javascript-at-eof \
	-rollup-assets
