GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

vuln:
	govulncheck ./...

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/pmtile cmd/pmtile/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/server cmd/server/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/update-hierarchies cmd/update-hierarchies/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/pip cmd/pip/main.go

server:
	go run -mod $(GOMOD) cmd/server/main.go \
		-enable-www \
		-server-uri http://localhost:8080 \
		-spatial-database-uri '$(DATABASE)' \
		-properties-reader-uri '{spatial-database-uri}'

lambda:
	@make lambda-server

lambda-server:
	if test -f bootstrap; then rm -f bootstrap; fi
	if test -f server.zip; then rm -f server.zip; fi
	GOARCH=arm64 GOOS=linux go build -mod $(GOMOD) -ldflags="-s -w" -tags lambda.norpc -o bootstrap cmd/server/main.go
	zip server.zip bootstrap
	rm -f bootstrap
