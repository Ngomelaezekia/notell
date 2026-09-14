CREATE TABLE IF NOT EXISTS subscriptions (
    id text PRIMARY KEY,
    user_id text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    price_id text NOT NULL,
    provider text NOT NULL DEFAULT 'stripe',
    provider_subscription_id text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'incomplete',
    current_period_start timestamptz,
    current_period_end timestamptz,
    cancel_at_period_end boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(user_id, resource_type, resource_id)
);
CREATE INDEX IF NOT EXISTS idx_subscriptions_resource ON subscriptions(resource_type, resource_id, status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id, created_at DESC);
