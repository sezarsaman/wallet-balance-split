package validation

import (
	"time"
	"wallet-simulator/internal/models"
)

func ValidateReleaseAt(releaseAt *time.Time) string {
	if releaseAt.Before(time.Now()) {
		return models.ErrReleaseAtMustBeFuture.Error()
	}
	return ""
}

func ValidateAmount(amount int64) string {
	if amount <= 0 {
		return models.ErrAmountCannotBeZero.Error()
	}
	return ""
}

func ValidateIdempotencyKey(idempotency_key string) string {
	if idempotency_key == "" {
		return models.ErrMissingIdempotencyKey.Error()
	}
	return ""
}

func ValidateUserID(user_id int) string {
	if user_id <= 0 {
		return models.ErrUserNotFound.Error()
	}
	return ""
}
