-- Seeder for transactions table

-- Transactions For User 1
INSERT INTO transactions (user_id, amount, type, status, idempotency_key, created_at, updated_at, release_at)
VALUES 
(1, 100000, 'charge', 'completed', 'seed_charge_1_001', NOW(), NOW(), NOW() + INTERVAL '3 hours'),
(1, 50000, 'charge', 'completed', 'seed_charge_1_002', NOW(), NOW(), NOW() + INTERVAL '3 hours'),
(1, 30000, 'withdraw', 'completed', 'seed_withdraw_1_001', NOW(), NOW(), NULL),
(1, 10000, 'withdraw', 'pending', 'seed_withdraw_1_002', NOW(), NOW(), NOW() + INTERVAL '2 hours')
ON CONFLICT (idempotency_key) DO NOTHING;

-- Transactions For User 2
INSERT INTO transactions (user_id, amount, type, status, idempotency_key, created_at, updated_at, release_at)
VALUES
(2, 200000, 'charge', 'completed', 'seed_charge_2_001', NOW(), NOW(), NOW() + INTERVAL '3 hours'),
(2, 75000, 'charge', 'completed', 'seed_charge_2_002', NOW(), NOW(), NOW() + INTERVAL '3 hours'),
(2, 40000, 'withdraw', 'completed', 'seed_withdraw_2_001', NOW(), NOW(), NULL),
(2, 25000, 'withdraw', 'failed', 'seed_withdraw_2_002', NOW(), NOW(), NULL)
ON CONFLICT (idempotency_key) DO NOTHING;

-- Transactions For User 3
INSERT INTO transactions (user_id, amount, type, status, idempotency_key, created_at, updated_at, release_at)
VALUES
(3, 150000, 'charge', 'completed', 'seed_charge_3_001', NOW(), NOW(), NOW() + INTERVAL '3 hours'),
(3, 60000, 'withdraw', 'completed', 'seed_withdraw_3_001', NOW(), NOW(), NULL),
(3, 20000, 'withdraw', 'pending', 'seed_withdraw_3_002', NOW(), NOW(), NOW() + INTERVAL '2 hours')
ON CONFLICT (idempotency_key) DO NOTHING;
