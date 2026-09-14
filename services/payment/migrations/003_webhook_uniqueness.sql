CREATE UNIQUE INDEX IF NOT EXISTS idx_webhook_events_provider_event
ON webhook_events(provider, provider_event_id);
