-- name: CreateRound :one
INSERT INTO rounds (operator_id, player_id, server_seed, client_seed, outcome)
VALUES ($1, $2, $3, $4, $5)
    RETURNING *;

-- name: GetRound :one
SELECT * FROM rounds
WHERE id = $1;