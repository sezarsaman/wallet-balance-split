package tasks

import (
	"context"
	"log"
	"math/rand"
	"time"
	"wallet-simulator/internal/models"
	"wallet-simulator/internal/repository"
)

// BankWithdrawalTask نمایندگی میکند یک async bank withdrawal
type BankWithdrawalTask struct {
	repo           *repository.Repository
	userID         int
	amount         int64
	idempotencyKey string
}

// NewBankWithdrawalTask ایجاد یک task جدید
func NewBankWithdrawalTask(repo *repository.Repository, userID int, amount int64, idempotencyKey string) *BankWithdrawalTask {
	return &BankWithdrawalTask{
		repo:           repo,
		userID:         userID,
		amount:         amount,
		idempotencyKey: idempotencyKey,
	}
}

// Execute انجام دادن task
func (t *BankWithdrawalTask) Execute(ctx context.Context) error {
	return t.withdrawWithRetries(ctx, 3)
}

func (t *BankWithdrawalTask) withdrawWithRetries(ctx context.Context, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			log.Printf("⏱️ Withdrawal cancelled for user %d", t.userID)
			return ctx.Err()
		default:
		}

		if rand.Float64() > 0.3 {
			log.Printf("✅ Bank withdrawal successful for user %d (attempt %d)", t.userID, attempt)
			if err := t.repo.UpdateWithdrawalStatus(t.idempotencyKey, "completed"); err != nil {
				log.Printf("⚠️ Failed to update withdrawal status: %v", err)
				return err
			}
			return nil
		}

		lastErr = models.ErrBankFailed
		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("⏳ Retry attempt %d/%d for user %d after %v", attempt, maxRetries, t.userID, backoff)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	log.Printf("❌ Bank withdrawal failed after %d retries for user %d", maxRetries, t.userID)
	if err := t.repo.UpdateWithdrawalStatus(t.idempotencyKey, "failed"); err != nil {
		log.Printf("⚠️ Failed to update withdrawal status: %v", err)
	}
	return lastErr
}
