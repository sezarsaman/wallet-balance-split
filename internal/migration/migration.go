package migration

import (
	"database/sql"
	"fmt"
	"log"
)

type Migrator struct {
	db *sql.DB
}

func New(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Up() error {
	log.Println("🔄 Running migrations...")

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "create_transactions_table",
			sql:  `CREATE TABLE IF NOT EXISTS transactions (id SERIAL PRIMARY KEY, idempotency_key VARCHAR(255) UNIQUE, user_id INTEGER NOT NULL, amount BIGINT NOT NULL, "type" VARCHAR(10) NOT NULL, created_at TIMESTAMP NOT NULL, release_at TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`,
		},
		{
			name: "add_status_column",
			sql:  `ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending';`,
		},
		{
			name: "create_idx_user_id",
			sql:  `CREATE INDEX IF NOT EXISTS idx_user_id ON transactions(user_id);`,
		},
		{
			name: "create_idx_created_at",
			sql:  `CREATE INDEX IF NOT EXISTS idx_created_at ON transactions(created_at);`,
		},
		{
			name: "create_idx_status",
			sql:  `CREATE INDEX IF NOT EXISTS idx_status ON transactions(status);`,
		},
		{
			name: "create_idx_idempotency_key",
			sql:  `CREATE INDEX IF NOT EXISTS idx_idempotency_key ON transactions(idempotency_key);`,
		},
	}

	for _, migration := range migrations {
		log.Printf("  ↳ Running: %s", migration.name)
		result, err := tx.Exec(migration.sql)
		if err != nil {
			return fmt.Errorf("migration '%s' failed: %w", migration.name, err)
		}
		rows, _ := result.RowsAffected()
		log.Printf("  ✅ %s completed (rows affected: %d)", migration.name, rows)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("✅ All migrations completed successfully")
	return nil
}

func (m *Migrator) Down() error {
	log.Println("⚠️  WARNING: Dropping all tables...")

	tables := []string{
		"transactions",
	}

	for _, table := range tables {
		log.Printf("  ↳ Dropping table: %s", table)
		if _, err := m.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		log.Printf("  ✅ Table %s dropped", table)
	}

	log.Println("✅ All tables dropped")
	return nil
}
