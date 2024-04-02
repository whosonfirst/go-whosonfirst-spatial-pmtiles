GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

vuln:
	govulncheck ./...

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/query cmd/query/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/pmtile cmd/pmtile/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/server cmd/server/main.go
