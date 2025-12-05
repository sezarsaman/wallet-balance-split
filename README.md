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

## 📚 Learn More

For detailed architecture, code explanations, and interview preparation guides, see `INTERVIEW_PREP.html`.

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
- See `INTERVIEW_PREP.html` for detailed concepts and interview tips
