CREATE TABLE IF NOT EXISTS payment_customers (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    provider TEXT NOT NULL DEFAULT 'stripe',
    provider_customer_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    provider TEXT NOT NULL DEFAULT 'stripe',
    provider_product_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_prices (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES payment_products(id),
    currency TEXT NOT NULL,
    unit_amount BIGINT NOT NULL CHECK (unit_amount >= 0),
    interval TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    provider_price_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    customer_id TEXT REFERENCES payment_customers(id),
    price_id TEXT REFERENCES payment_prices(id),
    amount BIGINT NOT NULL CHECK (amount >= 0),
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    provider TEXT NOT NULL DEFAULT 'stripe',
    provider_payment_id TEXT,
    idempotency_key TEXT NOT NULL,
    failure_code TEXT,
    failure_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (idempotency_key),
    UNIQUE (provider, provider_payment_id)
);

CREATE INDEX IF NOT EXISTS idx_payments_user_created ON payments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);

CREATE TABLE IF NOT EXISTS payment_webhook_events (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    provider_event_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'received',
    error_message TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_event_id)
);

CREATE TABLE IF NOT EXISTS payment_refunds (
    id TEXT PRIMARY KEY,
    payment_id TEXT NOT NULL REFERENCES payments(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    provider_refund_id TEXT,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_ledger_entries (
    id TEXT PRIMARY KEY,
    payment_id TEXT REFERENCES payments(id),
    user_id TEXT NOT NULL,
    entry_type TEXT NOT NULL,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    reference TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_audit_events (
    id TEXT PRIMARY KEY,
    payment_id TEXT REFERENCES payments(id),
    user_id TEXT,
    event_type TEXT NOT NULL,
    from_status TEXT,
    to_status TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_audit_payment ON payment_audit_events(payment_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_ledger_user ON payment_ledger_entries(user_id, created_at DESC);
