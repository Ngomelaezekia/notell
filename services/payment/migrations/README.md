# Payment migrations

Migrations are applied in lexical/version order by the service deployment process.

- `001_init.sql` creates the Phase 1 financial schema.
- `002_indexes.sql` adds operational indexes.

Migrations are forward-only. Never edit an applied migration; add the next numbered migration for schema changes.

The payment service uses its own PostgreSQL database. It must never connect to the Notell core database or another service database.
