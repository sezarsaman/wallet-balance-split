package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
	"wallet-simulator/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTransaction(tx *sql.Tx, userID int, amount int64, txType string, releaseAt *time.Time, idempotencyKey string) (int, error) {
	var id int
	query := `
		INSERT INTO transactions (user_id, amount, type, status, created_at, release_at, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`
	status := "completed"
	if txType == "withdraw" {
		status = "pending"
	}
	err := tx.QueryRow(query, userID, amount, txType, status, time.Now(), releaseAt, idempotencyKey).Scan(&id)
	return id, err
}

func (r *Repository) GetTransactions(userID, page, limit int) ([]models.Transaction, int, error) {
	offset := (page - 1) * limit
	rows, err := r.db.Query(`
		SELECT id, user_id, amount, type, status, created_at, release_at
		FROM transactions WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var releaseAt sql.NullTime
		err = rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Type, &t.Status, &t.CreatedAt, &releaseAt)
		if err != nil {
			return nil, 0, err
		}
		if releaseAt.Valid {
			t.ReleaseAt = &releaseAt.Time
		}
		transactions = append(transactions, t)
	}

	var total int
	err = r.db.QueryRow("SELECT COUNT(*) FROM transactions WHERE user_id = $1", userID).Scan(&total)
	return transactions, total, err
}

func (r *Repository) GetTotalBalance(userID int) (int64, error) {
	var total int64
	err := r.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1 AND status = 'completed'", userID).Scan(&total)
	return total, err
}

func (r *Repository) GetWithdrawableBalance(userID int) (int64, error) {
	var withdrawable int64
	now := time.Now()
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM transactions
		WHERE user_id = $1 AND status = 'completed' 
		AND (release_at <= $2 OR release_at IS NULL OR type = 'withdraw')
	`, userID, now).Scan(&withdrawable)
	return withdrawable, err
}

func (r *Repository) Charge(userID int, amount int64, releaseAt *time.Time, idempotencyKey string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var exists int
	err = tx.QueryRow("SELECT 1 FROM transactions WHERE idempotency_key = $1", idempotencyKey).Scan(&exists)
	if err == nil {
		tx.Rollback()
		return models.ErrDuplicateRequest
	} else if !errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return err
	}

	_, err = r.CreateTransaction(tx, userID, amount, "charge", releaseAt, idempotencyKey)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repository) Withdraw(ctx context.Context, userID int, amount int64, idempotencyKey string) error {

	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	tx, err := r.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	withdrawable, err := r.GetWithdrawableBalance(userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if withdrawable < amount {
		tx.Rollback()
		return models.ErrInsufficientBalance
	}

	var exists int
	err = tx.QueryRow("SELECT 1 FROM transactions WHERE idempotency_key = $1", idempotencyKey).Scan(&exists)
	if err == nil {
		tx.Rollback()
		return models.ErrDuplicateRequest
	} else if !errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return err
	}

	_, err = r.CreateTransaction(tx, userID, -amount, "withdraw", nil, idempotencyKey)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repository) UpdateWithdrawalStatus(idempotencyKey string, status string) error {
	query := `UPDATE transactions SET status = $1 WHERE idempotency_key = $2`
	result, err := r.db.Exec(query, status, idempotencyKey)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	log.Printf("✅ Updated withdrawal %s status to %s", idempotencyKey, status)
	return nil
}
