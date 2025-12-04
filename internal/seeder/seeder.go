package seeder

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Seeder مسئول اضافه کردن داده‌های آزمایشی است
type Seeder struct {
	db *sql.DB
}

// New یک نمونه جدید از Seeder می‌سازد
func New(db *sql.DB) *Seeder {
	return &Seeder{db: db}
}

// Seed تمام داده‌های آزمایشی را اضافه می‌کند
func (s *Seeder) Seed() error {
	log.Println("🌱 Seeding database...")

	// داده‌های آزمایشی
	testData := []struct {
		userID         int
		amount         int64
		txType         string
		status         string
		idempotencyKey string
	}{
		// User 1
		{1, 100000, "charge", "completed", "seed_charge_1_001"},
		{1, 50000, "charge", "completed", "seed_charge_1_002"},
		{1, 30000, "withdraw", "completed", "seed_withdraw_1_001"},
		{1, 10000, "withdraw", "pending", "seed_withdraw_1_002"},

		// User 2
		{2, 200000, "charge", "completed", "seed_charge_2_001"},
		{2, 75000, "charge", "completed", "seed_charge_2_002"},
		{2, 40000, "withdraw", "completed", "seed_withdraw_2_001"},
		{2, 25000, "withdraw", "failed", "seed_withdraw_2_002"},

		// User 3
		{3, 150000, "charge", "completed", "seed_charge_3_001"},
		{3, 60000, "withdraw", "completed", "seed_withdraw_3_001"},
		{3, 20000, "withdraw", "pending", "seed_withdraw_3_002"},
	}

	now := time.Now()
	for _, data := range testData {
		// set release_at: for charges, put a hold that releases in ~3 hours;
		// for pending withdrawals, set release a few hours ahead as well.
		var releaseAt interface{}
		if data.txType == "charge" {
			t := now.Add(3 * time.Hour)
			releaseAt = t
		} else if data.status == "pending" {
			t := now.Add(2 * time.Hour)
			releaseAt = t
		} else {
			releaseAt = nil
		}

		query := `
				INSERT INTO transactions (user_id, amount, type, status, idempotency_key, created_at, updated_at, release_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (idempotency_key) DO NOTHING
			`

		if _, err := s.db.Exec(
			query,
			data.userID,
			data.amount,
			data.txType,
			data.status,
			data.idempotencyKey,
			now,
			now,
			releaseAt,
		); err != nil {
			return fmt.Errorf("failed to seed transaction: %w", err)
		}

		// If this row already existed (ON CONFLICT DO NOTHING), or release_at was NULL,
		// make sure release_at is set for charges and pending withdrawals.
		if data.txType == "charge" {
			_, _ = s.db.Exec(`
						UPDATE transactions
						SET release_at = created_at + INTERVAL '3 hours'
						WHERE idempotency_key = $1 AND release_at IS NULL AND type = 'charge'
					`, data.idempotencyKey)
		} else if data.status == "pending" {
			_, _ = s.db.Exec(`
						UPDATE transactions
						SET release_at = created_at + INTERVAL '2 hours'
						WHERE idempotency_key = $1 AND release_at IS NULL AND status = 'pending'
					`, data.idempotencyKey)
		}

		log.Printf("  ✅ Seeded: User %d, %s %.2f (%s)",
			data.userID, data.txType, float64(data.amount)/100, data.status)
	}

	log.Println("✅ Database seeding completed successfully")
	return nil
}

// Clear تمام داده‌های آزمایشی را حذف می‌کند
func (s *Seeder) Clear() error {
	log.Println("🗑️  Clearing seed data...")

	seedIDs := []string{
		"seed_%",
	}

	for _, pattern := range seedIDs {
		query := `DELETE FROM transactions WHERE idempotency_key LIKE $1`
		result, err := s.db.Exec(query, pattern)
		if err != nil {
			return fmt.Errorf("failed to clear seed data: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		log.Printf("  ✅ Deleted %d rows", rowsAffected)
	}

	log.Println("✅ Seed data cleared successfully")
	return nil
}
