-- name: CreateComment :one
WITH ins AS (
  INSERT INTO event_comments (
    event_id,
    user_id,
    nick,
    text,
    rating_tenths
  )
  VALUES ($1, $2, $3, $4, $5)
  RETURNING
    id,
    event_id,
    user_id,
    nick,
    text,
    rating_tenths,
    created_at
),
agg AS (
  INSERT INTO event_rating_agg (
    event_id,
    sum_tenths,
    count,
    updated_at
  )
  SELECT
    event_id,
    rating_tenths,
    1,
    now()
  FROM ins
  ON CONFLICT (event_id) DO UPDATE
    SET
      sum_tenths = event_rating_agg.sum_tenths + EXCLUDED.sum_tenths,
      count      = event_rating_agg.count + 1,
      updated_at = now()
)
SELECT
  id,
  event_id,
  user_id,
  nick,
  text,
  rating_tenths,
  created_at
FROM ins;

-- name: ListCommentsByEvent :many
SELECT
  id,
  event_id,
  user_id,
  nick,
  text,
  rating_tenths,
  created_at
FROM event_comments
WHERE event_id = sqlc.arg(event_id)::bigint
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit)::bigint
OFFSET sqlc.arg(page_offset)::bigint;

-- name: GetRatingByEvent :one
SELECT
  e.event_id,
  (
    CASE
      WHEN COALESCE(a.count, 0) = 0 THEN 0
      ELSE (a.sum_tenths + (a.count / 2)) / a.count
    END
  )::bigint AS avg_tenths,
  COALESCE(a.count, 0)::bigint AS count
FROM (SELECT sqlc.arg(event_id)::bigint AS event_id) e
LEFT JOIN event_rating_agg a USING (event_id);
