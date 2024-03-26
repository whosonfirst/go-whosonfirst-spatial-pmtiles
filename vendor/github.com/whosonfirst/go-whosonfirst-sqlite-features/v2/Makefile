GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

LDFLAGS=-s -w

tools:
	@make cli

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/query cmd/query/main.go
