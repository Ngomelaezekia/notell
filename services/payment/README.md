# Notell Payment Service

Independent payment and billing foundation for Notell.

## Phase 1

- provider-neutral payment records
- customer and product/price records
- idempotency foundation
- webhook event persistence
- refund records
- immutable-style financial ledger entries
- audit-ready status lifecycle
- health/readiness endpoints
- independent PostgreSQL database

The service must not import the core backend, Channel, Message, or Music packages. Other services refer to payment records through stable IDs and API contracts.

## Planned provider

Stripe is the initial payment-provider integration target. Provider-specific code stays behind the payment service boundary so the application can add another provider later.

## Deployment

Render web service, independent PostgreSQL provider, no Docker/Kubernetes.
