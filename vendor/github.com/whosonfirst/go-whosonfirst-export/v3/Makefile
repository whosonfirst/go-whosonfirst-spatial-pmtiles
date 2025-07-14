GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

.PHONY: tools
tools:
	@make cli

.PHONY: test
test:
	go test -v ./

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/wof-export cmd/wof-export/main.go
