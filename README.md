# Wallet Balance Split Service

A high-performance wallet service for managing transactions and account balances with async processing, connection pooling, and idempotency support.

## ✨ Features

- **High Throughput**: Handles 10,000+ transactions per hour
- **Async Processing**: Withdrawal requests processed asynchronously
- **Connection Pooling**: Optimized database connection management
- **Worker Pool**: Fixed-size pool for concurrent task processing
- **Idempotency**: Duplicate requests handled gracefully
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

- Go 1.25
- PostgreSQL 15 (via Docker Compose)
- Docker & Docker Compose

## 🚀 Quick Start

```bash
# Initialize and Build and Run the services
make init

# Run the services
make run
```

The service will be available at `http://localhost:8080`.

## 📁 Project Structure

```
wbs/
├── cmd/
│   └── main.go              # HTTP API server entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── handlers/            # HTTP request handlers + test
│   ├── repository/          # Database layer (queries) + test
│   ├── worker/              # Worker pool for async tasks
│   ├── tasks/               # Async task definitions
│   ├── models/              # Data models
│   └── metrics/             # Prometheus metrics
├── docker-compose.yml       # Services (Postgres, Prometheus, Grafana, Swagger)
├── migrations/              # SQL migration files
├── Makefile                 # Build & lifecycle management
└── README.md                # This file
```

## 🏗️ Architecture Details

### Request Flow

1. **HTTP Handler** (`/cmd/main.go`): Receives requests, validates input, returns responses
2. **Repository Layer** (`/internal/repository`): Executes database queries with connection pooling
3. **Worker Pool** (`/internal/worker`): Async task queue for long-running operations (withdrawals)
4. **Database** (PostgreSQL): Persistent storage with indexed queries

### Key Components

- **Connection Pooling**: Uses `database/sql` with configurable pool size (default: 100 max connections)
- **Worker Pool**: Fixed-size goroutine pool (50 workers) for concurrent withdrawal processing
- **Idempotency**: `idempotency_key` prevents duplicate processing of same request
- **Metrics**: Prometheus integration tracks requests, errors, and worker queue stats

### Concurrency Model

- **Charge** (Synchronous): Immediate database update, instant response
- **Withdraw** (Asynchronous): HTTP returns immediately, worker processes in background
- **Safe**: Uses transactions and idempotency keys for data consistency

## 📖 API Documentation

You can explore the API in two ways:

### 1. Swagger UI (Interactive)
Visit `http://localhost:8282/` in your browser to see and test all endpoints interactively.

### 2. REST Endpoints

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Get User Balance
```bash
curl http://localhost:8080/balance/123
```

#### Charge Account (Synchronous)
```bash
curl -X POST http://localhost:8080/charge \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "amount": 5000,
    "idempotency_key": "charge-001"
  }'
```

#### Withdraw (Asynchronous)
```bash
curl -X POST http://localhost:8080/withdraw \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "amount": 1000,
    "idempotency_key": "withdraw-001"
  }'
```

#### Get Transactions
```bash
curl http://localhost:8080/transactions/123
```

## 🛠️ Useful Commands

- **Refresh database**: `make refresh_db`
- **Stop services**: `make stop`
- **Full cleanup**: `make reset`
- **Test**: `make test`
- **Test Coverage**: `make test-coverage`

## 🗄️ Database

Default credentials (docker-compose):
- User: `postgres`
- Password: `password`
- Database: `wpdb`
- Port: `5432`

## 📝 Notes

- Compiled binaries are ignored via `.gitignore` and should never be committed
- Source code only in version control
