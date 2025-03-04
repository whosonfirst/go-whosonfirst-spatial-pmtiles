GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/pip cmd/pip/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/mbr cmd/mbr/main.go
