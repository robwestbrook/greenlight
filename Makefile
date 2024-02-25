# Import .env file and include all .env variables
# if the .env file exists
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

#======================================================#
# HELPERS
#
#======================================================#

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

#======================================================#
# DEVELOPMENT
#======================================================#

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

#======================================================#
# QUALITY CONTROL
#
#======================================================#

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...

## test: run all application tests
.PHONY: test
test:
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

#======================================================#
# BUILD
#
#======================================================#

# Build Variables:
#	1.	current_time: Use the unix date command to generate the current time and store it in a current_time variable.
# 2.	git_description: make app version number from git commit
#	3.	linker_flags: Use the -s and -X flags
current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/api: build the cmd/api application
# Using the linker flag "-ldflags="-s"" instructs the
# Go linker to strip the DWARF debugging info and 
# symbol table from the binary that is built. This
# reduces the binary size by around 25%.
# The first "go build" builds a binary for the local
# machine.
# The second "go build" builds a binary for deploying
# to a Ubuntu Linux server.
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api