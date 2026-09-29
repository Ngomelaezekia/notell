ALTER TABLE refunds ADD COLUMN IF NOT EXISTS idempotency_key text NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_refunds_payment_idempotency_key
    ON refunds(payment_id, idempotency_key)
    WHERE idempotency_key <> '';
