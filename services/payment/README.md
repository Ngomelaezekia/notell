# Notell Payment Service

Independent payment, subscription, entitlement, and gift-token service for Notell.

## Responsibilities

- provider-neutral payment records
- customer and product/price records
- idempotent payment creation
- Stripe checkout, subscriptions, webhooks, and refunds
- subscription packages and usage/overage tracking
- entitlement lifecycle consumed by Channel and Live
- signed-in gift-token claims and redemptions
- health/readiness endpoints
- independent PostgreSQL database

The service must not import the core backend, Channel, Message, or Music packages. Other services use stable HTTP contracts and IDs.

## Channel-plan purchase flow

1. A local user can register and use normal Notell features without a paid channel plan.
2. Creating a Channel requires an active `platform/channel` subscription entitlement.
3. A package purchase can use a gift code on the first channel-plan subscription only.
4. Paid gifts are applied through Stripe as one-time coupons.
5. A 100% admin gift activates the subscription without Stripe checkout and creates the entitlement directly.
6. Channel creation resolves the purchased package to the matching Channel plan.
7. Live creation/start requires both the Channel live permission and active Payment entitlement.

## Gift token types

- `ADMIN_FULL`: 100% off, maximum two tokens in the system, one redemption each.
- `SOCIAL_MEDIA`: 40% off through generated share links.
- `USER_AFFILIATION`: 50% off for a referred user on their first channel-plan purchase; the token creator cannot redeem their own affiliation token.

Gift codes are six characters and are claimed by authenticated users. Unused claims are automatically applied to the user's first eligible channel-plan purchase.

## Required production environment

- `DATABASE_URL`
- `JWT_SECRET`
- `FRONTEND_URL`
- `PUBLIC_APP_URL`
- `INTERNAL_SERVICE_KEY`
- `GIFT_ADMIN_KEY`
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_PRICE_PKG_STARTER`
- `STRIPE_PRICE_PKG_CREATOR`
- `STRIPE_PRICE_PKG_PRO`
- `STRIPE_PRICE_PKG_247`
- `SUBSCRIPTION_SUCCESS_URL`
- `SUBSCRIPTION_CANCEL_URL`

`GIFT_ADMIN_KEY` is only for protected admin/social gift management and must not be exposed to the frontend.

## Deployment

Render web service, independent PostgreSQL provider, no Docker/Kubernetes. The repository root `render.yaml` wires the service URLs and declares the Payment environment variables above.
