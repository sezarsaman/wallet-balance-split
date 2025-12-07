package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"wallet-simulator/internal/models"
	"wallet-simulator/internal/utils"
)

func TestChargeHandler(t *testing.T) {
	repo := utils.SetupTestDB()
	r, _ := utils.SetupRouter(repo)

	tw := time.Now()
	twf := tw.Add(2 * time.Hour)
	reqBody, _ := json.Marshal(models.ChargeRequest{UserID: 1, Amount: 1000, IdempotencyKey: "test-1", ReleaseAt: &twf})
	req := httptest.NewRequest("POST", "/charge", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; resp: %s", w.Code, w.Body.String())
	}
}

func TestWithdrawHandler(t *testing.T) {
	repo := utils.SetupTestDB()
	r, pool := utils.SetupRouter(repo)

	tw := time.Now()
	repo.Charge(1, 100000, &tw, "testxyz")

	reqBody, _ := json.Marshal(models.WithdrawRequest{UserID: 1, Amount: 1000, IdempotencyKey: "test-2"})
	req := httptest.NewRequest("POST", "/withdraw", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	pool.Wait(1)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; resp: %s", w.Code, w.Body.String())
	}
}

func TestGetBalanceHandler(t *testing.T) {
	repo := utils.SetupTestDB()
	r, _ := utils.SetupRouter(repo)

	req := httptest.NewRequest("GET", "/balance?user_id=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; resp: %s", w.Code, w.Body.String())
	}
}

func TestGetTransactionsHandler(t *testing.T) {
	repo := utils.SetupTestDB()
	r, _ := utils.SetupRouter(repo)

	req := httptest.NewRequest("GET", "/transactions?user_id=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; resp: %s", w.Code, w.Body.String())
	}
}
