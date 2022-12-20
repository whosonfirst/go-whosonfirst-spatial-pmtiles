vuln:
	govulncheck ./...

cli:
	go build -mod vendor -o bin/query cmd/query/main.go
	go build -mod vendor -o bin/pmtile cmd/pmtile/main.go
