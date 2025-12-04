package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"wallet-simulator/internal/handlers"
	"wallet-simulator/internal/models"
	"wallet-simulator/internal/repository"
	"wallet-simulator/internal/worker"

	"github.com/go-chi/chi/v5"
)

func setupRouter(repo *repository.Repository) chi.Router {
	r := chi.NewRouter()
	pool := worker.NewWorkerPool(10)
	handlers.SetupRoutes(r, repo, pool)
	return r
}

func TestChargeHandler(t *testing.T) {
	repo := setupTestDB()
	r := setupRouter(repo)

	reqBody, _ := json.Marshal(models.ChargeRequest{UserID: 1, Amount: 1000, IdempotencyKey: "test-1"})
	req := httptest.NewRequest("POST", "/charge", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; resp: %s", w.Code, w.Body.String())
	}
}

func TestGetBalanceHandler(t *testing.T) {
	repo := setupTestDB()
	r := setupRouter(repo)

	req := httptest.NewRequest("GET", "/balance?user_id=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; resp: %s", w.Code, w.Body.String())
	}
}
