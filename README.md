# Wallet Balance Split Service

یک سرویس high-performance برای مدیریت تراکنش‌ها و balance حساب‌های کاربری.

## ✨ Features

- **High Throughput**: مدیریت 10,000+ تراکنش در ساعت
- **Async Processing**: درخواست‌های برداشت async پردازش می‌شوند
- **Connection Pooling**: بهینه‌سازی شده برای database connections
- **Worker Pool**: Fixed-size pool برای concurrent task processing
- **Idempotency**: درخواست‌های تکراری به‌طور معقول مدیریت می‌شوند
- **Graceful Shutdown**: Proper cleanup on termination

## 🏗️ Architecture

```
HTTP Requests
    ↓
API Routes (Chi Router)
    ├─ /charge      (Synchronous)
    ├─ /withdraw    (Async)
    ├─ /balance     (Synchronous)
    ├─ /transactions (Synchronous)
    └─ /health      (Status)
    ↓
Repository Layer
    ├─ Database (PostgreSQL)
    └─ Connection Pool (100 max)
    ↓
Worker Pool (50 workers)
    └─ Bank Withdrawal Tasks
```

## 📋 Prerequisites

- Go 1.25 (module set to go 1.25 in `go.mod`)
- PostgreSQL (containerized via `docker-compose` in repo uses Postgres 15)
- Docker & Docker Compose (recommended for local dev)

## 🚀 Quick Start

### 1. Clone & Setup
```bash
cd /home/saman/Projects/wbs
go mod download
```

### 2. Database & services (recommended)
The repo includes a `docker-compose.yml` that starts PostgreSQL, Prometheus, Grafana and a Swagger UI.

Start services with Make (recommended):
```bash
make db-up
```

Postgres will be available on `localhost:5433` (container maps 5432->5433). Default DB credentials used by the project are:

- user: `postgres`
- password: `password`
- database: `wallet`

If you prefer to run Postgres locally without Docker, create a database named `wallet` and set `TEST_DATABASE_URL`/`.env` accordingly.

### 3. Run Service

You can run the service directly or via the Makefile which wires up DB services automatically.

Run using Make (recommended):
```bash
make run
```

Or run directly (ensure DB is running and `.env` is configured):
```bash
go run ./cmd/main.go
```

## 📊 API Documentation

### 1. Charge (شارژ کردن)
```bash
curl -X POST http://localhost:8080/charge \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    # Wallet Balance Split Service

    This repository contains a small wallet service used for interview tasks. It's intentionally self-contained and includes a minimal Makefile to set up and run the project locally.

    ## What this repo contains

    - Service implemented in Go (module `wallet-simulator`).
    - HTTP API (Chi router) with endpoints: `/charge`, `/withdraw`, `/balance`, `/transactions`, `/health`.
    - PostgreSQL migrations and a small seeder.
    - Docker Compose for local development (Postgres, Prometheus, Grafana, Swagger UI).

    ## Quick overview of important concepts

    - Connection pooling: the DB layer uses `database/sql` connection pooling to limit concurrent connections and reuse them.
    - Worker pool: withdrawals are handled asynchronously by a fixed-size worker pool to avoid blocking HTTP handlers on long bank calls.
    - Idempotency: financial endpoints use `idempotency_key` to prevent duplicate processing.

    ## How to setup & run (Make-based)

    Prerequisites:

    - Docker & Docker Compose (recommended)
    - Go 1.25

    Steps:

    1. Start services and build the project:

    ```bash
    make init
    ```

    This will:
    - copy `.env.example` to `.env` if missing
    - start `docker compose up -d`
    - run migrations and seed data
    - build the binary into `bin/wallet`

    2. Run the service:

    ```bash
    make run
    ```

    3. Useful commands:

    - Refresh database (recreate schema + seed): `make refresh_db`
    - Stop services and app: `make stop`
    - Full clean (remove volumes & artifacts): `make clean_all`
    - Tail logs: `make logs`
    - Show docker status: `make status`
    - Help: `make help`

    Default DB connection used by tests and CI:

    ```
    postgres://postgres:password@localhost:5433/wallet?sslmode=disable
    ```

    ## Binaries and git

    Compiled binaries and build artifacts are ignored via `.gitignore` (`/bin/`). You should not commit binaries into Git; only source files and small config assets should be tracked.

    ## CI

    A simple GitHub Actions workflow is included in `.github/workflows/ci.yml` which starts a Postgres service, runs migrations, executes tests and builds the binary.

    ## Tests

    Run tests locally (ensure DB is running via `make db-up` or `make init`):

    ```bash
    make db-up
    go test ./... -v
    ```

    Or use `make run` and open the endpoints.

    ---

    If you need a more detailed overview for interview prep, see the generated `INTERVIEW_PREP.html` in the project root which contains explanations of the key components, design choices, concurrency considerations, and example exercises you may be asked to perform during an interview.
