package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wallet-simulator/internal/handlers"
	"wallet-simulator/internal/repository"
	"wallet-simulator/internal/worker"

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

	// ✅ Connection Pooling Configuration
	// For handling 10,000 transactions/hour (≈3 req/sec)
	db.SetMaxOpenConns(100)      // Maximum concurrent connections
	db.SetMaxIdleConns(25)       // Keep 25 idle connections ready
	db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections every 5 minutes

	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("✅ Database connected with connection pooling")

	repo := repository.NewRepository(db)

	// ✅ Create tables with proper schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			idempotency_key VARCHAR(255) UNIQUE,
			user_id INTEGER NOT NULL,
			amount BIGINT NOT NULL,
			type VARCHAR(10) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL,
			release_at TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_user_id ON transactions(user_id);
		CREATE INDEX IF NOT EXISTS idx_created_at ON transactions(created_at);
		CREATE INDEX IF NOT EXISTS idx_status ON transactions(status);
	`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✅ Database tables created")

	// ✅ Initialize Worker Pool
	// For 10k req/hour with 3-4 req/sec, 50 workers is optimal
	workerPool := worker.NewWorkerPool(50)
	defer func() {
		log.Println("🛑 Shutting down worker pool...")
		if err := workerPool.Shutdown(10 * time.Second); err != nil {
			log.Printf("Error shutting down worker pool: %v", err)
		}
	}()

	// ✅ Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	handlers.SetupRoutes(r, repo, workerPool)

	// ✅ HTTP Server Configuration
	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// ✅ Graceful Shutdown Setup
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("\n🛑 Shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// ✅ Log Configuration
	sep := "=================================================="
	log.Println(sep)
	log.Println("🚀 Wallet Balance Split Service")
	log.Println(sep)
	log.Println("📊 Configuration:")
	log.Println("   - Max Open Connections: 100")
	log.Println("   - Max Idle Connections: 25")
	log.Println("   - Worker Pool Size: 50")
	log.Println("   - Worker Queue Buffer: 100")
	log.Println(sep)
	log.Println("🌐 Server running on http://localhost:8080")
	log.Println("📌 Available endpoints:")
	log.Println("   POST   /charge        - شارژ کردن")
	log.Println("   POST   /withdraw      - برداشت")
	log.Println("   GET    /balance       - موجودی")
	log.Println("   GET    /transactions  - تاریخچه تراکنش‌ها")
	log.Println("   GET    /health        - وضعیت سرویس")
	log.Println(sep)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("✅ Server stopped")
}
