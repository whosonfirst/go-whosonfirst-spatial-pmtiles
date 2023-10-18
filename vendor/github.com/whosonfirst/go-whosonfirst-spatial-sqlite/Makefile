GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

cli:
	go build -ldflags="-s -w" -mod $(GOMOD) -o bin/query cmd/query/main.go
