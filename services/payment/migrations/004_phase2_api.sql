CREATE TABLE IF NOT EXISTS subscriptions (
 id TEXT PRIMARY KEY, user_id TEXT NOT NULL, package_id TEXT NOT NULL, status TEXT NOT NULL, provider TEXT NOT NULL DEFAULT 'stripe', provider_subscription_id TEXT, current_period_start TIMESTAMPTZ NOT NULL, current_period_end TIMESTAMPTZ NOT NULL, cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_provider_id ON subscriptions(provider, provider_subscription_id) WHERE provider_subscription_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS usage_events (
 id TEXT PRIMARY KEY, user_id TEXT NOT NULL, subscription_id TEXT, metric TEXT NOT NULL, quantity BIGINT NOT NULL CHECK(quantity >= 0), idempotency_key TEXT NOT NULL UNIQUE, occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_usage_events_user_metric ON usage_events(user_id, metric, occurred_at);

CREATE TABLE IF NOT EXISTS packages (
 id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, description TEXT NOT NULL DEFAULT '', currency TEXT NOT NULL, monthly_price BIGINT NOT NULL CHECK(monthly_price >= 0), active BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS package_allowances (
 id TEXT PRIMARY KEY, package_id TEXT NOT NULL REFERENCES packages(id) ON DELETE CASCADE, metric TEXT NOT NULL, included_quantity BIGINT NOT NULL CHECK(included_quantity >= 0), overage_unit_price BIGINT NOT NULL CHECK(overage_unit_price >= 0), UNIQUE(package_id, metric)
);

CREATE TABLE IF NOT EXISTS usage_overages (
 id TEXT PRIMARY KEY, user_id TEXT NOT NULL, subscription_id TEXT, package_id TEXT NOT NULL, metric TEXT NOT NULL, quantity BIGINT NOT NULL CHECK(quantity > 0), unit_price BIGINT NOT NULL CHECK(unit_price >= 0), amount BIGINT NOT NULL CHECK(amount >= 0), period_start TIMESTAMPTZ NOT NULL, period_end TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_usage_overages_user_period ON usage_overages(user_id, period_start, period_end);
