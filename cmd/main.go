package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wallet-simulator/docs"
	"wallet-simulator/internal/config"
	"wallet-simulator/internal/handlers"
	"wallet-simulator/internal/metrics"
	"wallet-simulator/internal/repository"
	"wallet-simulator/internal/worker"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Swagger info
var swaggerInfo = docs.SwaggerInfo

func main() {
	// Initialize swagger
	swaggerInfo.Host = "localhost:8080"
	swaggerInfo.BasePath = "/"

	// ✅ Load configuration from .env
	cfg := config.Load()
	log.Println(cfg.String())

	// ✅ Connect to database using config
	db, err := ConnectWithRetry(cfg.GetDSN())
	if err != nil {
		log.Fatal("[DB] failed:", err)
	}

	// ✅ Connection Pooling Configuration
	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetimeMin) * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("✅ Database connected with connection pooling")

	repo := repository.NewRepository(db)

	// ✅ Initialize Worker Pool
	workerPool := worker.NewWorkerPool(cfg.WorkerPool.Size)
	defer func() {
		log.Println("🛑 Shutting down worker pool...")
		if err := workerPool.Shutdown(10 * time.Second); err != nil {
			log.Printf("Error shutting down worker pool: %v", err)
		}
	}()

	// ✅ Initialize Metrics
	m := metrics.New()

	// ✅ Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json"))

	// ✅ CORS Middleware Setup
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Idempotency-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	r.Use(handlers.MetricsMiddleware(m))

	// ✅ Expose Metrics and Swagger endpoints FIRST (before API routes)
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	// ✅ Setup API routes
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
	log.Println("POST   /charge")
	log.Println("POST   /withdraw")
	log.Println("GET    /balance")
	log.Println("GET    /transactions")
	log.Println("GET    /health")
	log.Println(sep)
	log.Printf("🌐 Server running on http://%s:%s\n", cfg.Server.Host, cfg.Server.Port)
	log.Println(sep)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("✅ Server stopped")
}

func ConnectWithRetry(dsn string) (*sql.DB, error) {
	maxRetries := 30
	baseDelay := time.Second // 1s

	var db *sql.DB
	var err error

	for i := 1; i <= maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("[DB] attempt %d/%d failed: %v", i, maxRetries, err)
		} else {
			// test real connectivity
			pingErr := db.Ping()
			if pingErr == nil {
				log.Println("[DB] successfully connected.")
				return db, nil
			}
			err = pingErr
			log.Printf("[DB] attempt %d/%d failed (ping): %v", i, maxRetries, err)
		}

		// exponential backoff
		sleep := baseDelay * time.Duration(i)
		log.Printf("[DB] retrying in %v...", sleep)
		time.Sleep(sleep)
	}

	return nil, fmt.Errorf("could not connect to DB after %d attempts: %w", maxRetries, err)
}
