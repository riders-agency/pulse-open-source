<!-- Generated from registry/standards/database/postgres@1.1.0. Update the registry standard, then regenerate. -->

# PostgreSQL Data Contract Standard

This standard applies to PostgreSQL schema, migrations, seeds and data access expectations.

## May 2026 Baseline

1. PostgreSQL baseline: `18.x`.
2. Domain identifiers use `uuid`.
3. PostgreSQL 18 `uuidv7()` is preferred for new high-write chronological tables when available.
4. `gen_random_uuid()` remains acceptable for generic starter templates and older managed PostgreSQL services.
5. Migrations are append-only, reversible where practical and reviewed before production.

## Entity Key Principle

Entity name determines key name:

1. table `users` -> primary key `user_id`;
2. table `companies` -> primary key `company_id`;
3. table `sessions` -> primary key `session_id`.

Generic `id` columns are not used for domain entities.

## Table Naming

1. Use `snake_case`.
2. Use plural names for entity tables.
3. Use explicit names for event/audit tables.
4. Avoid overloaded tables that mix multiple business concepts.

Examples:

1. `users`
2. `companies`
3. `sessions`
4. `auth_events`

## Primary Keys

1. PK format: `<entity_singular>_id`.
2. PK type: `uuid`.
3. Default: `uuidv7()` for PostgreSQL 18 deployments that support it, otherwise `gen_random_uuid()`.

Forbidden:

1. `id`
2. `uuid`
3. `user_uuid`
4. `users_id`
5. `serial` or `bigserial` for domain entity IDs.

## Foreign Keys

1. FK name follows the target entity: `<target_entity_singular>_id`.
2. FK type is `uuid`.
3. Use explicit `ON DELETE` behavior.
4. Do not rely on application code as the only referential integrity layer.

Examples:

1. `sessions.user_id -> users.user_id`
2. `companies.owner_user_id -> users.user_id`
3. `users.current_company_id -> companies.company_id`

## Relation Tables

1. Relation table names use `snake_case`.
2. Relation tables use canonical FK names.
3. Composite PK from FKs is preferred.
4. Add a surrogate `*_id` only when the relation has independent lifecycle or external references.

## Version Tables

Business entities that need audit/history use version tables:

1. table: `<entity_singular>_versions`;
2. PK: `<entity_singular>_version_id`;
3. entity reference: `<entity_singular>_id`;
4. version number: `version_no integer`;
5. timestamps and actor metadata where available.

## Time And Soft Delete

Standard timestamp fields:

1. `created_at timestamptz not null default now()`
2. `updated_at timestamptz not null default now()`
3. `deleted_at timestamptz null` only when soft delete is intentional.

Store timestamps in UTC. Do not store local wall-clock time unless it is domain data.

## Constraints And Indexes

Naming:

1. indexes: `idx_<table>_<column>[_<suffix>]`;
2. unique indexes: `uidx_<table>_<columns>`;
3. FK constraints: `<table>_<column>_fkey`;
4. checks: `<table>_<meaning>_check`.

Rules:

1. enforce invariant constraints in the database when practical;
2. create indexes for FK columns used in joins;
3. avoid indexes that are not tied to a query path;
4. use partial indexes for soft-delete and status filters when useful;
5. document any intentionally missing FK.

## JSONB Policy

JSONB is allowed for:

1. external provider payload snapshots;
2. sparse metadata;
3. append-only audit context;
4. feature flags or settings where schema churn is high.

JSONB is not a replacement for relational modeling when fields are queried, joined, authorized or validated as core business data.

## Migration Safety

Production-safe migrations should:

1. be append-only files;
2. avoid long blocking locks;
3. backfill in batches for large tables;
4. split expand and contract phases for breaking changes;
5. avoid dropping columns in the same deploy that stops writing them;
6. include down migrations for local/test unless impossible and documented;
7. include verification queries for risky data changes.

## Security And Tenancy

1. Tenant boundaries must be explicit in schema or queries.
2. RLS may be used when the project has operational skill to maintain it.
3. Sensitive values are hashed, encrypted or tokenized according to the security standard.
4. Do not store raw access tokens unless absolutely necessary and encrypted.
5. Audit tables must not leak secrets into public examples.

## Operations

Production systems need:

1. backup and restore drill;
2. PITR plan where data matters;
3. migration rollback plan;
4. slow query visibility;
5. `pg_stat_statements` or provider equivalent;
6. index bloat and vacuum monitoring.

## Good Example

```sql
CREATE TABLE users (
  user_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
  session_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
```

## Bad Example

```sql
CREATE TABLE users (
  id bigserial PRIMARY KEY
);

CREATE TABLE sessions (
  id uuid PRIMARY KEY,
  user_uuid uuid
);
```

## Enforcement Reality

This standard is documented by default. It becomes linted/tested/CI-blocking only when the generated project includes real migration and schema checks.
