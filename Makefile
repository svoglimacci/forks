include .envrc

## run/api: run the cmd/api application
.PHONY: run/api
run/api:	
		go run ./cmd/api -db-dsn=${FORKS_DB_DSN}

## migrations
.PHONY: migrations
migrations:
	migrate -path=./migrations -database=${FORKS_DB_DSN} up
