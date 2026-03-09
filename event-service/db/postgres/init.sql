BEGIN;

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    performer_id BIGINT NOT NULL,
    venue_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    date_start TIMESTAMPTZ NOT NULL,
    sold_out BOOLEAN DEFAULT FALSE,
    post_date TIMESTAMPTZ NOT NULL,
    sale_start_date TIMESTAMPTZ NOT NULL,
    max_price_cof DOUBLE PRECISION NOT NULL,
    min_price_cof DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS tickets (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    venue_id BIGINT NOT NULL,
    sector_name VARCHAR(100) NOT NULL,
    row_number INT NOT NULL,
    seat_number INT NOT NULL,
    price BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
    user_id BIGINT DEFAULT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS promo (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    sector_name VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_event_promo_code UNIQUE (event_id, code)
);

CREATE TABLE IF NOT EXISTS early(
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    sector_name VARCHAR(100),
    valid_until TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_event_early_promo UNIQUE (event_id, code)
);

CREATE TABLE IF NOT EXISTS bundle(
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL,
    sector_name VARCHAR(100),
    bundle_buy_count BIGINT NOT NULL,
    bundle_get_count BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_event_bundle_promo UNIQUE (event_id, code)
);

CREATE INDEX IF NOT EXISTS idx_tickets_event_status ON tickets(event_id, status);

COMMIT;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'event_service') THEN
    EXECUTE 'ALTER ROLE event_service WITH REPLICATION';
  END IF;
END$$;

