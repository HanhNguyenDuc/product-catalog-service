-- migrations/001_initial_schema.sql
-- Initial schema for the product catalog service (Cloud Spanner DDL).
-- Apply with the Spanner Admin API or the gcloud CLI:
--
--   gcloud spanner databases ddl update <DATABASE> \
--     --instance=<INSTANCE> \
--     --ddl-file=migrations/001_initial_schema.sql

-- ─── Products ─────────────────────────────────────────────────────────────────
CREATE TABLE products (
    product_id              STRING(36)   NOT NULL,
    name                    STRING(255)  NOT NULL,
    description             STRING(MAX),
    category                STRING(100)  NOT NULL,
    base_price_numerator    INT64        NOT NULL,
    base_price_denominator  INT64        NOT NULL,
    discount_percent        NUMERIC,
    discount_start_date     TIMESTAMP,
    discount_end_date       TIMESTAMP,
    status                  STRING(20)   NOT NULL,
    created_at              TIMESTAMP    NOT NULL OPTIONS (allow_commit_timestamp=true),
    updated_at              TIMESTAMP    NOT NULL OPTIONS (allow_commit_timestamp=true),
    archived_at             TIMESTAMP,
) PRIMARY KEY (product_id);

CREATE INDEX idx_products_category ON products(category, status);

-- ─── Outbox Events ────────────────────────────────────────────────────────────
-- Rows are written atomically with business mutations and relayed to a broker.
CREATE TABLE outbox_events (
    event_id      STRING(36)   NOT NULL,
    event_type    STRING(100)  NOT NULL,
    aggregate_id  STRING(36)   NOT NULL,
    payload       JSON         NOT NULL,
    status        STRING(20)   NOT NULL,
    created_at    TIMESTAMP    NOT NULL OPTIONS (allow_commit_timestamp=true),
    processed_at  TIMESTAMP,
) PRIMARY KEY (event_id);

CREATE INDEX idx_outbox_status ON outbox_events(status, created_at);
