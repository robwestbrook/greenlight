# Import .env file and include all .env variables
# if the .env file exists
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

## help: print this help message
## .PHONY helps resolve any file naming conflicts
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# Create a confirm target.
.PHONY: confirm
confirm:
	@echo -n 'Are you sure [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/api: run cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## db/migrations/new name=$1; create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path=./migrations -database=${DB_DSN} -verbose up