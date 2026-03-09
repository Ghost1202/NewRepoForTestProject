-- name: GetPaymentByOrderID :one
SELECT * FROM payments
WHERE order_id = $1 LIMIT 1;

-- name: CreatePayment :exec
INSERT INTO payments (
    order_id, user_id, amount, currency, external_id, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
);

-- name: UpdatePaymentStatus :execresult
UPDATE payments
SET status = $2, external_id = $3, updated_at = NOW()
WHERE order_id = $1;

-- name: GetPaymentByExternalID :one
SELECT * FROM payments
WHERE external_id = $1
LIMIT 1;

-- name: UpdatePaymentStatusOnlyByExternalID :execresult
UPDATE payments
SET status = $2, updated_at = NOW()
WHERE external_id = $1;

-- name: CreateWaitlistEntry :exec
INSERT INTO waitlist (
    event_id, user_id, user_email
) VALUES (
    $1, $2, $3
);

-- name: GetNextWaitlistEntry :one
SELECT * FROM waitlist
WHERE event_id = $1 AND status = 'WAITING'
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: UpdateWaitlistStatus :exec
UPDATE waitlist
SET status = $2, updated_at = NOW()
WHERE id = $1;