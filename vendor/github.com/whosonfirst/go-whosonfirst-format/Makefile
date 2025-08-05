GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

.PHONY: build
build:
	@make cli
	@make wasmjs

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-format ./cmd/wof-format/main.go

wasmjs:
	GOOS=js GOARCH=wasm \
		go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -tags wasmjs \
		-o www/wasm/wof_format.wasm \
		cmd/wof-format-wasm/main.go

.PHONY: test
test:
	go test -v ./...

expected:
	@make cli
	@make expected-single NAME=collapsed_arrays
	@make expected-single NAME=fully_formatted
	@make expected-single NAME=one_line_hierarchy
	@make expected-single NAME=uglify_geometry

expected-single:
	bin/wof-format fixtures/$(NAME).geojson > fixtures/$(NAME).expected.geojson

