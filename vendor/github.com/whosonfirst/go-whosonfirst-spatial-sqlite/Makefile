GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/query cmd/query/main.go
	go build -ldflags="$(LDFLAGS)" -mod $(GOMOD) -o bin/server cmd/server/main.go
