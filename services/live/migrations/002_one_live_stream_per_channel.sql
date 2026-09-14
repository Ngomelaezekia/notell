-- A channel has one active live session at a time.
-- Drafts and ended sessions remain fully supported.
CREATE UNIQUE INDEX IF NOT EXISTS ux_live_stream_one_active_per_channel
ON streams (channel_id)
WHERE status = 'live';
