GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

TAGS=mattn
ENGINE=sqlite3
DSN=fixtures/sfomuseum-architecture.db
INITIAL_VIEW=-122.384292,37.621131,13

cli:
	go build -tags $(TAGS) -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/http-server cmd/http-server/main.go
	go build -tags $(TAGS) -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/grpc-server cmd/grpc-server/main.go
	go build -tags $(TAGS) -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/grpc-client cmd/grpc-client/main.go
	go build -tags $(TAGS) -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/update-hierarchies cmd/update-hierarchies/main.go
	go build -tags $(TAGS) -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/pip cmd/pip/main.go
	go build -tags $(TAGS) -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/intersects cmd/intersects/main.go

http-server:
	go run -tags $(TAGS) -mod $(GOMOD) \
		cmd/http-server/main.go \
		-enable-www \
		-initial-view '$(INITIAL_VIEW)' \
		-spatial-database-uri "sqlite://$(ENGINE)?dsn=$(DSN)"

grpcd:
	go run -tags $(TAGS) -mod $(GOMOD) \
		cmd/grpc-server/main.go \
		'sqlite://$(ENGINE)?dsn=$(DSN)'
