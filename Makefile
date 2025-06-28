GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

CWD=$(shell pwd)

DATABASE=pmtiles://?tiles=file://$(CWD)/fixtures&database=sf&zoom=13&enable-cache=true&layer=whosonfirst
INITIAL_VIEW=-122.384292,37.621131,13

vuln:
	govulncheck ./...

test:
	git lfs fetch fixtures/sf.pmtiles
	git lfs checkout fixtures/sf.pmtiles
	go test -v ./...

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/pmtile cmd/pmtile/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/http-server cmd/http-server/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/grpc-server cmd/grpc-server/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/grpc-client cmd/grpc-client/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/update-hierarchies cmd/update-hierarchies/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/pip cmd/pip/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/intersects cmd/intersects/main.go

http-server:
	go run -mod $(GOMOD) cmd/http-server/main.go \
		-enable-www \
		-initial-view '$(INITIAL_VIEW)' \
		-server-uri http://localhost:8080 \
		-spatial-database-uri '$(DATABASE)' \
		-properties-reader-uri '{spatial-database-uri}'

grpc-server:
	go run -mod $(GOMOD) cmd/grpc-server/main.go \
		-spatial-database-uri '$(DATABASE)'

lambda:
	@make lambda-server

lambda-server:
	if test -f bootstrap; then rm -f bootstrap; fi
	if test -f server.zip; then rm -f server.zip; fi
	GOARCH=arm64 GOOS=linux go build -mod $(GOMOD) -ldflags="-s -w" -tags lambda.norpc -o bootstrap cmd/http-server/main.go
	zip server.zip bootstrap
	rm -f bootstrap
