GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/server cmd/server/main.go
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/update-hierarchies cmd/update-hierarchies/main.go
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/pip cmd/pip/main.go

# For example:
# make server DSN=modernc:///PATH/TO/SQLITE.db

server:
	go run cmd/server/main.go \
		-enable-www \
		-spatial-database-uri "sqlite://?dsn=$(DSN)"
