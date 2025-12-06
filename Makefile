include .env.example
export $(shell sed 's/=.*//' .env.example)

.PHONY: init run stop refresh-db reset test test-coverage set-env

init:
	$(MAKE) set-env
	$(MAKE) build
	$(MAKE) run
	sleep 5
	$(MAKE) refresh-db
	$(MAKE) build-swagger


build:	
	docker compose build

run:
	docker compose up -d  --remove-orphans

stop:
	docker compose down

refresh-db:
	docker compose exec -T postgres psql -U postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	docker compose exec -T postgres psql -U postgres -c "CREATE DATABASE $(DB_NAME);"
	docker compose exec -T postgres psql -U postgres $(DB_NAME) < db/migrations/001_init.sql
	docker compose exec -T postgres psql -U postgres $(DB_NAME) < db/seed/001_transaction_seeder.sql

reset:
	docker compose down -v

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

set-env:
	@if [ ! -f .env ]; then cp .env.example .env; else echo ".env already exists"; fi

build-swagger:
	docker run --rm -v $(PWD)/docs:/local openapitools/openapi-generator-cli generate \
	-i /local/swagger.yaml \
	-g openapi \
	-o /local


