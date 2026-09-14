CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    product_id TEXT,
    price_id TEXT,
    provider TEXT NOT NULL,
    provider_subscription_id TEXT,
    status TEXT NOT NULL,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_provider_subscription
ON subscriptions(provider, provider_subscription_id)
WHERE provider_subscription_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_status ON subscriptions(user_id, status);

CREATE TABLE IF NOT EXISTS subscription_events (
    id TEXT PRIMARY KEY,
    subscription_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    provider_event_id TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_events_provider_event
ON subscription_events(provider_event_id)
WHERE provider_event_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS platform_fee_configs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    basis_points BIGINT NOT NULL CHECK (basis_points >= 0 AND basis_points <= 10000),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cost_rates (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    metric TEXT NOT NULL,
    unit TEXT NOT NULL,
    rate_micros BIGINT NOT NULL CHECK (rate_micros >= 0),
    currency CHAR(3) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cost_rates_active_metric
ON cost_rates(provider, metric, unit, currency)
WHERE active = TRUE;

CREATE TABLE IF NOT EXISTS packages (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    currency CHAR(3) NOT NULL,
    monthly_price BIGINT NOT NULL CHECK (monthly_price >= 0),
    platform_margin_bps BIGINT NOT NULL CHECK (platform_margin_bps >= 0 AND platform_margin_bps <= 10000),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS package_allowances (
    id TEXT PRIMARY KEY,
    package_id TEXT NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    metric TEXT NOT NULL,
    included_units BIGINT NOT NULL CHECK (included_units >= 0),
    unit TEXT NOT NULL,
    overage_rate_micros BIGINT NOT NULL DEFAULT 0 CHECK (overage_rate_micros >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(package_id, metric)
);

CREATE TABLE IF NOT EXISTS usage_events (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    package_id TEXT,
    event_key TEXT NOT NULL UNIQUE,
    metric TEXT NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity >= 0),
    unit TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_usage_events_user_metric_time ON usage_events(user_id, metric, occurred_at);

CREATE TABLE IF NOT EXISTS account_cost_snapshots (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    package_id TEXT,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    infrastructure_cost_micros BIGINT NOT NULL CHECK (infrastructure_cost_micros >= 0),
    platform_fee_micros BIGINT NOT NULL CHECK (platform_fee_micros >= 0),
    customer_price_micros BIGINT NOT NULL CHECK (customer_price_micros >= 0),
    currency CHAR(3) NOT NULL,
    breakdown JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_cost_snapshots_user_period ON account_cost_snapshots(user_id, period_start, period_end);
