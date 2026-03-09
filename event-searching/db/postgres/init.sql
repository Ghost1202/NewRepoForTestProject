BEGIN;

CREATE TABLE IF NOT EXISTS event_comments (
  id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id      BIGINT NOT NULL,
  user_id       BIGINT NOT NULL,
  nick          TEXT   NOT NULL,
  text          TEXT   NOT NULL,
  rating_tenths BIGINT NOT NULL CHECK (rating_tenths BETWEEN 0 AND 50),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS event_rating_agg (
  event_id    BIGINT PRIMARY KEY,
  sum_tenths  BIGINT NOT NULL,
  count       BIGINT NOT NULL,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
