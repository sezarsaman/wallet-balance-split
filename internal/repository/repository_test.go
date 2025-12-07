package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"wallet-simulator/internal/models"
	"wallet-simulator/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetTotalBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)

	userID := 1
	expectedBalance := int64(500)

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM transactions").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(expectedBalance))

	balance, err := repo.GetTotalBalance(userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestCharge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)

	userID := 1
	amount := int64(200)
	releaseAt := time.Now().Add(48 * time.Hour)
	idempotencyKey := "charge-key-456"

	// Check for duplicate
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 FROM transactions").
		WithArgs(idempotencyKey).
		WillReturnError(sql.ErrNoRows)

	// Insert transaction
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(userID, amount, "charge", "completed", sqlmock.AnyArg(), &releaseAt, idempotencyKey).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err = repo.Charge(userID, amount, &releaseAt, idempotencyKey)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)

	userID := 1
	amount := int64(300)
	idempotencyKey := "withdraw-key-789"

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM transactions").
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"withdrawable"}).AddRow(100)) // Less than amount
	mock.ExpectRollback()

	err = repo.Withdraw(context.Background(), userID, amount, idempotencyKey)
	assert.Equal(t, models.ErrInsufficientBalance, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
