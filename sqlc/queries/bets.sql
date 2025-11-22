-- name: CreateBet :one
INSERT INTO bets (
    operator_id, player_id, round_id,
    amount, outcome, win_amount,
    status, idempotency_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING *;

-- name: GetBetByIdempotency :one
SELECT * FROM bets
WHERE operator_id = $1 AND idempotency_key = $2
    LIMIT 1;

-- name: GetBetsByRound :many
SELECT * FROM bets
WHERE round_id = $1
ORDER BY id;

-- name: MarkBetAsWon :exec
UPDATE bets
SET status = 'won'
WHERE id = $1
    RETURNING *;

-- name: UpdateBetStatus :one
UPDATE bets
SET
    status = $2,
    win_amount = $3,
    updated_at = NOW()
WHERE id = $1
    RETURNING *;