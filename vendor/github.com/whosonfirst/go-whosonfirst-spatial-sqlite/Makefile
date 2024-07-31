GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/http-server cmd/http-server/main.go
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/grpc-server cmd/grpc-server/main.go
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/grpc-client cmd/grpc-client/main.go
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/update-hierarchies cmd/update-hierarchies/main.go
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/pip cmd/pip/main.go

# For example:
# make server DSN=modernc:///PATH/TO/SQLITE.db

httpd:
	go run cmd/http-server/main.go \
		-enable-www \
		-spatial-database-uri "sqlite://?dsn=$(DSN)"

grpcd:
	go run cmd/grpc-server/main.go \
		'sqlite://?dsn=$(DSN)'
