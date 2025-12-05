package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"wallet-simulator/internal/config"
	"wallet-simulator/internal/migration"
	"wallet-simulator/internal/seeder"

	_ "github.com/lib/pq"
)

func main() {

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	cfg := config.Load()
	fmt.Println(cfg)

	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}
	log.Println("✅ Database connected")

	switch command {
	case "migrate":
		handleMigrate(db, os.Args[2:])
	case "seed":
		handleSeed(db)
	case "refresh":
		handleRefresh(db)
	case "clear":
		handleClear(db)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleMigrate(db *sql.DB, args []string) {
	m := migration.New(db)

	if len(args) > 0 && args[0] == "down" {
		if err := m.Down(); err != nil {
			log.Fatalf("❌ Migration down failed: %v", err)
		}
	} else {
		if err := m.Up(); err != nil {
			log.Fatalf("❌ Migration up failed: %v", err)
		}
	}
}

func handleSeed(db *sql.DB) {
	s := seeder.New(db)
	if err := s.Seed(); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}
}

func handleRefresh(db *sql.DB) {
	log.Println("🔄 Refreshing database (Down -> Up -> Seed)...")

	m := migration.New(db)
	s := seeder.New(db)

	// Down
	if err := m.Down(); err != nil {
		log.Fatalf("❌ Migration down failed: %v", err)
	}

	// Up
	if err := m.Up(); err != nil {
		log.Fatalf("❌ Migration up failed: %v", err)
	}

	// Seed
	if err := s.Seed(); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Database refresh completed")
}

func handleClear(db *sql.DB) {
	s := seeder.New(db)
	if err := s.Clear(); err != nil {
		log.Fatalf("❌ Clear failed: %v", err)
	}
}

func printUsage() {
	fmt.Print(`
╔═════════════════════════════════════════════════════════════╗
║         Wallet Service - Database CLI Tool                 ║
╚═════════════════════════════════════════════════════════════╝

Usage:
  go run cmd/cli/main.go <command>

Commands:
  migrate          Run all migrations (create tables, indexes)
  migrate down     Drop all tables
  seed             Insert test data
  clear            Remove all test data
  refresh          Drop tables, create tables, insert data
                   (Complete database reset)

Examples:
  go run cmd/cli/main.go migrate
  go run cmd/cli/main.go migrate down
  go run cmd/cli/main.go seed
  go run cmd/cli/main.go refresh
  go run cmd/cli/main.go clear

Configuration:
  Environment variables are loaded from .env file
  Key variables: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME

Note:
  - Always run 'migrate' before 'seed'
  - Use 'refresh' to reset database completely
  - Seed data uses idempotency keys (safe to run multiple times)
`)
}
