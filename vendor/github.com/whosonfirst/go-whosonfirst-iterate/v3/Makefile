GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/count cmd/count/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/emit cmd/emit/main.go
