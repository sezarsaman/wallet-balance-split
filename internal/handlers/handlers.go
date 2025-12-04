package handlers

import (
"encoding/json"
"log"
"net/http"
"strconv"
"wallet-simulator/internal/models"
"wallet-simulator/internal/repository"
"wallet-simulator/internal/tasks"
"wallet-simulator/internal/worker"

"github.com/go-chi/chi/v5"
)

// HandlerConfig حاوی dependencies برای handlers
type HandlerConfig struct {
	Repo       *repository.Repository
	WorkerPool *worker.WorkerPool
}

func SetupRoutes(r chi.Router, repo *repository.Repository, pool *worker.WorkerPool) {
	config := &HandlerConfig{
		Repo:       repo,
		WorkerPool: pool,
	}

	r.Post("/charge", ChargeHandler(config))
	r.Get("/transactions", GetTransactionsHandler(config))
	r.Get("/balance", GetBalanceHandler(config))
	r.Post("/withdraw", WithdrawHandler(config))
	r.Get("/health", HealthHandler(config))
}

func ChargeHandler(cfg *HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ChargeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.IdempotencyKey == "" {
			http.Error(w, models.ErrMissingIdempotencyKey.Error(), http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 {
			http.Error(w, models.ErrInvalidAmount.Error(), http.StatusBadRequest)
			return
		}

		err := cfg.Repo.Charge(req.UserID, req.Amount, req.ReleaseAt, req.IdempotencyKey)
		if err != nil {
			if err == models.ErrDuplicateRequest {
				http.Error(w, err.Error(), http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "charged", "idempotency_key": req.IdempotencyKey})
	}
}

func GetTransactionsHandler(cfg *HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 {
			limit = 10
		}

		transactions, total, err := cfg.Repo.GetTransactions(userID, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.TransactionsResponse{
Transactions: transactions,
Total:        total,
Page:         page,
Limit:        limit,
})
	}
}

func GetBalanceHandler(cfg *HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.Atoi(r.URL.Query().Get("user_id"))

		total, err := cfg.Repo.GetTotalBalance(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		withdrawable, err := cfg.Repo.GetWithdrawableBalance(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.Balance{Total: total, Withdrawable: withdrawable})
	}
}

func WithdrawHandler(cfg *HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.WithdrawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Amount <= 0 {
			http.Error(w, models.ErrInvalidAmount.Error(), http.StatusBadRequest)
			return
		}

		if req.IdempotencyKey == "" {
			http.Error(w, models.ErrMissingIdempotencyKey.Error(), http.StatusBadRequest)
			return
		}

		err := cfg.Repo.Withdraw(req.UserID, req.Amount, req.IdempotencyKey)
		if err != nil {
			if err == models.ErrDuplicateRequest {
				http.Error(w, err.Error(), http.StatusConflict)
			} else if err == models.ErrInsufficientBalance {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Submit bank withdrawal task to worker pool asynchronously
		task := tasks.NewBankWithdrawalTask(cfg.Repo, req.UserID, req.Amount, req.IdempotencyKey)
		if err := cfg.WorkerPool.Submit(task); err != nil {
			log.Printf("⚠️ Failed to submit withdrawal task: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
"message": "withdrawal request submitted",
"idempotency_key": req.IdempotencyKey,
"status": "pending",
})
	}
}

// HealthHandler برای health checks
func HealthHandler(cfg *HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queueLen := cfg.WorkerPool.GetQueueLength()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
"status": "ok",
"queue_length": queueLen,
})
	}
}
