CREATE UNIQUE INDEX IF NOT EXISTS idx_ledger_payment_capture
ON ledger_entries(payment_id, entry_type)
WHERE entry_type = 'payment_captured';
