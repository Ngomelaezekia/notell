ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_subscriptions_provider_sync
    ON subscriptions(provider, provider_subscription_id, last_synced_at);
