GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/brooklynt cmd/brooklynt/main.go

