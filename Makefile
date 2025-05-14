include .envrc

## run/api: run the cmd/api application
.PHONY: run/api
run/api:	
		@go run ./cmd/api -db-dsn=${DB_DSN} -smtp-host=${SMTP_HOST} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD}

## migrations
.PHONY: migrations
migrations:
	@migrate -path=./migrations -database=${DB_DSN} up


