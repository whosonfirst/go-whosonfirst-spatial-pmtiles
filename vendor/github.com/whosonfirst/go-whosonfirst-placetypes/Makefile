cli:
	go build -mod vendor -o bin/wof-placetype-ancestors cmd/wof-placetype-ancestors/main.go
	go build -mod vendor -o bin/wof-placetype-children cmd/wof-placetype-children/main.go
	go build -mod vendor -o bin/wof-placetype-descendants cmd/wof-placetype-descendants/main.go
	go build -mod vendor -o bin/wof-valid-placetype cmd/wof-valid-placetype/main.go

spec:
	curl -o placetypes.json https://raw.githubusercontent.com/whosonfirst/whosonfirst-placetypes/master/data/placetypes-spec-latest.json
