package tests

import (
	"database/sql"
	"testing"
	"time"
	"wallet-simulator/internal/repository"

	"testing/synctest" // New in 1.25 for concurrent testing

	_ "github.com/lib/pq"
)

func setupTestDB() *repository.Repository {
	db, _ := sql.Open("postgres", "postgres://user:pass@localhost:5432/test_wallet?sslmode=disable")
	db.Exec("DROP TABLE IF EXISTS transactions")
	db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			amount BIGINT NOT NULL,
			idempotency_key VARCHAR(255) UNIQUE,
			type VARCHAR(10) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			release_at TIMESTAMP
		)
	`)
	return repository.NewRepository(db)
}

func TestChargeAndBalance(t *testing.T) {
	repo := setupTestDB()

	now := time.Now()
	future := now.Add(24 * time.Hour)

	repo.Charge(1, 1000, nil, "key1")
	repo.Charge(1, 500, &future, "key2")

	total, _ := repo.GetTotalBalance(1)
	withdrawable, _ := repo.GetWithdrawableBalance(1)
	if total != 1500 || withdrawable != 1000 {
		t.Errorf("expected 1500/1000, got %d/%d", total, withdrawable)
	}
}

func TestWithdrawConcurrent(t *testing.T) {
	repo := setupTestDB()
	repo.Charge(1, 2000, nil, "key1")

	// Use synctest for deterministic concurrent withdraw
	synctest.Test(t, func(t *testing.T) {
		repo.Withdraw(1, int64(1000), "key3")
		repo.Withdraw(1, int64(500), "key4")
	})

	total, _ := repo.GetTotalBalance(1)
	if total != 500 {
		t.Errorf("expected 500, got %d", total)
	}
}
