# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/web: run the cmd/web app
.PHONY: run/web
run/web:
	go run ./cmd/web -db-dsn=${CALENDAR_APP_DB_DSN}

## lint: run the linter
.PHONY: lint
lint:
	golangci-lint run ./...

## db/psql: connect to db using psql
.PHONY: db/psql
db/psql:
	psql ${CALENDAR_APP_DB_DSN}

## db/migrations/new name=$1: create a new db migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up db migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${CALENDAR_APP_DB_DSN} up

## db/migrations/down: apply all down db migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migrations...'
	migrate -path ./migrations -database ${CALENDAR_APP_DB_DSN} down



# ==================================================================================== #
# BUILD
# ==================================================================================== #
## build/api: build the cmd/api application
.PHONY: build/web
build/web:
	@echo 'Building cmd/web...'
	go build -ldflags='-s' -o=./bin/web ./cmd/web
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/web ./cmd/web
