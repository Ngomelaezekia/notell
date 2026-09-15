CREATE TABLE IF NOT EXISTS entitlements (
 id TEXT PRIMARY KEY,
 user_id TEXT NOT NULL,
 resource_type TEXT NOT NULL,
 resource_id TEXT NOT NULL,
 plan_id TEXT NOT NULL,
 status TEXT NOT NULL,
 starts_at TIMESTAMPTZ NOT NULL,
 ends_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 UNIQUE(user_id, resource_type, resource_id)
);

CREATE TABLE IF NOT EXISTS entitlement_limits (
 id TEXT PRIMARY KEY,
 entitlement_id TEXT NOT NULL REFERENCES entitlements(id) ON DELETE CASCADE,
 metric TEXT NOT NULL,
 limit_value BIGINT NOT NULL CHECK(limit_value >= 0),
 UNIQUE(entitlement_id, metric)
);

CREATE TABLE IF NOT EXISTS subscription_events (
 id TEXT PRIMARY KEY,
 subscription_id TEXT NOT NULL,
 event_type TEXT NOT NULL,
 old_status TEXT,
 new_status TEXT,
 metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entitlements_resource ON entitlements(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_subscription_events_subscription ON subscription_events(subscription_id);
