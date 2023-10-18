GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

vuln:
	govulncheck ./...

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/query cmd/query/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/pmtile cmd/pmtile/main.go
