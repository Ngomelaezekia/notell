CREATE INDEX IF NOT EXISTS idx_payment_customers_provider_customer_id ON payment_customers(provider, provider_customer_id);
CREATE INDEX IF NOT EXISTS idx_payment_products_provider_product_id ON payment_products(provider, provider_product_id);
CREATE INDEX IF NOT EXISTS idx_payment_prices_provider_price_id ON payment_prices(provider_price_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status_created_at ON webhook_events(status, created_at);
CREATE INDEX IF NOT EXISTS idx_refunds_provider_refund_id ON refunds(provider_refund_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_payment_id ON ledger_entries(payment_id);
