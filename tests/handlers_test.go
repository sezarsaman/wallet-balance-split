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

	"github.com/go-chi/chi/v5"
)

func setupRouter(repo *repository.Repository) chi.Router {
	r := chi.NewRouter()
	handlers.SetupRoutes(r, repo)
	return r
}

func TestChargeHandler(t *testing.T) {
	repo := setupTestDB()
	r := setupRouter(repo)

	reqBody, _ := json.Marshal(models.ChargeRequest{UserID: 1, Amount: 1000})
	req := httptest.NewRequest("POST", "/charge", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetBalanceHandler(t *testing.T) {
	repo := setupTestDB()
	r := setupRouter(repo)

	req := httptest.NewRequest("GET", "/balance?user_id=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
