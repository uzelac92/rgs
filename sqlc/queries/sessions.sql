INSERT INTO operators (name, api_key)
VALUES ($1, $2)
    RETURNING *;

SELECT * FROM operators
WHERE api_key = $1
    LIMIT 1;

INSERT INTO players (operator_id, external_player_id, jurisdiction)
VALUES ($1, $2, $3)
    RETURNING *;

SELECT * FROM players
WHERE operator_id = $1 AND external_player_id = $2
    LIMIT 1;

INSERT INTO sessions (id, operator_id, player_id, launch_token, expires_at)
VALUES ($1, $2, $3, $4, $5)
    RETURNING *;

SELECT * FROM sessions
WHERE id = $1
    LIMIT 1;

SELECT * FROM sessions
WHERE id = $1
  AND revoked = FALSE
  AND expires_at > NOW()
    LIMIT 1;

UPDATE sessions
SET revoked = TRUE
WHERE id = $1;