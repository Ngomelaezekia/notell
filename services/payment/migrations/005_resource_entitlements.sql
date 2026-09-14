ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS resource_type TEXT;
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS resource_id TEXT;
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS price_id TEXT;
UPDATE subscriptions SET resource_type = COALESCE(resource_type, 'account'), resource_id = COALESCE(resource_id, user_id), price_id = COALESCE(price_id, package_id) WHERE resource_type IS NULL OR resource_id IS NULL OR price_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_resource ON subscriptions(resource_type, resource_id);
