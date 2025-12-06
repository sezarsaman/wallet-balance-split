package models

import (
	"errors"
	"time"
)

type Transaction struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Amount    int64      `json:"amount"` // positive for charge, negative for withdraw
	Type      string     `json:"type"`   // "charge" or "withdraw"
	Status    string     `json:"status"` // "pending", "completed", "failed"
	CreatedAt time.Time  `json:"created_at"`
	ReleaseAt *time.Time `json:"release_at,omitempty"` // optional for charge
}

type Balance struct {
	Total        int64 `json:"total"`
	Withdrawable int64 `json:"withdrawable"`
}

type ChargeRequest struct {
	UserID         int        `json:"user_id"`
	Amount         int64      `json:"amount"`
	IdempotencyKey string     `json:"idempotency_key"`
	ReleaseAt      *time.Time `json:"release_at,omitempty"`
}

type WithdrawRequest struct {
	UserID         int    `json:"user_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}

type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
	Total        int           `json:"total"`
	Page         int           `json:"page"`
	Limit        int           `json:"limit"`
}

type ChargeResponse struct {
	Message        string `json:"message"`
	IdempotencyKey string `json:"idempotency_key"`
}

type WithdrawResponse struct {
	Message        string `json:"message"`
	IdempotencyKey string `json:"idempotency_key"`
	Status         string `json:"status"`
}

type BalanceResponse struct {
	Total        int64 `json:"total"`
	Withdrawable int64 `json:"withdrawable"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

// Custom errors
var (
	ErrDuplicateRequest      = errors.New("duplicate request - idempotency key already exists")
	ErrInsufficientBalance   = errors.New("insufficient withdrawable balance")
	ErrBankFailed            = errors.New("bank withdrawal failed")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrMissingIdempotencyKey = errors.New("missing idempotency_key")
	ErrUserNotFound          = errors.New("user not found")
	ErrAmountMustBePositive  = errors.New("amount must be positive")
)
