CREATE TABLE IF NOT EXISTS gift_tokens (
 id TEXT PRIMARY KEY,
 code VARCHAR(6) NOT NULL UNIQUE,
 type VARCHAR(32) NOT NULL,
 discount_percent INT NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),
 max_redemptions INT NOT NULL DEFAULT 1 CHECK (max_redemptions > 0),
 redeemed_count INT NOT NULL DEFAULT 0 CHECK (redeemed_count >= 0),
 created_by_user_id TEXT,
 provider_coupon_id TEXT UNIQUE,
 active BOOLEAN NOT NULL DEFAULT TRUE,
 expires_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS gift_claims (
 id TEXT PRIMARY KEY,
 gift_token_id TEXT NOT NULL REFERENCES gift_tokens(id) ON DELETE CASCADE,
 user_id TEXT NOT NULL,
 claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 used_at TIMESTAMPTZ,
 subscription_id TEXT,
 UNIQUE(gift_token_id, user_id)
);

CREATE TABLE IF NOT EXISTS gift_redemptions (
 id TEXT PRIMARY KEY,
 gift_token_id TEXT NOT NULL REFERENCES gift_tokens(id) ON DELETE CASCADE,
 user_id TEXT NOT NULL,
 subscription_id TEXT NOT NULL,
 redeemed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 UNIQUE(gift_token_id, user_id),
 UNIQUE(gift_token_id, subscription_id)
);

CREATE INDEX IF NOT EXISTS idx_gift_tokens_active ON gift_tokens(active, expires_at);
CREATE INDEX IF NOT EXISTS idx_gift_claims_user ON gift_claims(user_id, claimed_at DESC);
CREATE INDEX IF NOT EXISTS idx_gift_redemptions_subscription ON gift_redemptions(subscription_id);

ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS gift_token_id TEXT REFERENCES gift_tokens(id) ON DELETE SET NULL;
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS gift_discount_percent INT NOT NULL DEFAULT 0 CHECK (gift_discount_percent >= 0 AND gift_discount_percent <= 100);
CREATE INDEX IF NOT EXISTS idx_subscriptions_gift_token ON subscriptions(gift_token_id);
