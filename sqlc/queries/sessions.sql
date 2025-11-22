-- name: CreateOperator :one
INSERT INTO operators (name, api_key)
VALUES ($1, $2)
    RETURNING *;

-- name: GetOperatorByApiKey :one
SELECT * FROM operators
WHERE api_key = $1
    LIMIT 1;

-- name: GetOperatorByID :one
SELECT * FROM operators WHERE id = $1 LIMIT 1;

-- name: CreatePlayer :one
INSERT INTO players (operator_id, external_player_id, jurisdiction)
VALUES ($1, $2, $3)
    RETURNING *;

-- name: GetPlayer :one
SELECT * FROM players
WHERE operator_id = $1 AND external_player_id = $2
    LIMIT 1;

-- name: GetPlayerByID :one
SELECT *
FROM players
WHERE id = $1
    LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (id, operator_id, player_id, launch_token, expires_at)
VALUES ($1, $2, $3, $4, $5)
    RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1
    LIMIT 1;

-- name: VerifySessionByToken :one
SELECT *
FROM sessions
WHERE launch_token = $1
  AND operator_id = $2
  AND revoked = FALSE
  AND expires_at > NOW()
    LIMIT 1;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked = TRUE
WHERE id = $1;