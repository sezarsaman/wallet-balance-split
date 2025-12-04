# Monitoring Quick Reference

## Starting the Complete Stack

```bash
# Terminal 1: Start all services
make db-up        # PostgreSQL + Prometheus + Grafana
make migrate      # Create tables
make seed         # Insert test data

# Terminal 2: Start the wallet API
make run

# Terminal 3 (optional): Monitor logs
make logs
```

## Accessing the Services

```bash
# Check if services are running
docker ps | grep wallet

# View service logs
docker logs wallet_postgres_v2    # Database
docker logs wallet_prometheus      # Metrics DB
docker logs wallet_grafana         # Dashboard
```

## Prometheus Queries

Access http://localhost:9090 and try these queries:

### Request Rate (requests per second)
```promql
rate(http_requests_total[1m])
```

### Request Duration (p95 latency)
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

### Error Rate
```promql
rate(http_errors_total[1m])
```

### Error Rate by Endpoint
```promql
rate(http_errors_total{route!="/health"}[1m])
```

### Database Query Time (p99)
```promql
histogram_quantile(0.99, rate(db_query_time_seconds_bucket[5m]))
```

### Worker Queue Depth
```promql
worker_queue_length
```

### Total Charges Sum
```promql
sum(charge_amount)
```

### Total Withdrawals Sum
```promql
sum(withdraw_amount)
```

## Grafana Dashboards

1. **Login**: http://localhost:3000
   - Username: `admin`
   - Password: `admin`

2. **Add Prometheus Data Source**:
   - Configuration → Data Sources → Add Prometheus
   - URL: `http://prometheus:9090`
   - Click "Save & Test"

3. **Create Dashboard**:
   - Click "+" → Create Dashboard
   - Add panels with queries from above

4. **Sample Dashboard Panels**:
   - **Request Rate**: `rate(http_requests_total[1m])`
   - **Latency p95**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
   - **Error Rate**: `rate(http_errors_total[1m])`
   - **Worker Queue**: `worker_queue_length`

## Testing with Load

```bash
# Generate traffic to see metrics
for i in {1..100}; do
  curl -X POST http://localhost:8080/charge \
    -H "Content-Type: application/json" \
    -d "{\"user_id\": \"user$((RANDOM % 10))\", \"amount\": $((1000 + RANDOM % 9000))}" &
done

# Monitor concurrent requests
for i in {1..50}; do
  curl http://localhost:8080/balance?user_id=user1 &
done
```

## Stopping Services

```bash
# Stop wallet API
make stop

# Stop all containers
make db-down

# Clean up everything
make clean
docker volume prune  # Remove unused volumes
```

## Monitoring Files

- `internal/metrics/metrics.go` - Metric definitions
- `internal/handlers/middleware.go` - HTTP tracking middleware
- `prometheus.yml` - Prometheus config
- `grafana-dashboard.json` - Sample dashboard
- `docker-compose.yml` - Service definitions

## Customizing Metrics

To add more metrics, edit `internal/metrics/metrics.go`:

```go
// Add in the Metrics struct
NewCounter: prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "my_custom_counter",
        Help: "Description",
    },
    []string{"label1", "label2"},
),

// Record in your code
metrics.NewCounter.WithLabelValues("value1", "value2").Inc()
```

Then update `cmd/main.go` to pass metrics to handlers if needed.

## Troubleshooting

### Prometheus not scraping metrics
- Check http://localhost:9090/targets
- Verify wallet API is running on 8080
- Check prometheus.yml has correct target config

### Grafana can't connect to Prometheus
- Verify both containers are on same network: `docker network ls`
- Check Prometheus URL is `http://prometheus:9090` (not localhost)
- Verify Prometheus is accessible: `curl http://prometheus:9090/api/v1/targets`

### Services won't start
- Check port conflicts: `lsof -i :8080 :5433 :9090 :3000`
- Check Docker resources: `docker stats`
- View service logs: `docker logs wallet_postgres_v2`

## Performance Tips

- Keep default scrape interval (15s) for low overhead
- Use rate functions with 5m+ windows for stable graphs
- Index frequently-queried metrics in Prometheus config
- Regularly backup Prometheus volumes for data retention

---

**For more info, see MONITORING_SETUP.md**
