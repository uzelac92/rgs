-- name: InsertAuditLog :exec
INSERT INTO audit_logs (operator_id, player_id, action, details)
VALUES (
           $1,
           $2,
           $3,
           $4
       );

-- name: GetOperatorLimits :one
SELECT *
FROM operator_limits
WHERE operator_id = $1;

-- name: UpsertOperatorLimits :one
INSERT INTO operator_limits (operator_id, max_bet, allowed_jurisdictions, daily_loss_limit, daily_win_limit)
VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (operator_id)
DO UPDATE SET
    max_bet = EXCLUDED.max_bet,
    allowed_jurisdictions = EXCLUDED.allowed_jurisdictions,
    daily_loss_limit = EXCLUDED.daily_loss_limit,
    daily_win_limit = EXCLUDED.daily_win_limit
    RETURNING *;