# 🚀 Wallet Balance Split Service - Performance Optimization

## خلاصه تحسینات برای مقیاسپذیری

این سرویس برای **مدیریت 10,000 تراکنش + 2,000 درخواست برداشت در ساعت** بهینه‌سازی شده است.

---

## 📊 معماری و بهبودی‌ها

### 1. **Connection Pooling** ✅
```go
db.SetMaxOpenConns(100)      // حداکثر 100 اتصال همزمان
db.SetMaxIdleConns(25)       // 25 اتصال idle آماده نگاه‌داری
db.SetConnMaxLifetime(5 * time.Minute) // بازیافت اتصال هر 5 دقیقه
```

**چرا مهم است:**
- هر تراکنش database کنندشی از اتصال موجود استفاده میکند
- جلوگیری از "connection leak" و بیش‌بار connection
- برای 10k req/hour (≈3 req/sec): **100 conn کافی است**

**محاسبه:**
```
Peak Connections = (Peak Requests/sec) × (Avg Query Time) + Buffer
                 = (4 req/sec) × (50ms) + 50
                 = ~70 connections (100 کمی احتیاط دارد)
```

---

### 2. **Worker Pool Pattern** ✅
```go
workerPool := worker.NewWorkerPool(50)
```

**چرا مهم است:**
- **Async Bank Processing**: Withdraw requests به صورت ناهمزمان در background پردازش می‌شوند
- **Goroutine Management**: بجای ایجاد unlimited goroutines، از fixed pool استفاده می‌کنیم
- **Queue Buffering**: هر worker از 2x اندازه pool size buffer دارد

**محاسبه اپتیمال:**
```
Workers = Peak Requests/sec × (Avg Processing Time)
        = 4 req/sec × 10sec (bank timeout)
        = 40 workers (50 کمی احتیاط دارد)
```

**Task Processing:**
```go
task := tasks.NewBankWithdrawalTask(cfg.Repo, userID, amount, idempotencyKey)
if err := cfg.WorkerPool.Submit(task); err != nil {
    // Queue is full - fail gracefully
}
```

---

### 3. **Async/Non-Blocking Withdraw** ✅
**قبل (Blocking):**
```go
// Synchronous - Response blocked during bank retry
for retry := 0; retry < 3; retry++ {
    if bankCall() { success = true; break }
    time.Sleep(time.Second)
}
// Client waits 3+ seconds
```

**بعد (Async):**
```go
// Immediate response
{"status": "pending", "message": "withdrawal request submitted"}

// Bank processing happens in background
// Update database status when complete
```

---

### 4. **Transaction Status Tracking** ✅
```sql
CREATE TABLE transactions (
    ...
    status VARCHAR(20) DEFAULT 'pending',  -- pending, completed, failed
    ...
)
```

**Status Flow:**
```
Charge Request → CREATE (status='completed') → Response
Withdraw Request → CREATE (status='pending') → Response
                 → Worker Processing → UPDATE status='completed'|'failed'
```

---

### 5. **Database Indexing** ✅
```sql
CREATE INDEX idx_user_id ON transactions(user_id);      -- Lookup by user
CREATE INDEX idx_created_at ON transactions(created_at); -- Sorting
CREATE INDEX idx_status ON transactions(status);         -- Filter pending
```

---

## 🔢 Performance Calculations

### Peak Load Handling
```
Scenario: 10,000 transactions + 2,000 withdrawals per hour

Breakdown:
- Average: 10,000/3600 ≈ 2.8 req/sec
- Peak (assuming 2x average): ~5-6 req/sec

Charges (7,000/hour):
- 7,000/3600 ≈ 2 req/sec
- Direct database write

Withdrawals (3,000/hour):
- 3,000/3600 ≈ 0.8 req/sec
- Queue → Worker processing → Async bank call
```

### Database Connections
```
Total Concurrent Transactions:
= (Charge req/sec × query time) + (Withdraw req/sec × query time)
= (2 req/sec × 0.1s) + (0.8 req/sec × 0.1s)
= 0.2 + 0.08 = 0.28 connections (on average)

Peak (2x):
= 0.56 connections

Safety Margin:
MaxOpenConns = 100 ✅ (172x peak usage)
```

### Worker Pool Processing
```
Withdrawal Queue:
= 0.8 req/sec × 10 workers worth of capacity
= 8 concurrent worker tasks

Queue Buffer:
= 50 workers × 2 = 100 task buffer ✅ (125x peak)
```

---

## 🎯 Endpoint Examples

### Charge Request
```bash
curl -X POST http://localhost:8080/charge \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "amount": 5000,
    "idempotency_key": "charge-2024-001",
    "release_at": "2024-01-15T12:00:00Z"
  }'

Response:
{
  "message": "charged",
  "idempotency_key": "charge-2024-001"
}
```

### Withdraw Request
```bash
curl -X POST http://localhost:8080/withdraw \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "amount": 1000,
    "idempotency_key": "withdraw-2024-001"
  }'

Response (Immediate):
{
  "message": "withdrawal request submitted",
  "idempotency_key": "withdraw-2024-001",
  "status": "pending"
}

## Bank processing happens async in background
## Check status via GET /transactions?user_id=123
```

