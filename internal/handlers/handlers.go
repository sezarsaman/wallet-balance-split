package handlers

import (
	"encoding/json" // use standard encoding/json for compatibility
	"net/http"
	"strconv"
	"wallet-simulator/internal/models"
	"wallet-simulator/internal/repository"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router, repo *repository.Repository) {
	r.Post("/charge", ChargeHandler(repo))
	r.Get("/transactions", GetTransactionsHandler(repo))
	r.Get("/balance", GetBalanceHandler(repo))
	r.Post("/withdraw", WithdrawHandler(repo))
}

func ChargeHandler(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ChargeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.IdempotencyKey == "" {
			http.Error(w, "missing idempotency_key", 400)
			return
		}
		if req.Amount <= 0 {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}
		err := repo.Charge(req.UserID, req.Amount, req.ReleaseAt, req.IdempotencyKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "charged"})
	}
}

func GetTransactionsHandler(repo *repository.Repository) http.HandlerFunc {
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
		transactions, total, err := repo.GetTransactions(userID, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(models.TransactionsResponse{
			Transactions: transactions,
			Total:        total,
			Page:         page,
			Limit:        limit,
		})
	}
}

func GetBalanceHandler(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
		total, err := repo.GetTotalBalance(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		withdrawable, err := repo.GetWithdrawableBalance(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(models.Balance{Total: total, Withdrawable: withdrawable})
	}
}

func WithdrawHandler(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.WithdrawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}
		if req.IdempotencyKey == "" {
			http.Error(w, "missing idempotency_key", 400)
			return
		}
		err := repo.Withdraw(req.UserID, req.Amount, req.IdempotencyKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "withdrawn"})
	}
}
