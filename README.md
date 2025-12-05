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
# Initialize services and build
make init

# Run the service
make run
```

The service will be available at `http://localhost:8080`.

## 📖 API Documentation

You can explore the API in two ways:

### 1. Swagger UI (Interactive)
Visit `http://localhost:8080/swagger` in your browser to see and test all endpoints interactively.

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
- **View logs**: `make logs`
- **Show status**: `make status`
- **Full cleanup**: `make clean_all`
- **Help**: `make help`

## 🗄️ Database

Default credentials (docker-compose):
- User: `postgres`
- Password: `password`
- Database: `wallet`
- Port: `5433`

## 📝 Notes

- Compiled binaries are ignored via `.gitignore` and should never be committed
- Source code only in version control
