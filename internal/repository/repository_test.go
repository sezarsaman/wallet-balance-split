package repository_test

import (
	"testing"
	"time"
	"wallet-simulator/internal/utils"

	"testing/synctest"

	_ "github.com/lib/pq"
)

func TestChargeAndBalance(t *testing.T) {
	repo := utils.SetupTestDB()

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
	repo := utils.SetupTestDB()
	if err := repo.Charge(1, 2000, nil, "key1"); err != nil {
		t.Fatalf("Charge error: %v", err)
	}

	// Use synctest for deterministic concurrent withdraw
	synctest.Test(t, func(t *testing.T) {
		repo.Withdraw(t.Context(), 1, int64(1000), "key3")
		repo.Withdraw(t.Context(), 1, int64(500), "key4")
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
