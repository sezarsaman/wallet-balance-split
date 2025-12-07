package utils

import (
	"database/sql"
	"os"
	"wallet-simulator/internal/handlers"
	"wallet-simulator/internal/repository"
	"wallet-simulator/internal/worker"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi/v5"
)

func SetupTestDB() *repository.Repository {
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
	db.Exec("TRUNCATE transactions RESTART IDENTITY CASCADE")
	db.Exec(`
		CREATE TABLE transactions (id SERIAL PRIMARY KEY, idempotency_key VARCHAR(255) UNIQUE, user_id INTEGER NOT NULL, amount BIGINT NOT NULL, "type" VARCHAR(10) NOT NULL, created_at TIMESTAMP NOT NULL, release_at TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
		ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending';
		CREATE INDEX IF NOT EXISTS idx_user_id ON transactions(user_id);
		CREATE INDEX IF NOT EXISTS idx_created_at ON transactions(created_at);
		CREATE INDEX IF NOT EXISTS idx_status ON transactions(status);
		CREATE INDEX IF NOT EXISTS idx_idempotency_key ON transactions(idempotency_key);
	`)
	return repository.NewRepository(db)
}

func SetupRouter(repo *repository.Repository) (chi.Router, *worker.WorkerPool) {
	r := chi.NewRouter()
	pool := worker.NewWorkerPool(1)
	handlers.SetupRoutes(r, repo, pool)
	return r, pool
}
