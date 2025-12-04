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
"wallet-simulator/internal/config"
"wallet-simulator/internal/handlers"
"wallet-simulator/internal/migration"
"wallet-simulator/internal/repository"
"wallet-simulator/internal/worker"

"github.com/go-chi/chi/v5"
"github.com/go-chi/chi/v5/middleware"
_ "github.com/lib/pq"
)

func main() {
	// ✅ Load configuration from .env
	cfg := config.Load()
	log.Println(cfg.String())

	// ✅ Connect to database using config
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ✅ Connection Pooling Configuration
	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetimeMin) * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("✅ Database connected with connection pooling")

	// ✅ Run migrations
	m := migration.New(db)
	if err := m.Up(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	repo := repository.NewRepository(db)

	// ✅ Initialize Worker Pool
	workerPool := worker.NewWorkerPool(cfg.WorkerPool.Size)
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
		Addr:           cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeoutSec) * time.Second,
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
	log.Println("📌 Available endpoints:")
	log.Println("   POST   /charge        - شارژ کردن")
	log.Println("   POST   /withdraw      - برداشت")
	log.Println("   GET    /balance       - موجودی")
	log.Println("   GET    /transactions  - تاریخچه تراکنش‌ها")
	log.Println("   GET    /health        - وضعیت سرویس")
	log.Println(sep)
	log.Printf("🌐 Server running on http://%s:%s\n", cfg.Server.Host, cfg.Server.Port)
	log.Println(sep)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("✅ Server stopped")
}
