-- name: InsertWebhookEvent :one
INSERT INTO webhook_events (
    operator_id,
    event_type,
    payload,
    status,
    retries,
    next_retry_at
    )
VALUES ($1, $2, $3, 'pending', 0, NOW())
RETURNING *;

-- name: GetPendingWebhookEvents :many
SELECT *
FROM webhook_events
WHERE status = 'pending'
AND next_retry_at <= NOW()
ORDER BY created_at
LIMIT 50; -- small batch

-- name: MarkWebhookProcessing :one
UPDATE webhook_events
SET status = 'processing',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkWebhookCompleted :exec
UPDATE webhook_events
SET status = 'completed',
    updated_at = NOW()
WHERE id = $1;

-- name: MarkWebhookFailed :exec
UPDATE webhook_events
SET status = 'failed',
    updated_at = NOW(),
    error_message = $2
WHERE id = $1;

-- name: UpdateWebhookRetry :one
UPDATE webhook_events
SET retries = retries + 1,
    next_retry_at = $2,
    status = 'pending',
    updated_at = NOW(),
    error_message = $3
WHERE id = $1
RETURNING *;

-- name: GetWebhookEventByID :one
SELECT *
FROM webhook_events
WHERE id = $1
    LIMIT 1;

-- name: ResetWebhookForRetry :exec
UPDATE webhook_events
SET
    status = 'pending',
    retries = 0,
    next_retry_at = NOW(),
    error_message = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: ListWebhooksByOperator :many
SELECT *
FROM webhook_events
WHERE operator_id = $1
ORDER BY id DESC
    LIMIT 200;

-- name: ListWebhooksByOperatorStatus :many
SELECT *
FROM webhook_events
WHERE operator_id = $1
  AND status = $2
ORDER BY id DESC
    LIMIT 200;