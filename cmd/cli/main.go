package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	Migrate()
}

func Migrate() {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/wbs_db_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	_, err = db.Exec(`
		DROP TABLE IF EXISTS transactions CASCADE;
		CREATE TABLE transactions (id SERIAL PRIMARY KEY, idempotency_key VARCHAR(255) UNIQUE, user_id INTEGER NOT NULL, amount BIGINT NOT NULL, "type" VARCHAR(10) NOT NULL, created_at TIMESTAMP NOT NULL, release_at TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
		ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending';
		CREATE INDEX IF NOT EXISTS idx_user_id ON transactions(user_id);
		CREATE INDEX IF NOT EXISTS idx_created_at ON transactions(created_at);
		CREATE INDEX IF NOT EXISTS idx_status ON transactions(status);
		CREATE INDEX IF NOT EXISTS idx_idempotency_key ON transactions(idempotency_key);
	`)

	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
}
