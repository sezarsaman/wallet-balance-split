package main

import (
	"database/sql"
	"log"
	"net/http"
	"wallet-simulator/internal/handlers"
	"wallet-simulator/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/wallet?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			idempotency_key VARCHAR(255) UNIQUE,
			user_id INTEGER NOT NULL,
			amount BIGINT NOT NULL,
			type VARCHAR(10) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			release_at TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	handlers.SetupRoutes(r, repo)
	http.ListenAndServe(":8080", r)
}
