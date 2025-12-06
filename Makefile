include .env.example
export $(shell sed 's/=.*//' .env.example)

.PHONY: init run stop refresh-db reset test test-coverage set-env

init:
	$(MAKE) set-env
	$(MAKE) build
	$(MAKE) run-db
	sleep 5
	$(MAKE) refresh-db
	sleep 5
	$(MAKE) run-app
	$(MAKE) run-swagger
	$(MAKE) run-prometheus
	$(MAKE) run-grafana
	

build:	
	docker compose build

run-db:
	docker compose up -d postgres

run-swagger:
	docker compose up -d swagger-ui

run-prometheus:
	docker compose up -d prometheus

run-grafana:
	docker compose up -d grafana

run-app:
	docker compose up -d app

run:
	docker compose up -d

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



