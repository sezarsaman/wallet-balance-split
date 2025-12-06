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

func SetupRouter(repo *repository.Repository) chi.Router {
	r := chi.NewRouter()
	pool := worker.NewWorkerPool(10)
	handlers.SetupRoutes(r, repo, pool)
	return r
}
