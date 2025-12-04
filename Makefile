.PHONY: help setup build run migrate seed refresh clear clean test docker-up docker-down docker-logs stop

# متغیرهای رنگی برای خروجی
COLOR_RESET=\033[0m
COLOR_BLUE=\033[34m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m

help:
	@echo "$(COLOR_BLUE)╔═════════════════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)║   Wallet Service - Makefile Commands                       ║$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)╚═════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)Project Setup:$(COLOR_RESET)"
	@echo "  make setup           → Initialize project (.env, docker, everything)"
	@echo "  make deps            → Download Go dependencies"
	@echo ""
	@echo "$(COLOR_GREEN)Build:$(COLOR_RESET)"
	@echo "  make build           → Build application binary"
	@echo "  make docker-build    → Build Docker image"
	@echo ""
	@echo "$(COLOR_GREEN)Database:$(COLOR_RESET)"
	@echo "  make db-up           → Start PostgreSQL (docker)"
	@echo "  make db-down         → Stop PostgreSQL (docker)"
	@echo "  make migrate         → Run all database migrations"
	@echo "  make migrate-down    → Rollback all migrations (DROP tables)"
	@echo "  make seed            → Insert test data"
	@echo "  make refresh         → Reset DB (migrate down/up + seed)"
	@echo "  make clear-seed      → Remove test data only"
	@echo ""
	@echo "$(COLOR_GREEN)Running:$(COLOR_RESET)"
	@echo "  make run             → Run application (requires DB to be ready)"
	@echo "  make dev             → Run with auto-reload (requires air)"
	@echo ""
	@echo "$(COLOR_GREEN)Testing:$(COLOR_RESET)"
	@echo "  make test            → Run tests"
	@echo "  make test-coverage   → Run tests with coverage report"
	@echo ""
	@echo "$(COLOR_GREEN)Utilities:$(COLOR_RESET)"
	@echo "  make clean           → Remove binaries and build artifacts"
	@echo "  make logs            → Show docker logs"
	@echo "  make status          → Show docker containers status"
	@echo ""

# ==================== Setup ====================
setup: .env db-up deps migrate seed build
	@echo "$(COLOR_GREEN)✅ Project setup completed!$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Next step: make run$(COLOR_RESET)"

.env:
	@echo "$(COLOR_YELLOW)📝 Creating .env from .env.example...$(COLOR_RESET)"
	@if [ ! -f .env ]; then cp .env.example .env; echo "$(COLOR_GREEN)✅ .env created$(COLOR_RESET)"; else echo "$(COLOR_YELLOW)⚠️  .env already exists$(COLOR_RESET)"; fi

deps:
	@echo "$(COLOR_YELLOW)📦 Downloading dependencies...$(COLOR_RESET)"
	@go mod download
	@go mod tidy
	@echo "$(COLOR_GREEN)✅ Dependencies ready$(COLOR_RESET)"

# ==================== Build ====================
build: bin/wallet
	@echo "$(COLOR_GREEN)✅ Build completed$(COLOR_RESET)"

bin/wallet:
	@echo "$(COLOR_YELLOW)🔨 Building application...$(COLOR_RESET)"
	@mkdir -p bin
	@go build -o bin/wallet ./cmd/main.go
	@echo "$(COLOR_GREEN)✅ Binary created: bin/wallet$(COLOR_RESET)"

docker-build:
	@echo "$(COLOR_YELLOW)🐳 Building Docker image...$(COLOR_RESET)"
	@docker build -t wallet-service:latest .
	@echo "$(COLOR_GREEN)✅ Docker image built$(COLOR_RESET)"

# ==================== Database ====================
db-up:
	@echo "$(COLOR_YELLOW)🐳 Starting PostgreSQL...$(COLOR_RESET)"
	@docker compose up -d postgres || docker compose up -d
	@echo "$(COLOR_YELLOW)⏳ Waiting for database to be ready...$(COLOR_RESET)"
	@sleep 3
	@echo "$(COLOR_GREEN)✅ Database service started$(COLOR_RESET)"

db-down:
	@echo "$(COLOR_YELLOW)🛑 Stopping database services...$(COLOR_RESET)"
	@docker compose down
	@echo "$(COLOR_GREEN)✅ Database services stopped$(COLOR_RESET)"

db-logs:
	@docker compose logs -f

db-clean:
	@echo "$(COLOR_YELLOW)🗑️  Removing database volumes...$(COLOR_RESET)"
	@docker compose down -v
	@echo "$(COLOR_GREEN)✅ Database volumes removed$(COLOR_RESET)"

# ==================== Migration & Seeding ====================
migrate:
	@echo "$(COLOR_YELLOW)🔄 Running migrations...$(COLOR_RESET)"
	@go run cmd/cli/main.go migrate

migrate-down:
	@echo "$(COLOR_YELLOW)⚠️  WARNING: Dropping all tables...$(COLOR_RESET)"
	@go run cmd/cli/main.go migrate down

seed:
	@echo "$(COLOR_YELLOW)🌱 Seeding database...$(COLOR_RESET)"
	@go run cmd/cli/main.go seed

refresh: migrate-down migrate seed
	@echo "$(COLOR_GREEN)✅ Database refresh completed$(COLOR_RESET)"

clear-seed:
	@echo "$(COLOR_YELLOW)🗑️  Removing seed data...$(COLOR_RESET)"
	@go run cmd/cli/main.go clear

# ==================== Running ====================

run: bin/wallet
	@echo "$(COLOR_YELLOW)🚀 Starting application (ensures DB is up)...$(COLOR_RESET)"
	@echo "Starting database services (if not running)..."
	@$(MAKE) db-up >/dev/null 2>&1 || true
	@# read DB port from .env fallback to 5433
	@PORT=$$(grep -E '^DB_PORT=' .env 2>/dev/null | cut -d'=' -f2 || echo 5433); \
		echo "Waiting for Postgres on localhost:$$PORT..."; \
		counter=0; \
		until ss -ltn | grep -q ":$$PORT"; do \
			if [ $$counter -ge 60 ]; then echo "Timed out waiting for Postgres"; exit 1; fi; \
			sleep 1; counter=$$((counter+1)); \
		done; \
		echo "Postgres appears to be listening on port $$PORT"; \
		echo "Launching app..."; \
		./bin/wallet

dev:
	@echo "$(COLOR_YELLOW)🔄 Running with auto-reload (requires air)...$(COLOR_RESET)"
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	@air

# ==================== Testing ====================
test:
	@echo "$(COLOR_YELLOW)🧪 Running tests...$(COLOR_RESET)"
	@go test -v ./...

test-coverage:
	@echo "$(COLOR_YELLOW)📊 Running tests with coverage...$(COLOR_RESET)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✅ Coverage report: coverage.html$(COLOR_RESET)"

# ==================== Utilities ====================
clean:
	@echo "$(COLOR_YELLOW)🧹 Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✅ Clean completed$(COLOR_RESET)"

logs:
	@docker compose logs -f

status:
	@docker ps -a

# Stop local background app (if any) and docker containers
stop:
	@echo "$(COLOR_YELLOW)🛑 Stopping local app (if running) and docker containers...$(COLOR_RESET)"
	@if [ -f ./wallet.pid ]; then \
		PID=$$(cat ./wallet.pid); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "Killed process $$PID" || echo "Failed to kill $$PID"; \
		fi; \
		rm -f ./wallet.pid; \
	else \
		echo "No wallet.pid found"; \
	fi
	@echo "$(COLOR_YELLOW)🛑 Stopping docker compose services...$(COLOR_RESET)"
	@docker compose down
	@echo "$(COLOR_GREEN)✅ Stopped local app and containers$(COLOR_RESET)"

fmt:
	@echo "$(COLOR_YELLOW)📐 Formatting code...$(COLOR_RESET)"
	@go fmt ./...
	@echo "$(COLOR_GREEN)✅ Code formatted$(COLOR_RESET)"

lint:
	@echo "$(COLOR_YELLOW)🔍 Running linter...$(COLOR_RESET)"
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...

# ==================== Shortcuts ====================
.DEFAULT_GOAL := help
