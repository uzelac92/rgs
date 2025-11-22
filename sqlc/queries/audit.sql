-- name: ListAuditLogsByOperator :many
SELECT *
FROM audit_logs
WHERE operator_id = $1
ORDER BY created_at DESC
    LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByOperatorPlayer :many
SELECT *
FROM audit_logs
WHERE operator_id = $1
  AND player_id = $2
ORDER BY created_at DESC
    LIMIT $3 OFFSET $4;