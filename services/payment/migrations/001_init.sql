CREATE TABLE IF NOT EXISTS payment_customers (
    id text PRIMARY KEY,
    user_id text NOT NULL UNIQUE,
    provider text NOT NULL DEFAULT '',
    provider_customer_id text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payment_products (
    id text PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    active boolean NOT NULL DEFAULT true,
    provider text NOT NULL DEFAULT '',
    provider_product_id text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payment_prices (
    id text PRIMARY KEY,
    product_id text NOT NULL REFERENCES payment_products(id),
    currency varchar(3) NOT NULL,
    unit_amount bigint NOT NULL CHECK (unit_amount >= 0),
    interval text NOT NULL DEFAULT '',
    active boolean NOT NULL DEFAULT true,
    provider_price_id text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payment_prices_product_id ON payment_prices(product_id);

CREATE TABLE IF NOT EXISTS payments (
    id text PRIMARY KEY,
    user_id text NOT NULL,
    customer_id text NOT NULL DEFAULT '',
    price_id text NOT NULL DEFAULT '',
    amount bigint NOT NULL CHECK (amount >= 0),
    currency varchar(3) NOT NULL,
    status text NOT NULL,
    provider text NOT NULL,
    provider_payment_id text NOT NULL DEFAULT '',
    idempotency_key text NOT NULL UNIQUE,
    failure_code text NOT NULL DEFAULT '',
    failure_message text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payments_user_id_created_at ON payments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_provider_payment_id ON payments(provider_payment_id);

CREATE TABLE IF NOT EXISTS webhook_events (
    id text PRIMARY KEY,
    provider text NOT NULL,
    provider_event_id text NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    status text NOT NULL,
    error_message text NOT NULL DEFAULT '',
    processed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(provider, provider_event_id)
);

CREATE TABLE IF NOT EXISTS refunds (
    id text PRIMARY KEY,
    payment_id text NOT NULL REFERENCES payments(id),
    amount bigint NOT NULL CHECK (amount > 0),
    currency varchar(3) NOT NULL,
    status text NOT NULL,
    provider_refund_id text NOT NULL DEFAULT '',
    reason text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_refunds_payment_id ON refunds(payment_id);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id text PRIMARY KEY,
    payment_id text REFERENCES payments(id),
    user_id text NOT NULL,
    entry_type text NOT NULL,
    amount bigint NOT NULL,
    currency varchar(3) NOT NULL,
    reference text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_user_id_created_at ON ledger_entries(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS audit_events (
    id text PRIMARY KEY,
    payment_id text REFERENCES payments(id),
    user_id text NOT NULL DEFAULT '',
    event_type text NOT NULL,
    from_status text NOT NULL DEFAULT '',
    to_status text NOT NULL DEFAULT '',
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_events_payment_id_created_at ON audit_events(payment_id, created_at DESC);
