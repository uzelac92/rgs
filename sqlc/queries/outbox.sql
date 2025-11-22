-- name: InsertOutbox :one
INSERT INTO outbox (bet_id, operator_id, player_id, amount)
VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: GetPendingOutbox :many
SELECT *
FROM outbox
WHERE processed = FALSE
ORDER BY id
    LIMIT 10;

-- name: MarkOutboxProcessed :one
UPDATE outbox
SET processed = TRUE
WHERE id = $1
    RETURNING *;

-- name: ListOutboxByOperator :many
SELECT *
FROM outbox
WHERE operator_id = $1
ORDER BY id DESC
    LIMIT 200;

-- name: ListOutboxByOperatorStatus :many
SELECT *
FROM outbox
WHERE operator_id = $1
  AND processed = $2
ORDER BY id DESC
    LIMIT 200;