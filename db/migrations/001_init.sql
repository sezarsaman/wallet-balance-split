CREATE TABLE IF NOT EXISTS transactions (id SERIAL PRIMARY KEY, idempotency_key VARCHAR(255) UNIQUE, user_id INTEGER NOT NULL, amount BIGINT NOT NULL, "type" VARCHAR(10) NOT NULL, created_at TIMESTAMP NOT NULL, release_at TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending';
CREATE INDEX IF NOT EXISTS idx_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_idempotency_key ON transactions(idempotency_key);