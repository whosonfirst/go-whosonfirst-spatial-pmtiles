dynamo-local:
	docker run --rm -it -p 8000:8000 amazon/dynamodb-local

dynamo-tables-local:
	go run -mod vendor cmd/create-dynamodb-tables/main.go \
		-refresh \
		-table-prefix '$(TABLE_PREFIX)' \
		-dynamodb-client-uri 'awsdynamodb://?local=true'
