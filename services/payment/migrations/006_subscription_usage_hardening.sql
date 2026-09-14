CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_one_active_per_user_package
ON subscriptions(user_id, package_id)
WHERE status IN ('incomplete','active','past_due','paused');

CREATE INDEX IF NOT EXISTS idx_usage_events_subscription
ON usage_events(subscription_id, occurred_at);
