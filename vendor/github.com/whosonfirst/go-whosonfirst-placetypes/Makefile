cli:
	go build -mod vendor -ldflags="-s -w" -o bin/wof-placetype-ancestors cmd/wof-placetype-ancestors/main.go
	go build -mod vendor -ldflags="-s -w" -o bin/wof-placetype-children cmd/wof-placetype-children/main.go
	go build -mod vendor -ldflags="-s -w" -o bin/wof-placetype-descendants cmd/wof-placetype-descendants/main.go
	go build -mod vendor -ldflags="-s -w" -o bin/wof-placetypes cmd/wof-placetypes/main.go
	go build -mod vendor -ldflags="-s -w" -o bin/wof-valid-placetype cmd/wof-valid-placetype/main.go

spec:
	go run cmd/wof-compile-spec/main.go > placetypes.json.tmp
	mv placetypes.json.tmp placetypes.json
	go run cmd/wof-render-spec/main.go -path docs/images/placetypes.png

spec-old:
	curl -o placetypes.json https://raw.githubusercontent.com/whosonfirst/whosonfirst-placetypes/master/data/placetypes-spec-latest.json
