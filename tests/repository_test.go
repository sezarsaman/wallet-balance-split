package tests

import (
	"database/sql"
	"os"
	"testing"
	"time"
	"wallet-simulator/internal/repository"

	"testing/synctest" // New in 1.25 for concurrent testing

	_ "github.com/lib/pq"
)

func setupTestDB() *repository.Repository {
	// Allow overriding test DB URL via TEST_DATABASE_URL env var.
	// Default matches docker-compose postgres credentials and port mapping.
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5433/wallet?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	db.Exec("DROP TABLE IF EXISTS transactions")
	db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			amount BIGINT NOT NULL,
			type VARCHAR(10) NOT NULL,
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			release_at TIMESTAMP,
			idempotency_key VARCHAR(255) UNIQUE
		)
	`)
	return repository.NewRepository(db)
}

func TestChargeAndBalance(t *testing.T) {
	repo := setupTestDB()

	now := time.Now()
	future := now.Add(24 * time.Hour)

	if err := repo.Charge(1, 1000, nil, "key1"); err != nil {
		t.Fatalf("Charge error: %v", err)
	}
	if err := repo.Charge(1, 500, &future, "key2"); err != nil {
		t.Fatalf("Charge error: %v", err)
	}

	total, err := repo.GetTotalBalance(1)
	if err != nil {
		t.Fatalf("GetTotalBalance error: %v", err)
	}
	withdrawable, err := repo.GetWithdrawableBalance(1)
	if err != nil {
		t.Fatalf("GetWithdrawableBalance error: %v", err)
	}
	if total != 1500 || withdrawable != 1000 {
		t.Errorf("expected 1500/1000, got %d/%d", total, withdrawable)
	}
}

func TestWithdrawConcurrent(t *testing.T) {
	repo := setupTestDB()
	if err := repo.Charge(1, 2000, nil, "key1"); err != nil {
		t.Fatalf("Charge error: %v", err)
	}

	// Use synctest for deterministic concurrent withdraw
	synctest.Test(t, func(t *testing.T) {
		repo.Withdraw(1, int64(1000), "key3")
		repo.Withdraw(1, int64(500), "key4")
	})

	// Simulate background worker completing the withdrawals
	if err := repo.UpdateWithdrawalStatus("key3", "completed"); err != nil {
		t.Fatalf("UpdateWithdrawalStatus error: %v", err)
	}
	if err := repo.UpdateWithdrawalStatus("key4", "completed"); err != nil {
		t.Fatalf("UpdateWithdrawalStatus error: %v", err)
	}

	total, err := repo.GetTotalBalance(1)
	if err != nil {
		t.Fatalf("GetTotalBalance error: %v", err)
	}
	if total != 500 {
		t.Errorf("expected 500, got %d", total)
	}
}
