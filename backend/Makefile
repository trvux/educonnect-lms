include .env
export

MIGRATE := $(shell go env GOPATH)/bin/migrate

AIR := $(shell go env GOPATH)/bin/air

.PHONY: up down migrate-install migrate-up migrate-down migrate-create test run build dev

up: ## Start Postgres in Docker
	docker compose up -d
	@echo "Waiting for Postgres to be healthy..."
	@until docker compose ps postgres | grep -q "healthy"; do sleep 1; done

down: ## Stop Docker services
	docker compose down

migrate-install: ## Install the golang-migrate CLI (postgres + file sources only)
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up: ## Apply all pending migrations
	$(MIGRATE) -database "$(DATABASE_URL)" -path migrations up

migrate-down: ## Roll back the last migration
	$(MIGRATE) -database "$(DATABASE_URL)" -path migrations down 1

migrate-create: ## Usage: make migrate-create name=add_assignments_table
	$(MIGRATE) create -ext sql -dir migrations -seq $(name)

test: ## Run unit + integration tests (needs `make up` + `make migrate-up` first for repository tests)
	go test ./... -v

run: ## Run the API locally (1 lần, không tự reload)
	go run ./cmd/api

dev: ## Chạy API với air — tự build lại & restart mỗi khi sửa file .go
	$(AIR) -c .air.toml

build:
	go build -o bin/api ./cmd/api