### Get Balance
```bash
curl http://localhost:8080/balance?user_id=123

Response:
{
  "total": 10000,        # جمع تمام تراکنش‌ها
  "withdrawable": 8000   # قابل برداشت (release_at passed)
}
```

### Get Transactions
```bash
curl http://localhost:8080/transactions?user_id=123&page=1&limit=10

Response:
{
  "transactions": [
    {
      "id": 1,
      "user_id": 123,
      "amount": 5000,
      "type": "charge",
      "status": "completed",
      "created_at": "2024-01-10T10:00:00Z",
      "release_at": "2024-01-15T12:00:00Z"
    },
    ...
  ],
  "total": 45,
  "page": 1,
  "limit": 10
}
```

### Health Check
```bash
curl http://localhost:8080/health

Response:
{
  "status": "ok",
  "queue_length": 3  # Pending withdrawal tasks
}
```

---

## 🛠️ Idempotency Handling

**چرا مهم است:** اگر client درخواست را دوبار بفرستد، یکی از دو اتفاق میافتد:

1. **اول‌بار:** Transaction ایجاد می‌شود
2. **دوم‌بار:** Error 409 Conflict (duplicate idempotency key)

```go
// Database prevents duplicate insertions
CREATE TABLE transactions (
    idempotency_key VARCHAR(255) UNIQUE,
    ...
)

// Application checks before insert
err = tx.QueryRow(
    "SELECT 1 FROM transactions WHERE idempotency_key = $1",
    idempotencyKey,
).Scan(&exists)
if err == nil {
    return ErrDuplicateRequest // 409 Conflict
}
```

---

## 📈 Scaling Strategy

### اگر بتوانیم بیش‌تر مقیاس‌بندی کنیم:

| Level | Requirement | Solution |
|-------|-------------|----------|
| **10k req/h** | Current | ✅ Connection Pool + Worker Pool |
| **100k req/h** | 10x | Load Balance + Multiple Instances |
| **1M req/h** | 100x | Database Sharding + Cache Layer |

### Multi-Instance Setup:
```
              ┌─────────────────────┐
              │   Load Balancer     │
              │  (Round Robin)      │
              └──────────┬──────────┘
                ┌────────┼────────┐
         ┌──────▼──────┐ │ ┌──────▼──────┐
         │ Instance 1  │ │ │ Instance 2  │
         │ Workers: 50 │ │ │ Workers: 50 │
         └──────┬──────┘ │ └──────┬──────┘
                └────────┼────────┘
                         │
                   ┌─────▼─────┐
                   │ PostgreSQL │
                   │ (Master)   │
                   └───────────┘
```

---

## 🔐 Error Handling

```go
// Custom errors برای بهتر handling
var (
    ErrDuplicateRequest    = errors.New("duplicate request")
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrBankFailed          = errors.New("bank withdrawal failed")
    ErrMissingIdempotencyKey = errors.New("missing idempotency_key")
)

// Response codes:
409 Conflict           // Duplicate idempotency key
400 Bad Request        // Invalid amount, missing fields
500 Internal Server    // Database errors
503 Service Unavailable // Worker pool queue full
```

---

## 📋 Checklist برای Production

- [ ] Database backups configured
- [ ] Connection pool monitored (max_used_connections)
- [ ] Worker queue length monitored (alert if > 80%)
- [ ] Graceful shutdown configured
- [ ] Request timeouts set (30s)
- [ ] Logging centralized (ELK stack)
- [ ] Rate limiting per user (future)
- [ ] Redis caching for balance queries (future)
- [ ] Database replicas for read scaling (future)

---

## 🚀 Starting the Service

```bash
# Build
go build -o wallet-simulator ./cmd/main.go

# Run (requires PostgreSQL)
./wallet-simulator

# Output:
# ==================================================
# 🚀 Wallet Balance Split Service
# ==================================================
# 📊 Configuration:
#    - Max Open Connections: 100
#    - Max Idle Connections: 25
#    - Worker Pool Size: 50
#    - Worker Queue Buffer: 100
# ==================================================
# 🌐 Server running on http://localhost:8080
# ==================================================
```

---

## 📞 Contact & Questions

اگر سوالی دارید در مورد architecture یا optimization:

1. **Connection Pooling:** `internal/handlers/handlers.go`
2. **Worker Pool:** `internal/worker/pool.go`
3. **Async Tasks:** `internal/tasks/bank_withdrawal.go`
4. **Repository:** `internal/repository/repository.go`

---

**نتیجه:** سرویس با این optimization می‌تواند:
- ✅ 10,000+ تراکنش در ساعت مدیریت کند
- ✅ 2,000+ درخواست برداشت در ساعت پردازش کند
- ✅ Bank API retry‌ها را asynchronously انجام دهد
- ✅ Database connections را efficiently مدیریت کند
- ✅ Gracefully scale به چند instance
